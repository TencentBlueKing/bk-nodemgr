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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

// TryCreateInstallDir try to create install dir.
func TryCreateInstallDir(_ context.Context, setupPath string, overwrite bool) error {
	logger.Infof(constant.StepInstallNode, constant.StateRunning,
		"try to create install dir, path(%s), overwrite(%t)", setupPath, overwrite)

	// 1. check setup path is valid.
	if setupPath = strings.TrimSpace(setupPath); setupPath == "" {
		return errors.New("setup path cannot be empty")
	}

	// 2. check if setup path exists.
	exists, err := utils.PathExists(setupPath)
	if err != nil {
		return fmt.Errorf("failed to check path existence: %w", err)
	}

	if exists {
		if !overwrite {
			if err := backupExistedSetupDir(setupPath); err != nil {
				return err
			}
		}

		logger.Warnf(constant.StepInstallNode, constant.StateRunning,
			"start to clean directory, dir-path(%s)", setupPath)
		paths, err := utils.CleanDirectory(setupPath)
		if err != nil {
			return fmt.Errorf("failed to clean entry: %w", err)
		}

		for _, path := range paths {
			logger.Warnf(constant.StepInstallNode, constant.StateRunning, "clean exist directory: %s", path)
		}
	}

	// 4. create directories.
	if err := createDirectories(setupPath); err != nil {
		return fmt.Errorf("create directories failed: %w", err)
	}

	return nil
}

func backupExistedSetupDir(setupPath string) error {
	backDir, errs := utils.BackupExistingDirIgnoreErr(setupPath)
	if len(errs) > 0 {
		for _, err := range errs {
			logger.Warnf(constant.StepInstallNode, constant.StateRunning,
				"ignore err, err: %s", err)
		}
	}

	if backDir == "" {
		return errors.New("failed to backup exist setup files")
	}

	logger.Infof(constant.StepInstallNode, constant.StateRunning,
		"backup exist setup files to %s", backDir)

	return nil
}

// createDirectories create directories.
func createDirectories(setupPath string) error {
	// create main directory.
	if err := os.MkdirAll(setupPath, 0700); err != nil { // nolint: mnd
		return fmt.Errorf("failed to create main directory: %w", err)
	}

	// create etc directory.
	etcPath := filepath.Join(setupPath, "etc")
	if err := os.MkdirAll(etcPath, 0700); err != nil { // nolint: mnd
		return fmt.Errorf("failed to create etc directory: %w", err)
	}

	return nil
}
