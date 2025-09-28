/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodeupgrader this package is used to upgrade gse agent/proxy.
package nodeupgrader

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

// Step upgrade node.
type Step struct {
	args StepArgs
}

// StepArgs args for step.
type StepArgs struct {
	AgentHandler agenthandler.IAgentHandler

	PkgPath      string
	SrcConfigDir string
	Backup       bool

	// SelectUpgrades set false by default, will upgrade all things.
	// set true, then will only upgrade the enabled ones following.
	SelectUpgrades              bool
	EnableUpgradeReleasePackage bool
	EnableUpgradeConfig         bool
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("pkg-path(%s), src-config-dir(%s)", args.PkgPath, args.SrcConfigDir)
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to upgrade node.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(node.StepUpgradeNode, "start to upgrade node. %s", step.args.String())

	// 1. init file-system architecture.
	if err := step.args.AgentHandler.FS().Init(); err != nil {
		logger.Errorf(node.StepUpgradeNode, "failed to init file-system: %v", err)

		return err
	}
	logger.Info(node.StepUpgradeNode, "inited file-system")

	if step.args.Backup {
		// 1.1. backup old files.
		if err := step.args.AgentHandler.FS().Backup(ctx); err != nil {
			logger.Warnf(node.StepUpgradeNode, "failed to backup: %v", err)
		} else {
			logger.Info(node.StepUpgradeNode, "backuped node")
		}
	}

	if !step.args.SelectUpgrades || step.args.EnableUpgradeReleasePackage {
		// 2. unpack release package files into installed file-system.
		if err := step.args.AgentHandler.FS().UnpackReleasePackage(ctx, step.args.PkgPath, true); err != nil {
			logger.Errorf(node.StepUpgradeNode, "failed to unpack release pkg: %v", err)

			return err
		}
		logger.Info(node.StepUpgradeNode, "unpacked release pkg")
	}

	if !step.args.SelectUpgrades || step.args.EnableUpgradeConfig {
		// 3. copy config files to installed file-system.
		if err := step.args.AgentHandler.FS().CopyConfigDir(ctx, step.args.SrcConfigDir); err != nil {
			logger.Errorf(node.StepUpgradeNode, "failed to copy config dir: %v", err)

			return err
		}
		logger.Info(node.StepUpgradeNode, "copied config dir")
	}

	logger.Info(node.StepUpgradeNode, "upgraded node, wait for restarting")

	return nil
}
