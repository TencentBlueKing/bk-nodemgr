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

package syncdata

import (
	"fmt"
	"slices"
	"time"

	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncStoppedPluginProcessV2Info defines the action name.
	ActionNameSyncStoppedPluginProcessV2Info = "sync_stopped_plugin_process_v2_info"
)

// NewActionSyncStoppedPluginProcessV2Info creates a new sync stopped v2 plugin process info action.
func NewActionSyncStoppedPluginProcessV2Info(capability *Capability) action.Definition {
	return &actionSyncStoppedPluginProcessV2Info{
		gseHandler: capability.GSEHandler,
		hostStg:    capability.StorageTopo,
		processStg: capability.StoragePlugin,
	}
}

// ActionParamSyncStoppedPluginProcessV2Info the action's param.
type ActionParamSyncStoppedPluginProcessV2Info struct {
	syncDataUtils.SyncDataActionStandardParam

	HostIDs []int64 `json:"host_ids"`
}

type actionSyncStoppedPluginProcessV2Info struct {
	gseHandler gse.IHandler
	hostStg    topoStg.IStorageHost
	processStg plugin.IDaoProcessV2
}

// Name returns the name of the action.
func (act *actionSyncStoppedPluginProcessV2Info) Name() string {
	return ActionNameSyncStoppedPluginProcessV2Info
}

// Version returns the version of the action.
func (act *actionSyncStoppedPluginProcessV2Info) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncStoppedPluginProcessV2Info) Description() string {
	return "sync stopped v2 plugin process state from gse"
}

// Timeout returns the timeout of this action.
func (act *actionSyncStoppedPluginProcessV2Info) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionSyncStoppedPluginProcessV2Info) MaxRetryCount() uint {
	return 2 // nolint: mnd
}

// DelayFn returns the delay of this action.
func (act *actionSyncStoppedPluginProcessV2Info) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Tags returns the tags of this action.
func (act *actionSyncStoppedPluginProcessV2Info) Tags() []action.Tag {
	return []action.Tag{}
}

// Queue returns the queue of the action.
func (act *actionSyncStoppedPluginProcessV2Info) Queue() action.Queue {
	return action.QueueDefault
}

// Do the action.
func (act *actionSyncStoppedPluginProcessV2Info) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamSyncStoppedPluginProcessV2Info)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
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
		logger.G.Sys().Ctx(std.Context()).
			With("host-ids", param.HostIDs).
			Info("no hosts found to sync stopped v2 process info")

		return nil
	}

	stoppedProcesses, err := act.listProcessesToSync(std.Context(), param.HostIDs)
	if err != nil {
		return err
	}

	if len(stoppedProcesses) == 0 {
		logger.G.Sys().Ctx(std.Context()).
			With("host-ids-count", len(param.HostIDs)).
			Info("no stopped v2 plugin process found to sync")

		return nil
	}

	hostIDAgentIDMap := make(map[int64]string, len(hosts))
	agentIDHostIDMap := make(map[string]int64, len(hosts))
	for _, host := range hosts {
		hostIDAgentIDMap[host.HostID] = host.Dynamic.AgentID
		agentIDHostIDMap[host.Dynamic.AgentID] = host.HostID
	}

	agentAliveProcesses, agentNotAliveProcesses, err := act.splitProcessesByAgentState(
		std.Context(), hostIDAgentIDMap, stoppedProcesses)
	if err != nil {
		return err
	}

	processInfoDeltas, queryErr := act.queryProcessInfoDeltas(std.Context(), hostIDAgentIDMap, agentIDHostIDMap, agentAliveProcesses)

	// The recorded stopped state of these processes can no longer be confirmed, so it is downgraded to unknown.
	agentNotAliveDeltas := conv.SliceToSlice(agentNotAliveProcesses, func(proc *types.Process) *types.ProcessInfoDelta {
		return newUnknownProcessInfoDelta(proc, hostIDAgentIDMap[proc.HostID])
	})
	processInfoDeltas = slices.Concat(processInfoDeltas, agentNotAliveDeltas)

	if len(processInfoDeltas) == 0 {
		if queryErr != nil {
			return queryErr
		}

		logger.G.Sys().Ctx(std.Context()).
			With("stopped-proc-count", len(stoppedProcesses)).
			Info("no stopped v2 plugin process info needs update")

		return nil
	}

	if err = batchHandleProcessInfoDeltas(std.Context(), agentNotAliveDeltas,
		func(processInfoDeltas ...*types.ProcessInfoDelta) error {
			return act.processStg.UpdateManyProcessInfoV2(std.Context(), processInfoDeltas)
		}); err != nil {
		return err
	}

	if queryErr != nil {
		return queryErr
	}

	logger.G.Sys().Ctx(std.Context()).
		With("need-update-proc-infos", len(processInfoDeltas)).
		Info("sync stopped v2 plugin process info success")

	return nil
}

// listProcessesToSync lists the stopped v2 processes whose runtime state still needs to be confirmed from GSE.
func (act *actionSyncStoppedPluginProcessV2Info) listProcessesToSync(
	nCtx contextx.IContext, hostIDs []int64) ([]*types.Process, error) {

	processes, _, err := act.processStg.ListProcessesV2(nCtx, types.UnlimitedPage(), hostIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list v2 processes: %w", err)
	}

	stoppedProcesses := make([]*types.Process, 0, len(processes))
	for _, proc := range processes {
		if proc.Info.Status == types.ProcessStatusStopped {
			stoppedProcesses = append(stoppedProcesses, proc)
		}
	}

	return stoppedProcesses, nil
}

// splitProcessesByAgentState splits the stopped v2 processes by whether their agent is reported as running.
func (act *actionSyncStoppedPluginProcessV2Info) splitProcessesByAgentState(
	nCtx contextx.IContext, hostIDAgentIDMap map[int64]string, processes []*types.Process) (
	[]*types.Process, []*types.Process, error) {

	agentIDs := conv.SliceToSlice(processes, func(proc *types.Process) string {
		return hostIDAgentIDMap[proc.HostID]
	})
	result, err := batchexecutor.Collect(nCtx, conv.SliceUnique(agentIDs),
		func(nCtx contextx.IContext, batchAgentIDs []string) ([]*types.AgentState, error) {
			return act.gseHandler.ListAgentState(nCtx, batchAgentIDs...)
		},
		batchexecutor.WithBatchSize(gse.ListAgentStatePageSize),
		batchexecutor.WithTimeout(act.Timeout()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list agent states for stopped v2 processes: %w", err)
	}

	agentStates, err := conv.SliceToMap(result.Items, func(state *types.AgentState) string { return state.AgentID })
	if err != nil {
		return nil, nil, fmt.Errorf("failed to map agent states by agent id: %w", err)
	}

	agentAliveProcesses := make([]*types.Process, 0, len(processes))
	agentNotAliveProcesses := make([]*types.Process, 0)
	for _, proc := range processes {
		state, ok := agentStates[hostIDAgentIDMap[proc.HostID]]
		if !ok || state.NodeStatus == types.NodeStatusRunning {
			agentAliveProcesses = append(agentAliveProcesses, proc)

			continue
		}

		agentNotAliveProcesses = append(agentNotAliveProcesses, proc)
	}

	return agentAliveProcesses, agentNotAliveProcesses, nil
}

// queryProcessInfoDeltas retains confirmed process info and reports failed or unfinished queries separately.
func (act *actionSyncStoppedPluginProcessV2Info) queryProcessInfoDeltas(
	nCtx contextx.IContext,
	hostIDAgentIDMap map[int64]string,
	agentIDHostIDMap map[string]int64,
	processes []*types.Process,
) ([]*types.ProcessInfoDelta, error) {

	if len(processes) == 0 {
		return nil, nil
	}

	pluginNameAgentIDListMap := make(map[string][]string)
	pluginNameProcessNameMap := make(map[string]string)
	for _, proc := range processes {
		pluginNameAgentIDListMap[proc.PluginName] = append(pluginNameAgentIDListMap[proc.PluginName],
			hostIDAgentIDMap[proc.HostID])

		// In node manager v2 the process name is the plugin name, the identity name is only used as a fallback
		// source when the record does not carry it.
		processName := proc.Identity.Name
		if processName == "" {
			processName = proc.PluginName
		}
		pluginNameProcessNameMap[proc.PluginName] = processName
	}

	agentGroups := make([]*types.ProcessAgentGroup, 0, len(pluginNameAgentIDListMap))
	for pluginName, agentIDList := range pluginNameAgentIDListMap {
		agentGroups = append(agentGroups, &types.ProcessAgentGroup{
			PluginName:  pluginName,
			ProcessName: pluginNameProcessNameMap[pluginName],
			AgentIDList: agentIDList,
		})
	}

	results := act.gseHandler.NewHandlerProc(gse.WithProcNameSpace(v2PluginProcNameSpace)).
		QueryMultiProcessInfoManyStream(nCtx, agentGroups...)

	return collectProcessInfoDeltas(nCtx, results, agentIDHostIDMap, func(deltas ...*types.ProcessInfoDelta) error {
		return act.processStg.UpdateManyProcessInfoV2(nCtx, deltas)
	})
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionSyncStoppedPluginProcessV2Info) DisplayNameZh() string {
	return "同步已停止 V2 插件进程信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionSyncStoppedPluginProcessV2Info) DisplayNameEn() string {
	return "Sync Stopped V2 Plugin Process Info"
}
