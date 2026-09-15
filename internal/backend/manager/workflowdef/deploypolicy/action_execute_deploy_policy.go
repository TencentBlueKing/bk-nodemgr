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

package deploypolicy

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/dpmgr"
	deployPolicyUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/deploypolicy/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameExecuteDeployPolicy defines the action name.
	ActionNameExecuteDeployPolicy = "execute_deploy_policy"
)

// NewActionExecuteDeployPolicy creates a new actionExecuteDeployPolicy.
func NewActionExecuteDeployPolicy(capability *Capability) action.Definition {
	return &actionExecuteDeployPolicy{
		dpMgr:                   capability.DPMgr,
		daoDeployPolicy:         capability.StorageDeployPolicy,
		daoDeployPolicyWorkflow: capability.StorageDeployPolicy,
	}
}

// ActionParamExecuteDeployPolicy the action's param.
type ActionParamExecuteDeployPolicy struct {
	deployPolicyUtils.DeployPolicyActionStandardParam
	DeployPolicyIDs []int64 `json:"deploy_policy_ids"`
}

type actionExecuteDeployPolicy struct {
	dpMgr                   dpmgr.IHandler
	daoDeployPolicy         deploypolicy.IDaoDeployPolicy
	daoDeployPolicyWorkflow deploypolicy.IDaoDeployPolicyWorkflow
}

// Name returns the name of the action.
func (act *actionExecuteDeployPolicy) Name() string {
	return ActionNameExecuteDeployPolicy
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionExecuteDeployPolicy) DisplayNameZh() string { return "执行部署策略" }

// DisplayNameEn returns the English display name of the action.
func (act *actionExecuteDeployPolicy) DisplayNameEn() string { return "Execute Deploy Policy" }

// Version returns the version of the action.
func (act *actionExecuteDeployPolicy) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionExecuteDeployPolicy) Description() string {
	return "execute deploy policy"
}

// Timeout returns the timeout of this action.
func (act *actionExecuteDeployPolicy) Timeout() time.Duration {
	return 1 * time.Minute
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionExecuteDeployPolicy) MaxRetryCount() uint {
	return 0
}

// DelayFn returns the delay of this action.
func (act *actionExecuteDeployPolicy) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Tags returns the tags of this action.
func (act *actionExecuteDeployPolicy) Tags() []action.Tag {
	return []action.Tag{}
}

// Do the action.
func (act *actionExecuteDeployPolicy) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamExecuteDeployPolicy)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := deployPolicyUtils.NewDeployPolicyActionStandarder()
	if err = std.Initialize(ctx, param.DeployPolicyActionStandardParam); err != nil {
		return err
	}

	nCtx := std.Context()
	execution := dpmgr.ExecutionParam{
		OperationID: ctx.Data.OperationID,
		TriggerID:   ctx.Data.TriggerID,
		WorkflowIDs: make(map[int64]string),
	}
	if err := ensurePolicyWorkflows(nCtx, act.daoDeployPolicyWorkflow, execution, std.Operator(), param.DeployPolicyIDs); err != nil {
		return err
	}
	cond := &types.DeployPolicyCondition{
		ExactInclude: &types.DeployPolicyExactFields{
			DeployPolicyID: param.DeployPolicyIDs,
			Enabled:        []bool{true},
		},
	}
	deployPolicies, total, err := act.daoDeployPolicy.ListDeployPolicies(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(err).With("tenant-id", std.TenantID()).
			Error("failed to list deploy policies")

		return fmt.Errorf("failed to list deploy policies: %w", err)
	}

	if total == 0 {
		logger.G.Sys().Ctx(nCtx).With("tenant-id", std.TenantID()).
			Info("no deploy policy found")

		return fmt.Errorf("no deploy policy found")
	}

	if err := act.dpMgr.Do(nCtx, execution, deployPolicies...); err != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(err).With("tenant-id", std.TenantID()).
			Error("failed to execute deploy policy")

		return fmt.Errorf("failed to execute deploy policy: %w", err)
	}

	logger.G.Sys().Ctx(nCtx).With("tenant-id", std.TenantID()).
		Info("executed deploy policy")

	return nil
}

func ensurePolicyWorkflows(nCtx contextx.IContext, storage deploypolicy.IDaoDeployPolicyWorkflow,
	execution dpmgr.ExecutionParam, operator string, policyIDs []int64) error {

	var recordErr error
	for _, policyID := range policyIDs {
		parent, err := storage.EnsureDeployPolicyWorkflow(nCtx, execution.OperationID, execution.TriggerID, policyID, operator)
		if err != nil {
			recordErr = errors.Join(recordErr, fmt.Errorf("failed to ensure policy %d workflow: %w", policyID, err))

			continue
		}
		execution.WorkflowIDs[policyID] = parent.WorkflowID
	}

	return recordErr
}
