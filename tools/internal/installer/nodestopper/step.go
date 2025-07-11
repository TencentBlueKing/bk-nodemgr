/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodestopper provides step for stop node.
package nodestopper

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
)

// Step stop node step.
type Step struct {
	args StepArgs
}

// StepArgs this define the args for step.
type StepArgs struct {
	AgentHandler agenthandler.IAgentHandler

	Force bool
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("force(%t)", args.Force)
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run step to stop node.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(installer.StepStopNode, "start to stop node. %s", step.args.String())

	hasStdout := false
	stdoutF := func(content string) {
		hasStdout = true
		logger.Infof(installer.StepStopNode, "agent output: %s", content)
	}
	var stderrMsg string
	stderrF := func(content string) {
		stderrMsg += content + "\n"
	}

	if err := step.args.AgentHandler.Process().Stop(ctx,
		step.args.Force,
		&agenthandler.AsyncOutputOptions{
			Stdout: stdoutF,
			Stderr: stderrF,
		}); err != nil {

		if !step.args.Force {
			logger.Errorf(installer.StepStopNode, "failed to stop node by pid-file. err-output(%s): %v", stderrMsg, err)
			return fmt.Errorf("failed to stop node: %v", err)
		}

		logger.Warnf(installer.StepStopNode, "failed to stop node by pid-file. err-output(%s): %v", stderrMsg, err)

		// force kill.
		if err := step.args.AgentHandler.Process().ForceKill(ctx); err != nil {
			logger.Errorf(installer.StepStopNode, "failed to force kill node: %v", err)
			return fmt.Errorf("failed to stop node: %v", err)
		}

		logger.Info(installer.StepStopNode, "force killed node")
	}

	if !hasStdout {
		logger.Warnf(installer.StepStopNode, "process stopped without output")
	}

	logger.Infof(installer.StepStopNode, "stopped node. %s", step.args.String())

	return nil
}
