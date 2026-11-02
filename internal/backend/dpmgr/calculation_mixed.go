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

var _ policyCalculator = &mixedCalculation{}

type mixedCalculation struct {
	forward policyCalculator
	absent  policyCalculator
}

// Calculate partitions policies by direction and combines their calculation results.
func (calculation *mixedCalculation) Calculate(
	nCtx contextx.IContext, policies []*types.DeployPolicy,
) (*policyCalculation, error) {

	forwardPolicies := make([]*types.DeployPolicy, 0, len(policies))
	absentPolicies := make([]*types.DeployPolicy, 0, len(policies))
	for _, policy := range policies {
		if policy.EnsureAbsent {
			absentPolicies = append(absentPolicies, policy)
			continue
		}
		forwardPolicies = append(forwardPolicies, policy)
	}

	absent, err := calculation.absent.Calculate(nCtx, absentPolicies)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate absent policies: %w", err)
	}

	forward, err := calculation.forward.Calculate(nCtx, forwardPolicies)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate forward policies: %w", err)
	}

	return &policyCalculation{
		originalUnits: append(forward.originalUnits, absent.originalUnits...),
		resolvedUnits: append(forward.resolvedUnits, absent.resolvedUnits...),
		changeTasks:   append(forward.changeTasks, absent.changeTasks...),
	}, nil
}
