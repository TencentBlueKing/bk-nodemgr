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
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	// ActionNameGenOperSyncHost defines the action name.
	ActionNameGenOperSyncHost = "gen_oper_sync_host"
)

// NewActionGenOperSyncHost this action will create host sync operation for all business.
func NewActionGenOperSyncHost(capability *Capability) action.Definition {
	return &actionGenOperSyncHost{
		storageBusiness: capability.StorageTopo,
		workflowCtl:     capability.WorkflowCtl,
	}
}

// ActionParamGenOperSyncHost defines the action's param.
type ActionParamGenOperSyncHost struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionGenOperSyncHost struct {
	storageBusiness topoStg.IStorageBusiness
	workflowCtl     workflow.IController
}

// Name returns the name of the action.
func (act *actionGenOperSyncHost) Name() string {
	return ActionNameGenOperSyncHost
}

// Version returns the version of the action.
func (act *actionGenOperSyncHost) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenOperSyncHost) Description() string {
	return "reads all business information from the database," +
		"and creates host synchronization tasks on a business-by-business basis"
}

// Timeout returns the timeout of the action.
func (act *actionGenOperSyncHost) Timeout() time.Duration {
	return time.Second * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionGenOperSyncHost) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionGenOperSyncHost) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionGenOperSyncHost) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionGenOperSyncHost) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamGenOperSyncHost)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	bizs, _, err := act.storageBusiness.ListBusinesses(std.Context(), types.UnlimitedPage())
	if err != nil {
		return err
	}
	logger.G.Sys().Ctx(std.Context()).With("action", act.Name()).Info("found business: %d", len(bizs))

	if len(bizs) == 0 {
		return nil
	}

	// create trigger for handling sync host operations.
	meta := trigger.NewMetadataOnce()
	meta.CleanPolicy = trigger.MetadataCleanPolicy{
		MaxDays: 1,
	}
	trigCtl, err := act.workflowCtl.CreateTrigger(std.Context(), trigger.CategoryOnce, meta)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).With("action", act.Name()).Error("failed to create trigger for handling sync host operations")

		return err
	}

	for _, biz := range bizs {
		if err = act.executeOper(std, trigCtl, biz); err != nil {
			return err
		}
	}

	if err = trigCtl.ActivateTrigger(std.Context()); err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).With("action", act.Name()).Error("failed to run trigger for handling sync host operations")

		return err
	}

	logger.G.Sys().Ctx(std.Context()).With("action", act.Name()).Info("executed sync host operation for %d business", len(bizs))

	return nil
}

// executeOper create an operation to sync all host from cmdb and then execute it.
func (act *actionGenOperSyncHost) executeOper(std *syncDataUtils.SyncDataActionStandarder, trigCtl workflow.ITriggerCtl, biz *types.Business) error {
	operationDef := NewOperSyncHost(OperParamSyncHost{
		TenantID: biz.TenantID,
		Operator: std.Operator(),
		BizID:    biz.BizID,
	})

	operationParam := operationDef.DefaultParameters()
	operationParam.ParentOperationID = std.InstanceData().OperationID

	operCtl, err := trigCtl.CreateOperation(std.Context(), operationDef, operationParam)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).
			WithErr(err).
			With("action", act.Name(), "tenant-id", biz.TenantID, "biz-id", biz.BizID).
			Error("failed to create sync host operation for business")

		return err
	}

	logger.G.Sys().Ctx(std.Context()).
		With("action", act.Name(), "tenant-id", biz.TenantID, "biz-id", biz.BizID, "operation-id", operCtl.GetOperationID()).
		Info("created sync host operation for business")

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionGenOperSyncHost) DisplayNameZh() string { return "生成同步主机任务" }

// DisplayNameEn returns the English display name of the action.
func (act *actionGenOperSyncHost) DisplayNameEn() string { return "Generate Sync Host Operation" }
