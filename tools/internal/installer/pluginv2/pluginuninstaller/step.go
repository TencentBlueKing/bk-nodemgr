/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package pluginuninstaller is used to uninstall bk-nodemgr plugins in a V2 way.
package pluginuninstaller

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginv2handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

// Step uninstall plugin in compatible mode.
type Step struct {
	args StepArgs
}

// StepArgs define args for step.
type StepArgs struct {
	PluginHandler pluginv2handler.IPluginHandler
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to uninstall plugin.
func (step *Step) Run(ctx context.Context) error {
	logger.Info(pluginv2.StepUninstallPlugin, "start to uninstall v2 plugin")

	if err := step.args.PluginHandler.FS().Purge(ctx); err != nil {
		logger.Errorf(pluginv2.StepUninstallPlugin, "failed to purge file-system: %v", err)

		return fmt.Errorf("failed to purge file-system: %w", err)
	}

	logger.Info(pluginv2.StepUninstallPlugin, "uninstalled v2 plugin")

	return nil
}
