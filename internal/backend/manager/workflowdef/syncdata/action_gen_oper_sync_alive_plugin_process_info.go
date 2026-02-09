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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"

	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	// ActionNameGenOperSyncAlivePluginProcessInfo defines the action name.
	ActionNameGenOperSyncAlivePluginProcessInfo = "gen_oper_sync_alive_plugin_process_info"

	// syncAgentStateMaxPageSize defines the max page size for page executor.
	// In the scenario of 40,000 hosts, a single request for 1,000 hosts requires 400 table lookups.
	syncAlivePluginProcessStatusMaxPageSize = 1000
)

// NewActionGenOperSyncAlivePluginProcessInfo this action will create host sync operation for all business.
func NewActionGenOperSyncAlivePluginProcessInfo(capability *Capability) action.Definition {
	return &actionGenOperSyncAlivePluginProcessInfo{
		topoStg:     capability.StorageTopo,
		processStg:  capability.StoragePlugin,
		workflowCtl: capability.WorkflowCtl,
	}
}

// ActionParamGenOperSyncAlivePluginProcessInfo defines the action's param.
type ActionParamGenOperSyncAlivePluginProcessInfo struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionGenOperSyncAlivePluginProcessInfo struct {
	topoStg     topoStg.IStorageHost
	processStg  plugin.IDaoProcess
	workflowCtl workflow.IController
}

// Name returns the name of the action.
func (act *actionGenOperSyncAlivePluginProcessInfo) Name() string {
	return ActionNameGenOperSyncAlivePluginProcessInfo
}

// Version returns the version of the action.
func (act *actionGenOperSyncAlivePluginProcessInfo) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenOperSyncAlivePluginProcessInfo) Description() string {
	return "create sync alive plugin process status operation for all hosts"
}

// Timeout returns the timeout of the action.
func (act *actionGenOperSyncAlivePluginProcessInfo) Timeout() time.Duration {
	return time.Second * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionGenOperSyncAlivePluginProcessInfo) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionGenOperSyncAlivePluginProcessInfo) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionGenOperSyncAlivePluginProcessInfo) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
// nolint:gocognit
func (act *actionGenOperSyncAlivePluginProcessInfo) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamGenOperSyncAlivePluginProcessInfo)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	var trigCtl workflow.ITriggerCtl
	executor := pageexecutor.NewPageExecutor[*types.Host](syncAlivePluginProcessStatusMaxPageSize, 1*time.Hour)
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		hosts, err := act.topoStg.FindHostWithDynamic(nCtx, p, &types.HostCondition{
			DynamicExactExclude: &types.HostDynamicExactFields{
				AgentID: []string{""},
			},
		})
		if err != nil {
			return nil, err
		}

		if len(hosts) == 0 {
			return nil, nil
		}

		hostIDs := make([]int64, 0, len(hosts))
		for _, host := range hosts {
			hostIDs = append(hostIDs, host.HostID)
		}

		if trigCtl == nil {
			// create trigger for handling sync alive plugin process info operations.
			meta := trigger.NewMetadataOnce()
			meta.CleanPolicy = trigger.MetadataCleanPolicy{
				MaxDays: 1,
			}
			trigCtl, err = act.workflowCtl.CreateTrigger(nCtx, trigger.CategoryOnce, meta)
			if err != nil {
				logger.G.Sys().WithErr(err).With("action", act.Name()).Error("failed to create trigger for handling sync alive plugin process info operations")
				return nil, err
			}
		}

		if err = act.executeOper(std, trigCtl, hostIDs...); err != nil {
			return nil, err
		}

		return hosts, nil
	}

	result, err := executor.Execute(std.Context(), types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	if trigCtl != nil {
		if err = trigCtl.ActivateTrigger(std.Context()); err != nil {
			logger.G.Sys().WithErr(err).With("action", act.Name()).Error("failed to run trigger for handling sync alive plugin process info operations")
			return err
		}
	}

	logger.G.Sys().With("action", act.Name()).Info("executed sync alive plugin process info operation for %d hosts", result.Total)

	return nil
}

// executeOper create an operation to sync alive plugin process info and then execute it.
func (act *actionGenOperSyncAlivePluginProcessInfo) executeOper(
	std *syncDataUtils.SyncDataActionStandarder, trigCtl workflow.ITriggerCtl, hostIDs ...int64) error {

	operationDef := NewOperSyncAlivePluginProcessInfo(OperParamSyncAlivePluginProcessInfo{
		TenantID: std.TenantID(),
		Operator: std.Operator(),
		HostIDs:  hostIDs,
	})

	operationParam := operationDef.DefaultParameters()
	operationParam.ParentOperationID = std.InstanceData().OperationID

	operCtl, err := trigCtl.CreateOperation(std.Context(), operationDef, operationParam)
	if err != nil {
		logger.G.Sys().
			WithErr(err).
			With("action", act.Name(), "tenant-id", std.TenantID()).
			Error("failed to create sync alive plugin process info operation")

		return err
	}

	logger.G.Sys().
		With("action", act.Name(), "tenant-id", std.TenantID(), "operation-id", operCtl.GetOperationID()).
		Info("created sync alive plugin process info operation")

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionGenOperSyncAlivePluginProcessInfo) DisplayNameZh() string {
	return "生成同步存活插件进程信息任务"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionGenOperSyncAlivePluginProcessInfo) DisplayNameEn() string {
	return "Generate Sync Alive Plugin Process Info Operation"
}
