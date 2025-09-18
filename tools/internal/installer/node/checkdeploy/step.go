/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package checkdeploy provides step to check this gse node is deployed or not.
package checkdeploy

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

// Step this step is used to check this gse node is deploy or not.
type Step struct {
	args StepArgs
}

// StepArgs defines the args for step.
type StepArgs struct {
	AgentHandler agenthandler.IAgentHandler
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("node-role(%s)", args.AgentHandler.Role())
}

// NewStep new step to check this gse node is deploy or not.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run the step to check this gse node is deploy or not.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(node.StepCheckDeploy, "start to check deploy result. %s", step.args.String())

	err := step.checkDeploy(ctx)
	if err != nil {
		logger.Errorf(node.StepCheckDeploy, "failed to check deploy result. %s: %v", step.args.String(), err)
		return fmt.Errorf("failed to check deploy result. %s: %w", step.args.String(), err)
	}

	logger.Infof(node.StepCheckDeploy, "successfully checked deploy, it is running")

	return nil
}

func (step *Step) checkDeploy(ctx context.Context) error {
	processNode, err := step.args.AgentHandler.Process().GetProcess(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if node is running: %w", err)
	}

	if !processNode.IsAllRunning() {
		return fmt.Errorf("there are some node process not running: %v", processNode.Dead)
	}

	return nil
}
