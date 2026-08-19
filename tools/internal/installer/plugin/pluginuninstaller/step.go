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

package pluginuninstaller

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

// Step uninstall plugin.
type Step struct {
	args StepArgs
}

// StepArgs define args for step.
type StepArgs struct {
	PluginHandler pluginhandler.IPluginHandler
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to uninstall plugin.
func (step *Step) Run(ctx context.Context) error {
	logger.Info(plugin.StepUninstallPlugin, "start to uninstall plugin")

	if err := step.args.PluginHandler.FS().Purge(ctx); err != nil {
		logger.Errorf(plugin.StepUninstallPlugin, "failed to purge file-system: %v", err)

		return fmt.Errorf("failed to purge file-system: %w", err)
	}

	logger.Info(plugin.StepUninstallPlugin, "uninstalled plugin")

	return nil
}
