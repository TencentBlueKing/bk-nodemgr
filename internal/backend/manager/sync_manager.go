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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
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

	// LaunchSyncAlivePluginProcessInfo launch a task to sync alive plugin process info. returns the workflow-id.
	LaunchSyncAlivePluginProcessInfo(ctx contextx.IContext, hostIDs ...int64) (string, error)

	// LaunchSyncAllAlivePluginProcessInfo launch a task to sync all alive plugin process info. returns the workflow-id.
	LaunchSyncAllAlivePluginProcessInfo(ctx contextx.IContext) (string, error)
}

// LaunchSyncAllAgentState launch a task to sync all agent state.
func (mgr *Manager) LaunchSyncAllAgentState(nCtx contextx.IContext) (string, error) {
	tenantID := nCtx.TenantID()
	operator := nCtx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncAllAgentState(syncdata.OperParamSyncAllAgentState{
		TenantID: tenantID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID()).
		Info("launched sync all agent state task")

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncAgentState launch a task to sync agent state.
func (mgr *Manager) LaunchSyncAgentState(ctx contextx.IContext, hostIDs ...int64) (string, error) {
	if len(hostIDs) == 0 {
		return "", errors.New("hostIDs cannot be empty")
	}

	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	hosts, err := mgr.conf.StorageTopo.FindHostWithDynamic(ctx, types.UnlimitedPage(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
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

	operationDef := syncdata.NewOperSyncAgentState(syncdata.OperParamSyncAgentState{
		TenantID: tenantID,
		Hosts:    hostAgentID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.ActivateTrigger(ctx); err != nil {
		return "", err
	}

	logger.G.Sys().
		With("tenant-id", tenantID, "trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID()).
		Info("launched sync agent state task")

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncBizAndHost launch a task to sync biz and host.
func (mgr *Manager) LaunchSyncBizAndHost(ctx contextx.IContext) (string, error) {
	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncBizAndHost(syncdata.OperParamSyncBizAndHost{
		TenantID: tenantID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.ActivateTrigger(ctx); err != nil {
		return "", err
	}

	logger.G.Sys().
		With("tenant-id", tenantID, "trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID()).
		Info("launched sync biz and host task")

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncHostByBizID launch a task to sync host.
func (mgr *Manager) LaunchSyncHostByBizID(ctx contextx.IContext, bizID int64) (string, error) {
	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncHost(syncdata.OperParamSyncHost{
		TenantID: tenantID,
		BizID:    bizID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.ActivateTrigger(ctx); err != nil {
		return "", err
	}

	logger.G.Sys().
		With("tenant-id", tenantID, "trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID()).
		Info("launched sync host task")

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncNetworkArea launch a task to sync networkarea.
func (mgr *Manager) LaunchSyncNetworkArea(ctx contextx.IContext) (string, error) {
	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncNetworkArea(syncdata.OperParamSyncNetworkArea{
		TenantID: tenantID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.ActivateTrigger(ctx); err != nil {
		return "", err
	}

	logger.G.Sys().
		With("tenant-id", tenantID, "trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID()).
		Info("launched sync networkarea task")

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncAliveHostAgentInfo launch a task to sync all agent state.
func (mgr *Manager) LaunchSyncAliveHostAgentInfo(ctx contextx.IContext) (string, error) {
	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncAliveHostAgentInfo(syncdata.OperParamSyncAliveHostAgentInfo{
		TenantID: tenantID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.ActivateTrigger(ctx); err != nil {
		return "", err
	}

	logger.G.Sys().
		With("tenant-id", tenantID, "trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID()).
		Info("launched sync alive host agent info task")

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncAgentInfo launch a task to sync agent info.
func (mgr *Manager) LaunchSyncAgentInfo(ctx contextx.IContext, hostIDs ...int64) (string, error) {
	if len(hostIDs) == 0 {
		return "", errors.New("hostIDs cannot be empty")
	}

	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	hosts, err := mgr.conf.StorageTopo.FindHostWithDynamic(ctx, types.UnlimitedPage(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
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

	operationDef := syncdata.NewOperSyncAgentInfo(syncdata.OperParamSyncAgentInfo{
		TenantID: tenantID,
		Hosts:    hostAgentID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.ActivateTrigger(ctx); err != nil {
		return "", err
	}

	logger.G.Sys().
		With("tenant-id", tenantID, "trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID()).
		Info("launched sync agent info task")

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncAlivePluginProcessInfo launch a task to sync alive plugin process info.
func (mgr *Manager) LaunchSyncAlivePluginProcessInfo(ctx contextx.IContext, hostIDs ...int64) (string, error) {
	if len(hostIDs) == 0 {
		return "", errors.New("hostIDs cannot be empty")
	}

	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncAlivePluginProcessInfo(syncdata.OperParamSyncAlivePluginProcessInfo{
		TenantID: tenantID,
		Operator: operator,
		HostIDs:  hostIDs,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.ActivateTrigger(ctx); err != nil {
		return "", err
	}

	logger.G.Sys().
		With("tenant-id", tenantID, "trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID()).
		Info("launched sync alive plugin process info task")

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncAllAlivePluginProcessInfo launch a task to sync all alive plugin process info.
func (mgr *Manager) LaunchSyncAllAlivePluginProcessInfo(ctx contextx.IContext) (string, error) {
	tenantID := ctx.TenantID()
	operator := ctx.BKUsername()

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncAllAlivePluginProcessInfo(syncdata.OperParamSyncAllAlivePluginProcessInfo{
		TenantID: tenantID,
		Operator: operator,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.ActivateTrigger(ctx); err != nil {
		return "", err
	}

	logger.G.Sys().
		With("tenant-id", tenantID, "trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID()).
		Info("launched sync all alive plugin process info task")

	return triggerCtl.GetTriggerID(), nil
}
