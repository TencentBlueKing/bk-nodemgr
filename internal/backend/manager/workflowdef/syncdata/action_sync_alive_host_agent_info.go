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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncAliveHostAgentInfo defines the action name.
	ActionNameSyncAliveHostAgentInfo = "sync_alive_host_agent_info"

	// syncAgentInfoMaxPageSize defines the max page size for page executor.
	// In the scenario of 40,000 hosts, a single request for 1,000 hosts requires 400 table lookups.
	syncAgentInfoMaxPageSize = 100
)

// NewActionSyncAliveHostAgentInfo this action will create host sync operation for all business.
func NewActionSyncAliveHostAgentInfo(topoStg topo.IStorageHost, workflowCtl workflow.IController) action.Definition {
	return &actionSyncAliveHostAgentInfo{
		topoStg:     topoStg,
		workflowCtl: workflowCtl,
	}
}

// SyncAliveHostAgentInfoParam ...
type SyncAliveHostAgentInfoParam struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

type actionSyncAliveHostAgentInfo struct {
	topoStg     topo.IStorageHost
	workflowCtl workflow.IController
}

// Name returns the name of the action.
func (act *actionSyncAliveHostAgentInfo) Name() string {
	return ActionNameSyncAliveHostAgentInfo
}

// Version returns the version of the action.
func (act *actionSyncAliveHostAgentInfo) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncAliveHostAgentInfo) Description() string {
	return "create sync agent info operation for alive hosts"
}

// Timeout returns the timeout of the action.
func (act *actionSyncAliveHostAgentInfo) Timeout() time.Duration {
	return time.Second * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionSyncAliveHostAgentInfo) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionSyncAliveHostAgentInfo) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionSyncAliveHostAgentInfo) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionSyncAliveHostAgentInfo) Do(ctx *action.InstanceContext) error {
	param := new(SyncAliveHostAgentInfoParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	nCtx := contextx.New(ctx.Ctx, contextx.WithTenantID(param.TenantID), contextx.WithBKUsername(param.Operator))
	executor := pageexecutor.NewPageExecutor[*types.Host](syncAgentInfoMaxPageSize, 1*time.Hour)
	fn := func(_ context.Context, p types.Page) ([]*types.Host, error) {
		cond := &types.HostCondition{
			ExactInclude: &types.HostExactFields{
				NodeStatus: []types.NodeStatus{types.NodeStatusRunning},
			},
			ExactExclude: &types.HostExactFields{
				AgentID: []string{""},
			},
		}

		hosts, err := act.topoStg.FindHostWithDynamic(nCtx, p, cond)
		if err != nil {
			return nil, err
		}

		if err = act.executeOper(nCtx, ctx.Data, hosts...); err != nil {
			return nil, err
		}

		return hosts, nil
	}

	result, err := executor.Execute(ctx.Ctx, types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	ctx.Data.LogI(fmt.Sprintf("executed sync agent info operation for %d hosts", result.Total))

	return nil
}

// executeOper create an operation to sync agent info for the given hosts and then execute it.
func (act *actionSyncAliveHostAgentInfo) executeOper(
	ctx contextx.IContext, actionInstData *action.InstanceData, hosts ...*types.Host) error {

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
	operationDef := NewOperSyncAgentInfoFromGSE(OperParamSyncAgentInfoFromGSE{
		TenantID: tenantID,
		Hosts:    hostAgentID,
		Operator: operator,
	})

	operationParam := operationDef.DefaultParameters()
	operationParam.ParentOperationID = actionInstData.OperationID
	operCtl, err := trigCtl.CreateOperation(ctx, operationDef, operationParam)
	if err != nil {
		actionInstData.LogE(
			fmt.Sprintf("failed to create sync agent info operation, tenant-id(%s), operation-id(%s): %s",
				tenantID, actionInstData.OperationID, err.Error()))

		return err
	}

	actionInstData.LogI(
		fmt.Sprintf("created sync agent info operation, tenant-id(%s), operation-id(%s)",
			tenantID, operCtl.GetOperationID()))

	return nil
}
