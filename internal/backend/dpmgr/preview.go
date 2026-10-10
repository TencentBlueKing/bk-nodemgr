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
	"cmp"
	"fmt"
	"slices"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Preview calculates the requested policy's states without executing or persisting changes.
func (h *Handler) Preview(
	nCtx contextx.IContext, policy *types.DeployPolicy,
) ([]*types.DeployPolicySpecPreview, error) {

	policies, err := h.policyDiscovery.Discover(nCtx, policy)
	if err != nil {
		return nil, fmt.Errorf("failed to discover related deploy policies: %w", err)
	}

	calculation, err := h.calculateChanges(nCtx, policies)
	if err != nil {
		return nil, err
	}

	return buildPolicyPreview(policy, calculation), nil
}

func buildPolicyPreview(policy *types.DeployPolicy, calculation *policyCalculation) []*types.DeployPolicySpecPreview {
	targets := make([]*types.Target, 0)
	for _, unit := range calculation.originalUnits {
		if unit.DeployPolicyID == policy.DeployPolicyID {
			targets = unit.Targets
			break
		}
	}

	managed := make(map[string]struct{})
	for _, unit := range calculation.resolvedUnits {
		if unit.DeployPolicyID != policy.DeployPolicyID {
			continue
		}
		for _, target := range unit.Targets {
			managed[target.UniqueID()] = struct{}{}
		}
	}

	unsatisfied := make(map[int]map[string]struct{})
	for _, task := range calculation.changeTasks {
		if task.DeployPolicyID != policy.DeployPolicyID {
			continue
		}
		if _, ok := unsatisfied[task.specIndex]; !ok {
			unsatisfied[task.specIndex] = make(map[string]struct{})
		}
		unsatisfied[task.specIndex][task.Target.UniqueID()] = struct{}{}
	}

	items := make([]*types.DeployPolicySpecPreview, len(policy.Specs))
	for idx, spec := range policy.Specs {
		items[idx] = &types.DeployPolicySpecPreview{
			Spec:    spec,
			Results: previewTargets(targets, managed, unsatisfied[idx]),
		}
	}

	return items
}

func previewTargets(
	targets []*types.Target, managed, unsatisfied map[string]struct{},
) []*types.DeployPolicyTargetPreview {

	results := make([]*types.DeployPolicyTargetPreview, 0, len(targets))
	for _, target := range targets {
		result := &types.DeployPolicyTargetPreview{
			Target: target,
			Status: types.DeployPolicyPreviewStatusSatisfied,
		}
		results = append(results, result)
		key := target.UniqueID()
		if _, ok := managed[key]; !ok {
			result.Status = types.DeployPolicyPreviewStatusUnmanaged
			continue
		}
		if _, ok := unsatisfied[key]; ok {
			result.Status = types.DeployPolicyPreviewStatusUnsatisfied
		}
	}

	// Keep response order stable using the existing target identity, without changing deduplication.
	slices.SortFunc(results, func(left, right *types.DeployPolicyTargetPreview) int {
		if order := cmp.Compare(left.Target.Host.HostID, right.Target.Host.HostID); order != 0 {
			return order
		}

		return cmp.Compare(left.Target.ServiceInstance.ModuleID, right.Target.ServiceInstance.ModuleID)
	})

	return results
}
