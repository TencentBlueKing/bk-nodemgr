/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodeuninstaller the package define the uninstall agent step.
package nodeuninstaller

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
)

// Step uninstall step.
type Step struct {
	args StepArgs
}

// StepArgs define args for step.
type StepArgs struct {
	AgentHandler agenthandler.IAgentHandler

	Backup bool
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step tp uninstall agent.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(installer.StepUninstallNode, "start to uninstall node")

	if step.args.Backup {
		if err := step.args.AgentHandler.FS().Backup(ctx); err != nil {
			logger.Warnf(installer.StepUninstallNode, "failed to backup: %v", err)
		} else {
			logger.Info(installer.StepUninstallNode, "backuped node")
		}
	}

	if err := step.uninstallAutoStartup(ctx); err != nil {
		return err
	}
	logger.Info(installer.StepUninstallNode, "uninstalled auto-startup")

	if err := step.args.AgentHandler.FS().Purge(ctx); err != nil {
		logger.Errorf(installer.StepUninstallNode, "failed to purge file-system: %v", err)

		return fmt.Errorf("failed to purge file-system: %w", err)
	}

	logger.Infof(installer.StepUninstallNode, "uninstalled node")

	return nil
}

func (step *Step) uninstallAutoStartup(ctx context.Context) error {
	hasStdout := false
	stdoutF := func(content string) {
		hasStdout = true
		logger.Infof(installer.StepUninstallNode, "node output: %s", content)
	}
	var stderrMsg string
	stderrF := func(content string) {
		stderrMsg += content + "\n"
	}

	if err := step.args.AgentHandler.Process().UninstallAutoStartup(ctx,
		&agenthandler.AsyncOutputOptions{
			Stdout: stdoutF,
			Stderr: stderrF,
		}); err != nil {
		logger.Errorf(installer.StepUninstallNode, "failed to uninstall auto startup. err-output(%s): %v", stderrMsg, err)

		return fmt.Errorf("failed to uninstall auto startup: %w", err)
	}

	if !hasStdout {
		logger.Warnf(installer.StepUninstallNode, "uninstalled auto startup without output")
	}

	return nil
}
