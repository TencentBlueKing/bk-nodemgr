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

// Package cleantmp provides step to clean tmp files in fs.
package cleantmp

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

// Step this step is used to clean tmp files in fs.
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

// NewStep new step to clean tmp files in fs.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run the step to clean tmp files in fs.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(node.StepCleanTmp, "start to clean tmp files. %s", step.args.String())

	err := step.cleanTmp(ctx)
	if err != nil {
		logger.Errorf(node.StepCleanTmp, "failed to clean tmp files. %s: %v", step.args.String(), err)
		return fmt.Errorf("failed to clean tmp files. %s: %w", step.args.String(), err)
	}

	logger.Infof(node.StepCleanTmp, "successfully cleaned tmp files")

	return nil
}

func (step *Step) cleanTmp(ctx context.Context) error {
	if err := step.args.AgentHandler.FS().Clean(ctx); err != nil {
		return fmt.Errorf("failed to clean tmp files: %w", err)
	}

	return nil
}
