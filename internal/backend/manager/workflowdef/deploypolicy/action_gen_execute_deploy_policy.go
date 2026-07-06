/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deploypolicy

import (
	"fmt"
	"time"

	deployPolicyUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/deploypolicy/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	// ActionNameGenOperExecuteDeployPolicy defines the action name.
	ActionNameGenOperExecuteDeployPolicy = "gen_oper_execute_deploy_policy"

	deployPolicyExecutionInterval = 12 * time.Hour
)

// NewActionGenOperExecuteDeployPolicy creates a new actionGenOperExecuteDeployPolicy.
func NewActionGenOperExecuteDeployPolicy(capability *Capability) action.Definition {
	return &actionGenOperExecuteDeployPolicy{
		daoDeployPolicy: capability.StorageDeployPolicy,
		workflowCtl:     capability.WorkflowCtl,
	}
}

// ActionParamGenOperExecuteDeployPolicy the action's param.
type ActionParamGenOperExecuteDeployPolicy struct {
	deployPolicyUtils.DeployPolicyActionStandardParam
}

type actionGenOperExecuteDeployPolicy struct {
	daoDeployPolicy deploypolicy.IDaoDeployPolicy
	workflowCtl     workflow.IController
}

// Name returns the name of the action.
func (act *actionGenOperExecuteDeployPolicy) Name() string {
	return ActionNameGenOperExecuteDeployPolicy
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionGenOperExecuteDeployPolicy) DisplayNameZh() string {
	return "生成执行部署策略任务"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionGenOperExecuteDeployPolicy) DisplayNameEn() string {
	return "Generate Execute Deploy Policy Operation"
}

// Version returns the version of the action.
func (act *actionGenOperExecuteDeployPolicy) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenOperExecuteDeployPolicy) Description() string {
	return fmt.Sprintf("this action will find the deploy policy that has not been executed in the last (%f) hours.",
		deployPolicyExecutionInterval.Hours())
}

// Timeout returns the timeout of this action.
func (act *actionGenOperExecuteDeployPolicy) Timeout() time.Duration {
	return 30 * time.Minute // nolint: mnd
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionGenOperExecuteDeployPolicy) MaxRetryCount() uint {
	return 2 // nolint: mnd
}

// DelayFn returns the delay of this action.
func (act *actionGenOperExecuteDeployPolicy) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Minute)
	}
}

// Tags returns the tags of this action.
func (act *actionGenOperExecuteDeployPolicy) Tags() []action.Tag {
	return []action.Tag{}
}

// Do the action.
func (act *actionGenOperExecuteDeployPolicy) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamGenOperExecuteDeployPolicy)
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

	// list deploy policies that have not been executed within deployPolicyExecutionInterval.
	// endTime represents the time point deployPolicyExecutionInterval ago, policies executed before this time need to be executed again.
	endTime := time.Now().Add(-deployPolicyExecutionInterval)
	timeRange := types.BeforeTimeRange(endTime)

	cond := &types.DeployPolicyCondition{
		ExecutedTimeRange: &timeRange,
		ExactInclude: &types.DeployPolicyExactFields{
			Enabled: []bool{true},
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
			Info("no deploy policy found that needs to be executed")

		return nil
	}

	dsuGroup := make(map[int64][]int64)
	for _, dp := range deployPolicies {
		dsuGroup[dp.DsuID] = append(dsuGroup[dp.DsuID], dp.DeployPolicyID)
	}

	// create trigger for handling execute deploy policy
	meta := trigger.NewMetadataOrdered(10) // nolint: mnd
	meta.CleanPolicy = trigger.MetadataCleanPolicy{
		MaxDays: 30, // nolint: mnd
	}
	trigCtl, err := act.workflowCtl.CreateTrigger(nCtx, trigger.CategoryOrdered, meta)
	if err != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(err).With("action", act.Name()).Error("failed to create trigger for handling execute deploy policy")
		return err
	}

	for dsuID, policyIDs := range dsuGroup {
		if err = act.executeOper(std, trigCtl, policyIDs...); err != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(err).With("dsu-id", dsuID, "policy-ids", policyIDs).
				Error("failed to execute deploy policy")

			return fmt.Errorf("failed to execute deploy policy: %w", err)
		}
	}

	logger.G.Sys().Ctx(nCtx).With("action", act.Name(), "tenant-id", std.TenantID(),
		"total-policies", total, "dsu-count", len(dsuGroup)).
		Info("executed deploy policy")

	return nil
}

func (act *actionGenOperExecuteDeployPolicy) executeOper(std *deployPolicyUtils.DeployPolicyActionStandarder, trigCtl workflow.ITriggerCtl,
	policyIDs ...int64) error {

	operationDef := NewOperExecuteDeployPolicy(OperParamExecuteDeployPolicy{
		TenantID:        std.TenantID(),
		Operator:        std.Operator(),
		DeployPolicyIDs: policyIDs,
	})
	operationParam := operationDef.DefaultParameters()
	operationParam.ParentOperationID = std.InstanceData().OperationID
	operationParam.ParentOperInstID = std.InstanceData().OperationInstanceID

	operCtl, err := trigCtl.CreateOperation(std.Context(), operationDef, operationParam)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).
			WithErr(err).
			With("action", act.Name(), "tenant-id", std.TenantID()).
			Error("failed to create execute deploy policy operation")

		return err
	}

	logger.G.Sys().Ctx(std.Context()).
		With("action", act.Name(), "tenant-id", std.TenantID(), "operation-id", operCtl.GetOperationID()).
		Info("created execute deploy policy operation")

	return nil
}
