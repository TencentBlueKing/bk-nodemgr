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
	"fmt"
	"time"

	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	// ActionNameGenOperSyncHostTopoRelation defines the action name.
	ActionNameGenOperSyncHostTopoRelation = "gen_oper_sync_host_topo_relation"
)

// NewActionGenOperSyncHostTopoRelation creates host topo relation sync operations for all businesses.
func NewActionGenOperSyncHostTopoRelation(capability *Capability) action.Definition {
	return &actionGenOperSyncHostTopoRelation{
		storageBusiness: capability.StorageTopo,
		workflowCtl:     capability.WorkflowCtl,
	}
}

// ActionParamGenOperSyncHostTopoRelation defines the action's param.
type ActionParamGenOperSyncHostTopoRelation struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionGenOperSyncHostTopoRelation struct {
	storageBusiness topoStg.IStorageBusiness
	workflowCtl     workflow.IController
}

// Name returns the name of the action.
func (act *actionGenOperSyncHostTopoRelation) Name() string {
	return ActionNameGenOperSyncHostTopoRelation
}

// Version returns the version of the action.
func (act *actionGenOperSyncHostTopoRelation) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenOperSyncHostTopoRelation) Description() string {
	return "reads all business information from the database," +
		"and creates host topo relation synchronization tasks on a business-by-business basis"
}

// Timeout returns the timeout of the action.
func (act *actionGenOperSyncHostTopoRelation) Timeout() time.Duration {
	return time.Minute * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionGenOperSyncHostTopoRelation) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionGenOperSyncHostTopoRelation) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionGenOperSyncHostTopoRelation) DelayFn(_ int) func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionGenOperSyncHostTopoRelation) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamGenOperSyncHostTopoRelation)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	bizs, _, err := act.storageBusiness.ListBusinesses(std.Context(), types.UnlimitedPage())
	if err != nil {
		return err
	}
	ctx.Data.Log().
		Zh("查询到 %d 个业务需要生成同步主机拓扑关系任务", len(bizs)).
		En("found %d businesses to generate sync host topo relation operations", len(bizs)).
		Info()

	if len(bizs) == 0 {
		ctx.Data.Log().
			Zh("没有需要同步主机拓扑关系的业务，跳过生成任务").
			En("no business needs sync host topo relation operation, skip generating operations").
			Info()

		return nil
	}

	maxConcurrencyNum, err := conv.ToInt64(globalsettings.Get(
		std.Context(),
		globalsettings.OperSyncHostMaxConcurrencyNum,
		globalsettings.OperSyncHostMaxConcurrencyNumDefault,
	))
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", globalsettings.OperSyncHostMaxConcurrencyNum, err)
	}
	if maxConcurrencyNum <= 0 {
		return fmt.Errorf("%s must be positive, got %d",
			globalsettings.OperSyncHostMaxConcurrencyNum, maxConcurrencyNum)
	}

	meta := trigger.NewMetadataOrdered(int(maxConcurrencyNum))
	cleanPolicyMaxDays := syncDataCleanPolicyMaxDays(OperDefNameSyncHostTopoRelationTimeout)
	meta.CleanPolicy = trigger.MetadataCleanPolicy{
		MaxDays: cleanPolicyMaxDays,
	}
	trigCtl, err := act.workflowCtl.CreateTrigger(std.Context(), trigger.CategoryOrdered, meta)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).
			WithErr(err).
			With("action", act.Name()).
			Error("failed to create trigger for handling sync host topo relation operations")

		return err
	}
	ctx.Data.Log().
		Zh("已创建同步主机拓扑关系任务触发器，清理周期为 %.4f 天", cleanPolicyMaxDays).
		En("created sync host topo relation operation trigger, clean policy is %.4f days", cleanPolicyMaxDays).
		Info()

	gp := gopool.NewPool()
	gp.SetLimit(executeOperLimit)

	for idx := range bizs {
		biz := bizs[idx]
		gp.Go(func() error {
			if err := act.executeOper(std, trigCtl, biz); err != nil {
				return err
			}

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		ctx.Data.Log().
			Zh("生成同步主机拓扑关系任务失败").
			En("failed to generate sync host topo relation operations").
			Error()

		return fmt.Errorf("failed to generate sync host topo relation operations: %w", err)
	}

	ctx.Data.Log().
		Zh("已生成 %d 个同步主机拓扑关系任务", len(bizs)).
		En("generated %d sync host topo relation operations", len(bizs)).
		Info()

	if err = trigCtl.ActivateTrigger(std.Context()); err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).
			With("action", act.Name()).Error("failed to run trigger for handling sync host topo relation operations")

		return err
	}

	logger.G.Sys().Ctx(std.Context()).With("action", act.Name()).
		Info("executed sync host topo relation operation for %d business", len(bizs))

	return nil
}

// executeOper create an operation to sync host topo relation from cmdb and then execute it.
func (act *actionGenOperSyncHostTopoRelation) executeOper(
	std *syncDataUtils.SyncDataActionStandarder, trigCtl workflow.ITriggerCtl, biz *types.Business) error {

	operationDef := NewOperSyncHostTopoRelation(OperParamSyncHostTopoRelation{
		TenantID: biz.TenantID,
		Operator: std.Operator(),
		BizID:    biz.BizID,
	})

	operationParam := operationDef.DefaultParameters()
	operationParam.ParentOperationID = std.InstanceData().OperationID
	operationParam.ParentOperInstID = std.InstanceData().OperationInstanceID

	operCtl, err := trigCtl.CreateOperation(std.Context(), operationDef, operationParam)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).
			WithErr(err).
			With("action", act.Name(), "tenant-id", biz.TenantID, "biz-id", biz.BizID).
			Error("failed to create sync host topo relation operation for business")

		return err
	}

	logger.G.Sys().Ctx(std.Context()).
		With("action", act.Name(), "tenant-id", biz.TenantID,
			"biz-id", biz.BizID, "operation-id", operCtl.GetOperationID()).
		Info("created sync host topo relation operation for business")

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionGenOperSyncHostTopoRelation) DisplayNameZh() string {
	return "生成同步主机拓扑关系任务"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionGenOperSyncHostTopoRelation) DisplayNameEn() string {
	return "Generate Sync Host Topo Relation Operation"
}
