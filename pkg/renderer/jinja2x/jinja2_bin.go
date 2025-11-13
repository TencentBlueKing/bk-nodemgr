/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package jinja2x

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer/jinja2x/jinja2"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
)

// jinja2ExecBin is the jinja2 binary.
// nolint: gochecknoglobals
var jinja2ExecBin struct {
	binaryPath string
	once       sync.Once
	cleanup    func()
}

// jinja2ExecBinaryPath returns the path to the jinja2 binary.
func jinja2ExecBinaryPath() (string, error) {
	var initErr error
	jinja2ExecBin.once.Do(func() {
		tmpFile, err := tmp.NewTempFileWithSpecialName(io.NopCloser(bytes.NewBuffer(jinja2.Binary)), "jinja2_exec")
		if err != nil {
			initErr = fmt.Errorf("failed to create temporary file: %w", err)

			return
		}

		// make the temporary file executable.
		// nolint: gosec,mnd
		if err := os.Chmod(tmpFile.Path(), 0700); err != nil {
			_ = os.Remove(tmpFile.Path())
			initErr = fmt.Errorf("failed to make temporary file executable: %w", err)

			return
		}

		jinja2ExecBin.binaryPath = tmpFile.Path()
		jinja2ExecBin.cleanup = func() {
			_ = os.Remove(tmpFile.Path())
		}
	})

	if initErr != nil {
		return "", fmt.Errorf("failed to initialize jinja2 binary: %w", initErr)
	}

	return jinja2ExecBin.binaryPath, nil
}

func jinja2ExecRunCmd(ctx context.Context, args []string, envs []string) (string, string, error) {
	binPath, err := jinja2ExecBinaryPath()
	if err != nil {
		return "", "", err
	}

	// nolint: gosec
	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Env = envs

	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("commands cannot be executed using jinja2, stdOut(%s), stdErr(%s): %w",
			stdOut.String(), stdErr.String(), err)
	}

	return stdOut.String(), stdErr.String(), nil
}
