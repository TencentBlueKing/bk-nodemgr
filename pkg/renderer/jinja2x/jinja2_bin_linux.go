//go:build linux

/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package jinja2x

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer/jinja2x/jinja2"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
)

// jinja2ExecBin is the jinja2 binary.
// nolint: gochecknoglobals
var jinja2ExecBin struct {
	binaryPath string
	mu         sync.Mutex
	cleanup    func()
}

// jinja2ExecBinaryPath returns the path to the jinja2 binary.
func jinja2ExecBinaryPath() (string, error) {
	jinja2ExecBin.mu.Lock()
	defer jinja2ExecBin.mu.Unlock()

	if jinja2ExecBin.binaryPath == "" {
		path, cleanup, err := extractJinja2ExecBinary()
		if err != nil {
			return "", fmt.Errorf("failed to initialize jinja2 binary: %w", err)
		}
		jinja2ExecBin.binaryPath = path
		jinja2ExecBin.cleanup = cleanup

		return path, nil
	}

	if _, err := os.Stat(jinja2ExecBin.binaryPath); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("failed to stat cached jinja2 binary. path(%s): %w", jinja2ExecBin.binaryPath, err)
		}

		if jinja2ExecBin.cleanup != nil {
			jinja2ExecBin.cleanup()
		}

		path, cleanup, err := extractJinja2ExecBinary()
		if err != nil {
			return "", fmt.Errorf("failed to initialize jinja2 binary: %w", err)
		}
		jinja2ExecBin.binaryPath = path
		jinja2ExecBin.cleanup = cleanup
	}

	return jinja2ExecBin.binaryPath, nil
}

func extractJinja2ExecBinary() (string, func(), error) {
	tmpFile, err := tmp.NewTempFileWithSpecialName(io.NopCloser(bytes.NewBuffer(jinja2.Binary)), "jinja2_exec")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	// make the temporary file executable.
	// nolint: gosec,mnd
	if err := os.Chmod(tmpFile.Path(), 0700); err != nil {
		_ = os.Remove(tmpFile.Path())

		return "", nil, fmt.Errorf("failed to make temporary file executable: %w", err)
	}

	p := tmpFile.Path()
	cleanup := func() {
		if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
			_ = err
		}
		// NewTempFileWithSpecialName creates file under a temp dir; remove dir if empty.
		if err := os.Remove(filepath.Dir(p)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			_ = err
		}
	}

	return tmpFile.Path(), cleanup, nil
}

// The embedded process has a lifecycle independent from the starting request.
// nolint: contextcheck
func startServer(ctx context.Context, start, end int) (*serverProcess, error) {
	binPath, err := jinja2ExecBinaryPath()
	if err != nil {
		return nil, err
	}
	token := rand.Text()
	// The shared process must not retain the render request's context.
	processCtx, cancel := context.WithCancel(context.Background())
	// nolint: gosec
	cmd := exec.CommandContext(processCtx, binPath, "--serve", strconv.Itoa(start), strconv.Itoa(end))
	cmd.Env = append(os.Environ(), "JINJA2_SERVER_TOKEN="+token)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 0
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdout, writer, err := os.Pipe()
	if err != nil {
		cancel()
		_ = stdin.Close()

		return nil, err
	}
	cmd.Stdout = writer
	defer func() { _ = writer.Close() }()
	if err := cmd.Start(); err != nil {
		cancel()
		_ = stdin.Close()
		_ = stdout.Close()

		return nil, fmt.Errorf("failed to start jinja2 server: %w", err)
	}
	_ = writer.Close()
	process := &serverProcess{done: make(chan struct{})}
	process.stop = func() {
		_ = stdin.Close()
		const stopTimeout = 2 * time.Second
		timer := time.NewTimer(stopTimeout)
		defer timer.Stop()
		select {
		case <-process.done:
		case <-timer.C:
			cancel()
			<-process.done
		}
	}
	ready := make(chan string, 1)
	go func() {
		line, _ := bufio.NewReader(stdout).ReadString('\n')
		ready <- line
	}()
	go func() {
		_ = cmd.Wait()
		_ = stdin.Close()
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		_ = stdout.Close()
		reapServerGroup(cmd.Process.Pid)
		cancel()
		close(process.done)
	}()
	const startupTimeout = 10 * time.Second
	startupCtx, stopStartup := context.WithTimeout(ctx, startupTimeout)
	defer stopStartup()
	select {
	case line := <-ready:
		port, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || port < start || port > end {
			process.stop()
			return nil, fmt.Errorf("jinja2 server failed to bind configured port range %d-%d", start, end)
		}
		process.client, err = newClient(port, token)
		if err != nil {
			process.stop()
			return nil, err
		}
		if err := process.client.request(startupCtx, http.MethodGet, "/healthz", nil, &struct{}{}); err != nil {
			process.stop()
			return nil, fmt.Errorf("jinja2 server readiness check failed: %w", err)
		}

		return process, nil
	case <-process.done:
		return nil, fmt.Errorf("jinja2 server exited during startup")
	case <-startupCtx.Done():
		process.stop()
		return nil, fmt.Errorf("jinja2 server startup timed out: %w", startupCtx.Err())
	}
}

func reapServerGroup(pgid int) {
	const maxWaitAttempts = 100
	for range maxWaitAttempts {
		var status syscall.WaitStatus
		pid, err := syscall.Wait4(-pgid, &status, syscall.WNOHANG, nil)
		if err != nil || pid <= 0 {
			break
		}
	}
}
