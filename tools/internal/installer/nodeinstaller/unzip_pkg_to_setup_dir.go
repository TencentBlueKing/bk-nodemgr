/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstaller

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
)

const maxDecompressedSize = 100 * 1024 * 1024 // 100M

// UnzipPkgToSetupDir unzip gse pkg to setup dir.
// nolint: funlen,gocognit,gocyclo,cyclop
func UnzipPkgToSetupDir(ctx context.Context, pkgPath, setupDir string) error {
	logger.Infof(constant.StepInstallAgent, constant.StateRunning,
		"unzip pkg to setup dir, unzip-pkg-path(%s), setup-dir(%s)", pkgPath, setupDir)

	// check pkg exist or not.
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		logger.Errorf(constant.StepInstallAgent, constant.StateFailed,
			"unzip pkg to setup dir failed, pkg-path(%s) does not exist", pkgPath)

		return fmt.Errorf("pkg does not exist, pkg-path(%s), err: %v", pkgPath, err)
	}

	// check pkg format is tar.gz or not.
	if !strings.HasSuffix(strings.ToLower(pkgPath), ".tar.gz") {
		logger.Errorf(constant.StepInstallAgent, constant.StateFailed,
			"unzip pkg to setup dir failed, pkg-path(%s) format is not tar.gz", pkgPath)

		return fmt.Errorf("pkg format is not tar.gz, pkg-path(%s)", pkgPath)
	}

	// check setupDir exist or not.
	if _, err := os.Stat(setupDir); os.IsNotExist(err) {
		if err := os.MkdirAll(setupDir, 0700); err != nil { // nolint: mnd
			logger.Errorf(constant.StepInstallAgent, constant.StateFailed,
				"create setup directory failed, setup-dir(%s), err: %v", setupDir, err)

			return fmt.Errorf("create setup directory failed, setup-dir(%s), err: %v", setupDir, err)
		}

		logger.Infof(constant.StepInstallAgent, constant.StateRunning,
			"create setup directory, setup-dir(%s)", setupDir)
	}

	// make sure setupDir has write permission.
	testPath := filepath.Join(setupDir, fmt.Sprintf("write_test_%d", time.Now().UnixNano()))

	logger.Infof(constant.StepInstallAgent, constant.StateRunning,
		"try creating a file to make sure the setup directory has write permissions, "+
			"setup-dir(%s), test-path(%s)", setupDir, testPath)

	testFile, err := os.OpenFile(testPath, os.O_CREATE|os.O_WRONLY, 0600) // nolint: gosec,mnd
	if err != nil {
		logger.Errorf(constant.StepInstallAgent, constant.StateFailed,
			"setup directory has no write permission, setup-dir(%s), err: %v", setupDir, err)

		return fmt.Errorf("setup directory has no write permission, setup-dir(%s), err: %v", setupDir, err)
	}

	_ = testFile.Close()
	_ = os.Remove(testPath)

	logger.Infof(constant.StepInstallAgent, constant.StateRunning,
		"remove test file, setup-dir(%s), test-path(%s)", setupDir, testPath)

	// open pkg file.
	file, err := os.Open(pkgPath) // nolint: gosec
	if err != nil {
		logger.Errorf(constant.StepInstallAgent, constant.StateFailed,
			"cannot open pkg file, pkg-path(%s), err: %v", pkgPath, err)

		return fmt.Errorf("cannot open pkg file, pkg-path(%s), err: %v", pkgPath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	// create gzip reader.
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("cannot create gzip reader, pkg-path(%s), err: %v", pkgPath, err)
	}
	defer func() {
		_ = gzipReader.Close()
	}()

	// create tar reader.
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
			return fmt.Errorf("read tar content error, pkg-path(%s), err: %v", pkgPath, err)
		}

		// get target path.
		target := filepath.Join(setupDir, filepath.Clean(filepath.FromSlash(header.Name)))

		// make sure target path is within setupDir.
		relPath, err := filepath.Rel(setupDir, target)
		if err != nil || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) || relPath == ".." {
			return fmt.Errorf("invalid file path in archive, attempted path(%s)", target)
		}

		// handle each file by its type.
		switch header.Typeflag {
		case tar.TypeDir:
			// mkdir all parent directories.
			if err := os.MkdirAll(target, 0700); err != nil { // nolint: mnd,gosec
				return fmt.Errorf("create directory failed, path(%s), err: %v", target, err)
			}
		case tar.TypeReg:
			// make sure parent directories exist.
			if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil { // nolint: mnd,gosec
				return fmt.Errorf("create parent directory failed, path(%s), err: %v", filepath.Dir(target), err)
			}

			// create file.
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600) // nolint: mnd,gosec
			if err != nil {
				return fmt.Errorf("create file failed, path(%s), err: %v", target, err)
			}

			// write file content.
			limitedReader := &io.LimitedReader{
				R: tarReader,
				N: maxDecompressedSize,
			}
			if _, err := io.Copy(outFile, limitedReader); err != nil {
				_ = outFile.Close()

				return fmt.Errorf("write file content failed, path(%s), err: %v", target, err)
			}
			_ = outFile.Close()
		default:
			// log unsupported file type.
			logger.Infof(constant.StepInstallAgent, constant.StateRunning,
				"unzip file type is not supported, type(%d), path(%s)", header.Typeflag, target)
		}
	}

	logger.Info(constant.StepInstallAgent, constant.StateRunning, "successfully unzip gse pkg to setup dir")

	return nil
}
