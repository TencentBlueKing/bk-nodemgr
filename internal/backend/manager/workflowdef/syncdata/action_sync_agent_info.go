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
	// ActionNameSyncAgentInfo defines the action name.
	ActionNameSyncAgentInfo = "sync_agent_info"
)

// NewActionSyncAgentInfo creates a new syncAgentInfo.
func NewActionSyncAgentInfo(capability *Capability) action.Definition {
	return &actionSyncAgentInfo{
		gseHandler: capability.GSEHandler,
		topoStg:    capability.StorageTopo,
	}
}

// ActionParamSyncAgentInfo defines the action's param.
type ActionParamSyncAgentInfo struct {
	syncDataUtils.SyncDataActionStandardParam

	Hosts []*HostIDAgentID `json:"hosts"`
}

type actionSyncAgentInfo struct {
	gseHandler gse.IHandler
	topoStg    topoStg.IStorageHost
}

// Name returns the name of the action.
func (act *actionSyncAgentInfo) Name() string {
	return ActionNameSyncAgentInfo
}

// Version returns the version of the action.
func (act *actionSyncAgentInfo) Version() string {
	return "v1.0.0"
}

// Description returns the description of the action.
func (act *actionSyncAgentInfo) Description() string {
	return "sync agent info from gse"
}

// Timeout returns the timeout of the action.
func (act *actionSyncAgentInfo) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionSyncAgentInfo) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionSyncAgentInfo) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionSyncAgentInfo) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionSyncAgentInfo) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamSyncAgentInfo)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	agentIDs := make([]string, 0, len(param.Hosts))
	for _, host := range param.Hosts {
		agentIDs = append(agentIDs, host.AgentID)
	}

	result, err := act.gseHandler.ListAgentInfo(std.Context(), agentIDs...)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).With("agent-ids", agentIDs).Error("failed to list agent info")

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
		logger.G.Sys().Ctx(std.Context()).Info("no hosts to upsert")

		return nil
	}

	err = act.topoStg.UpdateHostDynamicFields(std.Context(), types.HostDynamicFields{
		NodeStatus:     true,
		NodeGeneration: true,
		NodeRole:       true,
		NodeVersion:    true,
		NodeCPUArch:    true,
		NodeOsType:     true,
	}, upsertHosts...)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to update host dynamic")

		return err
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionSyncAgentInfo) DisplayNameZh() string { return "同步 Agent 信息" }

// DisplayNameEn returns the English display name of the action.
func (act *actionSyncAgentInfo) DisplayNameEn() string { return "Sync Agent Info" }
