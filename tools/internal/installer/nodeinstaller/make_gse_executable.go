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

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

// MakeGseExecutable ...
func MakeGseExecutable(_ context.Context, agentPath, agentCtlPath string) error {
	if err := utils.MakeExecutable(agentPath); err != nil {
		logger.Errorf(constant.StepInstallAgent, constant.StateRunning,
			"make gse agent executable failed, agent-path(%s), err(%s)", agentPath, err)

		return err
	}
	logger.Infof(constant.StepInstallAgent, constant.StateRunning,
		"successfully make gse agent executable, agent-path(%s)", agentPath)

	if err := utils.MakeExecutable(agentCtlPath); err != nil {
		logger.Errorf(constant.StepInstallAgent, constant.StateRunning,
			"make gse agent ctl executable failed, agent-ctl-path(%s), err(%s)", agentCtlPath, err)

		return err
	}
	logger.Infof(constant.StepInstallAgent, constant.StateRunning,
		"successfully make gse agent executable, agent-ctl-path(%s)", agentCtlPath)

	return nil
}
