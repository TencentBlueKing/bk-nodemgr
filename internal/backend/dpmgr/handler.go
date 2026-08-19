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

// Package dpmgr provides the deploy policy manager.
package dpmgr

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler defines the handler interface.
type IHandler interface {
	Do(nCtx contextx.IContext, deployPolicies ...*types.DeployPolicy) error
}

var _ IHandler = &Handler{}

// Handler define the handler.
type Handler struct {
	_ struct{}

	policyDiscovery IPolicyDiscovery

	calculator       IScopeCalculator
	conflictResolver IConflictResolver

	analyzer IAnalyzer
	executor IExecutor

	domainDeployPolicyMgr deploypolicy.IDomainDeployPolicyMgr
}

const (
	defaultCalculateConcurrency = 10
)

// NewHandler creates a new handler.
func NewHandler(conf *Config) *Handler {
	policyDiscovery := NewPolicyDiscovery(&PolicyDiscoveryConfig{
		DomainDeployPolicyMgr: conf.DomainDeployPolicyMgr,
	})

	calculator := NewScopeCalculator(&CalculatorConfig{
		CmdbClient:           conf.CmdbHandler,
		CalculateConcurrency: defaultCalculateConcurrency,
	})
	conflictResolver := NewConflictResolver()

	analyzer := NewAnalyzer(&AnalyzerConfig{
		DaoProcess: conf.DaoProcess,
		DaoHost:    conf.DaoHost,
	})
	executor := NewExecutor(&ExecutorConfig{
		NodeManager:   conf.NodeManager,
		PluginManager: conf.PluginManager,
		DaoPlugin:     conf.DaoPlugin,
	})

	return &Handler{
		domainDeployPolicyMgr: conf.DomainDeployPolicyMgr,
		policyDiscovery:       policyDiscovery,
		calculator:            calculator,
		conflictResolver:      conflictResolver,
		analyzer:              analyzer,
		executor:              executor,
	}
}

// Do does the handler.
func (h *Handler) Do(nCtx contextx.IContext, originDeployPolicies ...*types.DeployPolicy) error {
	// 1. discover these deploy policies's related deploy policies.
	relatedDeployPolicies, err := h.policyDiscovery.Discover(nCtx, originDeployPolicies...)
	if err != nil {
		return fmt.Errorf("failed to discover related deploy policies: %w", err)
	}

	originDeployWorkUnits := make([]*DeployUnit, len(relatedDeployPolicies))

	// 2. convert scope to Targets and spec
	for idx, deployPolicy := range relatedDeployPolicies {
		targets, err := h.calculator.Calculate(nCtx, deployPolicy.Scopes...)
		if err != nil {
			return fmt.Errorf("failed to calculate targets for policy, deploy-policy(%v): %w", deployPolicy, err)
		}

		originDeployWorkUnits[idx] = &DeployUnit{
			DeployPolicyID: deployPolicy.DeployPolicyID,
			LifeCycle:      deployPolicy.LifeCycle,
			Targets:        targets,
			Specs:          deployPolicy.Specs,
		}
	}

	// 3. resolve the conflict of work units.
	unConflictDeployWorkUnits, err := h.conflictResolver.ResolveConflict(originDeployWorkUnits)
	if err != nil {
		return fmt.Errorf("failed to resolve conflict: %w", err)
	}

	// 4. analyze the work units and design the change tasks.
	changeTasks, err := h.analyzer.Analyze(nCtx, unConflictDeployWorkUnits...)
	if err != nil {
		return fmt.Errorf("failed to analyze work units: %w", err)
	}

	// 5. executor and execute the change tasks.
	if err := h.executor.Execute(nCtx, changeTasks...); err != nil {
		return fmt.Errorf("failed to execute change tasks: %w", err)
	}

	// 6. update the deploy policy status.
	for _, policy := range relatedDeployPolicies {
		policy.LifeCycle.ExecutedAt = time.Now()
	}

	err = h.domainDeployPolicyMgr.RefreshExecuteInfo(nCtx, relatedDeployPolicies...)
	if err != nil {
		return fmt.Errorf("failed to refresh execute info: %w", err)
	}

	return nil
}
