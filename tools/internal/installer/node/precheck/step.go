/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package precheck provides precheck step.
package precheck

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/retrier"
)

// Step precheck step.
type Step struct {
	args StepArgs
}

// StepArgs define args for step.
type StepArgs struct {
	AgentHandler agenthandler.IAgentHandler

	CheckList CheckList
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("check-list(%+v)", args.CheckList)
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to precheck.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(node.StepPreCheck, "start to precheck with config: %s", step.args.String())

	r := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	gp := gopool.NewPool()

	runCheck := func(name string, checkFn func() error) func() error {
		return func() error {
			logger.Infof(node.StepPreCheck, "start check %s", name)

			if err := r.Do(ctx, func(_ int) error {
				return checkFn()
			}); err != nil {
				logger.Errorf(node.StepPreCheck,
					"failed to check %s: %v", name, err)

				return err
			}

			logger.Infof(node.StepPreCheck, "done check %s", name)

			return nil
		}
	}

	gp.Go(runCheck("disk free space", func() error {
		return CheckDiskFreeSpace(step.args.CheckList.DiskRequires)
	}))
	gp.Go(runCheck("port policy", func() error {
		return CheckPortPolicies(ctx, step.args.CheckList.PortPolicies)
	}))
	gp.Go(runCheck("network policy", func() error {
		return CheckNetworkPolicies(ctx, step.args.CheckList.NetworkPolicies)
	}))
	gp.Go(runCheck("gse process", func() error {
		nodeProcess, err := step.args.AgentHandler.Process().GetProcess(ctx)
		if err != nil {
			return err
		}

		if !nodeProcess.IsAllDead() {
			return fmt.Errorf("there are some node process running: %v", nodeProcess.Running)
		}

		return nil
	}))

	if err := gp.Wait(); err != nil {
		logger.Errorf(node.StepPreCheck, "failed to do all precheck: %v", err)
		return err
	}

	logger.Infof(node.StepPreCheck, "done all precheck")

	return nil
}

// CheckList this is the list of precheck.
type CheckList struct {
	DiskRequires    []DiskRequire   `json:"disk_requires"`
	PortPolicies    []PortPolicy    `json:"port_policies"`
	NetworkPolicies []NetworkPolicy `json:"network_policies"`
}

// Validate validate precheck list.
func (list *CheckList) Validate() error {
	for _, require := range list.DiskRequires {
		if err := require.Validate(); err != nil {
			return err
		}
	}

	for _, policy := range list.PortPolicies {
		if err := policy.Validate(); err != nil {
			return err
		}
	}

	for _, policy := range list.NetworkPolicies {
		if err := policy.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// DefaultCheckList default check list.
func DefaultCheckList() *CheckList {
	return &CheckList{
		DiskRequires:    []DiskRequire{},
		PortPolicies:    []PortPolicy{},
		NetworkPolicies: []NetworkPolicy{},
	}
}

// LoadCheckList load check list from file.
func LoadCheckList(preCheckListPath string) (*CheckList, error) {
	if preCheckListPath == "" {
		return DefaultCheckList(), nil
	}

	// nolint: gosec
	bytes, err := os.ReadFile(preCheckListPath)
	if err != nil {
		return nil, err
	}

	list := new(CheckList)
	if err := json.Unmarshal(bytes, list); err != nil {
		return nil, err
	}

	if list.Validate() != nil {
		return nil, err
	}

	return list, nil
}
