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
	gseCtlPath string
}

// StepArgs this define the args for step.
type StepArgs struct {
	GseCtlPath string
}

// NewStep ...
func NewStep(args StepArgs) *Step {
	step := &Step{
		gseCtlPath: args.GseCtlPath,
	}

	return step
}

// Run run step to start node.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(constant.StepStartNode, "start to start node")

	// start gse agent
	if err := StartNode(ctx, step.gseCtlPath); err != nil {
		logger.Error(constant.StepStartNode, fmt.Sprintf("start agent failed, err: %v", err))

		return fmt.Errorf("start agent failed, err: %v", err)
	}
	logger.Infof(constant.StepStartNode, "successfully start agent")

	return nil
}
