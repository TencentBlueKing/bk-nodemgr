/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package wmix

import (
	"fmt"
	"os"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix/wmiexec"
)

// wmiBin is the wmiexec binary.
// nolint: gochecknoglobals
var wmiBin struct {
	binaryPath string
	once       sync.Once
	cleanup    func()
}

// wmiBinaryPath returns the path to the wmiexec binary.
func wmiBinaryPath() (string, error) {
	var initErr error
	wmiBin.once.Do(func() {
		tmpFile, err := os.CreateTemp("", "wmiexec-*")
		if err != nil {
			initErr = fmt.Errorf("failed to create temporary file, err: %w", err)

			return
		}

		tmpFilePath := tmpFile.Name()

		// write wmiexec to the temporary file.
		if _, err := tmpFile.Write(wmiexec.Binary); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpFilePath)
			initErr = fmt.Errorf("failed to write wmiexec binary to temporary file, err: %w", err)

			return
		}

		if err := tmpFile.Close(); err != nil {
			_ = os.Remove(tmpFilePath)
			initErr = fmt.Errorf("failed to close temporary file, err: %w", err)

			return
		}

		// make the temporary file executable.
		// nolint: gosec,mnd
		if err := os.Chmod(tmpFilePath, 0700); err != nil {
			_ = os.Remove(tmpFilePath)
			initErr = fmt.Errorf("failed to make temporary file executable, err: %w", err)

			return
		}

		wmiBin.binaryPath = tmpFilePath
		wmiBin.cleanup = func() {
			_ = os.Remove(tmpFilePath)
		}
	})

	if initErr != nil {
		return "", fmt.Errorf("failed to initialize wmiexec binary, err: %w", initErr)
	}

	return wmiBin.binaryPath, nil
}
