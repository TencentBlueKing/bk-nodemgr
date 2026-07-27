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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
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
	// ActionNameGenOperSyncAgentState defines the action name.
	ActionNameGenOperSyncAgentState = "gen_oper_sync_agent_state"

	// syncAgentStateMaxPageSize defines the operation batch size.
	// gse list api has a limit of 1000, we can accept one action execute 5 loops.
	syncAgentStateMaxPageSize = 5000

	// syncAgentStateCreateOperConcurrencyLimit defines the concurrency limit for creating sync agent state operations.
	syncAgentStateCreateOperConcurrencyLimit = 100
)

// NewActionGenOperSyncAgentState this action will create host sync operation for all business.
func NewActionGenOperSyncAgentState(capability *Capability) action.Definition {
	return &actionGenOperSyncAgentState{
		hostStg:     capability.StorageTopo,
		workflowCtl: capability.WorkflowCtl,
		businessStg: capability.StorageTopo,
	}
}

// ActionParamGenOperSyncAgentState defines the action's param.
type ActionParamGenOperSyncAgentState struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionGenOperSyncAgentState struct {
	hostStg     topoStg.IStorageHost
	workflowCtl workflow.IController
	businessStg topoStg.IStorageBusiness
}

// Name returns the name of the action.
func (act *actionGenOperSyncAgentState) Name() string {
	return ActionNameGenOperSyncAgentState
}

// Version returns the version of the action.
func (act *actionGenOperSyncAgentState) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenOperSyncAgentState) Description() string {
	return "create sync agent state operation for all hosts"
}

// Timeout returns the timeout of the action.
func (act *actionGenOperSyncAgentState) Timeout() time.Duration {
	return time.Minute * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionGenOperSyncAgentState) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionGenOperSyncAgentState) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionGenOperSyncAgentState) DelayFn(_ int) func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionGenOperSyncAgentState) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamGenOperSyncAgentState)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}
	maxConcurrencyNum, err := conv.ToInt64(globalsettings.Get(
		std.Context(),
		globalsettings.OperSyncAgentStateMaxConcurrencyNum,
		globalsettings.OperSyncAgentStateMaxConcurrencyNumDefault,
	))
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", globalsettings.OperSyncAgentStateMaxConcurrencyNum, err)
	}
	if maxConcurrencyNum <= 0 {
		return fmt.Errorf("%s must be positive, got %d",
			globalsettings.OperSyncAgentStateMaxConcurrencyNum, maxConcurrencyNum)
	}

	bizs, _, err := act.businessStg.ListBusinesses(std.Context(), types.UnlimitedPage())
	if err != nil {
		return err
	}
	ctx.Data.Log().
		Zh("查询到 %d 个业务需要生成 Agent 状态更新任务", len(bizs)).
		En("found %d businesses need sync agent state operation", len(bizs)).
		Info()

	if len(bizs) == 0 {
		ctx.Data.Log().
			Zh("没有需要更新 Agent 状态的业务，跳过生成任务").
			En("no business needs sync agent state operation, skip generating operations").
			Info()

		return nil
	}

	// create trigger for handling sync agent state operations.
	meta := trigger.NewMetadataOrdered(int(maxConcurrencyNum))
	meta.CleanPolicy = trigger.MetadataCleanPolicy{
		MaxDays: syncDataCleanPolicyMaxDays(OperDefNameSyncAgentStateTimeout),
	}
	trigCtl, err := act.workflowCtl.CreateTrigger(std.Context(), trigger.CategoryOrdered, meta)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to create trigger for handling sync agent state operations")

		return err
	}

	gp := gopool.NewPool()
	gp.SetLimit(syncAgentStateCreateOperConcurrencyLimit)
	for idx := range bizs {
		biz := bizs[idx]

		gp.Go(func() error {
			result, err := act.createOperForBusiness(std, trigCtl, biz)
			if err != nil {
				return fmt.Errorf("failed to create sync agent state operations for business %d: %w", biz.BizID, err)
			}

			ctx.Data.Log().
				Zh("已为业务 %d 的 %d 台主机执行同步 Agent 状态任务", biz.BizID, result).
				En("executed sync agent state operation for %d hosts in business %d", result, biz.BizID).
				Info()

			return nil
		})
	}

	if err = gp.Wait(); err != nil {
		ctx.Data.Log().
			Zh("生成同步 Agent 状态任务失败").
			En("failed to generate sync agent state operations").
			Error()

		return fmt.Errorf("failed to generate sync agent state operations: %w", err)
	}

	if err = trigCtl.ActivateTrigger(std.Context()); err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).
			With("action", act.Name()).
			Error("failed to run trigger for handling sync agent state operations")

		return err
	}

	return nil
}

func (act *actionGenOperSyncAgentState) createOperForBusiness(std *syncDataUtils.SyncDataActionStandarder,
	trigCtl workflow.ITriggerCtl, biz *types.Business) (int, error) {

	// find nodes and sync agent state. skip the empty-agent-id nodes.
	cond := &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			BizID: []int64{biz.BizID},
		},
		DynamicAgentIDNotEmpty: true,
	}

	scanCtx, cancel := contextx.WithTimeout(std.Context(), act.Timeout())
	defer cancel()

	hosts, err := act.hostStg.ScanAllHostWithFields(scanCtx, &types.HostFieldSelection{
		HostID:  true,
		AgentID: true,
	}, cond)
	if err != nil {
		return 0, err
	}

	if err = batchexecutor.Execute(scanCtx, hosts,
		func(_ contextx.IContext, batchHosts []*types.Host) error {
			return act.executeOper(std, trigCtl, batchHosts...)
		}, batchexecutor.WithBatchSize(syncAgentStateMaxPageSize), batchexecutor.WithTimeout(act.Timeout())); err != nil {
		return 0, err
	}

	return len(hosts), nil
}

// executeOper create an operation to sync agent state for the given hosts and then execute it.
func (act *actionGenOperSyncAgentState) executeOper(
	std *syncDataUtils.SyncDataActionStandarder, trigCtl workflow.ITriggerCtl, hosts ...*types.Host) error {

	hostAgentID := make([]*HostIDAgentID, 0, len(hosts))
	for _, host := range hosts {
		hostAgentID = append(hostAgentID, &HostIDAgentID{
			HostID:  host.HostID,
			AgentID: host.Dynamic.AgentID,
		})
	}
	operationDef := NewOperSyncAgentState(OperParamSyncAgentState{
		TenantID: std.TenantID(),
		Hosts:    hostAgentID,
		Operator: std.Operator(),
	})

	operationParam := operationDef.DefaultParameters()
	operationParam.ParentOperationID = std.InstanceData().OperationID
	operationParam.ParentOperInstID = std.InstanceData().OperationInstanceID

	operCtl, err := trigCtl.CreateOperation(std.Context(), operationDef, operationParam)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).
			WithErr(err).
			With("action", act.Name(), "tenant-id", std.TenantID()).
			Error("failed to create sync agent state operation")

		return err
	}

	logger.G.Sys().Ctx(std.Context()).
		With("action", act.Name(), "tenant-id", std.TenantID(), "operation-id", operCtl.GetOperationID()).
		Info("created sync agent state operation")

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionGenOperSyncAgentState) DisplayNameZh() string {
	return "生成同步 Agent 状态任务"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionGenOperSyncAgentState) DisplayNameEn() string {
	return "Generate Sync Agent State Operation"
}
