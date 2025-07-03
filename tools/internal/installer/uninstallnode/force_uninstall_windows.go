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

package uninstallnode

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/winapi"
)

const (
	gseAgentDaemonFileName = "gse_agent_daemon.exe"
	gseAgentFileName       = "gse_agent.exe"
)

// forceUninstall uninstall agent by force way.
func (step *Step) forceUninstall(_ context.Context) error {
	gseAgentDaemonFilePath := filepath.Join(step.binDirPath, gseAgentDaemonFileName)

	logger.Infof(constant.StepUninstallNode, "try to uninstall windows svc, filePath(%s)", gseAgentDaemonFilePath)

	gseAgentFilePath := filepath.Join(step.binDirPath, gseAgentFileName)

	err := winapi.KillProcessByNameAndPath(gseAgentDaemonFileName, gseAgentDaemonFilePath)
	if err != nil {
		return fmt.Errorf("kill gse_agent_daemon failed, err: %v", err)
	}

	err = winapi.KillProcessByNameAndPath(gseAgentFileName, gseAgentFilePath)
	if err != nil {
		return fmt.Errorf("kill gse_agent failed, err: %v", err)
	}

	logger.Infof(constant.StepUninstallNode, "successfully stop gse agent and gse agent daemon")

	return nil
}
