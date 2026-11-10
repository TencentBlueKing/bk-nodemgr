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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/syncdataaction"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
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
	// ActionNameGenOperSyncStoppedPluginProcessInfo defines the action name.
	ActionNameGenOperSyncStoppedPluginProcessInfo = "gen_oper_sync_stopped_plugin_process_info"

	// syncStoppedPluginProcessStatusMaxPageSize defines the operation batch size.
	// In the scenario of 40,000 hosts, a single request for 1,000 hosts requires 400 table lookups.
	syncStoppedPluginProcessStatusMaxPageSize = 1000
)

// NewActionGenOperSyncStoppedPluginProcessInfo this action will create host sync operation for all business.
func NewActionGenOperSyncStoppedPluginProcessInfo(capability *Capability) action.Definition {
	return &actionGenOperSyncStoppedPluginProcessInfo{
		topoStg:     capability.StorageTopo,
		processStg:  capability.StoragePlugin,
		workflowCtl: capability.WorkflowCtl,
	}
}

// ActionParamGenOperSyncStoppedPluginProcessInfo defines the action's param.
type ActionParamGenOperSyncStoppedPluginProcessInfo struct {
	syncdataaction.SyncDataActionStandardParam
}

type actionGenOperSyncStoppedPluginProcessInfo struct {
	topoStg     topoStg.IStorageHost
	processStg  plugin.IDaoProcess
	workflowCtl workflow.IController
}

// Name returns the name of the action.
func (act *actionGenOperSyncStoppedPluginProcessInfo) Name() string {
	return ActionNameGenOperSyncStoppedPluginProcessInfo
}

// Version returns the version of the action.
func (act *actionGenOperSyncStoppedPluginProcessInfo) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenOperSyncStoppedPluginProcessInfo) Description() string {
	return "create sync stopped plugin process status operation for all hosts"
}

// Timeout returns the timeout of this action.
func (act *actionGenOperSyncStoppedPluginProcessInfo) Timeout() time.Duration {
	return time.Minute * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionGenOperSyncStoppedPluginProcessInfo) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionGenOperSyncStoppedPluginProcessInfo) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionGenOperSyncStoppedPluginProcessInfo) DelayFn(_ int) func() {
	return func() {}
}

// Do this func define what the action will do.
// nolint:gocognit
func (act *actionGenOperSyncStoppedPluginProcessInfo) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamGenOperSyncStoppedPluginProcessInfo)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return err
	}

	// initialize standard data.
	std := syncdataaction.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}
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
			// create trigger for handling sync stopped plugin process info operations.
			meta := trigger.NewMetadataOrdered(int(maxConcurrencyNum))
			meta.CleanPolicy = trigger.MetadataCleanPolicy{
				MaxDays: syncDataCleanPolicyMaxDays(OperDefNameSyncStoppedPluginProcessInfoTimeout),
			}
			trigCtl, err = act.workflowCtl.CreateTrigger(nCtx, trigger.CategoryOrdered, meta)
			if err != nil {
				logger.G.Sys().Ctx(nCtx).
					WithErr(err).
					With("action", act.Name()).
					Error("failed to create trigger for handling sync stopped plugin process info operations")

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
				Error("failed to run trigger for handling sync stopped plugin process info operations")

			return err
		}
	}

	logger.G.Sys().Ctx(std.Context()).
		With("action", act.Name()).
		Info("executed sync stopped plugin process info operation for %d hosts", len(hosts))

	return nil
}

// executeOper create an operation to sync stopped plugin process info and then execute it.
func (act *actionGenOperSyncStoppedPluginProcessInfo) executeOper(
	std *syncdataaction.SyncDataActionStandarder, trigCtl workflow.ITriggerCtl, hostIDs ...int64) error {

	operationDef := NewOperSyncStoppedPluginProcessInfo(OperParamSyncStoppedPluginProcessInfo{
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
			Error("failed to create sync stopped plugin process info operation")

		return err
	}

	logger.G.Sys().Ctx(std.Context()).
		With("action", act.Name(), "tenant-id", std.TenantID(), "operation-id", operCtl.GetOperationID()).
		Info("created sync stopped plugin process info operation")

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionGenOperSyncStoppedPluginProcessInfo) DisplayNameZh() string {
	return "生成同步已停止插件进程信息任务"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionGenOperSyncStoppedPluginProcessInfo) DisplayNameEn() string {
	return "Generate Sync Stopped Plugin Process Info Operation"
}
