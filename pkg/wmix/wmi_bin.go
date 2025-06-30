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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
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
		tmpFile, err := tmp.NewTempFile(io.NopCloser(bytes.NewBuffer(wmiexec.Binary)), "wmiexec")
		if err != nil {
			initErr = fmt.Errorf("failed to create temporary file, err: %w", err)

			return
		}

		// make the temporary file executable.
		// nolint: gosec,mnd
		if err := os.Chmod(tmpFile.Path(), 0700); err != nil {
			_ = os.Remove(tmpFile.Path())
			initErr = fmt.Errorf("failed to make temporary file executable, err: %w", err)

			return
		}

		wmiBin.binaryPath = tmpFile.Path()
		wmiBin.cleanup = func() {
			_ = os.Remove(tmpFile.Path())
		}
	})

	if initErr != nil {
		return "", fmt.Errorf("failed to initialize wmiexec binary, err: %w", initErr)
	}

	return wmiBin.binaryPath, nil
}

func wmiRunCmd(ctx context.Context, args []string) (string, string, error) {
	binPath, err := wmiBinaryPath()
	if err != nil {
		return "", "", err
	}

	// nolint:gosec
	cmd := exec.CommandContext(ctx, binPath, args...)
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("commands cannot be executed using wmiexec, stdOut(%s), stdErr(%s), err: %w",
			stdOut.String(), stdErr.String(), err)
	}

	// because the wmiexec will output the license information in the first two lines,
	lines := strings.Split(stdOut.String(), "\n")[2:]
	stdOutStr := strings.Join(lines, "\n")

	return stdOutStr, stdErr.String(), nil
}
