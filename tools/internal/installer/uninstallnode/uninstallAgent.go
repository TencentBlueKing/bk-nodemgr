/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package uninstallnode the package define the uninstall agent step.
package uninstallnode

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
)

// Step uninstall agent step.
type Step struct {
	setupDirPath string
	gseCtlPath   string
}

// StepArgs define args for step.
type StepArgs struct {
	SetupDirPath string
	GseCtlPath   string
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	step := &Step{
		setupDirPath: args.SetupDirPath,
		gseCtlPath:   args.GseCtlPath,
	}

	// 1. stop agent
	// 2. remove setup dir

	return step
}

// Run run the step tp uninstall agent.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(constant.StepUninstallAgent, constant.StateStart, "start to uninstall agent")

	logger.Infof(constant.StepUninstallAgent, constant.StateRunning, "stop agent, gse-ctl(%s)", step.gseCtlPath)
	if err := StopAgent(ctx, step.gseCtlPath); err != nil {
		logger.Infof(constant.StepUninstallAgent, constant.StateFailed, "stop agent failed: %v", err)

		return err
	}
	logger.Infof(constant.StepUninstallAgent, constant.StateRunning, "successfully stop agent")

	logger.Infof(constant.StepUninstallAgent, constant.StateRunning, "remove setup dir(%s)", step.setupDirPath)
	if err := RemoveSetupDir(ctx, step.setupDirPath); err != nil {
		logger.Infof(constant.StepUninstallAgent, constant.StateFailed, "remove setup dir failed: %v", err)

		return err
	}

	logger.Infof(constant.StepUninstallAgent, constant.StateDone, "successfully uninstall agent")

	return nil
}
