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
	// ActionNameSyncAlivePluginProcessInfo defines the action name.
	ActionNameSyncAlivePluginProcessInfo = "sync_alive_plugin_process_info"

	processSyncListMaxPageSize = 500
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
func (act *actionSyncAlivePluginProcessInfo) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
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

	aliveProcess, err := act.listProcessesToSync(std.Context(), param.HostIDs)
	if err != nil {
		return err
	}

	if len(aliveProcess) > 0 {
		hostIDAgentIDMap := make(map[int64]string, len(hosts))
		agentIDHostIDMap := make(map[string]int64, len(hosts))
		for _, host := range hosts {
			hostIDAgentIDMap[host.HostID] = host.Dynamic.AgentID
			agentIDHostIDMap[host.Dynamic.AgentID] = host.HostID
		}

		agentAliveProcess, agentNotAliveProcess, err := act.splitProcessesByAgentState(std.Context(), hostIDAgentIDMap, aliveProcess)
		if err != nil {
			return err
		}

		needUpdateProcInfos, err := act.checkAliveProcess(std.Context(), hostIDAgentIDMap, agentIDHostIDMap, agentAliveProcess)
		if err != nil {
			return err
		}

		agentNotAliveProcInfos := conv.SliceToSlice(agentNotAliveProcess, func(proc *types.Process) *types.ProcessInfoDelta {
			return newUnknownProcessInfoDelta(proc, hostIDAgentIDMap[proc.HostID])
		})
		needUpdateProcInfos = slices.Concat(needUpdateProcInfos, agentNotAliveProcInfos)

		if err = batchHandleProcessInfoDeltas(std.Context(), needUpdateProcInfos, func(processInfoDeltas ...*types.ProcessInfoDelta) error {
			return act.processStg.UpdateManyProcessInfo(std.Context(), processInfoDeltas)
		}); err != nil {
			return err
		}

		logger.G.Sys().Ctx(std.Context()).
			With("need-update-proc-infos", len(needUpdateProcInfos)).
			Info("sync alive plugin process info success")
	}

	return nil
}

func (act *actionSyncAlivePluginProcessInfo) listProcessesToSync(nCtx contextx.IContext, hostIDs []int64) ([]*types.Process, error) {
	executor := pageexecutor.NewPageExecutor[*types.Process](processSyncListMaxPageSize, act.Timeout())
	listProcesses := func(cond *types.ProcessCondition) ([]*types.Process, error) {
		fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Process, error) {
			processes, _, err := act.processStg.ListProcesses(nCtx, p, cond)
			return processes, err
		}
		result, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
		if err != nil {
			return nil, err
		}

		return result.Items, nil
	}

	runningProcess, err := listProcesses(&types.ProcessCondition{
		ExactInclude: &types.ProcessExactFields{
			HostID:     hostIDs,
			InfoStatus: []types.ProcessStatus{types.ProcessStatusRunning},
		},
	})
	if err != nil {
		return nil, err
	}

	unknownAutoStartProcess, err := listProcesses(&types.ProcessCondition{
		ExactInclude: &types.ProcessExactFields{
			HostID:          hostIDs,
			InfoStatus:      []types.ProcessStatus{types.ProcessStatusUnknown},
			InfoTrusteeship: []bool{true},
		},
	})
	if err != nil {
		return nil, err
	}

	aliveProcess := make([]*types.Process, 0, len(runningProcess)+len(unknownAutoStartProcess))
	aliveProcessMap := make(map[types.ProcessUniqueKey]struct{}, len(runningProcess)+len(unknownAutoStartProcess))
	for _, processes := range [][]*types.Process{runningProcess, unknownAutoStartProcess} {
		for _, proc := range processes {
			if _, ok := aliveProcessMap[proc.GetUniqueKey()]; ok {
				continue
			}

			aliveProcessMap[proc.GetUniqueKey()] = struct{}{}
			aliveProcess = append(aliveProcess, proc)
		}
	}

	return aliveProcess, nil
}

func (act *actionSyncAlivePluginProcessInfo) checkAliveProcess(
	nCtx contextx.IContext, hostIDAgentIDMap map[int64]string, agentIDHostIDMap map[string]int64, aliveProcess []*types.Process) (
	[]*types.ProcessInfoDelta, error) {

	if len(aliveProcess) == 0 {
		return nil, nil
	}

	pluginNameAgentIDListMap := make(map[string][]string)
	pluginNameProcessNameMap := make(map[string]string)
	for _, proc := range aliveProcess {
		pluginName := proc.PluginName
		pluginNameAgentIDListMap[pluginName] = append(pluginNameAgentIDListMap[pluginName], hostIDAgentIDMap[proc.HostID])
		pluginNameProcessNameMap[pluginName] = proc.Identity.Name
	}

	pluginNameAgentIDList := make([]*types.ProcessAgentGroup, 0)
	for pluginName, agentIDList := range pluginNameAgentIDListMap {
		pluginNameAgentIDList = append(pluginNameAgentIDList, &types.ProcessAgentGroup{
			PluginName:  pluginName,
			ProcessName: pluginNameProcessNameMap[pluginName],
			AgentIDList: agentIDList,
		})
	}

	procInfos, err := act.gseHandler.NewHandlerProc().QueryMultiProcessInfoMany(nCtx, pluginNameAgentIDList...)
	if err != nil {
		return nil, err
	}

	underControlledProcInfos := make([]*types.ProcessInfoDelta, 0)
	for pluginName, infos := range procInfos {
		for idx := range infos {
			processInfo := infos[idx]
			hostID := agentIDHostIDMap[processInfo.AgentID]
			processInfoDelta := &types.ProcessInfoDelta{
				HostID:      hostID,
				PluginName:  pluginName,
				ProcessInfo: processInfo,
			}

			underControlledProcInfos = append(underControlledProcInfos, processInfoDelta)
		}
	}

	underControlledProcessMap, _ := conv.SliceToMap(underControlledProcInfos,
		func(v *types.ProcessInfoDelta) types.ProcessUniqueKey {
			return v.GetUniqueKey()
		})

	lostControlledProcInfos := make([]*types.ProcessInfoDelta, 0)
	for _, proc := range aliveProcess {
		if _, ok := underControlledProcessMap[proc.GetUniqueKey()]; ok {
			// proc is under controlled, skip it.
			continue
		}

		processInfoDelta := newUnknownProcessInfoDelta(proc, hostIDAgentIDMap[proc.HostID])
		lostControlledProcInfos = append(lostControlledProcInfos, processInfoDelta)
	}

	needUpdateProcInfos := slices.Concat(underControlledProcInfos, lostControlledProcInfos)

	return needUpdateProcInfos, nil
}

func (act *actionSyncAlivePluginProcessInfo) splitProcessesByAgentState(
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
		return nil, nil, fmt.Errorf("failed to list agent states for alive processes: %w", err)
	}

	agentStates, err := conv.SliceToMap(result.Items, func(state *types.AgentState) string { return state.AgentID })
	if err != nil {
		return nil, nil, fmt.Errorf("failed to map agent states by agent id: %w", err)
	}

	agentAliveProcess := make([]*types.Process, 0, len(processes))
	agentNotAliveProcess := make([]*types.Process, 0)
	for _, proc := range processes {
		agentID := hostIDAgentIDMap[proc.HostID]
		state, ok := agentStates[agentID]
		if !ok || state.NodeStatus == types.NodeStatusRunning {
			agentAliveProcess = append(agentAliveProcess, proc)
			continue
		}

		agentNotAliveProcess = append(agentNotAliveProcess, proc)
	}

	return agentAliveProcess, agentNotAliveProcess, nil
}

func newUnknownProcessInfoDelta(proc *types.Process, agentID string) *types.ProcessInfoDelta {
	return &types.ProcessInfoDelta{
		HostID:     proc.HostID,
		PluginName: proc.PluginName,
		ProcessInfo: types.ProcessInfo{
			Pid:        0,
			Version:    "",
			AgentID:    agentID,
			AutoStart:  proc.Info.AutoStart,
			Status:     types.ProcessStatusUnknown,
			LastSyncAt: time.Now(),
		},
	}
}

func batchHandleProcessInfoDeltas(nCtx contextx.IContext, processInfoDeltas []*types.ProcessInfoDelta,
	fn func(processInfoDeltas ...*types.ProcessInfoDelta) error) error {

	return batchexecutor.Execute(nCtx, processInfoDeltas,
		func(_ contextx.IContext, batchProcessInfoDeltas []*types.ProcessInfoDelta) error {
			return fn(batchProcessInfoDeltas...)
		},
		batchexecutor.WithBatchSize(syncHostDBBatchSize),
		batchexecutor.WithTimeout(10*time.Minute), // nolint: mnd
	)
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionSyncAlivePluginProcessInfo) DisplayNameZh() string {
	return "同步存活插件进程信息"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionSyncAlivePluginProcessInfo) DisplayNameEn() string {
	return "Sync Alive Plugin Process Info"
}
