/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package plugininstaller this package is used to install bk-nodemgr plugins in a V2 way.
package plugininstaller

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginv2handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

// Step install plugin.
type Step struct {
	args StepArgs
}

// StepArgs args for step.
type StepArgs struct {
	PluginHandler pluginv2handler.IPluginHandler

	PkgPath      string
	SrcConfigDir string
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("pkg-path(%s), src-config-dir(%s)", args.PkgPath, args.SrcConfigDir)
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to install plugin.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(pluginv2.StepInstallPlugin, "start to install plugin. %s", step.args.String())

	// 1. init file-system architecture.
	if err := step.args.PluginHandler.FS().Init(); err != nil {
		logger.Errorf(pluginv2.StepInstallPlugin, "failed to init file-system: %v", err)

		return err
	}
	logger.Info(pluginv2.StepInstallPlugin, "inited file-system")

	// 2. unpack release package file into installed file-system.
	if err := step.args.PluginHandler.FS().UnpackReleasePackage(ctx, step.args.PkgPath, false); err != nil {
		logger.Errorf(pluginv2.StepInstallPlugin, "failed to unpack release pkg: %v", err)

		return err
	}

	logger.Info(pluginv2.StepInstallPlugin, "unpacked release pkg")

	// 3. copy config files to installed file-system.
	if err := step.args.PluginHandler.FS().CopyConfigDir(ctx, step.args.SrcConfigDir); err != nil {
		logger.Errorf(pluginv2.StepInstallPlugin, "failed to copy config dir: %v", err)

		return err
	}
	logger.Info(pluginv2.StepInstallPlugin, "copied config dir")

	logger.Info(pluginv2.StepInstallPlugin, "installed plugin")

	return nil
}
