/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package unix

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (handler *PluginV2Handler) getAbsPath(relativePath string) string {
	return filepath.Clean(filepath.Join(handler.rootAbsDir, relativePath))
}

func (handler *PluginV2Handler) openFileForWrite(relativePath string) (*os.File, error) {
	if err := handler.checkPathInFS(relativePath); err != nil {
		return nil, fmt.Errorf("will not open path outside FS. file(%s): %w", relativePath, err)
	}

	if err := handler.mkdirAll(filepath.Dir(relativePath)); err != nil {
		return nil, fmt.Errorf("failed to open writing file, failed to make dirs: %w", err)
	}

	absPath := handler.getAbsPath(relativePath)

	// nolint: gosec
	file, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, overwriteMode)
	if err != nil {
		return nil, fmt.Errorf("failed to open writing file(%s): %w", absPath, err)
	}

	return file, nil
}
func (handler *PluginV2Handler) renameFileAsTmpFileIfExists(relativePath string) error {
	if err := handler.checkPathInFS(relativePath); err != nil {
		return fmt.Errorf("will not rename path outside FS. file(%s): %w", relativePath, err)
	}

	absPath := handler.getAbsPath(relativePath)

	_, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to stat file. file(%s): %w", absPath, err)
	}

	tmpFileAbsPath := filepath.Join(
		filepath.Dir(absPath),
		handler.genTmpFileName(filepath.Base(absPath)),
	)

	return os.Rename(absPath, tmpFileAbsPath)
}

// mkdirAll make all directories with relative-path.
func (handler *PluginV2Handler) mkdirAll(relativePath string) error {
	return os.MkdirAll(handler.getAbsPath(relativePath), mkdirMode)
}

func (handler *PluginV2Handler) overwriteFile(file io.Reader, targetRelativePath string) (err error) {
	targetAbsPath := handler.getAbsPath(targetRelativePath)
	if err := os.MkdirAll(filepath.Dir(targetAbsPath), mkdirMode); err != nil {
		return err
	}

	targetFile, err := handler.openFileForWrite(targetRelativePath)
	if err != nil {
		return fmt.Errorf("failed to overwrite file, failed to open file(%s): %w", targetAbsPath, err)
	}
	defer func() {
		if errClose := targetFile.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	if _, err := io.Copy(targetFile, file); err != nil {
		return fmt.Errorf("failed to overwrite file, failed to write file(%s): %w", targetAbsPath, err)
	}

	return nil
}

func (handler *PluginV2Handler) removeAll(relativePath string) error {
	if err := handler.checkPathInFS(relativePath); err != nil {
		return fmt.Errorf("will not remove path outside FS. file(%s): %w", relativePath, err)
	}

	absPath := handler.getAbsPath(relativePath)

	if absPath == "" ||
		absPath == "/" ||
		strings.HasPrefix(absPath, "/dev/") ||
		strings.HasPrefix(absPath, "/sys/") ||
		strings.HasPrefix(absPath, "/proc/") {

		return fmt.Errorf("failed to remove all, got invalid path(%s)", absPath)
	}

	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("failed to remove all(%s): %w", absPath, err)
	}

	return nil
}
