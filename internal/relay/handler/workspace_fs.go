/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type workspaceFS struct {
	rootDir string
}

func newWorkspaceFS(rootDir string) workspaceFS {
	return workspaceFS{rootDir: rootDir}
}

func (fs workspaceFS) ensure(op string) error {
	if err := os.MkdirAll(fs.rootDir, storageTmpDirMode); err != nil {
		return fmt.Errorf("failed to ensure transfer-file dir for %s, dir(%s): %w", op, fs.rootDir, err)
	}

	return nil
}

func (fs workspaceFS) removeFile(filename string) error {
	if err := fs.ensure("safe remove file"); err != nil {
		return err
	}

	absPath, err := fs.absPath(filename)
	if err != nil {
		return err
	}

	return fs.removeAll(absPath)
}

func (fs workspaceFS) absPath(filename string) (string, error) {
	cleanFilename := filepath.Clean(filename)
	if filename == "" || cleanFilename == "." {
		return "", fmt.Errorf("will not resolve empty filename in workspace. root(%s), filename(%s)", fs.rootDir, filename)
	}
	if filepath.IsAbs(filename) {
		return "", fmt.Errorf("will not resolve absolute path in workspace. root(%s), filename(%s)", fs.rootDir, filename)
	}

	absPath := filepath.Clean(filepath.Join(fs.rootDir, cleanFilename))
	if err := fs.checkPathInWorkspace(absPath); err != nil {
		return "", fmt.Errorf("will not resolve path outside workspace. root(%s), filename(%s): %w", fs.rootDir, filename, err)
	}

	return absPath, nil
}

func (fs workspaceFS) checkPathInWorkspace(absPath string) error {
	if fs.isProtectedPath(absPath) {
		return fmt.Errorf("failed to remove all, got invalid path. path(%s)", absPath)
	}

	rootDir := filepath.Clean(fs.rootDir)
	relPath, err := filepath.Rel(rootDir, absPath)
	if err != nil {
		return fmt.Errorf("failed to check if path in workspace. path(%s): %w", absPath, err)
	}

	if strings.HasPrefix(relPath, ".."+string(filepath.Separator)) || relPath == ".." || filepath.IsAbs(relPath) {
		return fmt.Errorf("path not in workspace. root(%s), path(%s)", rootDir, absPath)
	}

	return nil
}

func (fs workspaceFS) removeAll(absPath string) error {
	if fs.isProtectedPath(absPath) {
		return fmt.Errorf("failed to remove all, got invalid path. path(%s)", absPath)
	}

	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("failed to remove all. path(%s): %w", absPath, err)
	}

	return nil
}

func (fs workspaceFS) isProtectedPath(absPath string) bool {
	return absPath == "" ||
		absPath == "/" ||
		strings.HasPrefix(absPath, "/dev/") ||
		strings.HasPrefix(absPath, "/sys/") ||
		strings.HasPrefix(absPath, "/proc/")
}
