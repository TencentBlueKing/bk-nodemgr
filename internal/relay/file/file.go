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
	"sync/atomic"
	"time"

	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

const (
	fileRecoveryInterval  = 1 * time.Hour
	defaultExpirationTime = 24 * time.Hour

	dirDot = "."

	fileNumbers = 1
	fileIndex   = 0
)

type fileManagerImpl struct {
	baseDir   string
	baseGroup fileiface.FileGroup

	fileNameToKeyMap map[string]string     // filename -> key
	filesRegistryMap map[string]*cacheInfo // key -> cacheInfo
	mutex            sync.RWMutex

	dirSequence atomic.Int64

	logger logger.ILogger
}

type cacheInfo struct {
	fileName   string
	fileTmpDir fileiface.FileGroup

	lastAccessed     time.Time
	lastAccessedLock sync.Mutex
}

func (info *cacheInfo) updateLastAccessed() {
	info.lastAccessedLock.Lock()
	defer info.lastAccessedLock.Unlock()

	info.lastAccessed = time.Now()
}

func (info *cacheInfo) isExpired(cutoffTime time.Time) bool {
	info.lastAccessedLock.Lock()
	defer info.lastAccessedLock.Unlock()

	return info.lastAccessed.Before(cutoffTime)
}

// NewFileManager creates a new file manager.
func NewFileManager(ctx context.Context, baseDir string, logger logger.ILogger) IFileManager {
	if err := os.MkdirAll(baseDir, 0750); err != nil { // nolint: mnd
		logger.Errorf("failed to create base dir.basedir(%s): %v", baseDir, err)
	}

	baseGroup, err := local.NewLocalDir(baseDir, logger)
	if err != nil {
		logger.Errorf("failed to create root group.basedir(%s): %v", baseDir, err)
	}

	fm := &fileManagerImpl{
		baseDir:          baseDir,
		baseGroup:        baseGroup,
		fileNameToKeyMap: make(map[string]string),
		filesRegistryMap: make(map[string]*cacheInfo),

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

// restore restores the file manager.
func (fm *fileManagerImpl) restore(ctx context.Context) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	subGroups, err := fm.baseGroup.SubGroups(ctx)
	if err != nil {
		fm.logger.Errorf("failed to get sub groups: %v", err)
		return
	}

	for _, group := range subGroups {
		files, err := group.AllFiles(ctx)
		if err != nil || len(files) != fileNumbers {
			fm.logger.Errorf("failed to get files. groupname(%s), %v", group.Name(), err)
			continue
		}

		// if exist, check modtime.
		info := files[fileIndex].Info()
		filename := info.Name
		if gruopName, ok := fm.fileNameToKeyMap[filename]; ok {
			existsGroup := fm.filesRegistryMap[gruopName].fileTmpDir
			file, err := existsGroup.GetFile(ctx, filename)
			if err != nil {
				fm.logger.Errorf("failed to get file. filename(%s): %v", filename, err)
				continue
			}

			modtime := file.Info().ModTime
			if modtime.After(info.ModTime) {
				continue
			}
			fm.fileNameToKeyMap[filename] = gruopName
		}

		groupName := group.Name()

		fm.filesRegistryMap[groupName] = &cacheInfo{
			fileTmpDir:   group,
			lastAccessed: time.Now(),
			fileName:     info.Name,
		}
	}
}

// StoreFile store file form srcPath. return the cache file info.
func (fm *fileManagerImpl) StoreFile(ctx context.Context, srcPath, filename string) (*fileiface.FileInfo, error) {
	destDir, subGroup, err := fm.createNewLocalDir()
	if err != nil {
		return nil, err
	}

	shouldClean := true
	defer func() {
		if shouldClean {
			fm.safeRemove(destDir)
		}
	}()

	filepath := filepath.Join(srcPath, filename)
	srcFile, err := os.Open(filepath) // nolint: gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close() // nolint: errcheck

	if err := subGroup.Store(ctx, fileiface.FileInfo{Name: filename}, srcFile, true); err != nil {
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	file, err := subGroup.GetFile(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	fm.fileNameToKeyMap[filename] = destDir
	fm.filesRegistryMap[destDir] = &cacheInfo{
		fileTmpDir:   subGroup,
		lastAccessed: time.Now(),
		fileName:     filename,
	}

	storedInfo := file.Info()
	shouldClean = false

	return &storedInfo, nil
}

// GetFileInfo get file info, this func will refresh the file survival time.
func (fm *fileManagerImpl) GetFile(ctx context.Context, filename string) (fileiface.File, error) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	groupDir, exists := fm.fileNameToKeyMap[filename]
	if !exists {
		return nil, fmt.Errorf("file not found. fliename(%s)", filename)
	}

	info, ok := fm.filesRegistryMap[groupDir]
	if !ok || info == nil {
		return nil, fmt.Errorf("fileinfo not found. groupdir(%s)", groupDir)
	}

	info.updateLastAccessed()
	fm.logger.Infof("update last access time. filename(%s)", filename)

	return info.fileTmpDir.GetFile(ctx, filename)
}

// FileExists check file exists.
func (fm *fileManagerImpl) FileExists(ctx context.Context, filename, md5 string) bool {
	info, err := fm.GetFile(ctx, filename)
	if err != nil {
		fm.logger.Infof("file not exists. filename(%s): %v", filename, err)
		return false
	}

	fm.logger.Infof("check file exists. filename(%s) , expected md5(%s), actual md5(%s)", filename, md5, info.Info().MD5)

	return info.Info().MD5 == md5
}

func (fm *fileManagerImpl) runGC(_ context.Context, maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)

	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	keysToDelete := make([]string, 0)
	for key, info := range fm.filesRegistryMap {
		if info.isExpired(cutoff) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		info := fm.filesRegistryMap[key]
		groupname := info.fileTmpDir.Name()

		delete(fm.filesRegistryMap, key)
		go fm.safeRemove(local.GetLocalFileGroupAbsDirPath(info.fileTmpDir))

		fm.logger.Infof("removing expired group(%s)", groupname)
	}
}

// safeRemove remove file from file manager.check the file is in baseDir and legal dir.
func (fm *fileManagerImpl) safeRemove(groupPath string) {
	if !isSubPath(groupPath, fm.baseDir) {
		fm.logger.Errorf("attempted to remove file outside of basedir. path(%s)", groupPath)

		return
	}
	if err := removeAll(groupPath); err != nil {
		fm.logger.Errorf("failed to remove group. group path(%s): %v",
			groupPath, err)
	}
}

func isSubPath(targetPath, baseDir string) bool {
	rel, err := filepath.Rel(baseDir, targetPath)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, dirDot) && rel != dirDot
}

func removeAll(absPath string) error {
	if absPath == "" ||
		absPath == "/" ||
		strings.HasPrefix(absPath, "/dev/") ||
		strings.HasPrefix(absPath, "/sys/") ||
		strings.HasPrefix(absPath, "/proc/") {

		return fmt.Errorf("failed to remove all, got invalid path. path(%s)", absPath)
	}

	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("failed to remove all. path(%s): %w", absPath, err)
	}

	return nil
}

func (fm *fileManagerImpl) createNewLocalDir() (string, *local.LocalDir, error) {
	destDir := filepath.Join(fm.baseDir, fm.getStorageDirName())
	if err := os.MkdirAll(destDir, 0750); err != nil { // nolint: mnd
		return "", nil, fmt.Errorf("failed to create store dir. dest dir(%s): %w", destDir, err)
	}

	subGroup, err := local.NewLocalDir(destDir, fm.logger)
	if err != nil {
		defer fm.safeRemove(destDir)
		return "", nil, fmt.Errorf("failed to create sub group: %w", err)
	}

	return destDir, subGroup, nil
}

func (fm *fileManagerImpl) getStorageDirName() string {
	fm.dirSequence.Add(1)
	return fmt.Sprintf("%s_%d", time.Now().Format("20060102150405"), fm.dirSequence.Load())
}
