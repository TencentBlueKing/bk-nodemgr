//go:build windows

/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package windows

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginv2handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

const (
	mkdirMode     = 0755
	overwriteMode = 0644

	maxDecompressedSize = 100 * 1024 * 1024 // 100M

	tmpFilePrefix = ".bknm_tmp_"
)

var _ pluginv2handler.IPluginFSHandler = &PluginHandler{}

// Init implement pluginv2handler.IPluginFSHandler.
func (handler *PluginHandler) Init() error {
	if err := handler.mkdirAll(handler.setupDir); err != nil {
		return fmt.Errorf("failed to make setup-dir(%s): %w", handler.setupDir, err)
	}

	if err := handler.mkdirAll(handler.binDir); err != nil {
		return fmt.Errorf("failed to make bin-dir(%s): %w", handler.binDir, err)
	}

	if err := handler.mkdirAll(handler.etcDir); err != nil {
		return fmt.Errorf("failed to make etc-dir(%s): %w", handler.etcDir, err)
	}

	return nil
}

// CheckIntegrity implement pluginv2handler.IPluginFSHandler.
func (handler *PluginHandler) CheckIntegrity(_ context.Context) error {
	return nil
}

// Purge implement pluginv2handler.IPluginFSHandler.
// In compatible mode plugins/ is shared by all plugins, so only the current plugin's
// own files are removed: bin/<pluginName>, etc/<pluginName>, etc/<pluginName>.conf.
func (handler *PluginHandler) Purge(_ context.Context) error {
	pathsToRemove := []string{
		filepath.Join(handler.binDir, handler.pluginName),
		filepath.Join(handler.etcDir, handler.pluginName),
		filepath.Join(handler.etcDir, handler.pluginName+".conf"),
	}

	for _, p := range pathsToRemove {
		if err := handler.removeAll(p); err != nil {
			return err
		}
	}

	return nil
}

// Clean implement pluginv2handler.IPluginFSHandler.
func (handler *PluginHandler) Clean(_ context.Context) error {
	return handler.cleanAllTmpFiles()
}

// UnpackReleasePackage implement pluginv2handler.IPluginFSHandler.
func (handler *PluginHandler) UnpackReleasePackage(ctx context.Context, releasePkgAbsPath string, keepOldFileAsTmp bool) error {
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

// CopyConfigDir implement pluginv2handler.IPluginFSHandler.
func (handler *PluginHandler) CopyConfigDir(ctx context.Context, configFileAbsDir string) error {
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

		// copy file in config dir to etc-dir of plugin handler fs.
		targetRelativePath := filepath.Join(
			handler.etcDir,
			strings.TrimLeft(strings.TrimPrefix(path, configFileAbsDir), "/"),
		)

		return handler.overwriteFile(configFile, targetRelativePath)
	}); err != nil {
		return fmt.Errorf("failed to copy config dir: %w", err)
	}

	return nil
}

func (handler *PluginHandler) checkDirWritable(targetRelativeDir string) error {
	testRelativePath := filepath.Join(targetRelativeDir, handler.genTmpFileName("write_test"))

	// nolint: gosec
	testFile, err := handler.openFileForWrite(testRelativePath)
	if err != nil {
		return fmt.Errorf("target directory has no write permission. dir(%s): %v", testRelativePath, err)
	}
	_ = testFile.Close()

	_ = handler.removeFile(testRelativePath)

	return nil
}

func (handler *PluginHandler) genTmpFileName(tag string) string {
	return tmpFilePrefix + tag + "_" + strconv.FormatInt(time.Now().Unix(), 10)
}

func (handler *PluginHandler) checkPathInFS(relativePath string) error {
	absPath := handler.getAbsPath(relativePath)

	relPath, err := filepath.Rel(handler.rootAbsDir, absPath)
	if err != nil {
		return fmt.Errorf("failed to check if path in FS. path(%s): %w", absPath, err)
	}

	if strings.HasPrefix(relPath, ".."+string(filepath.Separator)) || relPath == ".." {
		return fmt.Errorf("path not in FS. root(%s), path(%s)", handler.rootAbsDir, absPath)
	}

	return nil
}

func (handler *PluginHandler) removeFile(relativePath string) error {
	if err := handler.checkPathInFS(relativePath); err != nil {
		return fmt.Errorf("will not remove path outside FS. file(%s): %w", relativePath, err)
	}

	absPath := handler.getAbsPath(relativePath)

	if err := os.Remove(absPath); err != nil {
		return fmt.Errorf("failed to remove file(%s): %w", absPath, err)
	}

	return nil
}
func (handler *PluginHandler) makeBinFilesExecutable() error {
	binFiles, err := utils.ListFiles(handler.getAbsPath(handler.binDir))
	if err != nil {
		return fmt.Errorf("failed to list bin files. dir(%s): %w", handler.binDir, err)
	}

	for _, binFile := range binFiles {
		if err := utils.MakeExecutable(binFile); err != nil {
			return fmt.Errorf("failed to make file executable. file(%s): %w", binFile, err)
		}
	}

	return nil
}

// extractTgz extract tgz file to target relative-dir.
// nolint: gocognit,funlen
func (handler *PluginHandler) extractTgz(ctx context.Context, sourceAbsPath string, targetRelativeDir string, keepOldFileAsTmp bool) (err error) {
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

func (handler *PluginHandler) cleanAllTmpFiles() error {
	if err := filepath.Walk(handler.rootAbsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), tmpFilePrefix) {
			tmpFileRelativePath := strings.TrimLeft(strings.TrimPrefix(path, handler.rootAbsDir), "/")
			_ = handler.removeFile(tmpFileRelativePath)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
