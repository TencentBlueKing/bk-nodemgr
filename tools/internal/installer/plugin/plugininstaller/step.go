/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package plugininstaller this package is used to install bk-nodemgr plugins.
package plugininstaller

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

// Step install plugin.
type Step struct {
	args StepArgs
}

// StepArgs args for step.
type StepArgs struct {
	PluginHandler pluginhandler.IPluginHandler

	PkgPath      string
	SrcConfigDir string
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("pkg-path(%s), src-config-dir(%s)", args.PkgPath, args.SrcConfigDir)
}

// StepResult result for step.
type StepResult struct {
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to install plugin.
func (step *Step) Run(ctx context.Context) (*StepResult, error) {
	logger.Infof(plugin.StepInstallPlugin, "start to install plugin. %s", step.args.String())

	// 1. init file-system architecture.
	if err := step.args.PluginHandler.FS().Init(); err != nil {
		logger.Errorf(plugin.StepInstallPlugin, "failed to init file-system: %v", err)

		return nil, err
	}
	logger.Info(plugin.StepInstallPlugin, "inited file-system")

	// 2. unpack release package file into installed file-system.
	if err := step.args.PluginHandler.FS().UnpackReleasePackage(ctx, step.args.PkgPath, false); err != nil {
		logger.Errorf(plugin.StepInstallPlugin, "failed to unpack release pkg: %v", err)

		return nil, err
	}

	logger.Info(plugin.StepInstallPlugin, "unpacked release pkg")

	// 3. copy config files to installed file-system.
	if err := step.args.PluginHandler.FS().CopyConfigDir(ctx, step.args.SrcConfigDir); err != nil {
		logger.Errorf(plugin.StepInstallPlugin, "failed to copy config dir: %v", err)

		return nil, err
	}
	logger.Info(plugin.StepInstallPlugin, "copied config dir")

	logger.Info(plugin.StepInstallPlugin, "installed node")

	return &StepResult{}, nil
}
