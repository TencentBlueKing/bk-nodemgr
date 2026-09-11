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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

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

func jinja2ExecRunCmd(ctx context.Context, args []string, envs []string) (string, string, error) {
	binPath, err := jinja2ExecBinaryPath()
	if err != nil {
		return "", "", err
	}

	// nolint: gosec
	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Env = envs
	// Create a new process group and set Pdeathsig to ensure all child processes
	// (including staticx-forked Python processes) are killed when the parent dies.
	// Pdeathsig ensures that if jinja2_exec is killed, its children receive SIGKILL
	// immediately, preventing them from becoming orphaned zombies adopted by backend.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,            // Create new process group
		Pdeathsig: syscall.SIGKILL, // Kill children when parent dies
	}
	// WaitDelay specifies how long to wait after the context is canceled
	// before sending SIGKILL to the process group.
	cmd.WaitDelay = 0 // Kill immediately after context cancellation
	// Kill the entire process group (negative PID) to terminate all descendants.
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	err = cmd.Run()
	// After cmd.Run() returns, the main jinja2_exec process has been waited.
	// However, staticx-forked Python children may have been reparented to backend
	// and killed by SIGKILL, becoming zombies. We need to wait for all processes
	// in the process group to prevent zombie accumulation.
	if cmd.Process != nil {
		// Wait for all remaining processes in the process group.
		// Use WNOHANG to avoid blocking if no zombies exist.
		// Add a safety limit to prevent infinite loops in edge cases.
		pgid := cmd.Process.Pid
		const maxWaitAttempts = 100 // Safety limit
		for i := 0; i < maxWaitAttempts; i++ {
			var status syscall.WaitStatus
			// Wait for any process in the process group (negative PID).
			// WNOHANG returns immediately if no zombie exists.
			pid, waitErr := syscall.Wait4(-pgid, &status, syscall.WNOHANG, nil)
			if waitErr != nil || pid <= 0 {
				// No more zombies in this process group, or error occurred.
				break
			}
			// Successfully reaped a zombie process, continue to check for more.
		}
	}
	if err != nil {
		return "", "", fmt.Errorf("commands cannot be executed using jinja2, stdOut(%s), stdErr(%s): %w",
			stdOut.String(), stdErr.String(), err)
	}

	return stdOut.String(), stdErr.String(), nil
}
