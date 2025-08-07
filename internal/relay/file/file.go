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
	"strconv"
	"strings"
	"sync"
	"time"

	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

const (
	fileRecoveryInterval  = 1 * time.Hour
	defaultExpirationTime = 24 * time.Hour
)

type fileManagerImpl struct {
	baseDir         string
	rootGroup       fileiface.FileGroup
	files           map[string]*cacheInfo
	mutex           sync.RWMutex
	pendingDeletion map[string]*cacheInfo
	logger          logger.Logger
}

type cacheInfo struct {
	fileGroup    fileiface.FileGroup
	lastAccessed time.Time
	fileName     string
	fileLock     sync.Mutex
}

func (info *cacheInfo) updateLastAccessed() {
	info.fileLock.Lock()
	defer info.fileLock.Unlock()
	info.lastAccessed = time.Now()
}

// NewFileManager creates a new file manager.
func NewFileManager(ctx context.Context, baseDir string, logger logger.Logger) IFileManager {
	if err := os.MkdirAll(baseDir, 0750); err != nil { // nolint: mnd
		logger.Errorf("failed to create base dir.basedir(%s): %v", err)
	}

	rootGroup, err := local.NewLocalDir(baseDir, logger)
	if err != nil {
		logger.Errorf("failed to create root group.basedir(%s): %v", err)
	}

	fm := &fileManagerImpl{
		baseDir:         baseDir,
		rootGroup:       rootGroup,
		files:           make(map[string]*cacheInfo),
		pendingDeletion: make(map[string]*cacheInfo),
		logger:          logger,
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
	subGroups, err := fm.rootGroup.SubGroups(ctx)
	if err != nil {
		fm.logger.Errorf("failed to get sub groups: %v", err)
		return
	}

	fileVersions := fm.extractFileVersions(ctx, subGroups)
	fm.sortAndPruneFileVersions(fileVersions)
}

type fileVersion struct {
	subGroup fileiface.FileGroup
	ModTime  time.Time
}

func (fm *fileManagerImpl) extractFileVersions(ctx context.Context,
	subGroups []fileiface.FileGroup) map[string][]fileVersion {

	fileVersions := make(map[string][]fileVersion)

	for _, group := range subGroups {
		storeTime, err := splitDirectoryTime(group.Name())
		if err != nil {
			continue
		}

		files, err := group.AllFiles(ctx)
		if err != nil || len(files) != 1 {
			fm.logger.Errorf("failed to get files: %s, %v", group.Name(), err)
			continue
		}

		info := files[0].Info()
		v := fileVersion{
			subGroup: group,
			ModTime:  storeTime,
		}
		fileVersions[info.Name] = append(fileVersions[info.Name], v)
	}

	return fileVersions
}

func (fm *fileManagerImpl) sortAndPruneFileVersions(fileVersions map[string][]fileVersion) {
	for filename, versions := range fileVersions {
		sort.Slice(versions, func(i, j int) bool {
			return versions[i].ModTime.After(versions[j].ModTime)
		})

		fm.files[filename] = &cacheInfo{
			fileGroup:    versions[0].subGroup,
			lastAccessed: time.Now(),
			fileName:     filename,
		}

		for _, old := range versions[1:] {
			fm.safeRemove(local.GetLocalFileGroupAbsDirPath(old.subGroup))
		}
	}
}

// StoreFile store file form srcPath.
func (fm *fileManagerImpl) StoreFile(ctx context.Context, srcPath, filename string) (*fileiface.FileInfo, error) {
	destDir, err := fm.cleanOldAndCreateNewLocalDir(filename)
	if err != nil {
		return nil, err
	}

	subGroup, err := local.NewLocalDir(destDir, fm.logger)
	if err != nil {
		fm.safeRemove(destDir)
		return nil, fmt.Errorf("failed to create sub group: %w", err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		fm.safeRemove(destDir)
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	fileInfo := fileiface.FileInfo{Name: filename}
	if err := subGroup.Store(ctx, fileInfo, srcFile, true); err != nil {
		fm.safeRemove(destDir)
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	file, err := subGroup.GetFile(ctx, filename)
	if err != nil {
		fm.safeRemove(destDir)
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	storedInfo := file.Info()

	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	fm.files[filename] = &cacheInfo{
		fileGroup:    subGroup,
		fileName:     filename,
		lastAccessed: time.Now(),
	}

	return &storedInfo, nil
}

// GetFileInfo get file info, this func will refresh the file survival time.
func (fm *fileManagerImpl) GetFile(ctx context.Context, filename string) (
	fileiface.File, error) {

	fm.mutex.RLock()
	info, exists := fm.files[filename]
	fm.mutex.RUnlock()

	if !exists {
		return nil, os.ErrNotExist
	}

	info.updateLastAccessed()

	return info.fileGroup.GetFile(ctx, filename)
}

// FileExists check file exists.
func (fm *fileManagerImpl) FileExists(ctx context.Context, filename, mD5 string) bool {
	info, err := fm.GetFile(ctx, filename)
	if err != nil {
		fm.logger.Errorf("failed to get file. filename(%s): %v", filename, err)
		return false
	}

	filePath := local.GetLocalFileAbsFilePath(info)
	stat, err := os.Stat(filePath)
	if err != nil {
		fm.logger.Errorf("failed to stat file. filepath(%s): %v", filePath, err)
		return false
	}

	currentMD5 := info.Info().MD5
	if info.Info().ModTime != stat.ModTime() {
		currentMD5, err = local.MD5SumWithBuffer(filePath)
		if err != nil {
			fm.logger.Errorf("failed to get file md5. filepath(%s): %v", filePath, err)
			return false
		}
		// TODO: write md5 and modtime to filegroup
	}

	return currentMD5 == mD5
}

func (fm *fileManagerImpl) runGC(_ context.Context, maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)

	fm.mutex.RLock()
	for _, info := range fm.files {
		info.fileLock.Lock()
		defer info.fileLock.Unlock()
		if !info.lastAccessed.Before(cutoff) {
			continue
		}
		delete(fm.files, info.fileName)
		go fm.safeRemove(local.GetLocalFileGroupAbsDirPath(info.fileGroup))
	}
	defer fm.mutex.RUnlock()

	for _, info := range fm.pendingDeletion {
		info.fileLock.Lock()
		defer info.fileLock.Unlock()
		if !info.lastAccessed.Before(cutoff) {
			continue
		}

		go fm.safeRemove(local.GetLocalFileGroupAbsDirPath(info.fileGroup))
	}
}

// safeRemove remove file from file manager.check the file is in baseDir and legal dir.
func (fm *fileManagerImpl) safeRemove(groupPath string) {
	if !isSubPath(groupPath, fm.baseDir) {
		fm.logger.Errorf("attempted to remove file outside of baseDir. path(%s)", groupPath)

		return
	}
	if err := removeAll(groupPath); err != nil {
		fm.logger.Errorf("failed to remove group. groupPath(%s): %v",
			groupPath, err)
	}
}

func isSubPath(targetPath, baseDir string) bool {
	rel, err := filepath.Rel(baseDir, targetPath)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, "..") && rel != ".."
}

func removeAll(absPath string) error {
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

func splitDirectoryTime(groupName string) (time.Time, error) {
	parts := strings.SplitN(groupName, "_", 2) // nolint: mnd
	if len(parts) < 2 {                        // nolint: mnd
		return time.Time{}, errors.New("invalid group name")
	}
	storeTime, err := time.Parse("20060102150405", parts[0])

	return storeTime, err
}

func getDirSeq(dirName string) int64 {
	splitDirectory := strings.Split(dirName, "_")
	if len(splitDirectory) < 2 { // nolint: mnd
		return 0
	}
	seqInt, _ := strconv.Atoi(splitDirectory[1])

	return int64(seqInt)
}

func (fm *fileManagerImpl) cleanOldAndCreateNewLocalDir(filename string) (string, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	var seq int64
	if oldInfo, exists := fm.files[filename]; exists {
		oldPath := local.GetLocalFileGroupAbsFilePath(oldInfo.fileGroup, filename)
		fm.pendingDeletion[oldPath] = oldInfo
		seq = getDirSeq(oldInfo.fileGroup.Name()) + 1
	}

	destDir := filepath.Join(fm.baseDir, getStorageDirName(seq))
	if err := os.MkdirAll(destDir, 0750); err != nil { // nolint: mnd
		return "", fmt.Errorf("failed to create store dir. destDir(%s): %w", destDir, err)
	}

	return destDir, nil
}

func getStorageDirName(seq int64) string {
	return fmt.Sprintf("%s_%d", time.Now().Format("20060102150405"), seq)
}
