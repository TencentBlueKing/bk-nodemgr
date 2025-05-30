//go:build linux || darwin || freebsd || aix

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package uninstallnode

import (
	"context"
	"fmt"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"path/filepath"
	"strings"
	"syscall"
)

// forceUninstall uninstall agent by force way.
func (step *Step) forceUninstall(_ context.Context) error {
	logger.Infof(constant.StepUninstallNode, constant.StateRunning, "force uninstalling ...")

	process, err := utils.GetSameSpaceProcesses()
	if err != nil {
		return fmt.Errorf("get same space processes failed, err: %v", err)
	}

	setupDirPath := filepath.Clean(step.setupDirPath)

	for _, p := range process {
		if !isGseBin(p.Name) {
			continue
		}

		if err := utils.CheckDirPathSafe(setupDirPath); err != nil {
			continue
		}

		if !strings.HasPrefix(p.FullPath, setupDirPath) {
			continue
		}

		logger.Infof(constant.StepUninstallNode, constant.StateRunning,
			"remnant process found, process number: %d, file path: %s", p.PID, p.FullPath)

		exist, err := utils.CheckPIDExist(p.PID)
		if err != nil {
			return fmt.Errorf("force kill failed, pid: %d, err: %v", p.PID, err)
		}

		if !exist {
			logger.Infof(constant.StepUninstallNode, constant.StateRunning,
				"process already exit, process number: %d, file path: %s", p.PID, p.FullPath)

			continue
		}

		// check if the pid again, make sure the pid is not a system pid.
		if p.PID <= 1 {
			return fmt.Errorf("process pid is too dangerous, pid: %d", p.PID)
		}

		err = syscall.Kill(p.PID, syscall.SIGKILL)
		if err != nil {
			logger.Errorf(constant.StepUninstallNode, constant.StateFailed,
				"force kill process failed, err: %v", err)

			return fmt.Errorf("force kill failed, pid: %d, err: %v", p.PID, err)
		}

		logger.Infof(constant.StepUninstallNode, constant.StateRunning,
			"force kill process success, process number: %d, file path: %s", p.PID, p.FullPath)
	}

	return nil
}

// gseBinList gse bin list.
func gseBinList() []string {
	return []string{
		"gse_agent",
		"gse_data",
		"gse_file",
	}
}

// isGseBin check if the name is a gse bin.
// notice: in order to kill the process which is not belong gse bin, we need to check the process name.
func isGseBin(name string) bool {
	for _, bin := range gseBinList() {
		if name == bin {
			return true
		}
	}

	return false
}
