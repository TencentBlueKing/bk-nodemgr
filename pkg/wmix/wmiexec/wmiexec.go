//go:build linux

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package wmiexec

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// WmiExecutor is the wmiexec executor.
type WmiExecutor struct {
	binaryPath string
	once       sync.Once
	cleanup    func()
	mu         sync.Mutex
}

// New creates a WmiExecutor instance.
func New() *WmiExecutor {
	return &WmiExecutor{}
}

// make sure the wmiexec binary is available and executable.
func (executor *WmiExecutor) ensureBinary() error {
	executor.mu.Lock()
	defer executor.mu.Unlock()

	var initErr error
	executor.once.Do(func() {
		tmpFile, err := os.CreateTemp("", "wmiexec-*")
		if err != nil {
			initErr = fmt.Errorf("failed to create temporary file, err: %v", err)

			return
		}

		tmpFilePath := tmpFile.Name()

		// write wmiexec to the temporary file.
		if _, err := tmpFile.Write(binary); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpFilePath)
			initErr = fmt.Errorf("failed to write wmiexec binary to temporary file, err: %v", err)

			return
		}

		if err := tmpFile.Close(); err != nil {
			_ = os.Remove(tmpFilePath)
			initErr = fmt.Errorf("failed to close temporary file, err: %v", err)

			return
		}

		// make the temporary file executable.
		if err := os.Chmod(tmpFilePath, 0755); err != nil {
			_ = os.Remove(tmpFilePath)
			initErr = fmt.Errorf("failed to make temporary file executable, err: %v", err)

			return
		}

		executor.binaryPath = tmpFilePath
		executor.cleanup = func() {
			_ = os.Remove(tmpFilePath)
		}
	})

	return initErr
}

// Execute a command on the target host.
func (executor *WmiExecutor) Execute(username, password, target, command string) ([]byte, error) {
	if err := executor.ensureBinary(); err != nil {
		return nil, err
	}

	args := []string{"-u", username, "-p", password, target, command}
	cmd := exec.Command(executor.binaryPath, args...)

	return cmd.CombinedOutput()
}

// Close close wmiexec.
func (executor *WmiExecutor) Close() error {
	executor.mu.Lock()
	defer executor.mu.Unlock()

	if executor.cleanup != nil {
		executor.cleanup()
		executor.cleanup = nil
		executor.binaryPath = ""
	}

	return nil
}
