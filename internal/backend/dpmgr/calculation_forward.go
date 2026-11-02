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

var _ policyCalculator = &forwardCalculation{}

type forwardCalculation struct {
	calculator       IScopeCalculator
	conflictResolver IConflictResolver
	analyzer         IAnalyzer
}

// Calculate computes forward changes through scope calculation, conflict resolution and analysis.
func (calculation *forwardCalculation) Calculate(
	nCtx contextx.IContext, policies []*types.DeployPolicy,
) (*policyCalculation, error) {

	originalUnits := make([]*DeployUnit, len(policies))
	for idx, policy := range policies {
		targets, err := calculation.calculator.Calculate(nCtx, policy.Scopes...)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate targets for policy, deploy-policy(%v): %w", policy, err)
		}

		originalUnits[idx] = &DeployUnit{
			DeployPolicyID: policy.DeployPolicyID,
			LifeCycle:      policy.LifeCycle,
			Targets:        targets,
			Specs:          policy.Specs,
		}
	}

	resolvedUnits, err := calculation.conflictResolver.ResolveConflict(originalUnits)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve conflict: %w", err)
	}

	changeTasks, err := calculation.analyzer.Analyze(nCtx, resolvedUnits...)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze work units: %w", err)
	}

	return &policyCalculation{
		originalUnits: originalUnits,
		resolvedUnits: resolvedUnits,
		changeTasks:   changeTasks,
	}, nil
}
