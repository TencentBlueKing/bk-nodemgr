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
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
)

const (
	mkdirMode     = 0755
	overwriteMode = 0644

	maxDecompressedSize = 100 * 1024 * 1024 // 100M

	tmpFilePrefix    = ".bknm_tmp_"
	backupFilePrefix = "bknm_bak_"

	backupFilesKeepNum = 3
)

// CheckIntegrity check the integrity of agent file system.
func (handler *AgentHandler) CheckIntegrity(_ context.Context) error {
	if err := handler.checkRegularFile(handler.agentBinFilePath); err != nil {
		return fmt.Errorf("failed to check agent-bin file(%s): %w", handler.agentBinFilePath, err)
	}

	if err := handler.checkRegularFile(handler.agentDaemonFilePath); err != nil {
		return fmt.Errorf("failed to check agent-daemon file(%s): %w", handler.agentBinFilePath, err)
	}

	if err := handler.checkRegularFile(handler.agentConfigFilePath); err != nil {
		return fmt.Errorf("failed to check agent-config file(%s): %w", handler.agentConfigFilePath, err)
	}

	return nil
}

// Init init agent file system.
func (handler *AgentHandler) Init() error {
	if err := handler.mkdirAll(handler.setupDir); err != nil {
		return fmt.Errorf("failed to make setup-dir(%s): %w", handler.setupDir, err)
	}

	if err := handler.mkdirAll(handler.backupDir); err != nil {
		return fmt.Errorf("failed to make backup-dir(%s): %w", handler.backupDir, err)
	}

	if err := handler.mkdirAll(handler.binDir); err != nil {
		return fmt.Errorf("failed to make bin-dir(%s): %w", handler.binDir, err)
	}

	if err := handler.mkdirAll(handler.etcDir); err != nil {
		return fmt.Errorf("failed to make etc-dir(%s): %w", handler.etcDir, err)
	}

	if err := handler.mkdirAll(handler.certDir); err != nil {
		return fmt.Errorf("failed to make cert-dir(%s): %w", handler.certDir, err)
	}

	return nil
}

// Backup backup agent file system.
func (handler *AgentHandler) Backup(ctx context.Context) error {
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(handler.backupDir, backupFilePrefix+string(types.NodeRoleAgent)+"_"+timestamp+".tgz")

	backupSrcs := []string{
		handler.agentBinFilePath,
		handler.agentDaemonFilePath,
		handler.agentConfigFilePath,
		handler.gseRuntimeFileProcFilePath,
		handler.gseRuntimeFileTaskFilePath,
	}

	if certFiles, err := handler.getCertFiles(); err == nil {
		backupSrcs = append(backupSrcs, certFiles...)
	}

	if err := handler.writeTgz(ctx, backupFile, backupSrcs); err != nil {
		return err
	}

	_ = handler.cleanBackupFiles()

	return nil
}

// Purge purge agent file system.
func (handler *AgentHandler) Purge(_ context.Context) error {
	return handler.removeAll(handler.setupDir)
}

// CopyConfigDir copy config file to agent file system.
func (handler *AgentHandler) CopyConfigDir(ctx context.Context, configFileAbsDir string) error {
	if err := filepath.Walk(configFileAbsDir, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		// nolint: gosec
		configFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file(%s): %w", path, err)
		}
		defer func() {
			if errClose := configFile.Close(); errClose != nil {
				err = errors.Join(err, errClose)
			}
		}()

		// copy file in config dir to etc-dir of agent handler fs.
		targetRelativePath := filepath.Join(
			handler.etcDir,
			strings.TrimLeft(strings.TrimPrefix(path, configFileAbsDir), "\\"),
		)

		return handler.overwriteFile(configFile, targetRelativePath)
	}); err != nil {
		return fmt.Errorf("failed to copy config dir: %w", err)
	}

	return nil
}

// UnpackReleasePackage unpack release package to agent file system.
func (handler *AgentHandler) UnpackReleasePackage(
	ctx context.Context, releasePkgAbsPath string, keepOldFileAsTmp bool) error {

	if err := handler.checkDirWritable(handler.setupDir); err != nil {
		return fmt.Errorf("setup dir is not writable: %w", err)
	}

	if err := handler.extractTgz(ctx, releasePkgAbsPath, handler.setupDir, keepOldFileAsTmp); err != nil {
		return fmt.Errorf("failed to extract tgz: %w", err)
	}

	if err := handler.makeBinFilesExecutable(); err != nil {
		return fmt.Errorf("failed to make bin files executable: %w", err)
	}

	return nil
}

// SaveGSERuntimeFile saves generated GSE runtime files before reinstall or upgrade overwrites them.
func (handler *AgentHandler) SaveGSERuntimeFile(ctx context.Context, opts agenthandler.PreserveOptions) error {
	return handler.saveGSERuntimeFiles(ctx, opts)
}

// RestoreGSERuntimeFile restores generated GSE runtime files saved by SaveGSERuntimeFile.
func (handler *AgentHandler) RestoreGSERuntimeFile(ctx context.Context, opts agenthandler.PreserveOptions) error {
	return handler.restoreGSERuntimeFiles(ctx, opts)
}

// Clean cleans all the tmp files and tools in agent file system.
func (handler *AgentHandler) Clean(_ context.Context) error {
	return handler.cleanAllTmpFiles()
}

func (handler *AgentHandler) saveGSERuntimeFiles(ctx context.Context, opts agenthandler.PreserveOptions) error {
	for _, runtimeFile := range handler.gseRuntimeFiles(opts) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := handler.copyRuntimeFileIfExists(runtimeFile.source, runtimeFile.target); err != nil {
			return err
		}
	}

	return nil
}

func (handler *AgentHandler) restoreGSERuntimeFiles(ctx context.Context, opts agenthandler.PreserveOptions) error {
	for _, runtimeFile := range handler.gseRuntimeFiles(opts) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := handler.copyRuntimeFileIfExists(runtimeFile.target, runtimeFile.source); err != nil {
			return err
		}
	}

	return nil
}

func (handler *AgentHandler) copyRuntimeFileIfExists(sourceRelativePath string, targetRelativePath string) error {
	sourceFile, err := handler.openFileForRead(sourceRelativePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to open gse runtime file(%s): %w", sourceRelativePath, err)
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	return handler.overwriteFile(sourceFile, targetRelativePath)
}

type gseRuntimeFile struct {
	source string
	target string
}

func (handler *AgentHandler) gseRuntimeFiles(opts agenthandler.PreserveOptions) []gseRuntimeFile {
	runtimeFiles := make([]gseRuntimeFile, 0, 2)
	if !opts.RenewGSEProc {
		runtimeFiles = append(runtimeFiles, handler.gseRuntimeFile(handler.gseRuntimeFileProcFilePath))
	}
	if !opts.RenewGSETask {
		runtimeFiles = append(runtimeFiles, handler.gseRuntimeFile(handler.gseRuntimeFileTaskFilePath))
	}

	return runtimeFiles
}

func (handler *AgentHandler) gseRuntimeFile(relatedFilePath string) gseRuntimeFile {
	return gseRuntimeFile{
		source: relatedFilePath,
		target: filepath.Join(handler.backupDir, tmpFilePrefix+filepath.Base(relatedFilePath)),
	}
}

func (handler *AgentHandler) getCertFiles() ([]string, error) {
	files := []string{}
	if err := filepath.Walk(handler.getAbsPath(handler.certDir), func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		files = append(files, filepath.Join(handler.certDir, info.Name()))

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

// writeTgz write srcs to target relative-path.
// nolint: gocognit
func (handler *AgentHandler) writeTgz(ctx context.Context, targetRelativePath string, srcs []string) (err error) {
	// create target file.
	targetFile, err := handler.openFileForWrite(targetRelativePath)
	if err != nil {
		return err
	}
	defer func() {
		if errClose := targetFile.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	// target gzip writer.
	gzipWriter := gzip.NewWriter(targetFile)
	defer func() {
		if errClose := gzipWriter.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	// target tar writer.
	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		if errClose := tarWriter.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	for _, src := range srcs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		stat, err := handler.statFile(src)
		if err != nil || !stat.Mode().IsRegular() {
			continue
		}

		// nolint: gosec
		source, err := handler.openFileForRead(src)
		if err != nil {
			continue
		}
		defer func() {
			if errClose := source.Close(); errClose != nil {
				err = errors.Join(err, errClose)
			}
		}()

		if err := tarWriter.WriteHeader(&tar.Header{
			Name:    src,
			Size:    stat.Size(),
			Mode:    int64(stat.Mode()),
			ModTime: time.Now(),
		}); err != nil {
			continue
		}

		limitedReader := &io.LimitedReader{
			R: source,
			N: maxDecompressedSize,
		}
		if _, err := io.Copy(tarWriter, limitedReader); err != nil {
			continue
		}
	}

	return nil
}

// extractTgz extract tgz file to target relative-dir.
// nolint: gocognit,funlen
func (handler *AgentHandler) extractTgz(
	ctx context.Context, sourceAbsPath string, targetRelativeDir string, keepOldFileAsTmp bool) (err error) {

	// nolint: gosec
	sourceFile, err := os.Open(sourceAbsPath)
	if err != nil {
		return fmt.Errorf("failed to open tgz file(%s): %w", sourceAbsPath, err)
	}
	defer func() {
		if errClose := sourceFile.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	// target gzip reader.
	gzipReader, err := gzip.NewReader(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader. file(%s): %w", sourceAbsPath, err)
	}
	defer func() {
		if errClose := gzipReader.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	// target tar reader.
	tarReader := tar.NewReader(gzipReader)

	// range read tar content.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			// end of tar archive.
			break
		}
		if err != nil {
			return fmt.Errorf("read tar content error, file(%s): %v", sourceAbsPath, err)
		}

		// get target path.
		targetRelativePath := filepath.Join(targetRelativeDir, filepath.Clean(filepath.FromSlash(header.Name)))

		// make sure target path is within target dir.
		if err = handler.checkPathInFS(targetRelativePath); err != nil {
			return fmt.Errorf("invalid path in archive. target-path(%s): %w", targetRelativePath, err)
		}

		// handle each file by its type.
		switch header.Typeflag {
		case tar.TypeDir:
			// mkdir all parent directories.
			if err := handler.mkdirAll(targetRelativePath); err != nil {
				return err
			}
		case tar.TypeReg:
			// make sure parent directories exist.
			if err := handler.mkdirAll(filepath.Dir(targetRelativePath)); err != nil {
				return err
			}

			if keepOldFileAsTmp {
				/**
				 * skip extract daemon when need keep old files.
				 * daemon will not be restart while upgrading, so the old files can not be cleaned.
				 * when upgrading node, the daemon will not be upgraded.
				 */
				if filepath.Base(header.Name) == gseAgentDaemonName {
					continue
				}

				if err = handler.renameFileAsTmpFileIfExists(targetRelativePath); err != nil {
					return fmt.Errorf("failed to rename old file(%s): %w", targetRelativePath, err)
				}
			}

			// create file.
			outFile, err := handler.openFileForWrite(targetRelativePath)
			if err != nil {
				return fmt.Errorf("create file failed, path(%s): %v", targetRelativePath, err)
			}

			// write file content.
			limitedReader := &io.LimitedReader{
				R: tarReader,
				N: maxDecompressedSize,
			}
			if _, err := io.Copy(outFile, limitedReader); err != nil {
				_ = outFile.Close()

				return fmt.Errorf("write file content failed, path(%s): %v", targetRelativePath, err)
			}
			_ = outFile.Close()
		default:
		}
	}

	return nil
}

func (handler *AgentHandler) cleanAllTmpFiles() error {
	if err := filepath.Walk(handler.rootAbsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), tmpFilePrefix) {
			tmpFileRelativePath := strings.TrimLeft(strings.TrimPrefix(path, handler.rootAbsDir), "\\")
			_ = handler.removeFile(tmpFileRelativePath)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

type fileList []fs.FileInfo

// Len is the number of elements in the collection.
func (f fileList) Len() int {
	return len(f)
}

// Less reports whether the element with index i
// must sort before the element with index j.
func (f fileList) Less(i, j int) bool {
	return f[i].ModTime().Before(f[j].ModTime())
}

// Swap swaps the elements with indexes i and j.
func (f fileList) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (handler *AgentHandler) cleanBackupFiles() error {
	files, err := os.ReadDir(handler.getAbsPath(handler.backupDir))
	if err != nil {
		return fmt.Errorf("failed to read backup-dir(%s): %w", handler.backupDir, err)
	}

	backupList := make(fileList, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasPrefix(file.Name(), backupFilePrefix) {
			info, err := file.Info()
			if err != nil {
				continue
			}

			backupList = append(backupList, info)
		}
	}

	if len(backupList) > backupFilesKeepNum {
		sort.Sort(backupList)

		for i := 0; i < len(backupList)-backupFilesKeepNum; i++ {
			_ = handler.removeFile(filepath.Join(handler.backupDir, backupList[i].Name()))
		}
	}

	return nil
}
