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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	// ActionNameGenOperSyncStoppedPluginProcessV2Info defines the action name.
	ActionNameGenOperSyncStoppedPluginProcessV2Info = "gen_oper_sync_stopped_plugin_process_v2_info"
)

// NewActionGenOperSyncStoppedPluginProcessV2Info this action will create v2 process sync operations for all hosts.
func NewActionGenOperSyncStoppedPluginProcessV2Info(capability *Capability) action.Definition {
	return &actionGenOperSyncStoppedPluginProcessV2Info{
		topoStg:     capability.StorageTopo,
		workflowCtl: capability.WorkflowCtl,
	}
}

// ActionParamGenOperSyncStoppedPluginProcessV2Info defines the action's param.
type ActionParamGenOperSyncStoppedPluginProcessV2Info struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionGenOperSyncStoppedPluginProcessV2Info struct {
	topoStg     topoStg.IStorageHost
	workflowCtl workflow.IController
}

// Name returns the name of the action.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) Name() string {
	return ActionNameGenOperSyncStoppedPluginProcessV2Info
}

// Version returns the version of the action.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) Description() string {
	return "create sync stopped v2 plugin process status operation for all hosts"
}

// Timeout returns the timeout of the action.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) Timeout() time.Duration {
	return time.Minute * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) Tags() []action.Tag {
	return []action.Tag{}
}

// Queue returns the queue of the action.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) Queue() action.Queue {
	return action.QueueDefault
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) DelayFn(_ int) func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamGenOperSyncStoppedPluginProcessV2Info)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	// The v2 process sync reuses the globally configured concurrency of the v3 stopped process sync,
	// so the v2 task adds no new setting surface.
	maxConcurrencyNum, err := conv.ToInt64(globalsettings.Get(
		std.Context(),
		globalsettings.OperSyncStoppedPluginProcessInfoMaxConcurrencyNum,
		globalsettings.OperSyncStoppedPluginProcessInfoMaxConcurrencyNumDefault,
	))
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w",
			globalsettings.OperSyncStoppedPluginProcessInfoMaxConcurrencyNum, err)
	}
	if maxConcurrencyNum <= 0 {
		return fmt.Errorf("%s must be positive, got %d",
			globalsettings.OperSyncStoppedPluginProcessInfoMaxConcurrencyNum, maxConcurrencyNum)
	}

	scanCtx, cancel := contextx.WithTimeout(std.Context(), 1*time.Hour)
	defer cancel()

	hosts, err := act.topoStg.ScanAllHostWithFields(scanCtx, &types.HostFieldSelection{
		HostID: true,
	}, &types.HostCondition{
		DynamicAgentIDNotEmpty: true,
	})
	if err != nil {
		return err
	}

	var trigCtl workflow.ITriggerCtl
	if err = batchexecutor.Execute(scanCtx, hosts, func(nCtx contextx.IContext, batchHosts []*types.Host) error {
		hostIDs := make([]int64, 0, len(batchHosts))
		for _, host := range batchHosts {
			hostIDs = append(hostIDs, host.HostID)
		}

		if trigCtl == nil {
			// create trigger for handling sync stopped v2 plugin process info operations.
			meta := trigger.NewMetadataOrdered(int(maxConcurrencyNum))
			meta.CleanPolicy = trigger.MetadataCleanPolicy{
				MaxDays: syncDataCleanPolicyMaxDays(OperDefNameSyncStoppedPluginProcessV2InfoTimeout),
			}
			trigCtl, err = act.workflowCtl.CreateTrigger(nCtx, trigger.CategoryOrdered, meta)
			if err != nil {
				logger.G.Sys().Ctx(nCtx).
					WithErr(err).
					With("action", act.Name()).
					Error("failed to create trigger for handling sync stopped v2 plugin process info operations")

				return err
			}
		}

		return act.executeOper(std, trigCtl, hostIDs...)
	}, batchexecutor.WithBatchSize(syncStoppedPluginProcessStatusMaxPageSize), batchexecutor.WithTimeout(1*time.Hour)); err != nil {
		return err
	}

	if trigCtl != nil {
		if err = trigCtl.ActivateTrigger(std.Context()); err != nil {
			logger.G.Sys().Ctx(std.Context()).
				WithErr(err).
				With("action", act.Name()).
				Error("failed to run trigger for handling sync stopped v2 plugin process info operations")

			return err
		}
	}

	logger.G.Sys().Ctx(std.Context()).
		With("action", act.Name()).
		Info("executed sync stopped v2 plugin process info operation for %d hosts", len(hosts))

	return nil
}

// executeOper create an operation to sync stopped v2 plugin process info and then execute it.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) executeOper(
	std *syncDataUtils.SyncDataActionStandarder, trigCtl workflow.ITriggerCtl, hostIDs ...int64) error {

	operationDef := NewOperSyncStoppedPluginProcessV2Info(OperParamSyncStoppedPluginProcessV2Info{
		TenantID: std.TenantID(),
		Operator: std.Operator(),
		HostIDs:  hostIDs,
	})

	operationParam := operationDef.DefaultParameters()
	operationParam.ParentOperationID = std.InstanceData().OperationID
	operationParam.ParentOperInstID = std.InstanceData().OperationInstanceID

	operCtl, err := trigCtl.CreateOperation(std.Context(), operationDef, operationParam)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).
			WithErr(err).
			With("action", act.Name(), "tenant-id", std.TenantID()).
			Error("failed to create sync stopped v2 plugin process info operation")

		return err
	}

	logger.G.Sys().Ctx(std.Context()).
		With("action", act.Name(), "tenant-id", std.TenantID(), "operation-id", operCtl.GetOperationID()).
		Info("created sync stopped v2 plugin process info operation")

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) DisplayNameZh() string {
	return "生成同步已停止 V2 插件进程信息任务"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionGenOperSyncStoppedPluginProcessV2Info) DisplayNameEn() string {
	return "Generate Sync Stopped V2 Plugin Process Info Operation"
}
