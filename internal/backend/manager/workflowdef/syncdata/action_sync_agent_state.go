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

	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncAgentState defines the action name.
	ActionNameSyncAgentState = "sync_agent_state"
)

// NewActionSyncAgentState creates a new syncAgentState.
func NewActionSyncAgentState(capability *Capability) action.Definition {
	return &actionSyncAgentState{
		gseHandler: capability.GSEHandler,
		topoStg:    capability.StorageTopo,
	}
}

// ActionParamSyncAgentState the action's param.
type ActionParamSyncAgentState struct {
	syncDataUtils.SyncDataActionStandardParam

	Hosts []*HostIDAgentID `json:"hosts"`
}

type actionSyncAgentState struct {
	gseHandler gse.IHandler
	topoStg    topoStg.IStorageHost
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
	return time.Minute * 5
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
	param := new(ActionParamSyncAgentState)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	agentIDs := conv.SliceToSlice(param.Hosts, func(hostIDAgentID *HostIDAgentID) string {
		return hostIDAgentID.AgentID
	})

	result := make([]*types.AgentState, 0, len(agentIDs))
	for start := 0; start < len(agentIDs); start += gse.ListAgentStatePageSize {
		end := min(start+gse.ListAgentStatePageSize, len(agentIDs))

		batchAgentIDs := agentIDs[start:end]
		batchResult, err := act.gseHandler.ListAgentState(std.Context(), batchAgentIDs...)
		if err != nil {
			logger.G.Sys().Ctx(std.Context()).WithErr(err).With("agent-ids", batchAgentIDs).Error("failed to list agent state")

			return err
		}

		result = append(result, batchResult...)
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
				NodeRole:       agentState.NodeRole,
				NodeGeneration: agentState.NodeGeneration,
				NodeVersion:    agentState.Version,
				NodeStatus:     agentState.NodeStatus,
			},
		})
	}

	if len(upsertHosts) == 0 {
		logger.G.Sys().Ctx(std.Context()).Info("no hosts to upsert")

		return nil
	}

	err = act.topoStg.UpdateHostDynamicFields(std.Context(), types.HostDynamicFields{
		NodeRole:       true,
		NodeGeneration: true,
		NodeVersion:    true,
		NodeStatus:     true,
	}, upsertHosts...)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to update host dynamic")

		return err
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionSyncAgentState) DisplayNameZh() string { return "同步 Agent 状态" }

// DisplayNameEn returns the English display name of the action.
func (act *actionSyncAgentState) DisplayNameEn() string { return "Sync Agent State" }
