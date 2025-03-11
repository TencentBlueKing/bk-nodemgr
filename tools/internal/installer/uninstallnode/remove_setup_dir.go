/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package uninstallnode ...
package uninstallnode

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
)

// RemoveSetupDir ...
func RemoveSetupDir(_ context.Context, setupDirPath string) error {
	logger.Warnf(constant.StepUninstallAgent, constant.StateRunning,
		"start to clean directory, dir-path(%s)", setupDirPath)

	var err error
	setupDirPath, err = filepath.Abs(setupDirPath)
	if err != nil {
		return fmt.Errorf("get abs path failed, err: %v", err)
	}

	setupDirPath = filepath.Clean(setupDirPath)
	if setupDirPath == "" {
		return errors.New("setup dir path is empty")
	}

	if _, err := os.Stat(setupDirPath); os.IsNotExist(err) {
		logger.Warnf(constant.StepUninstallAgent, constant.StateDone, "directory not exist, dir-path(%s)", setupDirPath)

		return nil
	}

	if err := os.RemoveAll(setupDirPath); err != nil {
		return fmt.Errorf("remove setup dir failed, err: %v", err)
	}

	logger.Warnf(constant.StepUninstallAgent, constant.StateRunning, "clean exist directory: %s", setupDirPath)

	return nil
}
