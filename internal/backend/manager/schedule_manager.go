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
	"context"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	// SyncCmdbHostWorkflowName defines the name of the sync host workflow.
	SyncCmdbHostWorkflowName = "schedule_sync_cmdb_host"
)

// ScheduleWorkflowFunc defines the function type for scheduling workflows.
type ScheduleWorkflowFunc func(ctx context.Context) error

// registerScheduleWorkflow registers a schedule workflow function.
func (mgr *manager) getScheduleWorkflow() map[string]ScheduleWorkflowFunc {
	return map[string]ScheduleWorkflowFunc{
		SyncCmdbHostWorkflowName: mgr.ScheduleSyncHostFromCMDB,
	}
}

// startScheduleWorkflow starts the scheduled workflows.
func (mgr *manager) startScheduleWorkflow(ctx context.Context) error {
	dbScheduleWorkflows, _, err := mgr.conf.StorageSchedule.ListScheduleWorkflow(ctx, types.UnlimitedPage(),
		&types.ScheduleWorkflowCondition{ExactInclude: &types.ScheduleWorkflowExactFields{}})
	if err != nil {
		return fmt.Errorf("failed to list schedule workflows from storage, err: %w", err)
	}

	dbScheduleWorkflowsMap, err := conv.SliceToMap(dbScheduleWorkflows, func(swo *types.ScheduleWorkflow) string {
		return swo.WorkflowName
	})
	if err != nil {
		return fmt.Errorf("failed to convert schedule workflows to map, err: %w", err)
	}

	scheduleWorkflowMaps := mgr.getScheduleWorkflow()
	unRegiteredWorkflows := make([]string, 0, len(mgr.getScheduleWorkflow()))
	for workflowName := range scheduleWorkflowMaps {
		dbSchedule, ok := dbScheduleWorkflowsMap[workflowName]
		if !ok {
			unRegiteredWorkflows = append(unRegiteredWorkflows, workflowName)
			continue
		}

		triggerCtl, err := mgr.workflowMgr.GetTrigger(ctx, dbSchedule.TriggerID)
		if err != nil {
			mgr.logger.Errorf("get schedule workflow(%s) trigger failed, err: %v", dbSchedule.WorkflowID, err)
			return fmt.Errorf("get schedule workflow(%s) trigger failed, err: %w", dbSchedule.WorkflowID, err)
		}

		err = triggerCtl.RunTrigger(ctx)
		if err != nil {
			mgr.logger.Errorf("run schedule workflow(%s) trigger(%s) failed, err: %v",
				dbSchedule.WorkflowID, dbSchedule.TriggerID, err)

			return fmt.Errorf("run schedule workflow(%s) trigger(%s) failed, err: %w",
				dbSchedule.WorkflowID, dbSchedule.TriggerID, err)
		}
	}

	for _, workflowName := range unRegiteredWorkflows {
		scheduleFn := scheduleWorkflowMaps[workflowName]
		if scheduleFn == nil {
			mgr.logger.Warnf("workflow-name(%s)'s function not found, skip", workflowName)
			continue
		}

		err := scheduleFn(ctx)
		if err != nil {
			mgr.logger.Errorf("register workflow-name(%s) failed, err: %v", workflowName, err)
			return fmt.Errorf("register workflow-name(%s) failed, err: %w", workflowName, err)
		}
	}

	return nil
}

// ScheduleSyncHostFromCMDB creates a new schedule workflow to sync hosts from CMDB.
func (mgr *manager) ScheduleSyncHostFromCMDB(ctx context.Context) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		mgr.logger.Errorf("get tenant id failed, err: %v", err)
		return fmt.Errorf("get tenant id failed, err: %w", err)
	}

	metadataPeriodic, err := trigger.NewMetadataPeriodic(scheduler.Midnight, false)
	if err != nil {
		mgr.logger.Errorf("create periodic metadata failed, err: %v", err)
		return fmt.Errorf("create periodic metadata failed, err: %w", err)
	}

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryPeriodic, metadataPeriodic)
	if err != nil {
		mgr.logger.Errorf("create periodic trigger failed, err: %v", err)
		return fmt.Errorf("create periodic trigger failed, err: %w", err)
	}

	scheduleWf := &types.ScheduleWorkflow{
		WorkflowID:   identifier.GenWorkflowID(),
		WorkflowName: SyncCmdbHostWorkflowName,
		TriggerID:    triggerCtl.GetTriggerID(),
		Operator:     runtime.SystemName,
		OperateTime:  time.Now(),
	}

	err = mgr.conf.StorageSchedule.CreateScheduleWorkflow(ctx, scheduleWf)
	if err != nil {
		mgr.logger.Errorf("create schedule workflow failed, err: %v", err)
		return fmt.Errorf("create schedule workflow failed, err: %w", err)
	}

	operationDef := schedule.NewOperScheduleSyncHostOperation(
		schedule.OperParamScheduleSyncHostOperation{TenantID: tenantID})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		mgr.logger.Errorf("create operation for schedule workflow(%s) failed, err: %v", scheduleWf.WorkflowID, err)
		return fmt.Errorf("create operation for schedule workflow(%s) failed, err: %w", scheduleWf.WorkflowID, err)
	}

	mgr.logger.InfoCtxf(ctx, "new schedule sync host task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	err = triggerCtl.RunTrigger(ctx)
	if err != nil {
		mgr.logger.Errorf("run schedule workflow(%s) failed, err: %v", scheduleWf.WorkflowID, err)
		return fmt.Errorf("run schedule workflow(%s) failed, err: %w", scheduleWf.WorkflowID, err)
	}

	return nil
}
