/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package file defines the file manager interface.
package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"

	"github.com/google/uuid"
)

type fileManagerImpl struct {
	baseDir   string
	rootGroup fileiface.FileGroup
	files     map[string]*fileiface.FileInfo
	mutex     sync.RWMutex
	refs      map[string]int
	refLock   sync.Mutex
	logger    logger.Logger
}

// NewFileManager creates a new file manager.
func NewFileManager(baseDir string, logger logger.Logger) FileManager {
	if err := os.MkdirAll(baseDir, 0750); err != nil { // nolint: mnd
		logger.Errorf("failed to create base dir.basedir(%s): %v", err)
	}

	rootGroup, err := local.NewLocalDir(baseDir, logger)
	if err != nil {
		logger.Warnf("failed to create root group.basedir(%s): %v", err)
	}

	fm := &fileManagerImpl{
		baseDir:   baseDir,
		rootGroup: rootGroup,
		files:     make(map[string]*fileiface.FileInfo),
		refs:      make(map[string]int),
		logger:    logger,
	}
	fm.restore()

	return fm
}

func (fm *fileManagerImpl) restore() {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	fm.files = make(map[string]*fileiface.FileInfo)

	subGroups, err := fm.rootGroup.SubGroups(context.Background())
	if err != nil {
		fm.logger.Warnf("failed to get sub groups: %v", err)
		return
	}

	fileInfos := make(map[string]*fileiface.FileInfo)
	for _, group := range subGroups {
		groupName := group.Name()
		parts := strings.SplitN(groupName, "_", 2) // nolint: mnd
		if len(parts) < 2 {                        // nolint: mnd
			continue
		}
		storeTime, _ := time.Parse("20060102150405", parts[0])

		files, err := group.AllFiles(context.Background())
		if err != nil {
			fm.logger.Warnf("failed to get files. groupname(%s): %v", groupName, err)
			continue
		}

		info := files[0].Info()

		if existing, exists := fileInfos[info.Name]; exists && storeTime.Before(existing.ModTime) {
			continue
		}

		absDirs := group.AbsDirs()
		filePath := filepath.Join(filepath.Join(absDirs...), info.Name)

		fileInfo := &fileiface.FileInfo{
			Name:         info.Name,
			Size:         info.Size,
			MD5:          info.MD5,
			ModTime:      storeTime, // specially set
			FullPath:     filePath,
			Description:  info.Description,
			ExtendFields: info.ExtendFields,
		}

		// Update or add the file information
		fileInfos[info.Name] = fileInfo
	}

	for filename, info := range fileInfos {
		fm.files[filename] = info
	}
}

// StoreFile store file form srcPath.
func (fm *fileManagerImpl) StoreFile(srcPath, filename string) (*fileiface.FileInfo, error) {
	destDir := getStoreDestDir(fm.baseDir)

	if err := os.MkdirAll(destDir, 0750); err != nil { // nolint: mnd
		return nil, fmt.Errorf("failed to create store dir. destDir(%s): %w", destDir, err)
	}

	subGroup, err := local.NewLocalDir(destDir, fm.logger)
	if err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("failed to create sub group. destDir(%s): %w", destDir, err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("failed to open file. srcPath(%s): %w", srcPath, err)
	}
	defer srcFile.Close()

	fileInfo := fileiface.FileInfo{
		Name: filename,
	}

	if err := subGroup.Store(context.Background(), fileInfo, srcFile, true); err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("failed to store file. filename(%s): %w", filename, err)
	}

	file, err := subGroup.GetFile(context.Background(), filename)
	if err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("failed to get file. filename(%s): %w", filename, err)
	}

	storedInfo := file.Info()

	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if oldInfo, exists := fm.files[filename]; exists {
		fm.safeRemoveFile(oldInfo.FullPath)
	}

	fm.files[filename] = &storedInfo

	return &storedInfo, nil
}

func (fm *fileManagerImpl) GetFileInfo(filename string) (*fileiface.FileInfo, error) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	info, exists := fm.files[filename]
	if !exists {
		return nil, os.ErrNotExist
	}

	return info, nil
}

func (fm *fileManagerImpl) DownloadFile(filename string) (iface.FileContent, *fileiface.FileInfo, error) {
	info, err := fm.GetFileInfo(filename)
	if err != nil {
		return nil, nil, err
	}

	return &fileContent{
		path: info.FullPath,
		fm:   fm,
	}, info, nil
}

func (fm *fileManagerImpl) FileExists(filename string) bool {
	_, err := fm.GetFileInfo(filename)
	return err == nil
}

func (fm *fileManagerImpl) RunGC(ctx context.Context, maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)

	toDelete := make([]*fileiface.FileInfo, 0)
	fm.mutex.RLock()
	for _, info := range fm.files {
		if !info.ModTime.Before(cutoff) {
			continue
		}
		toDelete = append(toDelete, info)
	}

	fm.mutex.RUnlock()

	for _, info := range toDelete {
		select {
		case <-ctx.Done():
			return
		default:
			if !fm.canSafelyRemove(info.FullPath) {
				continue
			}
			fm.safeRemoveFile(info.FullPath)
			fm.mutex.Lock()
			delete(fm.files, info.Name)
			fm.mutex.Unlock()

			fm.refLock.Lock()
			delete(fm.refs, info.FullPath)
			fm.refLock.Unlock()

			fm.logger.Infof("remove expired file. filename(%s),filepath(%s)", info.Name, info.FullPath)
		}
	}
}

func (fm *fileManagerImpl) safeRemoveFile(path string) {
	if fm.getRefCount(path) > 0 {
		return
	}

	dir := filepath.Dir(path)

	if err := os.RemoveAll(dir); err != nil {
		fm.logger.Errorf("failed to remove dir. dir(%s): %v", dir, err)
	}
}

func (fm *fileManagerImpl) canSafelyRemove(path string) bool {
	if fm.getRefCount(path) > 0 {
		return false
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func (fm *fileManagerImpl) decrementRef(path string) {
	fm.refLock.Lock()
	defer fm.refLock.Unlock()

	if count, exists := fm.refs[path]; exists {
		if count <= 1 {
			delete(fm.refs, path)
		} else {
			fm.refs[path] = count - 1
		}
	}
}

func (fm *fileManagerImpl) getRefCount(path string) int {
	fm.refLock.Lock()
	defer fm.refLock.Unlock()

	return fm.refs[path]
}

func getStoreDestDir(baseDir string) string {
	dirName := fmt.Sprintf("%s_%s", time.Now().Format("20060102150405"), uuid.New().String())
	return filepath.Join(baseDir, dirName)
}
