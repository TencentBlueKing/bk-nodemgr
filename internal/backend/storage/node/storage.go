/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package node provide storage for node.
package node

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoNodeDeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-deployment"
	daoNodeWorkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-workflow"
	daoOperation "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName defines the storage name.
const StorageName = "node"

const (
	recentMonitoredTime = 5 * time.Minute

	metricOperationGetNodeDeploymentNodeConf = "get_node_deployment_node_conf"
	metricOperationGetNodeDeploymentInfo     = "get_node_deployment_info"
	metricOperationListNodeDeployment        = "list_node_deployment"
	metricOperationSetNodeDeploymentNodeConf = "set_node_deployment_node_conf"
	metricOperationUpdateNodeDeploymentInfo  = "update_node_deployment_info"
	metricOperationCreateNodeDeployment      = "create_node_deployment"
	metricOperationListNodeWorkflow          = "list_node_workflow"
	metricOperationCountNodeWorkflow         = "count_node_workflow"
	metricOperationDistinctNodeWorkflow      = "distinct_node_workflow"
	metricOperationGetNodeWorkflow           = "get_node_workflow"
	metricOperationCreateNodeWorkflow        = "create_node_workflow"
	metricOperationUpdateNodeWorkflowStatus  = "update_node_workflow_status"
)

// NewStorage creates a new node workflow storage handler.
func NewStorage(client *mongo.Client, database string) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
		},
		monitoredWorkflows:      make(map[string]*types.NodeWorkflow),
		monitoredWorkflowsMutex: sync.RWMutex{},
	}

	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	err = s.registerScheduler()
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register scheduler")

		return nil, fmt.Errorf("register scheduler failed: %w", err)
	}

	return s, nil
}

// Storage provides a node workflow storage handler.
type Storage struct {
	basestorage.Storage

	// dao
	daoNodeDeployment daoNodeDeployment.IHandler
	daoNodeWorkflow   daoNodeWorkflow.IHandler
	daoOperation      daoOperation.IHandler

	monitoredWorkflows      map[string]*types.NodeWorkflow
	monitoredWorkflowsMutex sync.RWMutex
}

func (s *Storage) initDao() error {
	s.daoNodeWorkflow = daoNodeWorkflow.New(s.Database)
	s.daoOperation = daoOperation.New(s.Database)
	s.daoNodeDeployment = daoNodeDeployment.New(s.Database)

	return nil
}

func (s *Storage) registerScheduler() error {
	s.Scheduler = scheduler.NewScheduler()
	err := s.Scheduler.RegisterTask(scheduler.NewTask(
		"obtain monitored workflows",
		5*time.Second,  // nolint: mnd
		20*time.Second, // nolint: mnd
		s.obtainMonitoredWorkflows,
	))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register obtain monitored workflows task")

		return fmt.Errorf("register obtain monitored workflows task failed: %w", err)
	}

	err = s.Scheduler.RegisterTask(scheduler.NewTask(
		"monitor workflow status",
		1*time.Second,  // nolint: mnd
		10*time.Second, // nolint: mnd
		s.monitorWorkflowStatus,
	))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register monitor workflow status task")

		return fmt.Errorf("register monitor workflow status task failed: %w", err)
	}

	return nil
}

// obtainMonitoredWorkflows Obtain a list of workflows that need to be listened to.
func (s *Storage) obtainMonitoredWorkflows(nCtx contextx.IContext) error {
	tenantIDs := tenant.GetAllTenantIDs()

	type tenantResult struct {
		running        []*types.NodeWorkflow
		recentFinished []*types.NodeWorkflow
	}
	results := make([]tenantResult, len(tenantIDs))

	gp := gopool.NewPool()
	for i, tenantID := range tenantIDs {
		slot := &results[i]
		tenantCtx := contextx.From(nCtx, contextx.WithTenantID(tenantID))
		fn := func() error {
			runningWorkflows, _, err := s.daoNodeWorkflow.List(
				tenantCtx,
				types.UnlimitedPage(),
				daoNodeWorkflow.WithStatus(types.NodeWorkflowStatusRunning))
			if err != nil {
				return fmt.Errorf("query running workflows failed: %w", err)
			}

			recentFinishedWorkflows, _, err := s.daoNodeWorkflow.List(
				tenantCtx,
				types.UnlimitedPage(),
				daoNodeWorkflow.WithStatus(types.GetFinishedNodeWorkflowStatus()...),
				daoNodeWorkflow.WithOperateTimeRange(types.RecentTimeRange(recentMonitoredTime)),
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

	s.monitoredWorkflowsMutex.Lock()

	s.monitoredWorkflows = make(map[string]*types.NodeWorkflow)
	for i := range results {
		for _, workflow := range results[i].running {
			s.monitoredWorkflows[workflow.TriggerID] = workflow
		}
	}

	// recentFinished has higher priority and overwrites running entries
	for i := range results {
		for _, workflow := range results[i].recentFinished {
			s.monitoredWorkflows[workflow.TriggerID] = workflow
		}
	}

	s.monitoredWorkflowsMutex.Unlock()

	return nil
}

// nolint: gocognit
func (s *Storage) monitorWorkflowStatus(nCtx contextx.IContext) error {
	// Phase 1: RLock → deep copy snapshot → RUnlock
	s.monitoredWorkflowsMutex.RLock()
	if len(s.monitoredWorkflows) == 0 {
		s.monitoredWorkflowsMutex.RUnlock()
		return nil
	}
	snapshot := make(map[string]*types.NodeWorkflow, len(s.monitoredWorkflows))
	for k, v := range s.monitoredWorkflows {
		snapshot[k] = v
	}
	s.monitoredWorkflowsMutex.RUnlock()

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
		if oper.LatestInstBriefData == nil || !operation.CheckStateFinished(oper.LatestInstBriefData.Lifecycle.State) {
			unfinishedTriggerMap[oper.TriggerID] = struct{}{}
		}
	}

	triggersToDelete := make([]string, 0)
	for triggerID, opers := range triggerOpers {
		if _, unfinished := unfinishedTriggerMap[triggerID]; unfinished {
			continue
		}
		status, finishTime, zeroEndTime := calWorkflowStatusAndTime(opers)
		if zeroEndTime {
			logger.G.Sys().With("trigger-id", triggerID).
				Warn("all operation instances have zero end time, using current time as fallback")
		}

		nodeWorkflow, ok := snapshot[triggerID]
		if !ok {
			continue
		}

		updateCtx, cancel := contextx.WithTimeout(
			contextx.From(contextx.Background(), contextx.WithTenantID(nodeWorkflow.TenantID)),
			30*time.Second, // nolint: mnd
		)

		err = s.daoNodeWorkflow.UpdateStatus(updateCtx, nodeWorkflow.WorkflowID, status)
		if err != nil {
			cancel()
			logger.G.Sys().WithErr(err).
				With("trigger-id", triggerID, "workflow-id", nodeWorkflow.WorkflowID).
				Error("failed to update node workflow status")

			continue
		}

		err = s.daoNodeWorkflow.UpdateFinishTime(updateCtx, nodeWorkflow.WorkflowID, finishTime)
		cancel()
		if err != nil {
			logger.G.Sys().WithErr(err).
				With("trigger-id", triggerID, "workflow-id", nodeWorkflow.WorkflowID).
				Error("failed to update node workflow finish time")

			continue
		}

		triggersToDelete = append(triggersToDelete, triggerID)
	}

	// Phase 3: Lock → batch delete → Unlock (only if needed)
	if len(triggersToDelete) > 0 {
		s.monitoredWorkflowsMutex.Lock()
		for _, triggerID := range triggersToDelete {
			delete(s.monitoredWorkflows, triggerID)
		}
		s.monitoredWorkflowsMutex.Unlock()
	}

	return nil
}

func calWorkflowStatusAndTime(opers []*operation.Operation) (types.NodeWorkflowStatus, time.Time, bool) {
	successCount := 0
	failedCount := 0
	var latestEndTime time.Time

	for _, oper := range opers {
		if oper.LatestInstBriefData == nil || oper.LatestInstBriefData.Lifecycle == nil {
			continue
		}
		lc := oper.LatestInstBriefData.Lifecycle
		if lc.EndedAt.After(latestEndTime) {
			latestEndTime = lc.EndedAt
		}

		switch lc.State {
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
		return types.NodeWorkflowStatusSuccess, latestEndTime, zeroEndTime
	case failedCount == total:
		return types.NodeWorkflowStatusFailed, latestEndTime, zeroEndTime
	default:
		return types.NodeWorkflowStatusPartialFailed, latestEndTime, zeroEndTime
	}
}

func (s *Storage) check() error {
	if s.daoNodeWorkflow == nil {
		return errors.New("dao node workflow is nil")
	}

	if s.daoNodeDeployment == nil {
		return errors.New("dao node deployment is nil")
	}

	if s.daoOperation == nil {
		return errors.New("dao operation is nil")
	}

	return nil
}

// GetNodeDeploymentNodeConf get gse node conf.
func (s *Storage) GetNodeDeploymentNodeConf(nCtx contextx.IContext, token string) (*types.NodeConf, error) {
	var nodeConf *types.NodeConf
	err := s.WrapFn(nCtx, metricOperationGetNodeDeploymentNodeConf, func(nCtx contextx.IContext) error {
		var err error
		nodeConf, err = s.getNodeDeploymentNodeConf(nCtx, token)

		return err
	})

	return nodeConf, err
}

// GetNodeDeploymentInfo get node deployment info.
func (s *Storage) GetNodeDeploymentInfo(nCtx contextx.IContext, token string) (*types.DeploymentInfo, error) {
	var deployInfo *types.DeploymentInfo
	err := s.WrapFn(nCtx, metricOperationGetNodeDeploymentInfo, func(nCtx contextx.IContext) error {
		var err error
		deployInfo, err = s.getNodeDeploymentInfo(nCtx, token)

		return err
	})

	return deployInfo, err
}

// ListNodeDeployment lists node deployment by page and conditions.
func (s *Storage) ListNodeDeployment(nCtx contextx.IContext, page types.Page, conditions ...*types.NodeDeploymentCondition) (
	[]*types.NodeDeployment, int64, error) {

	var nodeDeployments []*types.NodeDeployment
	var num int64
	err := s.WrapFn(nCtx, metricOperationListNodeDeployment, func(nCtx contextx.IContext) error {
		var err error
		nodeDeployments, num, err = s.listNodeDeployment(nCtx, page, conditions...)

		return err
	})

	return nodeDeployments, num, err
}

// SetNodeDeploymentNodeConf set gse node conf.
func (s *Storage) SetNodeDeploymentNodeConf(nCtx contextx.IContext, token string, conf *types.NodeConf) error {
	return s.WrapFn(nCtx, metricOperationSetNodeDeploymentNodeConf, func(nCtx contextx.IContext) error {
		return s.seNodeDeploymenttNodeConf(nCtx, token, conf)
	})
}

// UpdateNodeDeploymentInfo update node deployment info.
func (s *Storage) UpdateNodeDeploymentInfo(nCtx contextx.IContext, token string, info *types.DeploymentInfo) error {
	return s.WrapFn(nCtx, metricOperationUpdateNodeDeploymentInfo, func(nCtx contextx.IContext) error {
		return s.updateNodeDeploymentInfo(nCtx, token, info)
	})
}

// CreateNodeDeployment createNodeDeployment a node deployment.
func (s *Storage) CreateNodeDeployment(nCtx contextx.IContext, nodeDeployment *types.NodeDeployment) error {
	return s.WrapFn(nCtx, metricOperationCreateNodeDeployment, func(nCtx contextx.IContext) error {
		return s.createNodeDeployment(nCtx, nodeDeployment)
	})
}

// ListNodeWorkflow lists node workflow by page and conditions.
func (s *Storage) ListNodeWorkflow(nCtx contextx.IContext, page types.Page, conditions ...*types.NodeWorkflowCondition) (
	[]*types.NodeWorkflow, int64, error) {

	var results []*types.NodeWorkflow
	var num int64
	err := s.WrapFn(nCtx, metricOperationListNodeWorkflow, func(nCtx contextx.IContext) error {
		var err error
		results, num, err = s.listNodeWorkflow(nCtx, page, conditions...)

		return err
	})

	return results, num, err
}

// CountNodeWorkflow counts node workflow by conditions.
func (s *Storage) CountNodeWorkflow(nCtx contextx.IContext, conditions ...*types.NodeWorkflowCondition) (int64, error) {
	var num int64
	err := s.WrapFn(nCtx, metricOperationCountNodeWorkflow, func(nCtx contextx.IContext) error {
		var err error
		num, err = s.countNodeWorkflow(nCtx, conditions...)

		return err
	})

	return num, err
}

// DistinctNodeWorkflow distincts node workflow fields.
func (s *Storage) DistinctNodeWorkflow(
	nCtx contextx.IContext, request types.NodeWorkflowDistinctRequest, conditions ...*types.NodeWorkflowCondition) (
	*types.NodeWorkflowDistinctResult, error) {

	var result *types.NodeWorkflowDistinctResult
	err := s.WrapFn(nCtx, metricOperationDistinctNodeWorkflow, func(nCtx contextx.IContext) error {
		var err error
		result, err = s.distinctNodeWorkflow(nCtx, request, conditions...)

		return err
	})

	return result, err
}

// GetNodeWorkflow gets a node workflow by workflow-id.
func (s *Storage) GetNodeWorkflow(nCtx contextx.IContext, workflowID string) (*types.NodeWorkflow, error) {
	var nodeWorkflow *types.NodeWorkflow
	err := s.WrapFn(nCtx, metricOperationGetNodeWorkflow, func(nCtx contextx.IContext) error {
		var err error
		nodeWorkflow, err = s.getNodeWorkflow(nCtx, workflowID)

		return err
	})

	return nodeWorkflow, err
}

// CreateNodeWorkflow creates a new node workflow.
func (s *Storage) CreateNodeWorkflow(nCtx contextx.IContext, workflow *types.NodeWorkflow) error {
	return s.WrapFn(nCtx, metricOperationCreateNodeWorkflow, func(nCtx contextx.IContext) error {
		return s.createNodeWorkflow(nCtx, workflow)
	})
}

// UpdateNodeWorkflowStatus updates the status of a node workflow.
func (s *Storage) UpdateNodeWorkflowStatus(nCtx contextx.IContext, workflowID string, status types.NodeWorkflowStatus) error {
	return s.WrapFn(nCtx, metricOperationUpdateNodeWorkflowStatus, func(nCtx contextx.IContext) error {
		return s.updateNodeWorkflowStatus(nCtx, workflowID, status)
	})
}
