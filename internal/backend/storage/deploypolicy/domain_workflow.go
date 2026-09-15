/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package deploypolicy

import (
	"fmt"
	"maps"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/deploypolicy-workflow"
	daoOperation "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	recentMonitoredTime       = 5 * time.Minute
	missingOperationGraceTime = time.Minute
)

func (s *Storage) registerScheduler() error {
	s.Scheduler = scheduler.NewScheduler()
	err := s.Scheduler.RegisterTask(scheduler.NewTask(
		"obtain monitored workflows",
		5*time.Second,  // nolint: mnd
		20*time.Second, // nolint: mnd
		s.obtainMonitoredWorkflows,
	))
	if err != nil {
		return fmt.Errorf("failed to register obtain monitored workflows task: %w", err)
	}
	err = s.Scheduler.RegisterTask(scheduler.NewTask(
		"monitor workflow status",
		time.Second,
		10*time.Second, // nolint: mnd
		s.monitorWorkflowStatus,
	))
	if err != nil {
		return fmt.Errorf("failed to register monitor workflow status task: %w", err)
	}

	return nil
}

func (s *Storage) obtainMonitoredWorkflows(nCtx contextx.IContext) error {
	tenantIDs, err := tenant.ListEnabledTenantIDs(nCtx)
	if err != nil {
		return fmt.Errorf("failed to list enabled tenant IDs: %w", err)
	}
	type tenantResult struct {
		running        []*types.DeployPolicyWorkflow
		recentFinished []*types.DeployPolicyWorkflow
	}
	results := make([]tenantResult, len(tenantIDs))
	gp := gopool.NewPool()
	for i, tenantID := range tenantIDs {
		slot := &results[i]
		tenantCtx := contextx.From(nCtx, contextx.WithTenantID(tenantID))
		gp.Go(func() error {
			running, _, err := s.daoDeployPolicyWorkflow.List(tenantCtx, types.UnlimitedPage(),
				deploypolicyworkflow.WithRunningOrMissingStatus())
			if err != nil {
				return fmt.Errorf("failed to query running deploy policy workflows: %w", err)
			}
			recentFinished, _, err := s.daoDeployPolicyWorkflow.List(tenantCtx, types.UnlimitedPage(),
				deploypolicyworkflow.WithStatus(types.GetFinishedDeployPolicyWorkflowStatus()...),
				deploypolicyworkflow.WithOperateTimeRange(types.RecentTimeRange(recentMonitoredTime)))
			if err != nil {
				return fmt.Errorf("failed to query recent finished deploy policy workflows: %w", err)
			}
			slot.running = running
			slot.recentFinished = recentFinished

			return nil
		})
	}
	if err := gp.Wait(); err != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(err).Error("failed to obtain monitored deploy policy workflows")
	}

	s.monitoredWorkflowsMutex.Lock()
	defer s.monitoredWorkflowsMutex.Unlock()
	s.monitoredWorkflows = make(map[string]*types.DeployPolicyWorkflow)
	for _, result := range results {
		for _, workflow := range result.running {
			s.monitoredWorkflows[workflow.WorkflowID] = workflow
		}
	}
	// Finished records take priority if a workflow completes between the two queries.
	for _, result := range results {
		for _, workflow := range result.recentFinished {
			s.monitoredWorkflows[workflow.WorkflowID] = workflow
		}
	}

	return nil
}

func (s *Storage) monitorWorkflowStatus(nCtx contextx.IContext) error {
	s.monitoredWorkflowsMutex.RLock()
	snapshot := make(map[string]*types.DeployPolicyWorkflow, len(s.monitoredWorkflows))
	maps.Copy(snapshot, s.monitoredWorkflows)
	s.monitoredWorkflowsMutex.RUnlock()
	if len(snapshot) == 0 {
		return nil
	}

	operationIDs := make([]string, 0, len(snapshot))
	for _, workflow := range snapshot {
		operationIDs = append(operationIDs, workflow.OperationID)
	}
	opers, _, err := s.daoOperation.List(nCtx, types.UnlimitedPage(), daoOperation.WithOperationID(operationIDs...))
	if err != nil {
		return fmt.Errorf("failed to query deploy policy workflow operations: %w", err)
	}
	operations := make(map[string]*operation.Operation, len(opers))
	for _, oper := range opers {
		operations[oper.OperationID] = oper
	}

	workflowsToDelete := make([]string, 0)
	for workflowID, workflow := range snapshot {
		finished, err := s.refreshWorkflowResult(nCtx, workflow, operations[workflow.OperationID])
		if err != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(err).
				With("workflow-id", workflowID, "operation-id", workflow.OperationID, "tenant-id", workflow.TenantID).
				Error("failed to update deploy policy workflow result")

			continue
		}
		if finished {
			workflowsToDelete = append(workflowsToDelete, workflowID)
		}
	}
	if len(workflowsToDelete) > 0 {
		s.monitoredWorkflowsMutex.Lock()
		for _, workflowID := range workflowsToDelete {
			// Do not remove a newer snapshot installed by the refresh task.
			if s.monitoredWorkflows[workflowID] == snapshot[workflowID] {
				delete(s.monitoredWorkflows, workflowID)
			}
		}
		s.monitoredWorkflowsMutex.Unlock()
	}

	return nil
}

func (s *Storage) refreshWorkflowResult(nCtx contextx.IContext, workflow *types.DeployPolicyWorkflow,
	oper *operation.Operation) (bool, error) {

	status := types.DeployPolicyWorkflowStatusRunning
	var finishTime time.Time // Running workflows have no finish time.
	if oper == nil {
		// An empty status identifies a legacy record, not a terminal workflow.
		if workflow.Status != "" && workflow.Status != types.DeployPolicyWorkflowStatusRunning {
			return true, nil
		}
		if workflow.OperateTime.IsZero() || time.Since(workflow.OperateTime) >= missingOperationGraceTime {
			status = types.DeployPolicyWorkflowStatusFailed
			finishTime = time.Now()
			logger.G.Sys().Ctx(nCtx).With("workflow-id", workflow.WorkflowID, "operation-id", workflow.OperationID).
				Warn("deploy policy workflow has no operation after grace time, fallback status to failed")
		}
	} else {
		status, finishTime = calWorkflowStatusAndTime(oper)
	}
	if workflow.Status == status && workflow.FinishTime.Equal(finishTime) {
		return status != types.DeployPolicyWorkflowStatusRunning, nil
	}
	if err := s.updateDeployPolicyWorkflowResult(nCtx, workflow, status, finishTime); err != nil {
		return false, fmt.Errorf("failed to update deploy policy workflow result: %w", err)
	}

	return status != types.DeployPolicyWorkflowStatusRunning, nil
}

func calWorkflowStatusAndTime(oper *operation.Operation) (types.DeployPolicyWorkflowStatus, time.Time) {
	if oper.LatestInstBriefData == nil || oper.LatestInstBriefData.Lifecycle == nil ||
		!operation.CheckStateFinished(oper.LatestInstBriefData.Lifecycle.State) {

		return types.DeployPolicyWorkflowStatusRunning, time.Time{}
	}
	lifecycle := oper.LatestInstBriefData.Lifecycle
	finishTime := lifecycle.EndedAt
	if finishTime.IsZero() {
		finishTime = time.Now()
		logger.G.Sys().With("operation-id", oper.OperationID).
			Warn("operation instance has zero end time, using current time as fallback")
	}
	switch lifecycle.State {
	case operation.StateSuccess:
		return types.DeployPolicyWorkflowStatusSuccess, finishTime
	case operation.StateFailed, operation.StateTimeout:
		return types.DeployPolicyWorkflowStatusFailed, finishTime
	default:
		return types.DeployPolicyWorkflowStatusPartialFailed, finishTime
	}
}

func (s *Storage) updateDeployPolicyWorkflowResult(nCtx contextx.IContext, workflow *types.DeployPolicyWorkflow,
	status types.DeployPolicyWorkflowStatus, finishTime time.Time) error {

	updateCtx, cancel := contextx.WithTimeout(contextx.From(nCtx, contextx.WithTenantID(workflow.TenantID)),
		30*time.Second) // nolint: mnd
	defer cancel()
	if err := s.daoDeployPolicyWorkflow.UpdateStatus(updateCtx, workflow.WorkflowID, status); err != nil {
		return fmt.Errorf("failed to update deploy policy workflow status: %w", err)
	}
	if err := s.daoDeployPolicyWorkflow.UpdateFinishTime(updateCtx, workflow.WorkflowID, finishTime); err != nil {
		return fmt.Errorf("failed to update deploy policy workflow finish time: %w", err)
	}

	return nil
}
