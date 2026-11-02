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

package dpmgr

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var _ policyCalculator = &absentCalculation{}

type absentCalculation struct{}

// Calculate derives reverse changes from specs without forward scope calculation.
func (calculation *absentCalculation) Calculate(
	_ contextx.IContext, policies []*types.DeployPolicy,
) (*policyCalculation, error) {

	changeTasks := make([]*ChangeTask, 0)
	for _, policy := range policies {
		tasks, err := calculation.calculateAbsentChanges(policy)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate absent changes for policy %d: %w", policy.DeployPolicyID, err)
		}
		changeTasks = append(changeTasks, tasks...)
	}

	return &policyCalculation{
		originalUnits: make([]*DeployUnit, 0),
		resolvedUnits: make([]*DeployUnit, 0),
		changeTasks:   changeTasks,
	}, nil
}

func (calculation *absentCalculation) calculateAbsentChanges(policy *types.DeployPolicy) ([]*ChangeTask, error) {
	changeTasks := make([]*ChangeTask, 0)
	for specIndex, spec := range policy.Specs {
		switch spec.Type() {
		case types.DeploySpecTypeSpecifyPlugin:
			// TODO: calculate uninstall tasks for matching plugins.
			return nil, fmt.Errorf("ensure_absent is not implemented for spec %d (%s)", specIndex, spec.Type())
		case types.DeploySpecTypeSpecifyPluginPkg:
			// TODO: calculate uninstall tasks for plugin packages owned by this spec.
			return nil, fmt.Errorf("ensure_absent is not implemented for spec %d (%s)", specIndex, spec.Type())
		case types.DeploySpecTypeProjectPluginPkgToHosts:
			// TODO: calculate uninstall tasks for projected plugin packages owned by this spec.
			return nil, fmt.Errorf("ensure_absent is not implemented for spec %d (%s)", specIndex, spec.Type())
		case types.DeploySpecTypeSpecifyPluginSubConfig:
			// TODO: calculate deletion tasks for sub-configs owned by this spec.
			return nil, fmt.Errorf("ensure_absent is not implemented for spec %d (%s)", specIndex, spec.Type())
		case types.DeploySpecTypeSpecifyPluginSubConfigTemplate:
			// TODO: calculate deletion tasks for rendered sub-configs owned by this spec.
			return nil, fmt.Errorf("ensure_absent is not implemented for spec %d (%s)", specIndex, spec.Type())
		default:
			return nil, fmt.Errorf("unsupported ensure_absent spec %d (%s)", specIndex, spec.Type())
		}
	}

	return changeTasks, nil
}
