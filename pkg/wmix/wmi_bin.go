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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix/wmiexec"
)

// wmiBin is the wmiexec binary.
// nolint: gochecknoglobals
var wmiBin struct {
	binaryPath string
	mu         sync.Mutex
	cleanup    func()
}

// wmiBinaryPath returns the path to the wmiexec binary.
func wmiBinaryPath() (string, error) {
	wmiBin.mu.Lock()
	defer wmiBin.mu.Unlock()

	if wmiBin.binaryPath == "" {
		path, cleanup, err := extractWMIBinary()
		if err != nil {
			return "", fmt.Errorf("failed to initialize wmiexec binary: %w", err)
		}

		wmiBin.binaryPath = path
		wmiBin.cleanup = cleanup

		return path, nil
	}

	if _, err := os.Stat(wmiBin.binaryPath); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("failed to stat cached wmiexec binary. path(%s): %w", wmiBin.binaryPath, err)
		}

		// clean old dir.
		if wmiBin.cleanup != nil {
			wmiBin.cleanup()
		}

		// if file no exist, we re extract it.
		path, cleanup, err := extractWMIBinary()
		if err != nil {
			return "", fmt.Errorf("failed to initialize wmiexec binary: %w", err)
		}

		wmiBin.binaryPath = path
		wmiBin.cleanup = cleanup
	}

	return wmiBin.binaryPath, nil
}

func extractWMIBinary() (string, func(), error) {
	tmpFile, err := tmp.NewTempFile(io.NopCloser(bytes.NewBuffer(wmiexec.Binary)), "wmiexec")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	// make the temporary file executable.
	// nolint: gosec,mnd
	if err := os.Chmod(tmpFile.Path(), 0700); err != nil {
		_ = os.Remove(tmpFile.Path())

		return "", nil, fmt.Errorf("failed to make temporary file executable: %w", err)
	}

	path := tmpFile.Path()
	cleanup := func() {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			_ = err
		}
	}

	return path, cleanup, nil
}

func wmiRunCmd(ctx context.Context, args []string, envs []string) (string, string, error) {
	binPath, err := wmiBinaryPath()
	if err != nil {
		return "", "", err
	}

	// nolint: gosec
	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Env = envs
	// Create a new process group and set Pdeathsig to ensure all child processes
	// (including staticx-forked Python processes) are killed when the parent dies.
	// Pdeathsig ensures that if wmiexec is killed, its children receive SIGKILL
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
	// After cmd.Run() returns, the main wmiexec process has been waited.
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
		return "", "", fmt.Errorf("commands cannot be executed using wmiexec, stdOut(%s), stdErr(%s): %w",
			stdOut.String(), stdErr.String(), err)
	}

	// because the wmiexec will output the license information in the first two lines,
	lines := strings.Split(stdOut.String(), "\n")[2:]
	stdOutStr := strings.Join(lines, "\n")

	return stdOutStr, stdErr.String(), nil
}
