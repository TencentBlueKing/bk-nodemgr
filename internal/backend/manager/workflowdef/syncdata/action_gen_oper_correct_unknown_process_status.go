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
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// ActionNameGenOperCorrectUnknownProcessStatus defines the action name.
const ActionNameGenOperCorrectUnknownProcessStatus = "gen_oper_correct_unknown_process_status"

// correctUnknownProcessStatusMaxPageSize defines the host batch size for process status correction operations.
const correctUnknownProcessStatusMaxPageSize = 1000

// NewActionGenOperCorrectUnknownProcessStatus creates the action that distributes process expiry checks.
func NewActionGenOperCorrectUnknownProcessStatus(capability *Capability) action.Definition {
	return &actionGenOperCorrectUnknownProcessStatus{
		topoStg:     capability.StorageTopo,
		workflowCtl: capability.WorkflowCtl,
	}
}

// ActionParamGenOperCorrectUnknownProcessStatus defines the generation parameters.
type ActionParamGenOperCorrectUnknownProcessStatus struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionGenOperCorrectUnknownProcessStatus struct {
	topoStg     topoStg.IStorageHost
	workflowCtl workflow.IController
}

// Name returns the action name.
func (act *actionGenOperCorrectUnknownProcessStatus) Name() string {
	return ActionNameGenOperCorrectUnknownProcessStatus
}

// Version returns the action version.
func (act *actionGenOperCorrectUnknownProcessStatus) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the action description.
func (act *actionGenOperCorrectUnknownProcessStatus) Description() string {
	return "create process expiry correction operations for all hosts"
}

// Timeout returns the action timeout.
func (act *actionGenOperCorrectUnknownProcessStatus) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// MaxRetryCount disables retries to avoid creating duplicate child operations.
func (act *actionGenOperCorrectUnknownProcessStatus) MaxRetryCount() uint {
	return 0
}

// DelayFn returns the retry delay function.
func (act *actionGenOperCorrectUnknownProcessStatus) DelayFn(_ int) func() {
	return func() {}
}

// Tags returns the action tags.
func (act *actionGenOperCorrectUnknownProcessStatus) Tags() []action.Tag {
	return []action.Tag{}
}

// Do creates one correction operation per host batch, including hosts without an AgentID.
func (act *actionGenOperCorrectUnknownProcessStatus) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamGenOperCorrectUnknownProcessStatus)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to parse correct unknown process status parameters: %w", err)
	}
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return fmt.Errorf("failed to initialize sync data action context: %w", err)
	}

	concurrencyValue := globalsettings.Get(std.Context(), globalsettings.OperCorrectUnknownProcessStatusMaxConcurrencyNum,
		globalsettings.OperCorrectUnknownProcessStatusMaxConcurrencyNumDefault)
	maxConcurrencyNum, err := conv.ToInt64(concurrencyValue)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", globalsettings.OperCorrectUnknownProcessStatusMaxConcurrencyNum, err)
	}
	if maxConcurrencyNum <= 0 {
		return fmt.Errorf("%s must be positive, got %d",
			globalsettings.OperCorrectUnknownProcessStatusMaxConcurrencyNum, maxConcurrencyNum)
	}

	scanCtx, cancel := contextx.WithTimeout(std.Context(), act.Timeout())
	defer cancel()
	hosts, err := act.topoStg.ScanAllHostWithFields(scanCtx, &types.HostFieldSelection{HostID: true}, &types.HostCondition{})
	if err != nil {
		return fmt.Errorf("failed to scan hosts for correct unknown process status operations: %w", err)
	}
	if len(hosts) == 0 {
		return nil
	}

	metadata := trigger.NewMetadataOrdered(int(maxConcurrencyNum))
	metadata.CleanPolicy = trigger.MetadataCleanPolicy{MaxDays: syncDataCleanPolicyMaxDays(act.Timeout())}
	triggerCtl, err := act.workflowCtl.CreateTrigger(scanCtx, trigger.CategoryOrdered, metadata)
	if err != nil {
		return fmt.Errorf("failed to create trigger for handling correct unknown process status operations: %w", err)
	}
	err = batchexecutor.Execute(scanCtx, hosts, func(nCtx contextx.IContext, batchHosts []*types.Host) error {
		hostIDs := conv.SliceToSlice(batchHosts, func(host *types.Host) int64 { return host.HostID })
		operDef := NewOperCorrectUnknownProcessStatus(OperParamCorrectUnknownProcessStatus{
			TenantID: std.TenantID(), Operator: std.Operator(), HostIDs: hostIDs,
		})
		operParam := operDef.DefaultParameters()
		operParam.ParentOperationID = std.InstanceData().OperationID
		operParam.ParentOperInstID = std.InstanceData().OperationInstanceID
		if _, createErr := triggerCtl.CreateOperation(nCtx, operDef, operParam); createErr != nil {
			return fmt.Errorf("failed to create correct unknown process status operation for hosts %v: %w", hostIDs, createErr)
		}

		return nil
	}, batchexecutor.WithBatchSize(correctUnknownProcessStatusMaxPageSize), batchexecutor.WithTimeout(act.Timeout()))
	if err != nil {
		return fmt.Errorf("failed to generate correct unknown process status operations: %w", err)
	}
	if err := triggerCtl.ActivateTrigger(scanCtx); err != nil {
		return fmt.Errorf("failed to activate correct unknown process status trigger %s: %w", triggerCtl.GetTriggerID(), err)
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionGenOperCorrectUnknownProcessStatus) DisplayNameZh() string {
	return "生成修正过期进程状态任务"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionGenOperCorrectUnknownProcessStatus) DisplayNameEn() string {
	return "Generate Correct Expired Process Status Operation"
}
