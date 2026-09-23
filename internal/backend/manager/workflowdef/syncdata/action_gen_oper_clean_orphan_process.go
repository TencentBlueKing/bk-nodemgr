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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	// ActionNameGenOperCleanOrphanProcess defines the action name.
	ActionNameGenOperCleanOrphanProcess = "gen_oper_clean_orphan_process"

	// cleanOrphanProcessBatchSize defines the batch size for generating cleanup operations.
	cleanOrphanProcessBatchSize = 500
	// Keep recently synced processes out of orphan cleanup.
	cleanOrphanProcessStaleAfter = 48 * time.Hour
)

// NewActionGenOperCleanOrphanProcess creates the stale process scan action.
func NewActionGenOperCleanOrphanProcess(capability *Capability) action.Definition {
	return &actionGenOperCleanOrphanProcess{
		processStg:  capability.StoragePlugin,
		workflowCtl: capability.WorkflowCtl,
	}
}

// ActionParamGenOperCleanOrphanProcess defines the generator parameters.
type ActionParamGenOperCleanOrphanProcess struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionGenOperCleanOrphanProcess struct {
	processStg  plugin.IDaoProcess
	workflowCtl workflow.IController
}

// Name returns the action name.
func (act *actionGenOperCleanOrphanProcess) Name() string {
	return ActionNameGenOperCleanOrphanProcess
}

// Version returns the action version.
func (act *actionGenOperCleanOrphanProcess) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the action description.
func (act *actionGenOperCleanOrphanProcess) Description() string {
	return "generate orphan process cleanup operations for stale processes"
}

// Timeout returns the scan timeout.
func (act *actionGenOperCleanOrphanProcess) Timeout() time.Duration {
	return 30 * time.Minute // nolint: mnd
}

// Tags returns the action tags.
func (act *actionGenOperCleanOrphanProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount disables automatic retries after partially generating operations.
func (act *actionGenOperCleanOrphanProcess) MaxRetryCount() uint {
	return 0
}

// DelayFn returns the retry delay function.
func (act *actionGenOperCleanOrphanProcess) DelayFn(_ int) func() {
	return func() {}
}

// Do scans up to the configured number of stale processes and generates cleanup operations in batches.
func (act *actionGenOperCleanOrphanProcess) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamGenOperCleanOrphanProcess)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to decode orphan process generator parameters: %w", err)
	}
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	pageSize, err := conv.ToInt64(globalsettings.Get(std.Context(), globalsettings.OperCleanOrphanProcessPageSize,
		globalsettings.OperCleanOrphanProcessPageSizeDefault))
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", globalsettings.OperCleanOrphanProcessPageSize, err)
	}
	if pageSize <= 0 {
		return fmt.Errorf("%s must be positive, got %d", globalsettings.OperCleanOrphanProcessPageSize, pageSize)
	}

	deadline := time.Now().Add(-cleanOrphanProcessStaleAfter)
	condition := &types.ProcessCondition{
		ExactInclude: &types.ProcessExactFields{
			InfoLastSyncAtBefore: &deadline,
		},
	}
	processes, err := act.processStg.ScanProcesses(std.Context(), pageSize, condition)
	if err != nil {
		return fmt.Errorf("failed to scan stale processes: %w", err)
	}
	if len(processes) == 0 {
		std.InstanceData().Log().
			Zh("未找到过期进程").
			En("no stale processes found").
			Info()

		return nil
	}

	var trigCtl workflow.ITriggerCtl
	if err := batchexecutor.Execute(std.Context(), processes, func(nCtx contextx.IContext, batchProcesses []*types.Process) error {
		if trigCtl == nil {
			meta := trigger.NewMetadataOrdered(1)
			meta.CleanPolicy = trigger.MetadataCleanPolicy{
				MaxDays: syncDataCleanPolicyMaxDays(act.Timeout()),
			}
			trigCtl, err = act.workflowCtl.CreateTrigger(nCtx, trigger.CategoryOrdered, meta)
			if err != nil {
				return fmt.Errorf("failed to create orphan process cleanup trigger: %w", err)
			}
		}

		return act.executeOper(nCtx, std, trigCtl, batchProcesses)
	}, batchexecutor.WithBatchSize(cleanOrphanProcessBatchSize), batchexecutor.WithTimeout(act.Timeout())); err != nil {
		return err
	}

	if trigCtl == nil {
		return nil
	}
	if err := trigCtl.ActivateTrigger(std.Context()); err != nil {
		return fmt.Errorf("failed to activate orphan process cleanup trigger: %w", err)
	}

	return nil
}

func (act *actionGenOperCleanOrphanProcess) executeOper(
	nCtx contextx.IContext, std *syncDataUtils.SyncDataActionStandarder, trigCtl workflow.ITriggerCtl, processes []*types.Process) error {

	operationDef := NewOperCleanOrphanProcess(OperParamCleanOrphanProcess{
		TenantID:  std.TenantID(),
		Operator:  std.Operator(),
		Processes: processes,
	})
	operationParam := operationDef.DefaultParameters()
	operationParam.ParentOperationID = std.InstanceData().OperationID
	operationParam.ParentOperInstID = std.InstanceData().OperationInstanceID

	operCtl, err := trigCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		return fmt.Errorf("failed to create orphan process cleanup operation: %w", err)
	}
	std.InstanceData().Log().
		Zh("已为 %d 个过期进程生成清理任务 %s", len(processes), operCtl.GetOperationID()).
		En("created cleanup operation %s for %d stale processes", operCtl.GetOperationID(), len(processes)).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name.
func (act *actionGenOperCleanOrphanProcess) DisplayNameZh() string {
	return "生成孤立进程清理任务"
}

// DisplayNameEn returns the English display name.
func (act *actionGenOperCleanOrphanProcess) DisplayNameEn() string {
	return "Generate Orphan Process Cleanup Operations"
}
