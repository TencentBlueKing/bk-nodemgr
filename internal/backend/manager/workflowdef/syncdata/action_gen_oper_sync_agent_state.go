/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package syncdata

import (
	"context"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameGenOperSyncAgentState defines the action name.
	ActionNameGenOperSyncAgentState = "gen_oper_sync_agent_state"

	// MaxPageSize defines the max page size for page executor.
	// In the scenario of 40,000 hosts, a single request for 1,000 hosts requires 400 table lookups.
	MaxPageSize = 1000
)

// NewActionGenOperSyncAgentState this action will create host sync operation for all business.
func NewActionGenOperSyncAgentState(topoStg topo.IStorageHost, workflowCtl workflow.IController) action.Definition {
	return &actionGenOperSyncAgentState{
		topoStg:     topoStg,
		workflowCtl: workflowCtl,
	}
}

// GenOperSyncAgentStateParam ...
type GenOperSyncAgentStateParam struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

type actionGenOperSyncAgentState struct {
	topoStg     topo.IStorageHost
	workflowCtl workflow.IController
}

// Name returns the name of the action.
func (act *actionGenOperSyncAgentState) Name() string {
	return ActionNameGenOperSyncAgentState
}

// Version returns the version of the action.
func (act *actionGenOperSyncAgentState) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenOperSyncAgentState) Description() string {
	return "create sync agent state operation for all hosts"
}

// Timeout returns the timeout of the action.
func (act *actionGenOperSyncAgentState) Timeout() time.Duration {
	return time.Second * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionGenOperSyncAgentState) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionGenOperSyncAgentState) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionGenOperSyncAgentState) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionGenOperSyncAgentState) Do(ctx *action.InstanceContext) error {
	param := new(GenOperSyncAgentStateParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	executor := runtime.NewPageExecutor[*types.Host](MaxPageSize, 1*time.Hour)
	fn := func(fnCtx context.Context, p types.Page) ([]*types.Host, error) {
		hosts, err := act.topoStg.FindHostWithDynamic(fnCtx, p, nil)
		if err != nil {
			return nil, err
		}

		tenantUserCtx := contextx.NewTenantUserContext(fnCtx, param.TenantID, param.Operator)

		if err = act.executeOper(tenantUserCtx, ctx.Data, hosts...); err != nil {
			return nil, err
		}

		return hosts, nil
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, param.TenantID)
	if err != nil {
		return err
	}

	result, err := executor.Execute(tenantCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	ctx.Data.LogI(fmt.Sprintf("executed sync agent state operation for %d hosts", result.Total))

	return nil
}

// executeOper create an operation to sync agent state for the given hosts and then execute it.
func (act *actionGenOperSyncAgentState) executeOper(
	ctx contextx.ITenantUserContext, actionInstData *action.InstanceData, hosts ...*types.Host) error {

	trigCtl, err := act.workflowCtl.GetTrigger(ctx, actionInstData.TriggerID)
	if err != nil {
		return fmt.Errorf("failed to get workflow trigger by trigger-id(%s): %w", actionInstData.TriggerID, err)
	}

	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	hostAgentID := make([]*HostIDAgentID, 0, len(hosts))
	for _, host := range hosts {
		hostAgentID = append(hostAgentID, &HostIDAgentID{
			HostID:  host.HostID,
			AgentID: host.Dynamic.AgentID,
		})
	}
	operationDef := NewOperSyncAgentStateFromGSE(OperParamSyncAgentStateFromGSE{
		TenantID: tenantID,
		Hosts:    hostAgentID,
		Operator: operator,
	})

	operationParam := operationDef.DefaultParameters()
	operationParam.ParentOperationID = actionInstData.OperationID
	operCtl, err := trigCtl.CreateOperation(ctx, operationDef, operationParam)
	if err != nil {
		actionInstData.LogE(
			fmt.Sprintf("failed to create sync agent state operation, tenant-id(%s), operation-id(%s): %s",
				tenantID, actionInstData.OperationID, err.Error()))

		return err
	}

	actionInstData.LogI(
		fmt.Sprintf("created sync agent state operation, tenant-id(%s), operation-id(%s)",
			tenantID, operCtl.GetOperationID()))

	return nil
}
