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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"

	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"

	"github.com/google/uuid"
)

const (
	fileRecoveryInterval  = 1 * time.Hour
	defaultExpirationTime = 24 * time.Hour
)

type fileManagerImpl struct {
	baseDir         string
	rootGroup       fileiface.FileGroup
	files           map[string]*fileiface.FileInfo
	mutex           sync.RWMutex
	refs            map[string]int
	refLock         sync.Mutex
	pendingDeletion map[string]*fileiface.FileInfo

	logger logger.Logger
}

// NewFileManager creates a new file manager.
func NewFileManager(ctx context.Context, baseDir string, logger logger.Logger) IFileManager {
	if err := os.MkdirAll(baseDir, 0750); err != nil { // nolint: mnd
		logger.Errorf("failed to create base dir.basedir(%s): %v", err)
	}

	rootGroup, err := local.NewLocalDir(baseDir, logger)
	if err != nil {
		logger.Warnf("failed to create root group.basedir(%s): %v", err)
	}

	fm := &fileManagerImpl{
		baseDir:         baseDir,
		rootGroup:       rootGroup,
		files:           make(map[string]*fileiface.FileInfo),
		refs:            make(map[string]int),
		pendingDeletion: make(map[string]*fileiface.FileInfo),

		logger: logger,
	}
	fm.restore(ctx)

	go func() {
		ticker := time.NewTicker(fileRecoveryInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fm.runGC(ctx, defaultExpirationTime)
			}
		}
	}()

	return fm
}

func (fm *fileManagerImpl) restore(ctx context.Context) {
	fm.files = make(map[string]*fileiface.FileInfo)

	subGroups, err := fm.rootGroup.SubGroups(ctx)
	if err != nil {
		fm.logger.Warnf("failed to get sub groups: %v", err)
		return
	}
	type fileVersion struct {
		GroupPath string
		ModTime   time.Time
		Info      *fileiface.FileInfo
	}
	fileVersions := make(map[string][]fileVersion)

	for _, group := range subGroups {
		storeTime, err := splitDirectoryTime(group.Name())
		if err != nil {
			continue
		}

		files, err := group.AllFiles(ctx)
		if err != nil || len(files) == 0 {
			fm.logger.Warnf("failed to get files in group %s: %v", group.Name(), err)
			continue
		}

		info := files[0].Info()
		absDirs := group.AbsDirs()
		filePath := filepath.Join(filepath.Join(absDirs...), info.Name)

		v := fileVersion{
			GroupPath: filepath.Join(fm.baseDir, group.Name()),
			ModTime:   storeTime,
			Info: &fileiface.FileInfo{
				Name:         info.Name,
				Size:         info.Size,
				MD5:          info.MD5,
				ModTime:      storeTime,
				FullPath:     filePath,
				Description:  info.Description,
				ExtendFields: info.ExtendFields,
			},
		}

		fileVersions[info.Name] = append(fileVersions[info.Name], v)
	}

	for filename, versions := range fileVersions {
		sort.Slice(versions, func(i, j int) bool {
			return versions[i].ModTime.After(versions[j].ModTime)
		})

		latest := versions[0]
		fm.files[filename] = latest.Info

		for _, old := range versions[1:] {
			if err := os.RemoveAll(old.GroupPath); err != nil {
				fm.logger.Warnf("failed to remove old file. groupPath(%s): %v", old.GroupPath, err)
			}
		}
	}
}

// StoreFile store file form srcPath.
func (fm *fileManagerImpl) StoreFile(ctx context.Context, srcPath, filename string) (*fileiface.FileInfo, error) {
	destDir := getStoreDestDir(fm.baseDir)

	if err := os.MkdirAll(destDir, 0750); err != nil { // nolint: mnd
		return nil, fmt.Errorf("failed to create store dir. destDir(%s): %w", destDir, err)
	}

	subGroup, err := local.NewLocalDir(destDir, fm.logger)
	if err != nil {
		fm.removeAll(destDir)
		return nil, fmt.Errorf("failed to create sub group. destDir(%s): %w", destDir, err)
	}

	srcFile, err := os.Open(srcPath) // nolint: gosec
	if err != nil {
		fm.removeAll(destDir)
		return nil, fmt.Errorf("failed to open file. srcPath(%s): %w", srcPath, err)
	}
	defer srcFile.Close() // nolint: errcheck

	if err := subGroup.Store(ctx, fileiface.FileInfo{Name: filename}, srcFile, true); err != nil {
		fm.removeAll(destDir)
		return nil, fmt.Errorf("failed to store file. filename(%s): %w", filename, err)
	}

	file, err := subGroup.GetFile(ctx, filename)
	if err != nil {
		fm.removeAll(destDir)
		return nil, fmt.Errorf("failed to get file. filename(%s): %w", filename, err)
	}

	storedInfo := file.Info()

	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if oldInfo, exists := fm.files[filename]; exists {
		fm.pendingDeletion[oldInfo.FullPath] = oldInfo
	}

	fm.files[filename] = &storedInfo

	return &storedInfo, nil
}

func (fm *fileManagerImpl) GetFileInfo(_ context.Context, filename string) (
	*fileiface.FileInfo, error) {

	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	info, exists := fm.files[filename]
	if !exists {
		return nil, os.ErrNotExist
	}

	return info, nil
}

func (fm *fileManagerImpl) DownloadFile(ctx context.Context, filename string) (fileiface.FileContent,
	*fileiface.FileInfo, error) {

	info, err := fm.GetFileInfo(ctx, filename)
	if err != nil {
		return nil, nil, err
	}

	return &fileContent{
		path: info.FullPath,
		fm:   fm,
	}, info, nil
}

// FileExists check file exists.
func (fm *fileManagerImpl) FileExists(ctx context.Context, filename, mD5 string) bool {
	info, err := fm.GetFileInfo(ctx, filename)
	if err != nil {
		fm.logger.Errorf("failed to get file info. filename(%s): %v", filename, err)
		return false
	}

	currentMD5 := info.MD5

	stat, err := os.Stat(info.FullPath)
	if err != nil {
		fm.logger.Errorf("failed to stat file. fullPath(%s): %v", info.FullPath, err)
		return false
	}

	if info.ModTime != stat.ModTime() {
		currentMD5, err := local.MD5SumWithBuffer(info.FullPath)
		if err != nil {
			fm.logger.Errorf("failed to get file md5. fullPath(%s): %v", info.FullPath, err)
			return false
		}

		info.MD5 = currentMD5
		info.ModTime = stat.ModTime()
		fm.logger.Infof("file has been modified. fullPath(%s)", info.FullPath)
	}

	if currentMD5 != mD5 {
		return false
	}

	return true
}

func (fm *fileManagerImpl) runGC(ctx context.Context, maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)

	allFiles := make([]*fileiface.FileInfo, 0)
	fm.mutex.RLock()
	for _, info := range fm.files {
		if !info.ModTime.Before(cutoff) {
			continue
		}
		allFiles = append(allFiles, info)
	}
	for _, info := range fm.pendingDeletion {
		if !info.ModTime.Before(cutoff) {
			continue
		}
		allFiles = append(allFiles, info)
	}

	fm.mutex.RUnlock()

	for _, info := range allFiles {
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
	dir := filepath.Dir(path)

	if err := os.RemoveAll(dir); err != nil {
		fm.logger.Errorf("failed to remove dir. dir(%s): %v", dir, err)
	}
}

func (fm *fileManagerImpl) canSafelyRemove(path string) bool {
	if fm.getRefCount(path) > 0 {
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

func (fm *fileManagerImpl) removeAll(path string) {
	if err := os.RemoveAll(path); err != nil {
		fm.logger.Errorf("failed to remove dir. dir(%s): %v", path, err)
	}
}

func splitDirectoryTime(groupName string) (time.Time, error) {
	parts := strings.SplitN(groupName, "_", 2) // nolint: mnd
	if len(parts) < 2 {                        // nolint: mnd
		return time.Time{}, errors.New("invalid group name")
	}
	storeTime, err := time.Parse("20060102150405", parts[0])

	return storeTime, err
}

func getStoreDestDir(baseDir string) string {
	dirName := fmt.Sprintf("%s_%s", time.Now().Format("20060102150405"), uuid.New().String())
	return filepath.Join(baseDir, dirName)
}
