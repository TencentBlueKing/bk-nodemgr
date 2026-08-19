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

// Package pluginrunner runs an arbitrary command inside the plugin deploy directory.
package pluginrunner

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
)

// Step runs a user-supplied command in the plugin deploy directory.
type Step struct {
	args StepArgs
}

// StepArgs args for step.
type StepArgs struct {
	// PluginHandler provides the plugin process operations.
	PluginHandler pluginhandler.IPluginHandler

	// DebugAction is the debug action type, e.g. start or stop.
	DebugAction types.PluginDebugAction

	// RunCmd is the shell command string to execute.
	RunCmd string

	// PidDir is the pid directory used to get plugin pid files.
	PidDir string
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("debug-action(%s), run-cmd(%s), pid-dir(%s)", args.DebugAction, args.RunCmd, args.PidDir)
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run executes the command and blocks until it exits.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(plugin.StepDebugPlugin, "start to run debug plugin. %s", step.args.String())
	if step.args.PluginHandler == nil {
		return fmt.Errorf("plugin handler is nil")
	}

	switch step.args.DebugAction {
	case types.PluginDebugActionStart:
		if err := step.args.PluginHandler.Process().RunCmd(ctx, step.args.RunCmd, step.args.PidDir); err != nil {
			logger.Errorf(plugin.StepDebugPlugin, "failed to start debug plugin command: %v", err)
			return fmt.Errorf("failed to start debug plugin command: %w", err)
		}
	case types.PluginDebugActionStop:
		if err := step.args.PluginHandler.Process().Stop(ctx, step.args.PidDir); err != nil {
			logger.Errorf(plugin.StepDebugPlugin, "failed to stop debug plugin command: %v", err)
			return fmt.Errorf("failed to stop debug plugin command: %w", err)
		}
	default:
		return fmt.Errorf("unsupported debug action: %s", step.args.DebugAction)
	}

	logger.Info(plugin.StepDebugPlugin, "debug plugin action finished successfully")
	return nil
}
