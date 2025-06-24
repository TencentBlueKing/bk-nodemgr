/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tmp

import (
	"fmt"
	"io"
	"os"
)

// File represents a temporary file with its path and a cleanup function.
type File struct {
	path    string
	cleanup func() error
}

// Exist returns true if the temporary file exists.
func (f *File) Exist() bool {
	_, err := os.Stat(f.path)

	return !os.IsNotExist(err)
}

// Path returns the path to the temporary file.
func (f *File) Path() string {
	return f.path
}

// CleanUp cleans up the temporary file.
func (f *File) CleanUp() error {
	return f.cleanup()
}

// NewTempFile creates a temporary file and returns its path.
func NewTempFile(data io.ReadCloser, name string) (file *File, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("create temporary file panic, recover(%v)", r)
		}
	}()

	tmpDir, err := GetTmpDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file, err: %w", err)
	}

	tmpFile, err := os.CreateTemp(tmpDir, name+"-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file, err: %w", err)
	}

	tmpFilePath := tmpFile.Name()

	// write data to the temporary file.
	if _, err = io.Copy(tmpFile, data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFilePath)

		return nil, fmt.Errorf("failed to write data to temporary file, err: %w", err)
	}

	if err = data.Close(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFilePath)

		return nil, fmt.Errorf("failed to close data, err: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFilePath)

		return nil, fmt.Errorf("failed to close temporary file, err: %w", err)
	}

	file = &File{
		path: tmpFilePath,
	}

	file.cleanup = func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("clean up temporary file panic, tmp-file-path(%s), recover(%v)", tmpFilePath, r)
			}
		}()

		if err = os.Remove(tmpFilePath); err != nil {
			return fmt.Errorf("failed to remove temporary file, err: %w", err)
		}

		return
	}

	return file, nil
}
