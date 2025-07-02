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

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
)

// softUninstall uninstall agent by soft way.
func (step *Step) softUninstall(ctx context.Context) error {
	logger.Infof(constant.StepUninstallNode, "stop agent, gse-ctl(%s)", step.gseCtlPath)
	if err := StopNode(ctx, step.gseCtlPath); err != nil {
		logger.Infof(constant.StepUninstallNode, "stop agent failed: %v", err)

		return err
	}
	logger.Infof(constant.StepUninstallNode, "successfully stop agent")

	logger.Infof(constant.StepUninstallNode, "remove setup dir(%s)", step.setupDirPath)
	if err := RemoveSetupDir(ctx, step.setupDirPath); err != nil {

		logger.Infof(constant.StepUninstallNode, "remove setup dir failed: %v", err)

		return err
	}

	return nil
}
