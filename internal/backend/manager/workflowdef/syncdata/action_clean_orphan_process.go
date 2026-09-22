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
	"time"

	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	// ActionNameCleanOrphanProcess defines the action name.
	ActionNameCleanOrphanProcess = "clean_orphan_process"
)

// NewActionCleanOrphanProcess creates a new clean orphan process action.
func NewActionCleanOrphanProcess(capability *Capability) action.Definition {
	return &actionCleanOrphanProcess{
		processStg:  capability.StoragePlugin,
		hostStg:     capability.StorageTopo,
		cmdbHandler: capability.CMDBHandler,
		workflowCtl: capability.WorkflowCtl,
	}
}

// ActionParamCleanOrphanProcess defines the action's param.
type ActionParamCleanOrphanProcess struct {
	syncDataUtils.SyncDataActionStandardParam
	Processes []*types.Process `json:"processes"`
}

type actionCleanOrphanProcess struct {
	processStg  plugin.IDaoProcess
	hostStg     topoStg.IStorageHost
	cmdbHandler cmdb.IHandler
	workflowCtl workflow.IController
}

// Name returns the name of the action.
func (act *actionCleanOrphanProcess) Name() string {
	return ActionNameCleanOrphanProcess
}

// Version returns the version of the action.
func (act *actionCleanOrphanProcess) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionCleanOrphanProcess) Description() string {
	return "delete the processes whose host no longer exists"
}

// Timeout returns the timeout of the action.
func (act *actionCleanOrphanProcess) Timeout() time.Duration {
	return 30 * time.Minute // nolint: mnd
}

// MaxRetryCount returns 0 because the action deletes in batches and the whole run
// must not be retried automatically.
func (act *actionCleanOrphanProcess) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionCleanOrphanProcess) DelayFn(_ int) func() {
	return func() {}
}

// Tags returns the tags of the action.
func (act *actionCleanOrphanProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// Do this func define what the action will do.
func (act *actionCleanOrphanProcess) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamCleanOrphanProcess)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return err
	}

	std := syncDataUtils.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	if err := act.cleanOrphanProcesses(std, param.Processes); err != nil {
		return fmt.Errorf("failed to clean orphan processes: %w", err)
	}

	return nil
}

// cleanOrphanProcesses deletes supplied processes with missing hosts and schedules correction for existing hosts.
func (act *actionCleanOrphanProcess) cleanOrphanProcesses(std *syncDataUtils.SyncDataActionStandarder, processes []*types.Process) error {
	if len(processes) == 0 {
		return nil
	}

	hostIDs := conv.SliceUnique(conv.SliceToSlice(processes, func(process *types.Process) int64 {
		return process.HostID
	}))
	missingHostIDs, err := act.findMissingHostIDs(std.Context(), hostIDs)
	if err != nil {
		return err
	}
	missingHosts := make(map[int64]struct{}, len(missingHostIDs))
	for _, hostID := range missingHostIDs {
		missingHosts[hostID] = struct{}{}
	}

	existingHostIDs := make([]int64, 0, len(hostIDs))
	for _, hostID := range hostIDs {
		if _, missing := missingHosts[hostID]; missing {
			continue
		}
		existingHostIDs = append(existingHostIDs, hostID)
	}
	if err := act.createCorrectionOperation(std, existingHostIDs); err != nil {
		return err
	}

	targets := make([]*types.Process, 0, len(processes))
	for _, process := range processes {
		if _, missing := missingHosts[process.HostID]; !missing {
			continue
		}
		targets = append(targets, process)
	}

	deletedCount := 0
	defer func() {
		std.InstanceData().Log().
			Zh("孤立进程清理结果: 输入 %d 个, 清理目标 %d 个, 已删除 %d 个", len(processes), len(targets), deletedCount).
			En("orphan process cleanup result: input %d, targets %d, deleted %d", len(processes), len(targets), deletedCount).
			Info()
	}()

	for _, process := range targets {
		if err := act.processStg.DeleteProcess(std.Context(), process.HostID, process.PluginName); err != nil {
			return fmt.Errorf("failed to delete process record, host-id(%d), plugin-name(%s): %w", process.HostID, process.PluginName, err)
		}
		deletedCount++
		std.InstanceData().Log().
			Zh("已删除孤立进程, 主机 ID %d, 插件 %s", process.HostID, process.PluginName).
			En("deleted orphan process, host id %d, plugin %s", process.HostID, process.PluginName).
			Info()
	}

	return nil
}

func (act *actionCleanOrphanProcess) createCorrectionOperation(
	std *syncDataUtils.SyncDataActionStandarder, hostIDs []int64) error {

	if len(hostIDs) == 0 {
		return nil
	}
	metadata := trigger.NewMetadataOrdered(1)
	metadata.CleanPolicy = trigger.MetadataCleanPolicy{MaxDays: syncDataCleanPolicyMaxDays(act.Timeout())}
	trigCtl, err := act.workflowCtl.CreateTrigger(std.Context(), trigger.CategoryOrdered, metadata)
	if err != nil {
		return fmt.Errorf("failed to create process status correction trigger: %w", err)
	}
	operDef := NewOperCorrectUnknownProcessStatus(OperParamCorrectUnknownProcessStatus{
		TenantID: std.TenantID(), Operator: std.Operator(), HostIDs: hostIDs,
	})
	operParam := operDef.DefaultParameters()
	operParam.ParentOperationID = std.InstanceData().OperationID
	operParam.ParentOperInstID = std.InstanceData().OperationInstanceID
	if _, err := trigCtl.CreateOperation(std.Context(), operDef, operParam); err != nil {
		return fmt.Errorf("failed to create process status correction operation: %w", err)
	}
	if err := trigCtl.ActivateTrigger(std.Context()); err != nil {
		return fmt.Errorf("failed to activate process status correction trigger: %w", err)
	}

	return nil
}

func (act *actionCleanOrphanProcess) findMissingHostIDs(nCtx contextx.IContext, hostIDs []int64) ([]int64, error) {
	executor := pageexecutor.NewPageExecutor[*types.Host](cleanOrphanProcessBatchSize, act.Timeout())
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		hosts, _, err := act.hostStg.ListHost(nCtx, p, &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{HostID: hostIDs},
		})

		return hosts, err
	}
	hosts, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return nil, fmt.Errorf("failed to query hosts for orphan process cleanup: %w", err)
	}

	aliveHostIDs := make(map[int64]struct{}, len(hosts.Items))
	for _, aliveHost := range hosts.Items {
		aliveHostIDs[aliveHost.HostID] = struct{}{}
	}

	missingHostIDs := make([]int64, 0, len(hostIDs))
	for _, hostID := range hostIDs {
		if _, ok := aliveHostIDs[hostID]; ok {
			continue
		}

		missingHostIDs = append(missingHostIDs, hostID)
	}

	if len(missingHostIDs) == 0 {
		return nil, nil
	}

	cmdbHosts, err := act.cmdbHandler.FindHostWithCondition(nCtx, types.UnlimitedPage(), &types.HostStaticExactCondition{
		StaticExactInclude: &types.HostStaticExactFields{HostID: missingHostIDs},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query CMDB hosts for orphan process cleanup: %w", err)
	}
	for _, host := range cmdbHosts {
		aliveHostIDs[host.HostID] = struct{}{}
	}

	orphanHostIDs := make([]int64, 0, len(missingHostIDs))
	for _, hostID := range missingHostIDs {
		if _, ok := aliveHostIDs[hostID]; ok {
			continue
		}
		orphanHostIDs = append(orphanHostIDs, hostID)
	}

	return orphanHostIDs, nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionCleanOrphanProcess) DisplayNameZh() string {
	return "清理孤立进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionCleanOrphanProcess) DisplayNameEn() string {
	return "Clean Orphan Processes"
}
