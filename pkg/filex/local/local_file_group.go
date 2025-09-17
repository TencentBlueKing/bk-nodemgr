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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/filelock"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/spf13/afero"
)

const defaultBufferSize = 32 * 1024 // 32KB usually has better performance.

// NewLocalDir creates a new LocalDir.
func NewLocalDir(fullPath string, logger logger.ILogger) (*LocalDir, error) {
	exists, err := afero.Exists(rFs(), fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file exists: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("file does not exist, fullPath(%s)", fullPath)
	}

	// check if path is a dir.
	isDir, err := afero.IsDir(rFs(), fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if path is dir: %w", err)
	}

	if !isDir {
		return nil, fmt.Errorf("path is not a dir, fullPath(%s)", fullPath)
	}

	group := &LocalDir{
		name:     filepath.Base(fullPath),
		fullPath: fullPath,
		absDirs:  fileiface.ConvertAbsPathToAbsDirs(fullPath),
		logger:   logger,
	}

	return group, nil
}

// LocalDir local file group.
// nolint: revive
type LocalDir struct {
	name     string
	fullPath string
	absDirs  []string
	logger   logger.ILogger
}

// Name the name of file group.
func (group *LocalDir) Name() string {
	return group.name
}

// SubGroups the sub groups of file group.
func (group *LocalDir) SubGroups(_ context.Context) ([]fileiface.FileGroup, error) {
	entries, err := afero.ReadDir(rFs(), group.fullPath)
	if err != nil {
		return nil, fmt.Errorf("read dir failed: %w", err)
	}

	subGroups := make([]fileiface.FileGroup, 0)
	for _, entry := range entries {
		fullPath := filepath.Join(group.fullPath, entry.Name())

		if entry.IsDir() {
			subDir, err := NewLocalDir(fullPath, group.logger)
			if err != nil {
				return nil, fmt.Errorf("failed to create local file group, subgroup(%s): %w", fullPath, err)
			}

			subGroups = append(subGroups, subDir)
		}
	}

	return subGroups, nil
}

// AllFiles the files of file group.
func (group *LocalDir) AllFiles(_ context.Context) ([]fileiface.File, error) {
	entries, err := afero.ReadDir(rFs(), group.fullPath)
	if err != nil {
		return nil, fmt.Errorf("read dir failed: %w", err)
	}

	files := make([]fileiface.File, 0)
	for _, entry := range entries {
		fullPath := filepath.Join(group.fullPath, entry.Name())

		if !entry.IsDir() {
			file, err := NewLocalFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create local file. file(%s): %w", fullPath, err)
			}

			files = append(files, file)
		}
	}

	return files, nil
}

// GetFile the func will get a file from the file group.
func (group *LocalDir) GetFile(_ context.Context, name string) (fileiface.File, error) {
	fullPath := filepath.Join(group.fullPath, name)
	file, err := NewLocalFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create local file. file(%s): %w", fullPath, err)
	}

	return file, nil
}

// Store the func will store a file into the file group.
func (group *LocalDir) Store(ctx context.Context, info fileiface.FileInfo, reader io.ReadCloser, overwrite bool) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	if reader == nil {
		return errors.New("file cannot be nil")
	}

	// check dir exist or not.
	exists, err := afero.DirExists(wFs(), group.fullPath)
	if err != nil {
		return fmt.Errorf("check directory existence failed: %w", err)
	}

	if !exists {
		err = wFs().MkdirAll(group.fullPath, 0755) // nolint:mnd
		if err != nil {
			return fmt.Errorf("create dir failed: %w", err)
		}

		group.logger.Infof("successfully create dir, path(%s)", group.fullPath)
	}

	if err != nil {
		return fmt.Errorf("get file info failed: %w", err)
	}

	fileFullPath := filepath.Join(group.fullPath, info.Name)

	// check file exist or not.
	if !overwrite {
		exists, err = afero.Exists(wFs(), fileFullPath)
		if err != nil {
			return fmt.Errorf("check file existence failed: %w", err)
		}

		if exists {
			return fmt.Errorf("file already exist, file-name(%s)", fileFullPath)
		}

		flock := filelock.NewFileLock(fileFullPath)
		if err = flock.TryLock(); err != nil {
			return fmt.Errorf("lock file failed: %w", err)
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

// AbsDirs the func will return the abs dirs of file group.
func (group *LocalDir) AbsDirs() []string {
	return group.absDirs
}

// Remove the func will delete a file from the file group.
func (group *LocalDir) Remove(_ context.Context, name string) error {
	fullPath := filepath.Join(group.fullPath, name)
	if fullPath == "" {
		return fmt.Errorf("file full path is empty, name(%s)", name)
	}

	return os.Remove(fullPath)
}
