// Tencent is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
// Copyright (C) 2017 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT

package workflow

import (
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoOperation "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	packageworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/package-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	packageWorkflowRecentMonitoredTime       = 5 * time.Minute
	packageWorkflowMissingOperationGraceTime = 1 * time.Minute
)

func (s *Storage) registerPackageWorkflowScheduler() error {
	if s.Scheduler == nil {
		return errors.New("scheduler is not initialized")
	}

	if err := s.Scheduler.RegisterTask(scheduler.NewTask(
		"obtain_monitored_package_workflows",
		5*time.Second,  // nolint:mnd
		20*time.Second, // nolint:mnd
		s.obtainMonitoredPackageWorkflows,
	)); err != nil {
		return fmt.Errorf("register obtain monitored package workflows task failed: %w", err)
	}

	if err := s.Scheduler.RegisterTask(scheduler.NewTask(
		"monitor_package_workflow_status",
		1*time.Second,  // nolint:mnd
		10*time.Second, // nolint:mnd
		s.monitorPackageWorkflowStatus,
	)); err != nil {
		return fmt.Errorf("register monitor package workflow status task failed: %w", err)
	}

	return nil
}

func (s *Storage) obtainMonitoredPackageWorkflows(nCtx contextx.IContext) error {
	tenantIDs, err := tenant.ListTenantIDs(nCtx)
	if err != nil {
		return fmt.Errorf("failed to list tenant IDs: %w", err)
	}

	type tenantResult struct {
		running        []*types.PackageWorkflow
		recentFinished []*types.PackageWorkflow
	}
	results := make([]tenantResult, len(tenantIDs))

	gp := gopool.NewPool()
	for i, tenantID := range tenantIDs {
		slot := &results[i]
		tenantCtx := contextx.From(nCtx, contextx.WithTenantID(tenantID))
		fn := func() error {
			runningWorkflows, _, err := s.daoPackageWorkflow.List(
				tenantCtx,
				types.UnlimitedPage(),
				packageworkflow.WithStatus(types.PackageWorkflowStatusRunning))
			if err != nil {
				return fmt.Errorf("query running workflows failed: %w", err)
			}

			recentFinishedWorkflows, _, err := s.daoPackageWorkflow.List(
				tenantCtx,
				types.UnlimitedPage(),
				packageworkflow.WithStatus(types.GetFinishedPackageWorkflowStatus()...),
				packageworkflow.WithOperateTimeRange(types.RecentTimeRange(packageWorkflowRecentMonitoredTime)),
			)
			if err != nil {
				return fmt.Errorf("query recent finished workflows failed: %w", err)
			}

			slot.running = runningWorkflows
			slot.recentFinished = recentFinishedWorkflows

			return nil
		}

		gp.Go(fn)
	}

	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to obtain monitored workflows")
	}

	s.monitoredPackageWorkflowsMutex.Lock()
	s.monitoredPackageWorkflows = make(map[string]*types.PackageWorkflow)
	for i := range results {
		for _, workflow := range results[i].running {
			s.monitoredPackageWorkflows[workflow.TriggerID] = workflow
		}
	}

	// recentFinished has higher priority and overwrites running entries
	for i := range results {
		for _, workflow := range results[i].recentFinished {
			s.monitoredPackageWorkflows[workflow.TriggerID] = workflow
		}
	}

	s.monitoredPackageWorkflowsMutex.Unlock()

	return nil
}

// nolint: gocognit
func (s *Storage) monitorPackageWorkflowStatus(nCtx contextx.IContext) error {
	// Phase 1: RLock → deep copy snapshot → RUnlock
	s.monitoredPackageWorkflowsMutex.RLock()
	if len(s.monitoredPackageWorkflows) == 0 {
		s.monitoredPackageWorkflowsMutex.RUnlock()
		return nil
	}
	snapshot := make(map[string]*types.PackageWorkflow, len(s.monitoredPackageWorkflows))
	maps.Copy(snapshot, s.monitoredPackageWorkflows)
	s.monitoredPackageWorkflowsMutex.RUnlock()

	// Phase 2: work with snapshot, no lock held
	operations, _, err := s.daoOperation.List(nCtx,
		types.UnlimitedPage(),
		daoOperation.WithTriggerID(conv.MapKeyToSlice(snapshot)...))
	if err != nil {
		return fmt.Errorf("query operations failed: %w", err)
	}
	unfinishedTriggerMap := make(map[string]struct{})
	triggerOpers := make(map[string][]*operation.Operation)
	for _, oper := range operations {
		triggerOpers[oper.TriggerID] = append(triggerOpers[oper.TriggerID], oper)
		if oper.LatestInstBriefData == nil || oper.LatestInstBriefData.Lifecycle == nil ||
			!operation.CheckStateFinished(oper.LatestInstBriefData.Lifecycle.State) {

			unfinishedTriggerMap[oper.TriggerID] = struct{}{}
		}
	}

	triggersToDelete := make([]string, 0)
	for triggerID, packageWorkflow := range snapshot {
		opers, hasOperations := triggerOpers[triggerID]
		if !hasOperations || len(opers) == 0 {
			// If the workflow is already in a finished state, its operations may have been cleaned
			// up by a scheduled purge job. Do not overwrite a legitimate terminal status.
			if packageWorkflow.Status != types.PackageWorkflowStatusRunning {
				triggersToDelete = append(triggersToDelete, triggerID)
				continue
			}

			if !packageWorkflow.OperateTime.IsZero() && time.Since(packageWorkflow.OperateTime) < packageWorkflowMissingOperationGraceTime {
				continue
			}

			if err := s.updatePackageWorkflowResult(nCtx, packageWorkflow, types.PackageWorkflowStatusFailed, time.Now()); err != nil {
				logger.G.Sys().WithErr(err).
					With("trigger-id", triggerID, "workflow-id", packageWorkflow.WorkflowID).
					Error("failed to fallback update package workflow status when operation is missing")

				continue
			}

			logger.G.Sys().With("trigger-id", triggerID, "workflow-id", packageWorkflow.WorkflowID).
				Warn("workflow has no operation after grace time, fallback status to failed")

			triggersToDelete = append(triggersToDelete, triggerID)

			continue
		}

		if _, unfinished := unfinishedTriggerMap[triggerID]; unfinished {
			continue
		}
		status, finishTime, zeroEndTime := calPackageWorkflowStatusAndTime(opers)
		if zeroEndTime {
			logger.G.Sys().With("trigger-id", triggerID).
				Warn("all operation instances have zero end time, using current time as fallback")
		}

		if err := s.updatePackageWorkflowResult(nCtx, packageWorkflow, status, finishTime); err != nil {
			logger.G.Sys().WithErr(err).
				With("trigger-id", triggerID, "workflow-id", packageWorkflow.WorkflowID).
				Error("failed to update package workflow status and finish time")

			continue
		}

		triggersToDelete = append(triggersToDelete, triggerID)
	}

	// Phase 3: Lock → batch delete → Unlock (only if needed)
	if len(triggersToDelete) > 0 {
		s.monitoredPackageWorkflowsMutex.Lock()
		for _, triggerID := range triggersToDelete {
			delete(s.monitoredPackageWorkflows, triggerID)
		}
		s.monitoredPackageWorkflowsMutex.Unlock()
	}

	return nil
}

func (s *Storage) updatePackageWorkflowResult(
	nCtx contextx.IContext, workflow *types.PackageWorkflow, status types.PackageWorkflowStatus, finishTime time.Time,
) error {

	updateCtx, cancel := contextx.WithTimeout(
		contextx.From(nCtx, contextx.WithTenantID(workflow.TenantID)),
		30*time.Second, // nolint:mnd
	)
	defer cancel()

	if err := s.daoPackageWorkflow.UpdateStatus(updateCtx, workflow.WorkflowID, status); err != nil {
		return fmt.Errorf("update package workflow status failed: %w", err)
	}

	if err := s.daoPackageWorkflow.UpdateFinishTime(updateCtx, workflow.WorkflowID, finishTime); err != nil {
		return fmt.Errorf("update package workflow finish time failed: %w", err)
	}

	return nil
}

func calPackageWorkflowStatusAndTime(opers []*operation.Operation) (types.PackageWorkflowStatus, time.Time, bool) {
	successCount := 0
	failedCount := 0
	var latestEndTime time.Time

	for _, oper := range opers {
		if oper.LatestInstBriefData == nil || oper.LatestInstBriefData.Lifecycle == nil {
			continue
		}
		lifecycle := oper.LatestInstBriefData.Lifecycle
		if lifecycle.EndedAt.After(latestEndTime) {
			latestEndTime = lifecycle.EndedAt
		}
		switch lifecycle.State {
		case operation.StateSuccess:
			successCount++
		case operation.StateFailed, operation.StateTimeout:
			failedCount++
		default:
		}
	}

	zeroEndTime := latestEndTime.IsZero()
	if zeroEndTime {
		latestEndTime = time.Now()
	}

	total := len(opers)
	switch {
	case successCount == total:
		return types.PackageWorkflowStatusSuccess, latestEndTime, zeroEndTime
	case failedCount == total:
		return types.PackageWorkflowStatusFailed, latestEndTime, zeroEndTime
	default:
		return types.PackageWorkflowStatusPartialFailed, latestEndTime, zeroEndTime
	}
}
