//go:build linux || darwin || freebsd || aix

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// unsafeCommandChars lists shell metacharacters that must not appear in the command name.
// Arguments are passed directly to exec and are not subject to shell expansion, but the
// command name itself may be logged or reused, so it is validated for defense in depth.
var unsafeCommandChars = []string{"|", "&", ";", "`", "$", ">", "<", "(", ")", "{", "}", "!", "\n", "\r", "\x00"}

// CheckCommandSafe validates that the command name contains no shell metacharacters and
// that no argument contains a null byte.
func CheckCommandSafe(name string, args []string) error {
	if name == "" {
		return errors.New("command name is empty")
	}

	for _, ch := range unsafeCommandChars {
		if strings.Contains(name, ch) {
			return fmt.Errorf("command name contains unsafe character: %q", ch)
		}
	}

	for i, arg := range args {
		if strings.ContainsRune(arg, 0) {
			return fmt.Errorf("arg[%d] contains null byte", i)
		}
	}

	return nil
}

// RunCommandStreaming executes name with args inside workDir and streams combined
// stdout+stderr output line by line through the returned output channel.
// The error channel receives the command's exit error (nil on success) and is closed
// after the process finishes.
//
// The caller must drain the output channel; otherwise the command will block on output.
// Cancelling ctx will terminate the process (SIGKILL) and cause both channels to close.
//
// Returns a non-nil error immediately if workDir or the command fail safety validation,
// or if the process cannot be started.
func RunCommandStreaming(ctx context.Context, workDir, name string, args []string) (<-chan string, <-chan error, error) {
	if err := CheckDirPathSafe(workDir); err != nil {
		return nil, nil, fmt.Errorf("unsafe working directory: %w", err)
	}

	if err := CheckCommandSafe(name, args); err != nil {
		return nil, nil, fmt.Errorf("unsafe command: %w", err)
	}

	pr, pw := io.Pipe()

	// nolint: gosec
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		_ = pw.Close()
		_ = pr.Close()

		return nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	outputCh := make(chan string, 64) // nolint: nmd
	errCh := make(chan error, 1)

	// Waiter: block until the process exits, then close the write end of the pipe so
	// the reader goroutine observes EOF.
	go func() {
		err := cmd.Wait()
		_ = pw.Close()
		errCh <- err
		close(errCh)
	}()

	// Reader: scan lines from the pipe and forward them to outputCh.
	go func() {
		defer close(outputCh)
		defer func() { _ = pr.Close() }()

		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case outputCh <- scanner.Text():
			}
		}

		if err := scanner.Err(); err != nil {
			outputCh <- fmt.Sprintf("[runner] output scan error: %v", err)
		}
	}()

	return outputCh, errCh, nil
}
