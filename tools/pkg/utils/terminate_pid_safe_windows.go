//go:build windows

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package utils ...
package utils

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

var (
	// ErrProcessNotInDir indicates that the process executable path is not in the specified directory.
	ErrProcessNotInDir = errors.New("process executable path is not in the specified directory")
)

const (
	processExitTimeout  = 5 * time.Second
	processPollInterval = 100 * time.Millisecond
)

// TerminateProcess terminate the process by PID.
func TerminateProcess(pid int, binaryDir string) error {
	procInfo, err := GetProcessInfo(pid)
	if err != nil {
		return fmt.Errorf("failed to get process info: %v", err)
	}

	if procInfo.FullPath == "" || !IsPathInDirectory(procInfo.FullPath, binaryDir) {
		return ErrProcessNotInDir
	}

	if err := terminateProcess(pid); err != nil {
		return fmt.Errorf("failed to terminate process %d: %w", pid, err)
	}

	return nil
}

// IsPathInDirectory checks if the given file path is located within the target directory, resolving any symbolic links.
func IsPathInDirectory(filePath, targetDir string) bool {
	realFilePath, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		realFilePath = filePath
	}

	realTargetDir, err := filepath.EvalSymlinks(targetDir)
	if err != nil {
		realTargetDir = targetDir
	}

	absFilePath, err := filepath.Abs(realFilePath)
	if err != nil {
		return false
	}

	absTargetDir, err := filepath.Abs(realTargetDir)
	if err != nil {
		return false
	}

	return strings.HasPrefix(absFilePath, absTargetDir+string(filepath.Separator)) ||
		absFilePath == absTargetDir
}

// terminateProcess terminates the process with the given PID.
// It first attempts a graceful shutdown via GenerateConsoleCtrlEvent (Ctrl+Break).
// If the process does not exit within the timeout, it forcefully terminates it via TerminateProcess.
func terminateProcess(pid int) error {
	// Attempt graceful shutdown: send Ctrl+Break to the process's console group.
	// This is the closest Windows equivalent of SIGTERM for console processes.
	// Errors here are non-fatal; we fall through to the forceful path on timeout.
	_ = windows.GenerateConsoleCtrlEvent(windows.CTRL_BREAK_EVENT, uint32(pid))

	return waitForProcessExit(pid, processExitTimeout, func(p int) error {
		handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(p))
		if err != nil {
			// If the process cannot be opened, it may have already exited.
			return nil
		}

		defer func() { _ = windows.CloseHandle(handle) }()

		return windows.TerminateProcess(handle, 1)
	})
}

// waitForProcessExit waits for the process to exit within the specified timeout.
// If the process is still running after the timeout, it invokes forceKill to terminate it.
func waitForProcessExit(pid int, timeout time.Duration, forceKill func(int) error) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		exists, err := CheckPIDExist(pid)
		if err != nil {
			return fmt.Errorf("failed to check if process %d exists: %v", pid, err)
		}

		if !exists {
			return nil
		}

		time.Sleep(processPollInterval)
	}

	return forceKill(pid)
}
