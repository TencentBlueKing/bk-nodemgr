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

package windows

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

func (handler *AgentHandler) getAbsPath(relativePath string) string {
	return filepath.Clean(filepath.Join(handler.rootAbsDir, relativePath))
}

func (handler *AgentHandler) statFile(relativePath string) (os.FileInfo, error) {
	absPath := handler.getAbsPath(relativePath)

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file. file(%s): %w", absPath, err)
	}

	return info, nil
}

// checkRegularFile check if relative-path is a regular file.
func (handler *AgentHandler) checkRegularFile(relativePath string) error {
	info, err := handler.statFile(relativePath)
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file. file(%s)", relativePath)
	}

	return nil
}

func (handler *AgentHandler) checkPathInFS(relativePath string) error {
	absPath := handler.getAbsPath(relativePath)

	relPath, err := filepath.Rel(handler.rootAbsDir, absPath)
	if err != nil {
		return fmt.Errorf("failed to check if path in FS. path(%s): %w", absPath, err)
	}

	if strings.HasPrefix(relPath, ".."+string(filepath.Separator)) || relPath == ".." {
		return fmt.Errorf("path not in FS. root(%s), path(%s)", handler.rootAbsDir, absPath)
	}

	return nil
}

func (handler *AgentHandler) openFileForWrite(relativePath string) (*os.File, error) {
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
		return nil, fmt.Errorf("failed to open writing file. file(%s): %w", absPath, err)
	}

	return file, nil
}

func (handler *AgentHandler) openFileForRead(relativePath string) (*os.File, error) {
	if err := handler.checkPathInFS(relativePath); err != nil {
		return nil, fmt.Errorf("will not open path outside FS. file(%s): %w", relativePath, err)
	}

	absPath := handler.getAbsPath(relativePath)

	// nolint: gosec
	file, err := os.OpenFile(absPath, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open reading file. file(%s): %w", absPath, err)
	}

	return file, nil
}

func (handler *AgentHandler) removeFile(relativePath string) error {
	if err := handler.checkPathInFS(relativePath); err != nil {
		return fmt.Errorf("will not remove path outside FS. file(%s): %w", relativePath, err)
	}

	absPath := handler.getAbsPath(relativePath)

	if err := os.Remove(absPath); err != nil {
		return fmt.Errorf("failed to remove file(%s): %w", absPath, err)
	}

	return nil
}

func (handler *AgentHandler) removeAll(relativePath string) error {
	if err := handler.checkPathInFS(relativePath); err != nil {
		return fmt.Errorf("will not remove path outside FS. file(%s): %w", relativePath, err)
	}

	absPath := handler.getAbsPath(relativePath)

	if absPath == "" ||
		absPath == "/" ||
		strings.HasPrefix(absPath, "/dev/") ||
		strings.HasPrefix(absPath, "/sys/") ||
		strings.HasPrefix(absPath, "/proc/") {

		return fmt.Errorf("failed to remove all, got invalid path. path(%s)", absPath)
	}

	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("failed to remove all(%s): %w", absPath, err)
	}

	return nil
}

// mkdirAll make all directories with relative-path.
func (handler *AgentHandler) mkdirAll(relativePath string) error {
	return os.MkdirAll(handler.getAbsPath(relativePath), mkdirMode)
}

func (handler *AgentHandler) renameFileAsTmpFileIfExists(relativePath string) error {
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

func (handler *AgentHandler) overwriteFile(file io.Reader, targetRelativePath string) (err error) {
	targetAbsPath := handler.getAbsPath(targetRelativePath)
	if err := os.MkdirAll(filepath.Dir(targetAbsPath), mkdirMode); err != nil {
		return err
	}

	targetFile, err := handler.openFileForWrite(targetRelativePath)
	if err != nil {
		return fmt.Errorf("failed to overwrite file, failed to open file. file(%s): %w", targetAbsPath, err)
	}
	defer func() {
		if errClose := targetFile.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	if _, err := io.Copy(targetFile, file); err != nil {
		return fmt.Errorf("failed to overwrite file, failed to write file. file(%s): %w", targetAbsPath, err)
	}

	return nil
}

func (handler *AgentHandler) checkDirWritable(targetRelativeDir string) error {
	testRelativePath := filepath.Join(targetRelativeDir, handler.genTmpFileName("write_test"))

	// nolint: gosec
	testFile, err := handler.openFileForWrite(testRelativePath)
	if err != nil {
		return fmt.Errorf("target directory has no write permission. dir(%s): %v", testRelativePath, err)
	}
	_ = testFile.Close()

	_ = handler.removeFile(testRelativePath)

	return nil
}

func (handler *AgentHandler) makeBinFilesExecutable() error {
	binFiles, err := utils.ListFiles(handler.getAbsPath(handler.binDir))
	if err != nil {
		return fmt.Errorf("failed to list bin files. dir(%s): %w", handler.binDir, err)
	}

	for _, binFile := range binFiles {
		if err := utils.MakeExecutable(binFile); err != nil {
			return fmt.Errorf("failed to make file executable. file(%s): %w", binFile, err)
		}
	}

	return nil
}

func (handler *AgentHandler) genTmpFileName(tag string) string {
	return tmpFilePrefix + tag + "_" + strconv.FormatInt(time.Now().Unix(), 10)
}
