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
	isProxy             bool
	agentPath           string
	agentCtlPath        string
	agentConfigPath     string
	dataProxyPath       string
	dataProxyConfigPath string
	fileProxyPath       string
	fileProxyConfigPath string
}

// StepArgs this define the args for step.
type StepArgs struct {
	IsProxy             bool
	AgentPath           string
	AgentCtlPath        string
	AgentConfigPath     string
	DataProxyPath       string
	DataProxyConfigPath string
	FileProxyPath       string
	FileProxyConfigPath string
}

// NewStep ...
func NewStep(args StepArgs) *Step {
	step := &Step{
		isProxy:             args.IsProxy,
		agentPath:           args.AgentPath,
		agentCtlPath:        args.AgentCtlPath,
		agentConfigPath:     args.AgentConfigPath,
		dataProxyPath:       args.DataProxyPath,
		dataProxyConfigPath: args.DataProxyConfigPath,
		fileProxyPath:       args.FileProxyPath,
		fileProxyConfigPath: args.FileProxyConfigPath,
	}

	return step
}

// Run run step to start node.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(constant.StepStartNode, constant.StateStart, "start to start node")

	logger.Infof(constant.StepStartNode, constant.StateRunning,
		"agent path(%s), config path(%s)", step.agentPath, step.agentConfigPath)

	//if step.isProxy {
	//	logger.Infof(constant.StepStartNode, constant.StateRunning,
	//		"data proxy path(%s), config path(%s)", step.dataProxyPath, step.dataProxyConfigPath)
	//	logger.Infof(constant.StepStartNode, constant.StateRunning,
	//		"file proxy path(%s), config path(%s)", step.fileProxyPath, step.fileProxyConfigPath)
	//
	//	err := StartProxy(ctx, step.dataProxyPath, step.dataProxyConfigPath, step.fileProxyPath, step.fileProxyConfigPath)
	//	if err != nil {
	//		logger.Error(constant.StepStartNode, constant.StateFailed,
	//			fmt.Sprintf("start proxy failed, err: %v", err))
	//	}
	//
	//	logger.Infof(constant.StepStartNode, constant.StateRunning, "successfully start data and file proxy")
	//}

	// check gse agent health
	healthState, err := CheckAgentHealth(ctx, step.agentPath, step.agentConfigPath)
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
	if err := StartNode(ctx, step.agentCtlPath); err != nil {
		logger.Error(constant.StepStartNode, constant.StateFailed, fmt.Sprintf("start agent failed, err: %v", err))

		return fmt.Errorf("start agent failed, err: %v", err)
	}
	logger.Infof(constant.StepStartNode, constant.StateRunning, "successfully start agent")

	logger.Infof(constant.StepStartNode, constant.StateDone, "successfully start node")

	return nil
}
