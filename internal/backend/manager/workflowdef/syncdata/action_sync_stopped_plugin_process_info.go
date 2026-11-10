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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncStoppedPluginProcessInfo defines the action name.
	ActionNameSyncStoppedPluginProcessInfo = "sync_stopped_plugin_process_info"

	// stoppedProcessSyncListMaxPageSize defines the page size when listing stopped processes.
	stoppedProcessSyncListMaxPageSize = 500
)

// NewActionSyncStoppedPluginProcessInfo creates a new sync stopped plugin process info action.
func NewActionSyncStoppedPluginProcessInfo(capability *Capability) action.Definition {
	return &actionSyncStoppedPluginProcessInfo{
		gseHandler: capability.GSEHandler,
		hostStg:    capability.StorageTopo,
		processStg: capability.StoragePlugin,
	}
}

// ActionParamSyncStoppedPluginProcessInfo the action's param.
type ActionParamSyncStoppedPluginProcessInfo struct {
	syncDataUtils.SyncDataActionStandardParam

	HostIDs []int64 `json:"host_ids"`
}

type actionSyncStoppedPluginProcessInfo struct {
	gseHandler gse.IHandler
	hostStg    topoStg.IStorageHost
	processStg plugin.IDaoProcess
}

// Name returns the name of the action.
func (act *actionSyncStoppedPluginProcessInfo) Name() string {
	return ActionNameSyncStoppedPluginProcessInfo
}

// Version returns the version of the action.
func (act *actionSyncStoppedPluginProcessInfo) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncStoppedPluginProcessInfo) Description() string {
	return "sync stopped process state from gse"
}

// Timeout returns the timeout of this action.
func (act *actionSyncStoppedPluginProcessInfo) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionSyncStoppedPluginProcessInfo) MaxRetryCount() uint {
	return 2 // nolint: mnd
}

// DelayFn returns the delay of this action.
func (act *actionSyncStoppedPluginProcessInfo) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Tags returns the tags of this action.
func (act *actionSyncStoppedPluginProcessInfo) Tags() []action.Tag {
	return []action.Tag{}
}

// Do the action.
func (act *actionSyncStoppedPluginProcessInfo) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamSyncStoppedPluginProcessInfo)
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
		logger.G.Sys().Ctx(std.Context()).With("host-ids", param.HostIDs).Info("no hosts found to sync stopped process info")

		return nil
	}

	stoppedProcesses, err := act.listProcessesToSync(std.Context(), param.HostIDs)
	if err != nil {
		return err
	}

	if len(stoppedProcesses) == 0 {
		logger.G.Sys().Ctx(std.Context()).
			With("host-ids-count", len(param.HostIDs)).
			Info("no stopped plugin process found to sync")

		return nil
	}

	hostIDAgentIDMap := make(map[int64]string, len(hosts))
	agentIDHostIDMap := make(map[string]int64, len(hosts))
	for _, host := range hosts {
		hostIDAgentIDMap[host.HostID] = host.Dynamic.AgentID
		agentIDHostIDMap[host.Dynamic.AgentID] = host.HostID
	}

	agentAliveProcesses, agentNotAliveProcesses, err := act.splitProcessesByAgentState(std.Context(), hostIDAgentIDMap, stoppedProcesses)
	if err != nil {
		return err
	}

	processInfoDeltas, err := act.queryProcessInfoDeltas(std.Context(), hostIDAgentIDMap, agentIDHostIDMap, agentAliveProcesses)
	if err != nil {
		return err
	}

	// The recorded stopped state of these processes can no longer be confirmed, so it is downgraded to unknown.
	agentNotAliveDeltas := conv.SliceToSlice(agentNotAliveProcesses, func(proc *types.Process) *types.ProcessInfoDelta {
		return newUnknownProcessInfoDelta(proc, hostIDAgentIDMap[proc.HostID])
	})
	processInfoDeltas = slices.Concat(processInfoDeltas, agentNotAliveDeltas)

	if len(processInfoDeltas) == 0 {
		logger.G.Sys().Ctx(std.Context()).
			With("stopped-proc-count", len(stoppedProcesses)).
			Info("no stopped plugin process info needs update")

		return nil
	}

	if err = batchHandleProcessInfoDeltas(std.Context(), processInfoDeltas,
		func(processInfoDeltas ...*types.ProcessInfoDelta) error {
			return act.processStg.UpdateManyProcessInfo(std.Context(), processInfoDeltas)
		}); err != nil {
		return err
	}

	logger.G.Sys().Ctx(std.Context()).
		With("need-update-proc-infos", len(processInfoDeltas)).
		Info("sync stopped plugin process info success")

	return nil
}

// listProcessesToSync lists the stopped processes whose runtime state still needs to be confirmed from GSE.
func (act *actionSyncStoppedPluginProcessInfo) listProcessesToSync(nCtx contextx.IContext, hostIDs []int64) ([]*types.Process, error) {
	executor := pageexecutor.NewPageExecutor[*types.Process](stoppedProcessSyncListMaxPageSize, act.Timeout())
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Process, error) {
		processes, _, err := act.processStg.ListProcesses(nCtx, p, &types.ProcessCondition{
			ExactInclude: &types.ProcessExactFields{
				HostID:     hostIDs,
				InfoStatus: []types.ProcessStatus{types.ProcessStatusStopped},
			},
		})

		return processes, err
	}

	result, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return nil, fmt.Errorf("failed to list stopped processes: %w", err)
	}

	return result.Items, nil
}

// splitProcessesByAgentState splits the stopped processes by whether their agent is reported as running.
func (act *actionSyncStoppedPluginProcessInfo) splitProcessesByAgentState(
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
		return nil, nil, fmt.Errorf("failed to list agent states for stopped processes: %w", err)
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

// queryProcessInfoDeltas queries the process info of the given stopped processes from GSE. A process that GSE no longer
// reports is marked unknown, consistent with the alive sync: an absent record cannot confirm the recorded stopped state.
func (act *actionSyncStoppedPluginProcessInfo) queryProcessInfoDeltas(
	nCtx contextx.IContext,
	hostIDAgentIDMap map[int64]string,
	agentIDHostIDMap map[string]int64,
	processes []*types.Process,
) ([]*types.ProcessInfoDelta, error) {

	pluginNameAgentIDListMap := make(map[string][]string)
	pluginNameProcessNameMap := make(map[string]string)
	for _, proc := range processes {
		pluginNameAgentIDListMap[proc.PluginName] = append(pluginNameAgentIDListMap[proc.PluginName], hostIDAgentIDMap[proc.HostID])
		pluginNameProcessNameMap[proc.PluginName] = proc.Identity.Name
	}

	agentGroups := make([]*types.ProcessAgentGroup, 0, len(pluginNameAgentIDListMap))
	for pluginName, agentIDList := range pluginNameAgentIDListMap {
		agentGroups = append(agentGroups, &types.ProcessAgentGroup{
			PluginName:  pluginName,
			ProcessName: pluginNameProcessNameMap[pluginName],
			AgentIDList: agentIDList,
		})
	}

	procInfos, err := act.gseHandler.NewHandlerProc().QueryMultiProcessInfoMany(nCtx, agentGroups...)
	if err != nil {
		return nil, fmt.Errorf("failed to query stopped plugin process info from gse: %w", err)
	}

	underControlledDeltas := make([]*types.ProcessInfoDelta, 0)
	underControlledProcessMap := make(map[types.ProcessUniqueKey]struct{})
	for pluginName, infos := range procInfos {
		for idx := range infos {
			delta := &types.ProcessInfoDelta{
				HostID:      agentIDHostIDMap[infos[idx].AgentID],
				PluginName:  pluginName,
				ProcessInfo: infos[idx],
			}

			underControlledProcessMap[delta.GetUniqueKey()] = struct{}{}
			underControlledDeltas = append(underControlledDeltas, delta)
		}
	}

	lostControlledDeltas := make([]*types.ProcessInfoDelta, 0)
	for _, proc := range processes {
		if _, ok := underControlledProcessMap[proc.GetUniqueKey()]; ok {
			continue
		}

		lostControlledDeltas = append(lostControlledDeltas, newUnknownProcessInfoDelta(proc, hostIDAgentIDMap[proc.HostID]))
	}

	return slices.Concat(underControlledDeltas, lostControlledDeltas), nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionSyncStoppedPluginProcessInfo) DisplayNameZh() string {
	return "同步已停止插件进程信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionSyncStoppedPluginProcessInfo) DisplayNameEn() string {
	return "Sync Stopped Plugin Process Info"
}
