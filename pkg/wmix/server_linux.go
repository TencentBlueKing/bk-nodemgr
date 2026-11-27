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

package wmix

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// The embedded process outlives the request that first starts it.
func startServer(ctx context.Context, start, end int) (*serverProcess, error) {
	binPath, err := wmiBinaryPath()
	if err != nil {
		return nil, err
	}
	token := rand.Text()
	processCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	// nolint: gosec
	cmd := exec.CommandContext(processCtx, binPath, "--serve", strconv.Itoa(start), strconv.Itoa(end))
	cmd.Env = append(os.Environ(), "WMI_SERVER_TOKEN="+token)
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

		return nil, fmt.Errorf("failed to start WMI server: %w", err)
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
			return nil, fmt.Errorf("WMI server failed to bind configured port range %d-%d", start, end)
		}
		process.client, err = newHTTPClient(port, token)
		if err != nil {
			process.stop()
			return nil, err
		}
		if err := process.client.request(startupCtx, http.MethodGet, "/healthz", nil, &struct{}{}); err != nil {
			process.stop()
			return nil, fmt.Errorf("WMI server readiness check failed: %w", err)
		}

		return process, nil
	case <-process.done:
		return nil, fmt.Errorf("WMI server exited during startup")
	case <-startupCtx.Done():
		process.stop()
		return nil, fmt.Errorf("WMI server startup timed out: %w", startupCtx.Err())
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
