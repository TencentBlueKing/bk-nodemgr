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
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler defines the handler interface.
type IHandler interface {
	// Do executes policies and records their participating workflows.
	Do(nCtx contextx.IContext, execution ExecutionParam, deployPolicies ...*types.DeployPolicy) error
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

	domainDeployPolicyMgr   deploypolicy.IDomainDeployPolicyMgr
	daoDeployPolicyWorkflow deploypolicy.IDaoDeployPolicyWorkflow
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
		DaoProcess:       conf.DaoProcess,
		DaoProcessConfig: conf.DaoProcessConfig,
		DaoHost:          conf.DaoHost,
	})
	executor := NewExecutor(&ExecutorConfig{
		NodeManager:             conf.NodeManager,
		PluginManager:           conf.PluginManager,
		DaoPlugin:               conf.DaoPlugin,
		DaoProcessConfig:        conf.DaoProcessConfig,
		DaoDeployPolicyWorkflow: conf.DaoDeployPolicyWorkflow,
	})

	return &Handler{
		domainDeployPolicyMgr:   conf.DomainDeployPolicyMgr,
		daoDeployPolicyWorkflow: conf.DaoDeployPolicyWorkflow,
		policyDiscovery:         policyDiscovery,
		calculator:              calculator,
		conflictResolver:        conflictResolver,
		analyzer:                analyzer,
		executor:                executor,
	}
}

// Do executes policies and records their participating workflows.
func (h *Handler) Do(nCtx contextx.IContext, execution ExecutionParam,
	originDeployPolicies ...*types.DeployPolicy) error {

	if execution.OperationID == "" || execution.WorkflowIDs == nil {
		return errors.New("deploy policy operation identity and workflow map are required")
	}
	if h.daoDeployPolicyWorkflow == nil {
		return errors.New("deploy policy workflow storage is required")
	}

	// 1. discover these deploy policies's related deploy policies.
	relatedDeployPolicies, err := h.policyDiscovery.Discover(nCtx, originDeployPolicies...)
	if err != nil {
		return fmt.Errorf("failed to discover related deploy policies: %w", err)
	}

	execution.PolicyGroups = make(map[int64]int64, len(relatedDeployPolicies))
	for _, policy := range relatedDeployPolicies {
		execution.PolicyGroups[policy.DeployPolicyID] = policy.DsuID
	}
	if err := h.ensureDiscoveredWorkflows(nCtx, execution, relatedDeployPolicies); err != nil {
		return err
	}

	// 2-4. Calculate targets, resolve conflicts and analyze changes.
	calculation, err := h.calculateChanges(nCtx, relatedDeployPolicies)
	if err != nil {
		return err
	}

	// 5. executor and execute the change tasks.
	if err := h.executor.Execute(nCtx, execution, calculation.changeTasks...); err != nil {
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

func (h *Handler) ensureDiscoveredWorkflows(nCtx contextx.IContext, execution ExecutionParam,
	policies []*types.DeployPolicy) error {

	var recordErr error
	for _, policy := range policies {
		parent, err := h.daoDeployPolicyWorkflow.EnsureDeployPolicyWorkflow(nCtx,
			execution.OperationID, execution.TriggerID, policy.DeployPolicyID, nCtx.BKUsername())
		if err != nil {
			recordErr = errors.Join(recordErr, fmt.Errorf("failed to ensure policy %d workflow: %w", policy.DeployPolicyID, err))

			continue
		}
		execution.WorkflowIDs[policy.DeployPolicyID] = parent.WorkflowID
	}

	return recordErr
}
