/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package dpmgr

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IPolicyDiscovery defines the interface of policy discovery.
type IPolicyDiscovery interface {
	// Discover discovers the related deploy policies.
	Discover(nCtx contextx.IContext, originDeployPolicies ...*types.DeployPolicy) ([]*types.DeployPolicy, error)
}

var _ IPolicyDiscovery = &PolicyDiscovery{}

// PolicyDiscovery defines the policy discovery.
type PolicyDiscovery struct {
	domainDeployPolicyMgr deploypolicy.IDomainDeployPolicyMgr
}

// PolicyDiscoveryConfig defines the config of policy discovery.
type PolicyDiscoveryConfig struct {
	DomainDeployPolicyMgr deploypolicy.IDomainDeployPolicyMgr
}

// NewPolicyDiscovery creates a new policy discovery.
func NewPolicyDiscovery(conf *PolicyDiscoveryConfig) *PolicyDiscovery {
	return &PolicyDiscovery{
		domainDeployPolicyMgr: conf.DomainDeployPolicyMgr,
	}
}

// Discover discovers the related deploy policies.
func (discover *PolicyDiscovery) Discover(nCtx contextx.IContext, originPolicies ...*types.DeployPolicy) ([]*types.DeployPolicy, error) {
	checkMap := make(map[int64]struct{})
	relatedPolicies := make([]*types.DeployPolicy, 0)
	for _, originPolicy := range originPolicies {
		if _, ok := checkMap[originPolicy.DeployPolicyID]; ok {
			continue
		}

		result, err := discover.discoverRelatedDeployPolicy(nCtx, originPolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to discover related deploy policies: %w", err)
		}

		for _, policy := range result {
			if _, ok := checkMap[policy.DeployPolicyID]; ok {
				continue
			}

			checkMap[policy.DeployPolicyID] = struct{}{}
			relatedPolicies = append(relatedPolicies, policy)
		}
	}

	return relatedPolicies, nil
}

// nolint: gocognit
func (discover *PolicyDiscovery) discoverRelatedDeployPolicy(nCtx contextx.IContext, policy *types.DeployPolicy) ([]*types.DeployPolicy, error) {
	result := []*types.DeployPolicy{policy}

	state := newBfsState()

	state.pushPolicyToQueue(policy)

	for len(state.queue) > 0 {
		currentPolicy := state.popPolicyFromQueue()
		if currentPolicy == nil {
			break
		}

		if state.isVisitedPolicy(currentPolicy) {
			continue
		}

		state.markPolicyAsVisited(currentPolicy)

		for _, spec := range currentPolicy.Specs {
			isSpecVisited, err := state.isVisitedSpec(spec)
			if err != nil {
				return nil, fmt.Errorf("failed to check if spec is visited: %w", err)
			}
			if isSpecVisited {
				continue
			}

			if err := state.markSpecAsVisited(spec); err != nil {
				return nil, fmt.Errorf("failed to mark spec as visited: %w", err)
			}

			relatedPolicies, err := discover.discoverPoliciesBySpec(nCtx, spec)
			if err != nil {
				return nil, fmt.Errorf("failed to find related policies: %w", err)
			}

			for _, relatedPolicy := range relatedPolicies {
				if state.isVisitedPolicy(relatedPolicy) {
					continue
				}

				result = append(result, relatedPolicy)
				state.pushPolicyToQueue(relatedPolicy)
			}
		}
	}

	return result, nil
}

type bfsState struct {
	queue           []*types.DeployPolicy
	visitedSpecs    map[string]struct{}
	visitedPolicies map[int64]struct{}
}

func newBfsState() *bfsState {
	return &bfsState{
		queue:           make([]*types.DeployPolicy, 0),
		visitedSpecs:    make(map[string]struct{}),
		visitedPolicies: make(map[int64]struct{}),
	}
}

func (state *bfsState) pushPolicyToQueue(policy *types.DeployPolicy) {
	state.queue = append(state.queue, policy)
}

func (state *bfsState) popPolicyFromQueue() *types.DeployPolicy {
	if len(state.queue) == 0 {
		return nil
	}

	policy := state.queue[0]
	state.queue = state.queue[1:]

	return policy
}

func (state *bfsState) isVisitedSpec(spec *types.DeploySpec) (bool, error) {
	uniqueID, err := spec.UniqueID()
	if err != nil {
		return false, fmt.Errorf("failed to get spec unique ID: %w", err)
	}
	_, ok := state.visitedSpecs[uniqueID]

	return ok, nil
}

func (state *bfsState) markSpecAsVisited(spec *types.DeploySpec) error {
	uniqueID, err := spec.UniqueID()
	if err != nil {
		return fmt.Errorf("failed to get spec unique ID: %w", err)
	}
	state.visitedSpecs[uniqueID] = struct{}{}

	return nil
}

func (state *bfsState) isVisitedPolicy(policy *types.DeployPolicy) bool {
	_, ok := state.visitedPolicies[policy.DeployPolicyID]
	return ok
}

func (state *bfsState) markPolicyAsVisited(policy *types.DeployPolicy) {
	state.visitedPolicies[policy.DeployPolicyID] = struct{}{}
}

func (discover *PolicyDiscovery) discoverPoliciesBySpec(nCtx contextx.IContext, targetSpec *types.DeploySpec) ([]*types.DeployPolicy, error) {
	switch targetSpec.Type() {
	case types.DeploySpecTypeSpecifyPlugin:
		param, err := targetSpec.GetSpecifyPluginParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get spec specify plugin param: %w", err)
		}

		policies, err := discover.domainDeployPolicyMgr.DiscoverPoliciesBySpecifyPlugin(nCtx, param)
		if err != nil {
			return nil, fmt.Errorf("failed to discover policies by specify plugin: %w", err)
		}

		return policies, nil
	case types.DeploySpecTypeSpecifyPluginPkg:
		// spec specify plugin pkg no need to discover other policies, it's a leaf node
		return []*types.DeployPolicy{}, nil
	case types.DeploySpecTypeSpecifyProxy:
		// TODO: implement specify proxy
		return nil, fmt.Errorf("specify proxy not supported yet")
	case types.DeploySpecTypeSpecifyAgent:
		// TODO: implement specify agent
		return nil, fmt.Errorf("specify agent not supported yet")
	default:
		return nil, fmt.Errorf("unsupported spec type: %s", targetSpec.Type())
	}
}
