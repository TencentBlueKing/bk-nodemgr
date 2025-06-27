/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package checkdeploy ...
package checkdeploy

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
)

// Step this step is used to check this gse node is deploy or not.
type Step struct {
	runDir    string
	nodeRole  constant.NodeRole
	deployEnv string
}

// StepArgs ...
type StepArgs struct {
	RunDir    string
	NodeRole  constant.NodeRole
	DeployEnv string
}

// NewStep new step to check this gse node is deploy or not.
func NewStep(args StepArgs) *Step {
	step := &Step{
		runDir:    args.RunDir,
		nodeRole:  args.NodeRole,
		deployEnv: args.DeployEnv,
	}

	return step
}

// Run the step to check this gse node is deploy or not.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(constant.StepCheckDeploy, "start check deploy result. role(%s), run-dir(%s)",
		step.nodeRole, step.runDir)

	err := step.checkDeploy(ctx)
	if err != nil {
		logger.Errorf(constant.StepCheckDeploy, "check deploy result failed, err: %s", err.Error())
		return fmt.Errorf("check deploy result failed, err: %w", err)
	}

	logger.Infof(constant.StepCheckDeploy, "successfully check deploy result")

	return nil
}
