/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package local

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/filelock"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/spf13/afero"
)

const defaultBufferSize = 32 * 1024 // 32KB usually has better performance.

// NewLocalDir creates a new LocalDir.
func NewLocalDir(fullPath string, logger logger.Logger) (*LocalDir, error) {
	exists, err := afero.Exists(rFs(), fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file exists, err: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("file does not exist, fullPath(%s)", fullPath)
	}

	// check if path is a dir.
	isDir, err := afero.IsDir(rFs(), fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if path is dir, err: %w", err)
	}

	if !isDir {
		return nil, fmt.Errorf("path is not a dir, fullPath(%s)", fullPath)
	}

	group := &LocalDir{
		name:     filepath.Base(fullPath),
		fullPath: fullPath,
		fileMap:  make(map[string]*LocalFile),
		subDirs:  make([]*LocalDir, 0),
		logger:   logger,
	}

	err = group.loadContent()
	if err != nil {
		return nil, fmt.Errorf("failed to load directory content, err: %w", err)
	}

	return group, nil
}

func (group *LocalDir) loadContent() error {
	entries, err := afero.ReadDir(rFs(), group.fullPath)
	if err != nil {
		return fmt.Errorf("read dir failed, err: %w", err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(group.fullPath, entry.Name())

		if entry.IsDir() {
			subDir, err := NewLocalDir(fullPath, group.logger)
			if err != nil {
				group.logger.Warnf("failed to create local file group, subgroup(%s), err: %w", fullPath, err)
				continue
			}

			group.subDirs = append(group.subDirs, subDir)
		} else {
			file, err := NewLocalFile(fullPath)
			if err != nil {
				group.logger.Warnf("failed to create local file, file(%s), err: %w", fullPath, err)
				continue
			}

			group.fileMap[entry.Name()] = file
		}
	}

	return nil
}

// LocalDir local file group.
// nolint: revive
type LocalDir struct {
	name     string
	fileMap  map[string]*LocalFile
	fullPath string
	subDirs  []*LocalDir
	logger   logger.Logger

	mutex sync.Mutex
}

// Name the name of file group.
func (group *LocalDir) Name() string {
	return group.name
}

// SubGroups the sub groups of file group.
func (group *LocalDir) SubGroups(_ context.Context) ([]iface.FileGroup, error) {
	subGroups := make([]iface.FileGroup, 0, len(group.subDirs))
	for _, subDir := range group.subDirs {
		subGroups = append(subGroups, subDir)
	}

	return subGroups, nil
}

// GetFile the func will get a file from the file group.
func (group *LocalDir) GetFile(_ context.Context, name string) (iface.File, error) {
	group.mutex.Lock()
	defer group.mutex.Unlock()

	file, ok := group.fileMap[name]
	if !ok {
		return nil, fmt.Errorf("file not found, name(%s)", name)
	}

	return file, nil
}

// AllFiles the files of file group.
func (group *LocalDir) AllFiles(_ context.Context) ([]iface.File, error) {
	group.mutex.Lock()
	defer group.mutex.Unlock()

	files := make([]iface.File, 0, len(group.fileMap))
	for _, file := range group.fileMap {
		files = append(files, file)
	}

	return files, nil
}

// Store the func will store a file into the file group.
func (group *LocalDir) Store(ctx context.Context, info iface.FileInfo, reader io.ReadCloser, overwrite bool) error {
	group.mutex.Lock()
	defer group.mutex.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	if reader == nil {
		return errors.New("file cannot be nil")
	}

	// check dir exist or not.
	exists, err := afero.DirExists(wFs(), group.fullPath)
	if err != nil {
		return fmt.Errorf("check directory existence failed, err: %w", err)
	}

	if !exists {
		err = wFs().MkdirAll(group.fullPath, 0755) // nolint:mnd
		if err != nil {
			return fmt.Errorf("create dir failed, err: %w", err)
		}

		group.logger.Infof("successfully create dir, path(%s)", group.fullPath)
	}

	if err != nil {
		return fmt.Errorf("get file info failed, err: %w", err)
	}

	fileFullPath := filepath.Join(group.fullPath, info.Name)

	// check file exist or not.
	if !overwrite {
		exists, err = afero.Exists(wFs(), fileFullPath)
		if err != nil {
			return fmt.Errorf("check file existence failed, err: %w", err)
		}

		if exists {
			return fmt.Errorf("file already exist, file-name(%s)", fileFullPath)
		}

		flock := filelock.NewFileLock(fileFullPath)
		if err = flock.TryLock(); err != nil {
			return fmt.Errorf("lock file failed, err: %w", err)
		}

		defer func() {
			_ = flock.Unlock()
		}()
	}

	defer func() {
		_ = reader.Close()
	}()

	// create file
	lfile, err := wFs().Create(fileFullPath)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}

	defer func() {
		_ = lfile.Close()
	}()

	if err := group.writeDataToFile(ctx, lfile, reader); err != nil {
		return fmt.Errorf("write file content failed: %w", err)
	}

	localFile, err := NewLocalFile(fileFullPath)
	if err != nil {
		return fmt.Errorf("create local file failed, err: %w", err)
	}

	group.fileMap[info.Name] = localFile

	return nil
}

// writeDataToFile write data to local file.
func (group *LocalDir) writeDataToFile(ctx context.Context, lfile afero.File, reader io.ReadCloser) error {
	// use bufio.NewWriter to improve performance.
	writer := bufio.NewWriter(lfile)
	defer func() {
		_ = writer.Flush()
	}()

	// read file content.
	buf := make([]byte, defaultBufferSize)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					if n > 0 {
						_, err = writer.Write(buf[:n])

						return fmt.Errorf("write final buffer failed: %w", err)
					}

					return nil
				}

				return fmt.Errorf("read content failed: %w", err)
			}

			_, err = writer.Write(buf[:n])
			if err != nil {
				return fmt.Errorf("failed to write file content: %w", err)
			}
		}
	}
}

// Remove the func will delete a file from the file group.
func (group *LocalDir) Remove(ctx context.Context, name string) error {
	group.mutex.Lock()
	defer group.mutex.Unlock()

	file, ok := group.fileMap[name]
	if !ok {
		return fmt.Errorf("file not found, name(%s)", name)
	}

	if file.fullPath == "" {
		return fmt.Errorf("file full path is empty, name(%s)", name)
	}

	// delete file whether the remove successfully.
	delete(group.fileMap, name)

	return os.Remove(file.fullPath)
}
