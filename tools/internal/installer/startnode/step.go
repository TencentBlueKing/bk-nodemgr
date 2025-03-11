/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package startnode ...
package startnode

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
)

// Step start node step.
type Step struct {
	agentPath    string
	agentCtlPath string
	configPath   string
}

// StepArgs this define the args for step.
type StepArgs struct {
	AgentPath    string
	AgentCtlPath string
	ConfigPath   string
}

// NewStep ...
func NewStep(args StepArgs) *Step {
	step := &Step{
		agentPath:    args.AgentPath,
		agentCtlPath: args.AgentCtlPath,
		configPath:   args.ConfigPath,
	}

	return step
}

// Run run step to start node.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(constant.StepStartNode, constant.StateStart, "start to start node")

	logger.Infof(constant.StepStartNode, constant.StateRunning,
		"agent path(%s), config path(%s)", step.agentPath, step.configPath)

	// check gse agent health
	healthState, err := CheckAgentHealth(ctx, step.agentPath, step.configPath)
	if err != nil {
		logger.Error(constant.StepStartNode, constant.StateFailed,
			fmt.Sprintf("check agent health failed, err: %v", err))

		return fmt.Errorf("check agent health failed, err: %v", err)
	}
	if !healthState.OK {
		logger.Error(constant.StepStartNode, constant.StateFailed,
			fmt.Sprintf("check agent health failed, health-state(%v)", healthState))

		return fmt.Errorf("check agent health failed, health-state(%v)", healthState)
	}
	logger.Infof(constant.StepStartNode, constant.StateRunning, "successfully check agent health")

	// start gse agent
	if err := StartAgent(ctx, step.agentCtlPath); err != nil {
		logger.Error(constant.StepStartNode, constant.StateFailed, fmt.Sprintf("start agent failed, err: %v", err))

		return fmt.Errorf("start agent failed, err: %v", err)
	}
	logger.Infof(constant.StepStartNode, constant.StateRunning, "successfully start agent")

	// TODO: start go proxy.

	logger.Infof(constant.StepStartNode, constant.StateDone, "successfully start node")

	return nil
}
