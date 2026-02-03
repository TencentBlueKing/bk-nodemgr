/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package bkrepo provides the BKRepo mock API storage.
package bkrepo

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bkrepo"
)

// storage provides thread-safe disk-backed storage for BKRepo mock data.
type storage struct {
	mu      sync.RWMutex
	baseDir string
}

// newStorage creates a new storage instance from config.
func newStorage(conf *Config) *storage {
	return &storage{
		baseDir: conf.BaseDir,
	}
}

// resolvePath cleans the repo path, builds the absolute filesystem path, and validates it.
func (s *storage) resolvePath(projectID, repoName, repoPath string) (string, string, error) {
	cleanedPath := cleanRepoPath(repoPath)
	absPath := getAbsPath(s.baseDir, projectID, repoName, cleanedPath)

	if err := checkPathInFS(s.baseDir, absPath); err != nil {
		return "", "", fmt.Errorf("failed to check path in fs: %w", err)
	}

	return cleanedPath, absPath, nil
}

// CreateDir creates a directory at the specified path.
func (s *storage) CreateDir(projectID, repoName, repoPath string) (*bkrepo.MkdirResp, error) {
	_, absPath, err := s.resolvePath(projectID, repoName, repoPath)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := mkdirAll(absPath); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &bkrepo.MkdirResp{}, nil
}

// QueryNodeInfo queries node information for the specified path.
func (s *storage) QueryNodeInfo(projectID, repoName, repoPath string) (*bkrepo.QueryNodeInfoResp, error) {
	cleanedPath, absPath, err := s.resolvePath(projectID, repoName, repoPath)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// get file info.
	info, err := statNode(absPath)
	if err != nil {
		return nil, err
	}

	name, err := extractNodeName(cleanedPath)
	if err != nil {
		return nil, err
	}

	parentPath := getParentPath(cleanedPath)

	// build node info.
	nodeInfo := &bkrepo.NodeInfo{
		ProjectID: projectID,
		RepoName:  repoName,
		Path:      parentPath,
		Name:      name,
		FullPath:  cleanedPath,
		Folder:    info.IsDir(),
		Size:      int(info.Size()),
	}

	return &bkrepo.QueryNodeInfoResp{
		NodeInfo: *nodeInfo,
	}, nil
}

// UploadFile uploads a file to the specified path.
func (s *storage) UploadFile(projectID, repoName, repoPath string, data io.Reader, overwrite bool) (*bkrepo.UploadFileResp, error) {
	cleanedPath, absPath, err := s.resolvePath(projectID, repoName, repoPath)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// check if file exists.
	exists, err := fileExists(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check file existence: %w", err)
	}

	if exists && !overwrite {
		return nil, fmt.Errorf("failed to upload file, file already exists: %w", os.ErrExist)
	}

	// write file to disk.
	size, err := writeFile(absPath, data)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	name, err := extractNodeName(cleanedPath)
	if err != nil {
		return nil, err
	}

	parentPath := getParentPath(cleanedPath)

	return &bkrepo.UploadFileResp{
		ProjectID: projectID,
		RepoName:  repoName,
		Path:      parentPath,
		Name:      name,
		FullPath:  cleanedPath,
		Folder:    false,
		Size:      size,
	}, nil
}

// DownloadFile reads a file from the specified path and returns its content reader and info.
func (s *storage) DownloadFile(projectID, repoName, repoPath string) (io.ReadCloser, os.FileInfo, error) {
	_, absPath, err := s.resolvePath(projectID, repoName, repoPath)
	if err != nil {
		return nil, nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// get file info.
	info, err := statNode(absPath)
	if err != nil {
		return nil, nil, err
	}

	// check if the path is a directory.
	if info.IsDir() {
		return nil, nil, fmt.Errorf("failed to download directory: %w", os.ErrInvalid)
	}

	// open file for reading.
	file, err := os.Open(absPath) // nolint:gosec
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, info, nil
}

// ListNode lists nodes in the specified directory with pagination.
func (s *storage) ListNode(projectID, repoName, repoPath string, pageNumber, pageSize int) (*bkrepo.ListNodeResp, error) {
	cleanedPath, absPath, err := s.resolvePath(projectID, repoName, repoPath)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// read directory.
	entries, err := readDir(absPath)
	if err != nil {
		return nil, err
	}

	// calculate pagination slice bounds.
	start, end := calcPaginationBounds(len(entries), pageNumber, pageSize)

	// build records.
	records := make([]bkrepo.NodeRecord, end-start)
	for idx, entry := range entries[start:end] {
		name := entry.Name()
		records[idx] = bkrepo.NodeRecord{
			Name:     name,
			FullPath: path.Join(cleanedPath, name),
			Folder:   entry.IsDir(),
		}
	}

	return &bkrepo.ListNodeResp{
		PageNumber: pageNumber,
		PageSize:   pageSize,
		Records:    records,
	}, nil
}

// calcPaginationBounds returns the slice bounds [start, end) for the given page number and page size.
// nolint:gocognit
func calcPaginationBounds(total, pageNumber, pageSize int) (int, int) {
	if pageNumber < 1 || pageSize < 1 || total <= 0 {
		return 0, 0
	}

	start := (pageNumber - 1) * pageSize
	end := start + pageSize

	if start > total {
		start = total
	}

	if end > total {
		end = total
	}

	return start, end
}
