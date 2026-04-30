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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncAlivePluginProcessInfo defines the action name.
	ActionNameSyncAlivePluginProcessInfo = "sync_alive_plugin_process_info"
)

// NewActionSyncAlivePluginProcessInfo creates a new syncAgentInfo.
func NewActionSyncAlivePluginProcessInfo(capability *Capability) action.Definition {
	return &actionSyncAlivePluginProcessInfo{
		gseHandler: capability.GSEHandler,
		hostStg:    capability.StorageTopo,
		processStg: capability.StoragePlugin,
	}
}

// ActionParamSyncAlivePluginProcessInfo the action's param.
type ActionParamSyncAlivePluginProcessInfo struct {
	syncDataUtils.SyncDataActionStandardParam

	HostIDs []int64 `json:"host_ids"`
}

type actionSyncAlivePluginProcessInfo struct {
	gseHandler gse.IHandler
	hostStg    topoStg.IStorageHost
	processStg plugin.IDaoProcess
}

// Name returns the name of the action.
func (act *actionSyncAlivePluginProcessInfo) Name() string {
	return ActionNameSyncAlivePluginProcessInfo
}

// Version returns the version of the action.
func (act *actionSyncAlivePluginProcessInfo) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncAlivePluginProcessInfo) Description() string {
	return "sync alive process state from gse"
}

// Timeout returns the timeout of this action.
func (act *actionSyncAlivePluginProcessInfo) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionSyncAlivePluginProcessInfo) MaxRetryCount() uint {
	return 2 // nolint: mnd
}

// DelayFn returns the delay of this action.
func (act *actionSyncAlivePluginProcessInfo) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Tags returns the tags of this action.
func (act *actionSyncAlivePluginProcessInfo) Tags() []action.Tag {
	return []action.Tag{}
}

// Do the action.
func (act *actionSyncAlivePluginProcessInfo) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamSyncAlivePluginProcessInfo)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	hosts, err := act.hostStg.FindHostWithDynamic(std.Context(), types.UnlimitedPage(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			HostID: param.HostIDs,
		},
	})
	if err != nil {
		return err
	}

	if len(hosts) == 0 {
		logger.G.Sys().Ctx(std.Context()).With("host-ids", param.HostIDs).Info("no hosts found to sync alive process info")
		return nil
	}

	cond := &types.ProcessCondition{
		ExactInclude: &types.ProcessExactFields{
			HostID:     param.HostIDs,
			InfoStatus: []types.ProcessStatus{types.ProcessStatusRunning},
		},
	}
	aliveProcess, _, err := act.processStg.ListProcesses(std.Context(), types.UnlimitedPage(), cond)
	if err != nil {
		return err
	}

	if len(aliveProcess) == 0 {
		logger.G.Sys().Ctx(std.Context()).With("host-ids", param.HostIDs).Info("hosts have no alive process need to sync")
		return nil
	}

	agentIDHostIDMap, pluginNameAgentIDList := aggregateHostsAndProcesses(hosts, aliveProcess)
	procInfos, err := act.gseHandler.QueryMultiProcessInfoMany(std.Context(), pluginNameAgentIDList...)
	if err != nil {
		return err
	}

	if len(procInfos) == 0 {
		logger.G.Sys().Ctx(std.Context()).Info("no process info need to sync")
		return nil
	}

	processInfoDeltas := make([]*types.ProcessInfoDelta, 0)
	for name, infos := range procInfos {
		for _, info := range infos {
			hostID := agentIDHostIDMap[info.AgentID]
			processInfoDelta := &types.ProcessInfoDelta{
				HostID:      hostID,
				PluginName:  name,
				ProcessInfo: info,
			}

			processInfoDeltas = append(processInfoDeltas, processInfoDelta)
		}
	}

	if err = act.processStg.UpdateManyProcessInfo(std.Context(), processInfoDeltas); err != nil {
		return err
	}

	logger.G.Sys().Ctx(std.Context()).Info("sync alive plugin process info success, process count: %d", len(procInfos))

	return nil
}

func aggregateHostsAndProcesses(hosts []*types.Host, processes []*types.Process) (map[string]int64, []*types.ProcessAgentGroup) {
	hostIDAgentIDMap := make(map[int64]string)
	agentIDHostIDMap := make(map[string]int64)

	for _, host := range hosts {
		hostIDAgentIDMap[host.HostID] = host.Dynamic.AgentID
		agentIDHostIDMap[host.Dynamic.AgentID] = host.HostID
	}

	pluginNameAgentIDListMap := make(map[string][]string)
	pluginNameProcessNameMap := make(map[string]string)
	for _, proc := range processes {
		pluginNameAgentIDListMap[proc.PluginName] = append(pluginNameAgentIDListMap[proc.PluginName], hostIDAgentIDMap[proc.HostID])
		pluginNameProcessNameMap[proc.PluginName] = proc.Identity.Name
	}

	pluginNameAgentIDList := make([]*types.ProcessAgentGroup, 0)
	for pluginName, agentIDList := range pluginNameAgentIDListMap {
		pluginNameAgentIDList = append(pluginNameAgentIDList, &types.ProcessAgentGroup{
			PluginName:  pluginName,
			ProcessName: pluginNameProcessNameMap[pluginName],
			AgentIDList: agentIDList,
		})
	}

	return agentIDHostIDMap, pluginNameAgentIDList
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionSyncAlivePluginProcessInfo) DisplayNameZh() string {
	return "同步存活插件进程信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionSyncAlivePluginProcessInfo) DisplayNameEn() string {
	return "Sync Alive Plugin Process Info"
}
