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
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
)

// Step precheck step.
type Step struct {
	preCheckListPath string
}

// StepArgs define args for step.
type StepArgs struct {
	PreCheckListPath string
}

// NewStep ...
func NewStep(args StepArgs) *Step {
	step := &Step{
		preCheckListPath: args.PreCheckListPath,
	}

	return step
}

// Run run the step to precheck.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(constant.StepPreCheck, constant.StateStart, "start precheck")

	list, err := loadCheckList(step.preCheckListPath)
	if err != nil {
		logger.Infof(constant.StepPreCheck, constant.StateFailed, "failed to load precheck list")
		return err
	}
	logger.Infof(constant.StepPreCheck, constant.StateRunning, "successfully loaded precheck list")

	gp := gopool.NewPool()
	gp.Go(func() error {
		logger.Infof(constant.StepPreCheck, constant.StateRunning, "start check disk free space")

		if err := CheckDiskFreeSpace(list.DiskRequires); err != nil {
			return err
		}

		logger.Infof(constant.StepPreCheck, constant.StateRunning, "successfully done check disk free space")

		return nil
	})

	gp.Go(func() error {
		logger.Infof(constant.StepPreCheck, constant.StateRunning, "start check port policy")

		if err := CheckPortPolicies(ctx, list.PortPolicies); err != nil {
			return err
		}

		logger.Infof(constant.StepPreCheck, constant.StateRunning, "successfully done check port policy")

		return nil
	})

	gp.Go(func() error {
		logger.Infof(constant.StepPreCheck, constant.StateRunning, "start check network policy")

		if err := CheckNetworkPolicies(ctx, list.NetworkPolicies); err != nil {
			return err
		}

		logger.Infof(constant.StepPreCheck, constant.StateRunning, "successfully done check network policy")

		return nil
	})

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
