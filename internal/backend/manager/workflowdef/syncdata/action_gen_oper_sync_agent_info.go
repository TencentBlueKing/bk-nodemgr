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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	// ActionNameGenOperSyncAgentInfo defines the action name.
	ActionNameGenOperSyncAgentInfo = "gen_oper_sync_agent_info"

	// syncAgentInfoMaxPageSize defines the max page size for page executor.
	// gse list api has a limit of 1000, we can accept one action execute 5 loops.
	syncAgentInfoMaxPageSize = 5000
)

// NewActionGenOperSyncAgentInfo this action will create host sync operation for all businessStg.
func NewActionGenOperSyncAgentInfo(capability *Capability) action.Definition {
	return &actionGenOperSyncAgentInfo{
		hostStg:     capability.StorageTopo,
		workflowCtl: capability.WorkflowCtl,
		businessStg: capability.StorageTopo,
	}
}

// ActionParamGenOperSyncAgentInfo defines the action's param.
type ActionParamGenOperSyncAgentInfo struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionGenOperSyncAgentInfo struct {
	hostStg     topoStg.IStorageHost
	workflowCtl workflow.IController
	businessStg topoStg.IStorageBusiness
}

// Name returns the name of the action.
func (act *actionGenOperSyncAgentInfo) Name() string {
	return ActionNameGenOperSyncAgentInfo
}

// Version returns the version of the action.
func (act *actionGenOperSyncAgentInfo) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenOperSyncAgentInfo) Description() string {
	return "create sync agent info operation for alive hosts"
}

// Timeout returns the timeout of the action.
func (act *actionGenOperSyncAgentInfo) Timeout() time.Duration {
	return time.Minute * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionGenOperSyncAgentInfo) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionGenOperSyncAgentInfo) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionGenOperSyncAgentInfo) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionGenOperSyncAgentInfo) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamGenOperSyncAgentInfo)
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
		globalsettings.OperSyncAgentInfoMaxConcurrencyNum,
		globalsettings.OperSyncAgentInfoMaxConcurrencyNumDefault,
	))
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", globalsettings.OperSyncAgentInfoMaxConcurrencyNum, err)
	}
	if maxConcurrencyNum <= 0 {
		return fmt.Errorf("%s must be positive, got %d",
			globalsettings.OperSyncAgentInfoMaxConcurrencyNum, maxConcurrencyNum)
	}

	bizs, _, err := act.businessStg.ListBusinesses(std.Context(), types.UnlimitedPage())
	if err != nil {
		return err
	}
	ctx.Data.Log().
		Zh("查询到 %d 个业务需要生成 Agent 信息更新任务", len(bizs)).
		En("found %d businesses need sync agent info operation", len(bizs)).
		Info()

	if len(bizs) == 0 {
		ctx.Data.Log().
			Zh("没有需要更新 Agent 信息的业务，跳过生成任务").
			En("no business needs sync agent info operation, skip generating operations").
			Info()

		return nil
	}

	var trigCtl workflow.ITriggerCtl
	// create trigger for handling sync agent state operations.
	meta := trigger.NewMetadataOrdered(int(maxConcurrencyNum))
	meta.CleanPolicy = trigger.MetadataCleanPolicy{
		MaxDays: 1,
	}
	trigCtl, err = act.workflowCtl.CreateTrigger(std.Context(), trigger.CategoryOrdered, meta)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to create trigger for handling sync agent info operations")

		return err
	}

	gp := gopool.NewPool()
	gp.SetLimit(100)
	for idx := range bizs {
		biz := bizs[idx]

		gp.Go(func() error {
			result, err := act.createOperForBusiness(std, trigCtl, biz)
			if err != nil {
				return err
			}

			ctx.Data.Log().
				Zh("已为业务 %d 的 %d 台主机执行同步 Agent 信息任务", biz.BizID, result.Total).
				En("executed sync agent info operation for %d hosts in business %d", result.Total, biz.BizID).
				Info()

			return nil
		})
	}

	if trigCtl != nil {
		if err = trigCtl.ActivateTrigger(std.Context()); err != nil {
			logger.G.Sys().Ctx(std.Context()).WithErr(err).
				With("action", act.Name()).
				Error("failed to run trigger for handling sync agent info operations")

			return err
		}
	}

	return nil
}

func (act *actionGenOperSyncAgentInfo) createOperForBusiness(std *syncDataUtils.SyncDataActionStandarder, trigCtl workflow.ITriggerCtl,
	biz *types.Business) (*pageexecutor.PageResult[*types.Host], error) {

	executor := pageexecutor.NewPageExecutor[*types.Host](syncAgentInfoMaxPageSize, act.Timeout())
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		// find alive nodes and sync agent info. skip the empty-agent-id nodes.
		cond := &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				BizID: []int64{biz.BizID},
			},
			DynamicExactInclude: &types.HostDynamicExactFields{
				NodeStatus: []types.NodeStatus{types.NodeStatusRunning},
			},
			DynamicExactExclude: &types.HostDynamicExactFields{
				AgentID: []string{""},
			},
		}

		hosts, err := act.hostStg.FindHostWithDynamic(nCtx, p, cond)
		if err != nil {
			return nil, err
		}

		if len(hosts) > 0 {
			if err = act.executeOper(std, trigCtl, hosts...); err != nil {
				return nil, err
			}
		}

		return hosts, nil
	}

	result, err := executor.Execute(std.Context(), types.UnlimitedPage(), fn)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeOper create an operation to sync agent info for the given hosts and then execute it.
func (act *actionGenOperSyncAgentInfo) executeOper(
	std *syncDataUtils.SyncDataActionStandarder, trigCtl workflow.ITriggerCtl, hosts ...*types.Host) error {

	hostAgentID := make([]*HostIDAgentID, 0, len(hosts))
	for _, host := range hosts {
		hostAgentID = append(hostAgentID, &HostIDAgentID{
			HostID:  host.HostID,
			AgentID: host.Dynamic.AgentID,
		})
	}
	operationDef := NewOperSyncAgentInfo(OperParamSyncAgentInfo{
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
			Error("failed to create sync agent info operation")

		return err
	}

	logger.G.Sys().Ctx(std.Context()).
		With("action", act.Name(), "tenant-id", std.TenantID(), "operation-id", operCtl.GetOperationID()).
		Info("created sync agent info operation")

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionGenOperSyncAgentInfo) DisplayNameZh() string {
	return "生成同步 Agent 信息任务"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionGenOperSyncAgentInfo) DisplayNameEn() string {
	return "Generate Sync Agent Info Operation"
}
