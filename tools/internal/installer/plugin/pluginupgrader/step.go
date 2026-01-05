/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package pluginupgrader provides the step to upgrade plugin.
package pluginupgrader

import (
	"context"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

// Step upgrade plugin.
type Step struct {
	args StepArgs
}

// StepArgs define args for step.
type StepArgs struct {
	PluginHandler pluginhandler.IPluginHandler

	PkgPath      string
	SrcConfigDir string
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to upgrade plugin.
func (step *Step) Run(ctx context.Context) error {
	// 1. purge existing file-system.
	if err := step.args.PluginHandler.FS().Purge(ctx); err != nil {
		logger.Errorf(plugin.StepUpgradePlugin, "failed to purge file-system: %v", err)

		return err
	}
	logger.Info(plugin.StepUpgradePlugin, "purged existing file-system")

	// 2. init file-system architecture.
	if err := step.args.PluginHandler.FS().Init(); err != nil {
		logger.Errorf(plugin.StepUpgradePlugin, "failed to init file-system: %v", err)

		return err
	}
	logger.Info(plugin.StepUpgradePlugin, "inited file-system")

	// 3. unpack release package file into inited file-system.
	if err := step.args.PluginHandler.FS().UnpackReleasePackage(ctx, step.args.PkgPath, false); err != nil {
		logger.Errorf(plugin.StepUpgradePlugin, "failed to unpack release pkg: %v", err)

		return err
	}

	logger.Info(plugin.StepUpgradePlugin, "unpacked release pkg")

	// 4. copy config files to inited file-system.
	if err := step.args.PluginHandler.FS().CopyConfigDir(ctx, step.args.SrcConfigDir); err != nil {
		logger.Errorf(plugin.StepUpgradePlugin, "failed to copy config dir: %v", err)

		return err
	}

	logger.Info(plugin.StepUpgradePlugin, "upgrade plugin")

	return nil
}
