/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package bkrepo provides the BKRepo mock API helper.
package bkrepo

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bkrepo"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/common"
	"github.com/gin-gonic/gin"
)

const (
	pathSeparator      = string(filepath.Separator)
	mkdirMode          = 0750
	repoPathRoot       = "/"
	repoPathCurrentDir = "."
	repoPathEmpty      = ""
)

// response operations
// ------------------------------------------------------------------------------------------------

// respondSuccess sends a success response with BKRepo format.
func respondSuccess[T any](gCtx *gin.Context, data T) {
	common.RespondJSON(gCtx, bkrepo.BaseBroker[T]{
		RespCommon: bkrepo.RespCommon{
			Code: CodeOK,
		},
		Data: data,
	})
}

// respondError sends an error response with BKRepo format.
func respondError(gCtx *gin.Context, code int, message string) {
	common.RespondJSON(gCtx, bkrepo.BaseBroker[any]{
		RespCommon: bkrepo.RespCommon{
			Code:    code,
			Message: message,
		},
		Data: nil,
	})
}

// respondFile streams file content to the response.
func respondFile(gCtx *gin.Context, reader io.Reader, size int64) {
	gCtx.DataFromReader(http.StatusOK, size, server.MIMETypeBin.String(), reader, nil)
}

// isNodeNotFound checks whether the error indicates a missing node.
func isNodeNotFound(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}

// file operations
// ------------------------------------------------------------------------------------------------

// cleanRepoPath cleans a repo path.
func cleanRepoPath(p string) string {
	return path.Clean(repoPathRoot + p)
}

// getAbsPath gets the absolute filesystem path for the given parameters.
func getAbsPath(baseDir, projectID, repoName, repoPath string) string {
	relativePath := strings.TrimPrefix(repoPath, repoPathRoot)

	return filepath.Clean(filepath.Join(baseDir, projectID, repoName, relativePath))
}

// checkPathInFS checks if the absolute path is within the base directory.
func checkPathInFS(baseDir, absPath string) error {
	cleanBase := filepath.Clean(baseDir)
	cleanAbs := filepath.Clean(absPath)

	if cleanAbs != cleanBase && !strings.HasPrefix(cleanAbs, cleanBase+pathSeparator) {
		return fmt.Errorf("path not in FS. base(%s), path(%s)", baseDir, absPath)
	}

	return nil
}

// extractNodeName extract the node name from repo path.
func extractNodeName(repoPath string) (string, error) {
	name := path.Base(repoPath)
	if name == repoPathRoot || name == repoPathCurrentDir || name == repoPathEmpty {
		return "", fmt.Errorf("invalid repo path: %s", repoPath)
	}

	return name, nil
}

// mkdirAll creates a directory and all parent directories if needed.
func mkdirAll(absPath string) error {
	if err := os.MkdirAll(absPath, mkdirMode); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// statNode gets file info for the specified path.
func statNode(absPath string) (os.FileInfo, error) {
	info, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("node not found. path(%s): %w", absPath, os.ErrNotExist)
		}

		return nil, fmt.Errorf("failed to stat node: %w", err)
	}

	return info, nil
}

// writeFile writes data to a file, creating parent directories if needed.
func writeFile(absPath string, data io.Reader) (int64, error) {
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, mkdirMode); err != nil {
		return 0, fmt.Errorf("failed to create parent directory: %w", err)
	}

	file, err := os.Create(absPath) // nolint:gosec
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close() // nolint:errcheck

	n, err := io.Copy(file, data)
	if err != nil {
		return 0, fmt.Errorf("failed to write file: %w", err)
	}

	return n, nil
}

// fileExists checks if a file exists and is not a directory.
func fileExists(absPath string) (bool, error) {
	info, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return !info.IsDir(), nil
}

// readDir reads a directory and returns file info for each entry.
func readDir(absPath string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("directory not found: %w", os.ErrNotExist)
		}

		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	return entries, nil
}

// getParentPath returns the parent path of a repo path.
func getParentPath(repoPath string) string {
	parent := path.Dir(repoPath)
	if parent == repoPathCurrentDir || parent == repoPathEmpty {
		return repoPathRoot
	}

	return parent
}
