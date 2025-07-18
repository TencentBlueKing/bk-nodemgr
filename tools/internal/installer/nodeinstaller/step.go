/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodeinstaller this package is used to install gse agent/proxy.
package nodeinstaller

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
)

// Step install agent.
type Step struct {
	args StepArgs
}

// StepArgs args for step.
type StepArgs struct {
	AgentHandler agenthandler.IAgentHandler

	AgentID         string
	ReRegisterAgent bool
	PkgPath         string
	SrcConfigDir    string
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("agent-id(%s), re-register-agent(%t), pkg-path(%s), src-config-dir(%s)",
		args.AgentID, args.ReRegisterAgent, args.PkgPath, args.SrcConfigDir)
}

// StepResult result for step.
type StepResult struct {
	AgentID string
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to install node.
func (step *Step) Run(ctx context.Context) (*StepResult, error) {
	logger.Infof(installer.StepInstallNode, "start to install node. %s", step.args.String())

	// 1. init file-system architecture.
	if err := step.args.AgentHandler.FS().Init(ctx); err != nil {
		logger.Errorf(installer.StepInstallNode, "failed to init file-system: %v", err)

		return nil, err
	}
	logger.Info(installer.StepInstallNode, "inited file-system")

	// 2. unpack release package files into installed file-system.
	if err := step.args.AgentHandler.FS().UnpackReleasePackage(ctx, step.args.PkgPath, false); err != nil {
		logger.Errorf(installer.StepInstallNode, "failed to unpack release pkg: %v", err)

		return nil, err
	}
	logger.Info(installer.StepInstallNode, "unpacked release pkg")

	// 3. copy config files to installed file-system.
	if err := step.args.AgentHandler.FS().CopyConfigDir(ctx, step.args.SrcConfigDir); err != nil {
		logger.Errorf(installer.StepInstallNode, "failed to copy config dir: %v", err)

		return nil, err
	}
	logger.Info(installer.StepInstallNode, "copied config dir")

	if step.args.ReRegisterAgent {
		// 3.1. unregister agent if necessary.
		if err := step.args.AgentHandler.Process().UnregisterAgentID(ctx); err != nil {
			logger.Errorf(installer.StepInstallNode, "failed to unregister agent: %v", err)

			return nil, err
		}
		logger.Info(installer.StepInstallNode, "unregistered agent")
	}

	// 4. register agent.
	agentID, err := step.args.AgentHandler.Process().RegisterAgentID(ctx, step.args.AgentID)
	if err != nil {
		logger.Error(installer.StepInstallNode, fmt.Sprintf("failed to register agent: %v", err))

		return nil, err
	}
	logger.Info(installer.StepInstallNode, "registered agent")

	if err := step.installAutoStartup(ctx); err != nil {
		return nil, err
	}
	logger.Info(installer.StepInstallNode, "installed auto-startup")

	logger.Info(installer.StepInstallNode, "installed node")

	return &StepResult{AgentID: agentID}, nil
}

func (step *Step) installAutoStartup(ctx context.Context) error {
	hasStdout := false
	stdoutF := func(content string) {
		hasStdout = true
		logger.Infof(installer.StepInstallNode, "node output: %s", content)
	}
	var stderrMsg string
	stderrF := func(content string) {
		stderrMsg += content + "\n"
	}

	if err := step.args.AgentHandler.Process().InstallAutoStartup(ctx,
		&agenthandler.AsyncOutputOptions{
			Stdout: stdoutF,
			Stderr: stderrF,
		}); err != nil {
		logger.Errorf(installer.StepInstallNode, "failed to install auto startup. err-output(%s): %v", stderrMsg, err)

		return fmt.Errorf("failed to install auto startup: %w", err)
	}

	if !hasStdout {
		logger.Warnf(installer.StepInstallNode, "installed auto startup without output")
	}

	return nil
}
