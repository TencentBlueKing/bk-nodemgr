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

	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncAgentInfo defines the action name.
	ActionNameSyncAgentInfo = "sync_agent_info"
)

// NewActionSyncAgentInfo ...
func NewActionSyncAgentInfo(gseHandler gse.IHandler, topoStg topoStg.IStorageHost) action.Definition {

	return &SyncAgentInfo{
		gseHandler: gseHandler,
		topoStg:    topoStg,
	}
}

// ActParamSyncAgentInfo ...
type ActParamSyncAgentInfo struct {
	TenantID string           `json:"tenant_id"`
	Hosts    []*HostIDAgentID `json:"hosts"`
	Operator string           `json:"operator"`
}

// SyncAgentInfo ...
type SyncAgentInfo struct {
	gseHandler gse.IHandler
	topoStg    topoStg.IStorageHost
}

// Name returns the name of the action.
func (act *SyncAgentInfo) Name() string {
	return ActionNameSyncAgentInfo
}

// Version returns the version of the action.
func (act *SyncAgentInfo) Version() string {
	return "v1.0.0"
}

// Description returns the description of the action.
func (act *SyncAgentInfo) Description() string {
	return "sync agent info from gse"
}

// Timeout returns the timeout of the action.
func (act *SyncAgentInfo) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *SyncAgentInfo) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *SyncAgentInfo) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *SyncAgentInfo) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *SyncAgentInfo) Do(ctx *action.InstanceContext) error {
	param := new(ActParamSyncAgentInfo)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	newCtx := contextx.From(ctx.Ctx, contextx.WithTenantID(param.TenantID), contextx.WithBKUsername(param.Operator))

	agentIDs := make([]string, 0, len(param.Hosts))
	for _, host := range param.Hosts {
		agentIDs = append(agentIDs, host.AgentID)
	}

	result, err := act.gseHandler.ListAgentInfo(newCtx, agentIDs...)
	if err != nil {
		logger.G.Sys().WithErr(err).With("agent-ids", agentIDs).Error("failed to list agent state")

		return err
	}

	agentInfos, err := conv.SliceToMap(result, func(state *types.AgentInfo) string { return state.AgentID })
	if err != nil {
		return err
	}

	upsertHosts := make([]*types.Host, len(param.Hosts))
	for idx := range param.Hosts {
		host := param.Hosts[idx]
		agentInfo, ok := agentInfos[host.AgentID]
		if !ok {
			continue
		}

		upsertHosts[idx] = &types.Host{
			HostID: host.HostID,
			Dynamic: &types.HostDynamic{
				NodeStatus:     agentInfo.NodeStatus,
				NodeGeneration: agentInfo.NodeGeneration,
				NodeRole:       agentInfo.NodeRole,
				NodeVersion:    agentInfo.Version,
				NodeCPUArch:    agentInfo.Arch,
				NodeOsType:     agentInfo.OSType,
			},
		}
	}

	if len(upsertHosts) == 0 {
		logger.G.Sys().Info("no hosts to upsert")

		return nil
	}

	err = act.topoStg.UpdateHostDynamicFields(newCtx, types.HostDynamicFields{
		NodeStatus:     true,
		NodeGeneration: true,
		NodeRole:       true,
		NodeVersion:    true,
		NodeCPUArch:    true,
		NodeOsType:     true,
	}, upsertHosts...)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to update host dynamic")

		return err
	}

	return nil
}
