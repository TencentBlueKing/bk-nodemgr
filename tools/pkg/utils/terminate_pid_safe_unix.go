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
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	// ErrProcessNotInDir indicates that the process executable path is not in the specified directory.
	ErrProcessNotInDir = errors.New("process executable path is not in the specified directory")
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

// terminateProcess terminates the process with the given PID by sending a SIGTERM signal.
// If the process does not exit within the timeout, it sends a SIGKILL signal to forcefully terminate it.
func terminateProcess(pid int) error {
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return err
	}

	return waitForProcessExit(pid, 5*time.Second, syscall.SIGKILL)
}

// waitForProcessExit waits for the process to exit within the specified timeout.
// If the process is still running after the timeout, it sends the final signal (e.g., SIGKILL) to forcefully terminate it.
func waitForProcessExit(pid int, timeout time.Duration, finalSignal syscall.Signal) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		exists, err := CheckPIDExist(pid)
		if err != nil {
			return fmt.Errorf("failed to check if process %d exists: %v", pid, err)
		}

		if !exists {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return syscall.Kill(pid, finalSignal)
}
