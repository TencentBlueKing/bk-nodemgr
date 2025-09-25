/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// ISyncManager defines the SyncManager interface.
type ISyncManager interface {
	// LaunchSyncBizAndHost launch a task to sync biz and host. returns the trigger-id.
	LaunchSyncBizAndHost(ctx contextx.IContext) (string, error)

	// LaunchSyncHostByBizID launch a task to sync host by biz-id. returns the trigger-id.
	LaunchSyncHostByBizID(ctx contextx.IContext, bizID int64) (string, error)

	// LaunchSyncNetworkArea launch a task to sync networkarea. returns the trigger-id.
	LaunchSyncNetworkArea(ctx contextx.IContext) (string, error)

	// LaunchSyncAgentState launch a task to sync agent state from gse. returns the workflow-id.
	LaunchSyncAgentState(ctx contextx.IContext, hostIDs ...int64) (string, error)

	// LaunchSyncAllAgentState launch a task to sync all agent state from gse. returns the workflow-id.
	LaunchSyncAllAgentState(ctx contextx.IContext) (string, error)

	// LaunchSyncAgentInfo launch a task to sync agent info from gse. returns the workflow-id.
	LaunchSyncAgentInfo(ctx contextx.IContext, hostIDs ...int64) (string, error)

	// LaunchSyncAliveHostAgentInfo launch a task to sync alive host agent info. returns the trigger-id.
	LaunchSyncAliveHostAgentInfo(ctx contextx.IContext) (string, error)
}

// LaunchSyncAllAgentState launch a task to sync all agent state.
func (mgr *Manager) LaunchSyncAllAgentState(ctx contextx.IContext) (string, error) {
	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncAllAgentStateFromGSE(syncdata.OperParamSyncAllAgentStateFromGSE{
		TenantID: tenantID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync all agent state task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncAgentState launch a task to sync agent state.
func (mgr *Manager) LaunchSyncAgentState(ctx contextx.IContext, hostIDs ...int64) (string, error) {
	if len(hostIDs) == 0 {
		return "", errors.New("hostIDs cannot be empty")
	}

	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	hosts, err := mgr.conf.StorageTopo.FindHostWithDynamic(ctx, types.UnlimitedPage(), &types.HostCondition{
		ExactInclude: &types.HostExactFields{
			HostID: hostIDs,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to find hosts with dynamic info: %w", err)
	}

	hostAgentID := make([]*syncdata.HostIDAgentID, 0, len(hosts))
	for _, host := range hosts {
		hostAgentID = append(hostAgentID, &syncdata.HostIDAgentID{
			HostID:  host.HostID,
			AgentID: host.Dynamic.AgentID,
		})
	}

	operationDef := syncdata.NewOperSyncAgentStateFromGSE(syncdata.OperParamSyncAgentStateFromGSE{
		TenantID: tenantID,
		Hosts:    hostAgentID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync agent state task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncBizAndHost launch a task to sync biz and host.
func (mgr *Manager) LaunchSyncBizAndHost(ctx contextx.IContext) (string, error) {
	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncBizAndHostFromCMDB(syncdata.OperParamSyncBizAndHostFromCMDB{
		TenantID: tenantID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync biz and host task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncHostByBizID launch a task to sync host.
func (mgr *Manager) LaunchSyncHostByBizID(ctx contextx.IContext, bizID int64) (string, error) {
	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncHostFromCMDB(syncdata.OperParamSyncHostFromCMDB{
		TenantID: tenantID,
		BizID:    bizID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync host task. tenant-id(%s), biz-id(%d), trigger-id(%s), operation-id(%s)",
		tenantID, bizID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncNetworkArea launch a task to sync networkarea.
func (mgr *Manager) LaunchSyncNetworkArea(ctx contextx.IContext) (string, error) {
	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncNetworkAreaFromCMDB(syncdata.OperParamSyncNetworkAreaFromCMDB{
		TenantID: tenantID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync networkarea task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncAliveHostAgentInfo launch a task to sync all agent state.
func (mgr *Manager) LaunchSyncAliveHostAgentInfo(ctx contextx.IContext) (string, error) {
	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncAliveHostAgentInfoFromGSE(syncdata.OperParamSyncAliveHostAgentInfoFromGSE{
		TenantID: tenantID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync alive host agent info task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncAgentInfo launch a task to sync agent info.
func (mgr *Manager) LaunchSyncAgentInfo(ctx contextx.IContext, hostIDs ...int64) (string, error) {
	if len(hostIDs) == 0 {
		return "", errors.New("hostIDs cannot be empty")
	}

	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	hosts, err := mgr.conf.StorageTopo.FindHostWithDynamic(ctx, types.UnlimitedPage(), &types.HostCondition{
		ExactInclude: &types.HostExactFields{
			HostID: hostIDs,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to find hosts with dynamic info: %w", err)
	}

	hostAgentID := make([]*syncdata.HostIDAgentID, 0, len(hosts))
	for _, host := range hosts {
		hostAgentID = append(hostAgentID, &syncdata.HostIDAgentID{
			HostID:  host.HostID,
			AgentID: host.Dynamic.AgentID,
		})
	}

	operationDef := syncdata.NewOperSyncAgentInfoFromGSE(syncdata.OperParamSyncAgentInfoFromGSE{
		TenantID: tenantID,
		Hosts:    hostAgentID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync agent info task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}
