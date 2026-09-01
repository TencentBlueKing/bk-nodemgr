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

package pkg

import (
	"errors"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoOperation "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	packagedeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/package-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/package-export"
	packageworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/package-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "pkg"

	packageWorkflowRecentMonitoredTime       = 5 * time.Minute
	packageWorkflowMissingOperationGraceTime = 1 * time.Minute

	metricOperationCreatePackageWorkflow    = "create_package_workflow"
	metricOperationGetPackageWorkflow       = "get_package_workflow"
	metricOperationCreatePackageDeployment  = "create_pkg_deployment"
	metricOperationListPackageDeployment    = "list_pkg_deployment"
	metricOperationGetPackageDeploymentInfo = "get_pkg_deployment_info"
	metricOperationUpdatePackageDeployment  = "update_pkg_deployment_info"
	metricOperationListPackageExport        = "list_package_export"
	metricOperationCreatePackageExport      = "create_package_export"
	metricOperationGetPackageExport         = "get_package_export"
	metricOperationDeletePackageExport      = "delete_package_export"
)

// NewStorage creates a new package storage.
func NewStorage(client *mongo.Client, database string) (IStorage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:      StorageName,
			Database:  client.Database(database),
			Scheduler: scheduler.NewScheduler(),
		},
	}
	if err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check)); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	return s, nil
}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoPackageWorkflow   packageworkflow.IHandler
	daoPackageDeployment packagedeployment.IHandler
	daoPackageExport     packageexport.IHandler
	daoOperation         daoOperation.IHandler

	// key is triggerID, value is the package workflow that needs to be monitored.
	monitoredPackageWorkflows      map[string]*types.PackageWorkflow
	monitoredPackageWorkflowsMutex sync.RWMutex
}

func (s *Storage) initDao() error {
	s.daoPackageWorkflow = packageworkflow.New(s.Database)
	s.daoPackageDeployment = packagedeployment.New(s.Database)
	s.daoPackageExport = packageexport.New(s.Database)
	s.daoOperation = daoOperation.New(s.Database)
	s.monitoredPackageWorkflows = make(map[string]*types.PackageWorkflow)

	if err := s.registerPackageWorkflowScheduler(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register package workflow scheduler")

		return fmt.Errorf("failed to register package workflow scheduler: %w", err)
	}

	return nil
}

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

// obtainMonitoredPackageWorkflows obtains the list of package workflows that need to be monitored.
func (s *Storage) obtainMonitoredPackageWorkflows(nCtx contextx.IContext) error {
	tenantIDs, err := tenant.ListEnabledTenantIDs(nCtx)
	if err != nil {
		return fmt.Errorf("failed to list enabled tenant IDs: %w", err)
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
	// Phase 1: RLock -> deep copy snapshot -> RUnlock
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

	// Phase 3: Lock -> batch delete -> Unlock (only if needed)
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

func (s *Storage) check() error {
	if s.daoPackageWorkflow == nil {
		return errors.New("dao package workflow is nil")
	}

	if s.daoPackageDeployment == nil {
		return errors.New("dao package deployment is nil")
	}

	if s.daoPackageExport == nil {
		return errors.New("dao package export is nil")
	}

	if s.daoOperation == nil {
		return errors.New("dao operation is nil")
	}

	return nil
}

// CreatePackageWorkflow creates a package workflow record.
func (s *Storage) CreatePackageWorkflow(nCtx contextx.IContext, workflow *types.PackageWorkflow) error {
	workflowID := ""
	if workflow != nil {
		workflowID = workflow.WorkflowID
	}

	return s.WrapFn(nCtx, metricOperationCreatePackageWorkflow, func(nCtx contextx.IContext) error {
		if err := s.createPackageWorkflow(nCtx, workflow); err != nil {
			logger.G.Sys().WithErr(err).With("workflow-id", workflowID).Error("failed to create package workflow")

			return fmt.Errorf("failed to create package workflow, workflow-id(%s): %w", workflowID, err)
		}

		return nil
	})
}

// GetPackageWorkflow gets a package workflow by workflow ID.
func (s *Storage) GetPackageWorkflow(nCtx contextx.IContext, workflowID string) (*types.PackageWorkflow, error) {
	var workflow *types.PackageWorkflow
	err := s.WrapFn(nCtx, metricOperationGetPackageWorkflow, func(nCtx contextx.IContext) error {
		var err error
		workflow, err = s.getPackageWorkflow(nCtx, workflowID)
		if err != nil {
			logger.G.Sys().WithErr(err).With("workflow-id", workflowID).Error("failed to get package workflow")

			return fmt.Errorf("failed to get package workflow, workflow-id(%s): %w", workflowID, err)
		}

		return nil
	})

	return workflow, err
}

// CreatePackageDeployment creates a package deployment record.
func (s *Storage) CreatePackageDeployment(nCtx contextx.IContext, deployment *types.PackageDeployment) error {
	if deployment == nil {
		return errors.New("package deployment is nil")
	}

	return s.WrapFn(nCtx, metricOperationCreatePackageDeployment, func(nCtx contextx.IContext) error {
		if err := s.createPackageDeployment(nCtx, deployment); err != nil {
			logger.G.Sys().WithErr(err).With("token", deployment.Token).Error("failed to create package deployment")

			return err
		}

		return nil
	})
}

// ListPackageDeployment lists package deployment records.
func (s *Storage) ListPackageDeployment(nCtx contextx.IContext, page types.Page, conditions ...*types.PackageDeploymentCondition) (
	[]*types.PackageDeployment, int64, error) {

	var deployments []*types.PackageDeployment
	var count int64
	err := s.WrapFn(nCtx, metricOperationListPackageDeployment, func(nCtx contextx.IContext) error {
		var err error
		deployments, count, err = s.listPackageDeployment(nCtx, page, conditions...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list package deployments")

			return err
		}

		return nil
	})

	return deployments, count, err
}

// GetPackageDeploymentInfo gets package deployment info by token.
func (s *Storage) GetPackageDeploymentInfo(nCtx contextx.IContext, token string) (*types.PackageDeploymentInfo, error) {
	var info *types.PackageDeploymentInfo
	err := s.WrapFn(nCtx, metricOperationGetPackageDeploymentInfo, func(nCtx contextx.IContext) error {
		var err error
		info, err = s.getPackageDeploymentInfo(nCtx, token)
		if err != nil {
			logger.G.Sys().WithErr(err).With("token", token).Error("failed to get package deployment info")

			return err
		}

		return nil
	})

	return info, err
}

// UpdatePackageDeploymentInfo updates package deployment info by token.
func (s *Storage) UpdatePackageDeploymentInfo(nCtx contextx.IContext, token string, info *types.PackageDeploymentInfo) error {
	return s.WrapFn(nCtx, metricOperationUpdatePackageDeployment, func(nCtx contextx.IContext) error {
		if err := s.updatePackageDeploymentInfo(nCtx, token, info); err != nil {
			logger.G.Sys().WithErr(err).With("token", token).Error("failed to update package deployment info")

			return err
		}

		return nil
	})
}

// ListPackageExport lists package export records.
func (s *Storage) ListPackageExport(
	nCtx contextx.IContext, page types.Page, conditions ...*types.PackageExportCondition) (
	[]*types.PackageExport, int64, error) {

	var exports []*types.PackageExport
	var count int64
	err := s.WrapFn(nCtx, metricOperationListPackageExport, func(nCtx contextx.IContext) error {
		var err error
		exports, count, err = s.listPackageExport(nCtx, page, conditions...)

		return err
	})

	return exports, count, err
}

// CreatePackageExport creates a package export record.
func (s *Storage) CreatePackageExport(nCtx contextx.IContext, exportData *types.PackageExport) error {
	return s.WrapFn(nCtx, metricOperationCreatePackageExport, func(nCtx contextx.IContext) error {
		return s.createPackageExport(nCtx, exportData)
	})
}

// GetPackageExport gets a package export by export ID.
func (s *Storage) GetPackageExport(nCtx contextx.IContext, exportID string) (*types.PackageExport, error) {
	var exportData *types.PackageExport
	err := s.WrapFn(nCtx, metricOperationGetPackageExport, func(nCtx contextx.IContext) error {
		var err error
		exportData, err = s.getPackageExport(nCtx, exportID)

		return err
	})

	return exportData, err
}

// DeletePackageExport deletes a package export by export ID.
func (s *Storage) DeletePackageExport(nCtx contextx.IContext, exportID string) error {
	return s.WrapFn(nCtx, metricOperationDeletePackageExport, func(nCtx contextx.IContext) error {
		return s.deletePackageExport(nCtx, exportID)
	})
}
