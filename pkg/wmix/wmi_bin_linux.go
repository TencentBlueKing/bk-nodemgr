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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"

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
