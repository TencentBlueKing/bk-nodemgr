/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package manager provides handlers to manage all the nodeman task operations
package manager

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	worksche "github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/schedule"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	// SyncCmdbHostWorkflowName defines the name of the sync host workflow.
	SyncCmdbHostWorkflowName = "schedule_sync_cmdb_host"

	// SyncCmdbNetworkAreaWorkflowName defines the name of the sync cmdb network area workflow.
	SyncCmdbNetworkAreaWorkflowName = "schedule_sync_cmdb_network_area"

	// SyncGseAgentStateWorkflowName defines the name of the sync GSE agent state workflow.
	SyncGseAgentStateWorkflowName = "schedule_sync_gse_agent_state"

	// SyncAliveHostAgentInfoWorkflowName defines the name of the sync alive host agent info workflow.
	SyncAliveHostAgentInfoWorkflowName = "schedule_sync_alive_host_agent_info"

	// WatchAndApplyCMDBResourceWorkflowName defines the name of the watch and apply cmdb resource workflow.
	WatchAndApplyCMDBResourceWorkflowName = "schedule_watch_and_apply_cmdb_resource"
)

// ScheduleWorkflowFunc defines the function type for scheduling workflows.
type ScheduleWorkflowFunc func(ctx contextx.IContext) error

// getScheduleWorkflow returns a map of workflow names to their corresponding scheduling functions.
func (mgr *Manager) getScheduleWorkflow() map[string]ScheduleWorkflowFunc {
	return map[string]ScheduleWorkflowFunc{
		SyncCmdbHostWorkflowName:              mgr.ScheduleSyncHostFromCMDB,
		SyncCmdbNetworkAreaWorkflowName:       mgr.ScheduleSyncNetworkAreaFromCMDB,
		SyncGseAgentStateWorkflowName:         mgr.ScheduleSyncAllAgentStateFromGSE,
		SyncAliveHostAgentInfoWorkflowName:    mgr.ScheduleSyncAliveHostAgentInfo,
		WatchAndApplyCMDBResourceWorkflowName: mgr.ScheduleWatchAndApplyCMDBResource,
	}
}

// startScheduleWorkflow starts the scheduled workflows.
func (mgr *Manager) startScheduleWorkflow(ctx contextx.IContext) error {
	dbScheduleWorkflows, _, err := mgr.conf.StorageWorkflow.ListScheduleWorkflow(ctx, types.UnlimitedPage(),
		&types.ScheduleWorkflowCondition{ExactInclude: &types.ScheduleWorkflowExactFields{}})
	if err != nil {
		return fmt.Errorf("failed to list schedule workflows from storage: %w", err)
	}

	dbScheduleWorkflowsMap, err := conv.SliceToMap(dbScheduleWorkflows, func(swo *worksche.Schedule) string {
		return swo.WorkflowName
	})
	if err != nil {
		return fmt.Errorf("failed to convert schedule workflows to map: %w", err)
	}

	scheduleWorkflowMaps := mgr.getScheduleWorkflow()
	unRegisteredWorkflows := make([]string, 0, len(scheduleWorkflowMaps))
	for workflowName := range scheduleWorkflowMaps {
		dbSchedule, ok := dbScheduleWorkflowsMap[workflowName]
		if !ok {
			unRegisteredWorkflows = append(unRegisteredWorkflows, workflowName)
			continue
		}

		triggerCtl, err := mgr.workflowMgr.GetTrigger(ctx, dbSchedule.TriggerID)
		if err != nil {
			logger.G.Sys().WithErr(err).With("workflow-id", dbSchedule.WorkflowID).Error("failed to get schedule workflow trigger")

			return fmt.Errorf("get schedule workflow(%s) trigger failed: %w", dbSchedule.WorkflowID, err)
		}

		err = triggerCtl.RunTrigger(ctx)
		if err != nil {
			logger.G.Sys().
				WithErr(err).
				With("workflow-id", dbSchedule.WorkflowID, "trigger-id", dbSchedule.TriggerID).
				Error("failed to run schedule workflow trigger")

			return fmt.Errorf("run schedule workflow(%s) trigger(%s) failed: %w",
				dbSchedule.WorkflowID, dbSchedule.TriggerID, err)
		}
	}

	for _, workflowName := range unRegisteredWorkflows {
		scheduleFn := scheduleWorkflowMaps[workflowName]
		if scheduleFn == nil {
			logger.G.Sys().With("workflow-name", workflowName).Warn("workflow function not found, skip")

			continue
		}

		err := scheduleFn(ctx)
		if err != nil {
			logger.G.Sys().WithErr(err).With("workflow-name", workflowName).Warn("failed to register workflow")

			return fmt.Errorf("register workflow-name(%s) failed: %w", workflowName, err)
		}
	}

	return nil
}

// createAndRunScheduleWorkflows creates and runs a scheduled workflow.
func (mgr *Manager) createAndRunScheduleWorkflows(ctx contextx.IContext, workflowName string, interval string) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to get tenant id")

		return fmt.Errorf("get tenant id failed: %w", err)
	}

	metadataPeriodic, err := trigger.NewMetadataPeriodic(interval, false)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to create periodic metadata")

		return fmt.Errorf("create periodic metadata failed: %w", err)
	}

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryPeriodic, metadataPeriodic)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to create periodic trigger")

		return fmt.Errorf("create periodic trigger failed: %w", err)
	}

	scheduleWf := &worksche.Schedule{
		WorkflowID:   identifier.GenWorkflowID(),
		WorkflowName: workflowName,
		TriggerID:    triggerCtl.GetTriggerID(),
		Operator:     access.GetVirtualUser(),
		OperateTime:  time.Now(),
	}

	err = mgr.conf.StorageWorkflow.CreateScheduleWorkflow(ctx, scheduleWf)
	if err != nil {
		logger.G.Sys().WithErr(err).With("schedule-workflow", scheduleWf.WorkflowName).Error("failed to create schedule workflow")

		return fmt.Errorf("create schedule-workflow(%s) failed: %w", scheduleWf.WorkflowName, err)
	}

	operationDef := schedule.NewOperScheduleOnceTriggerOperation(
		schedule.OperParamScheduleOnceTriggerOperation{TenantID: tenantID, Operator: access.GetVirtualUser()},
		workflowName)
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		logger.G.Sys().WithErr(err).With("schedule-workflow", scheduleWf.WorkflowName).Error("failed to create operation for schedule workflow")

		return fmt.Errorf("create operation for schedule-workflow(%s) failed: %w", scheduleWf.WorkflowName, err)
	}

	logger.G.Sys().
		With("tenant-id", tenantID).
		With("workflow-name", scheduleWf.WorkflowName, "trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID()).
		Info("new schedule watch and apply cmdb resource task")

	err = triggerCtl.RunTrigger(ctx)
	if err != nil {
		logger.G.Sys().WithErr(err).With("schedule-workflow", scheduleWf.WorkflowName).Error("failed to run schedule workflow")

		return fmt.Errorf("run schedule-workflow(%s) failed: %w", scheduleWf.WorkflowName, err)
	}

	return nil
}

// ScheduleSyncHostFromCMDB creates a new schedule workflow to sync hosts from CMDB.
func (mgr *Manager) ScheduleSyncHostFromCMDB(ctx contextx.IContext) error {
	return mgr.createAndRunScheduleWorkflows(ctx, SyncCmdbHostWorkflowName, scheduler.Midnight)
}

// ScheduleSyncNetworkAreaFromCMDB creates a new schedule workflow to sync network areas from CMDB.
func (mgr *Manager) ScheduleSyncNetworkAreaFromCMDB(ctx contextx.IContext) error {
	return mgr.createAndRunScheduleWorkflows(ctx, SyncCmdbNetworkAreaWorkflowName, scheduler.Every+"10s")
}

// ScheduleSyncAllAgentStateFromGSE creates a new schedule workflow to sync agent state from GSE.
func (mgr *Manager) ScheduleSyncAllAgentStateFromGSE(ctx contextx.IContext) error {
	return mgr.createAndRunScheduleWorkflows(ctx, SyncGseAgentStateWorkflowName, scheduler.Every+"10m")
}

// ScheduleWatchAndApplyCMDBResource creates a new schedule workflow to watch and apply CMDB resources.
func (mgr *Manager) ScheduleWatchAndApplyCMDBResource(ctx contextx.IContext) error {
	return mgr.createAndRunScheduleWorkflows(ctx, WatchAndApplyCMDBResourceWorkflowName, scheduler.Every+"10s")
}

// ScheduleSyncAliveHostAgentInfo creates a new schedule workflow to sync alive host agent info.
func (mgr *Manager) ScheduleSyncAliveHostAgentInfo(ctx contextx.IContext) error {
	return mgr.createAndRunScheduleWorkflows(ctx, SyncAliveHostAgentInfoWorkflowName, scheduler.Every+"10h")
}
