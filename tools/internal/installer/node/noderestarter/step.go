/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package noderestarter provides step to restart node.
package noderestarter

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

// Step restart node step.
type Step struct {
	args StepArgs
}

// StepArgs this define the args for step.
type StepArgs struct {
	AgentHandler agenthandler.IAgentHandler

	Force bool
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run step to restart node.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(node.StepRestartNode, "start to restart node")
	diagnostic, err := step.args.AgentHandler.Process().DiagnoseVersion(ctx)
	if err != nil {
		if diagnostic != nil {
			logger.Errorf(node.StepRestartNode, "failed to diagnose agent version: work_dir(%s) executable(%s) args(%v) stdout_raw(%s) stderr_raw(%s): %v", diagnostic.WorkDir, diagnostic.Executable, diagnostic.Args, diagnostic.Stdout, diagnostic.Stderr, err)
		} else {
			logger.Errorf(node.StepRestartNode, "failed to diagnose agent version: %v", err)
		}

		return fmt.Errorf("failed to diagnose agent version: %w", err)
	}
	logger.Infof(node.StepRestartNode, "agent version diagnostic: work_dir(%s) executable(%s) args(%v) stdout_raw(%s) stderr_raw(%s)", diagnostic.WorkDir, diagnostic.Executable, diagnostic.Args, diagnostic.Stdout, diagnostic.Stderr)

	hasStdout := false
	stdoutF := func(content string) {
		hasStdout = true
		logger.Infof(node.StepRestartNode, "node output: %s", content)
	}
	var stderrMsg string
	stderrF := func(content string) {
		stderrMsg += content + "\n"
	}

	if err := step.args.AgentHandler.Process().Restart(ctx,
		step.args.Force,
		&agenthandler.AsyncOutputOptions{
			Stdout: stdoutF,
			Stderr: stderrF,
		}); err != nil {
		logger.Errorf(node.StepRestartNode, "failed to restart node. err-output(%s): %v", stderrMsg, err)

		return fmt.Errorf("failed to restart node: %w", err)
	}

	if !hasStdout {
		logger.Warnf(node.StepRestartNode, "process restarted without output")
	}

	logger.Infof(node.StepRestartNode, "restarted node")

	return nil
}
