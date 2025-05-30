/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package precheck ...
package precheck

import (
	"context"
	"encoding/json"
	"os"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
)

// Step precheck step.
type Step struct {
	preCheckListPath string
	setupDirPath     string
}

// StepArgs define args for step.
type StepArgs struct {
	PreCheckListPath string
	SetupDirPath     string
}

// NewStep ...
func NewStep(args StepArgs) *Step {
	step := &Step{
		preCheckListPath: args.PreCheckListPath,
		setupDirPath:     args.SetupDirPath,
	}

	return step
}

// Run run the step to precheck.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(constant.StepPreCheck, constant.StateStart, "start precheck with config(%s)", step.preCheckListPath)

	list, err := loadCheckList(step.preCheckListPath)
	if err != nil {
		logger.Errorf(constant.StepPreCheck, constant.StateFailed,
			"failed to load precheck list from (%s), err: %v", step.preCheckListPath, err)
		return err
	}
	logger.Infof(constant.StepPreCheck, constant.StateRunning, "successfully loaded precheck list")

	r := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	gp := gopool.NewPool()

	runCheck := func(name string, checkFn func() error) func() error {
		return func() error {
			logger.Infof(constant.StepPreCheck, constant.StateRunning, "start check %s", name)

			if err := r.Do(ctx, func(_ int) error {
				return checkFn()
			}); err != nil {
				logger.Infof(constant.StepPreCheck, constant.StateFailed,
					"failed to check %s, err: %v", name, err)

				return err
			}

			logger.Infof(constant.StepPreCheck, constant.StateRunning, "successfully done check %s", name)

			return nil
		}
	}

	gp.Go(runCheck("disk free space", func() error { return CheckDiskFreeSpace(list.DiskRequires) }))
	gp.Go(runCheck("port policy", func() error { return CheckPortPolicies(ctx, list.PortPolicies) }))
	gp.Go(runCheck("network policy", func() error { return CheckNetworkPolicies(ctx, list.NetworkPolicies) }))
	gp.Go(runCheck("gse process", func() error {
		return CheckRemnantProcessInSetupDir(ctx, step.setupDirPath)
	}))

	if err := gp.Wait(); err != nil {
		logger.Infof(constant.StepPreCheck, constant.StateFailed, "failed to do all precheck, err: %s", err.Error())
		return err
	}

	logger.Infof(constant.StepPreCheck, constant.StateDone, "successfully done all precheck")

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

func loadCheckList(preCheckListPath string) (*CheckList, error) {
	if preCheckListPath == "" {
		return DefaultCheckList(), nil
	}

	bytes, err := os.ReadFile(preCheckListPath) // nolint: gosec
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
