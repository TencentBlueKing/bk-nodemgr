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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncAgentState defines the action name.
	ActionNameSyncAgentState = "sync_agent_state"
)

// NewActionSyncAgentState creates a new syncAgentState.
func NewActionSyncAgentState(gseHandler gse.IHandler, topoStg topo.IStorageHost,
	logger logger.ILogger) action.Definition {

	return &actionSyncAgentState{
		gseHandler: gseHandler,
		topoStg:    topoStg,
		logger:     logger,
	}
}

// SyncAgentStateParam the action's param.
type SyncAgentStateParam struct {
	TenantID string           `json:"tenant_id"`
	Hosts    []*HostIDAgentID `json:"hosts"`
	Operator string           `json:"operator"`
}

// HostIDAgentID defines the host ID and agent ID.
type HostIDAgentID struct {
	HostID  int64  `json:"host_id"`
	AgentID string `json:"agent_id"`
}

type actionSyncAgentState struct {
	gseHandler gse.IHandler
	topoStg    topo.IStorageHost
	logger     logger.ILogger
}

// Name returns the name of the action.
func (act *actionSyncAgentState) Name() string {
	return ActionNameSyncAgentState
}

// Version returns the version of the action.
func (act *actionSyncAgentState) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncAgentState) Description() string {
	return "sync agent state from gse"
}

// Timeout returns the timeout of this action.
func (act *actionSyncAgentState) Timeout() time.Duration {
	return 1 * time.Minute
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionSyncAgentState) MaxRetryCount() uint {
	return 2 // nolint: mnd
}

// DelayFn returns the delay of this action.
func (act *actionSyncAgentState) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Tags returns the tags of this action.
func (act *actionSyncAgentState) Tags() []action.Tag {
	return []action.Tag{}
}

// Do the action.
func (act *actionSyncAgentState) Do(ctx *action.InstanceContext) error {
	param := new(SyncAgentStateParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, param.TenantID)
	if err != nil {
		return err
	}

	newCtx := contextx.New(ctx.Ctx, contextx.WithTenantID(param.TenantID), contextx.WithBKUsername(param.Operator))

	agentIDs := make([]string, 0, len(param.Hosts))
	for _, host := range param.Hosts {
		agentIDs = append(agentIDs, host.AgentID)
	}

	result, err := act.gseHandler.ListAgentState(newCtx, agentIDs...)
	if err != nil {
		act.logger.Errorf("list agent state by agent-id-list(%v) failed: %v", agentIDs, err)
		return err
	}

	agentStates, err := conv.SliceToMap(result, func(state *types.AgentState) string { return state.AgentID })
	if err != nil {
		return err
	}

	upsertHosts := make([]*types.Host, 0, len(param.Hosts))
	for _, host := range param.Hosts {
		agentState, ok := agentStates[host.AgentID]
		if !ok {
			continue
		}

		upsertHosts = append(upsertHosts, &types.Host{
			HostID: host.HostID,
			Dynamic: &types.HostDynamic{
				NodeVersion: agentState.Version,
				NodeStatus:  agentState.StatusCode.ToNodeStatus(),
			},
		})
	}

	if len(upsertHosts) == 0 {
		act.logger.Info("no hosts to upsert")
		return nil
	}

	err = act.topoStg.UpdateHostDynamicFields(tenantCtx, types.HostDynamicFields{NodeVersion: true, NodeStatus: true}, upsertHosts...)
	if err != nil {
		act.logger.Errorf("failed to update host dynamic: %v", err)
		return err
	}

	return nil
}
