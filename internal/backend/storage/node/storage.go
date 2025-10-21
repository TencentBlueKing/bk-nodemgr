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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoBase "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoNodeDeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-deployment"
	daoNodeWorkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operinstdata"
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
	daoOperInstData   operinstdata.IHandler

	monitoredWorkflows      map[string]*types.NodeWorkflow
	monitoredWorkflowsMutex sync.RWMutex
}

func (s *Storage) initDao() error {
	s.daoNodeWorkflow = daoNodeWorkflow.New(s.Database)
	s.daoOperInstData = operinstdata.New(s.Database)
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

	var (
		runningWorkflowMap        map[string]*types.NodeWorkflow
		recentFinishedWorkflowMap map[string]*types.NodeWorkflow
	)

	gp := gopool.NewPool()
	for _, tenantID := range tenantIDs {
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
				daoBase.WithUpdateAtTimeRange(types.RecentTimeRange(recentMonitoredTime)),
			)
			if err != nil {
				return fmt.Errorf("query recent finished workflows failed: %w", err)
			}

			// we can sure that the workflow id is unique, so we can use map in concurrency.
			for _, workflow := range runningWorkflows {
				runningWorkflowMap[workflow.WorkflowID] = workflow
			}

			for _, workflow := range recentFinishedWorkflows {
				recentFinishedWorkflowMap[workflow.WorkflowID] = workflow
			}

			return nil
		}

		gp.Go(fn)
	}

	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to obtain monitored workflows")
	}

	s.monitoredWorkflowsMutex.Lock()

	s.monitoredWorkflows = make(map[string]*types.NodeWorkflow, len(s.monitoredWorkflows))
	for _, workflow := range runningWorkflowMap {
		s.monitoredWorkflows[workflow.TriggerID] = workflow
	}

	// notice: When these maps intersect,
	// you need to ensure that the workflow in the recentFinishedWorkflowMap has a higher priority
	for _, workflow := range recentFinishedWorkflowMap {
		s.monitoredWorkflows[workflow.TriggerID] = workflow
	}

	s.monitoredWorkflowsMutex.Unlock()

	return nil
}

func (s *Storage) monitorWorkflowStatus(nCtx contextx.IContext) error {
	s.monitoredWorkflowsMutex.RLock()
	defer s.monitoredWorkflowsMutex.RUnlock()

	if len(s.monitoredWorkflows) == 0 {
		return nil
	}

	operInst, err := s.daoOperInstData.ListAllLastOperInst(nCtx,
		operinstdata.WithTriggerID(conv.MapKeyToSlice(s.monitoredWorkflows)...))
	if err != nil {
		return fmt.Errorf("query last operation instance failed: %w", err)
	}

	unfinishedTriggerMap := make(map[string]struct{}, len(operInst))
	for _, inst := range operInst {
		if !operation.CheckStateFinished(inst.Lifecycle.State) {
			unfinishedTriggerMap[inst.Metadata.TriggerID] = struct{}{}
		}
	}

	finishedTriggerOperInstsMap := make(map[string][]*operation.InstanceBriefData, len(operInst))
	for _, inst := range operInst {
		if _, ok := unfinishedTriggerMap[inst.Metadata.TriggerID]; !ok {
			finishedTriggerOperInstsMap[inst.Metadata.TriggerID] =
				append(finishedTriggerOperInstsMap[inst.Metadata.TriggerID], inst)
		}
	}

	for triggerID, operInsts := range finishedTriggerOperInstsMap {
		status, finishTime := calWorkflowStatusAndTime(operInsts)

		err = s.daoNodeWorkflow.UpdateStatus(nCtx, s.monitoredWorkflows[triggerID].WorkflowID, status)
		if err != nil {
			return fmt.Errorf("update node workflow status failed: %w", err)
		}

		err = s.daoNodeWorkflow.UpdateFinishTime(
			nCtx, s.monitoredWorkflows[triggerID].WorkflowID, finishTime)
		if err != nil {
			return fmt.Errorf("update node workflow finish time failed: %w", err)
		}

		delete(s.monitoredWorkflows, triggerID)
	}

	return nil
}

func calWorkflowStatusAndTime(operationInsts []*operation.InstanceBriefData) (types.NodeWorkflowStatus, time.Time) {
	successCount := 0
	failedCount := 0
	var latestEndTime time.Time

	for _, inst := range operationInsts {
		if inst.Lifecycle.EndedAt.After(latestEndTime) {
			latestEndTime = inst.Lifecycle.EndedAt
		}

		switch inst.Lifecycle.State {
		case operation.StateSuccess:
			successCount++
		case operation.StateFailed, operation.StateTimeout:
			failedCount++
		default:
		}
	}

	total := len(operationInsts)
	switch {
	case successCount == total:
		return types.NodeWorkflowStatusSuccess, latestEndTime
	case failedCount == total:
		return types.NodeWorkflowStatusFailed, latestEndTime
	default:
		return types.NodeWorkflowStatusPartialFailed, latestEndTime
	}
}

func (s *Storage) check() error {
	if s.daoNodeWorkflow == nil {
		return errors.New("dao node workflow is nil")
	}

	if s.daoNodeDeployment == nil {
		return errors.New("dao node deployment is nil")
	}

	if s.daoOperInstData == nil {
		return errors.New("dao operation instance data is nil")
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}

// GetNodeDeploymentNodeConf get gse node conf.
func (s *Storage) GetNodeDeploymentNodeConf(nCtx contextx.IContext, token string) (*types.NodeConf, error) {
	var (
		nodeConf *types.NodeConf
		err      error
	)

	// record metric.
	metric := s.metric().Start("get_node_deployment_node_conf")
	defer metric.End(err)

	nodeConf, err = s.getNodeDeploymentNodeConf(nCtx, token)

	return nodeConf, err
}

// GetNodeDeploymentInfo get node deployment info.
func (s *Storage) GetNodeDeploymentInfo(nCtx contextx.IContext, token string) (*types.DeploymentInfo, error) {
	var (
		deployInfo *types.DeploymentInfo
		err        error
	)

	// record metric.
	metric := s.metric().Start("get_node_deployment_info")
	defer metric.End(err)

	deployInfo, err = s.getNodeDeploymentInfo(nCtx, token)

	return deployInfo, err
}

// SetNodeDeploymentNodeConf set gse node conf.
func (s *Storage) SetNodeDeploymentNodeConf(nCtx contextx.IContext, token string, conf *types.NodeConf) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("set_node_deployment_node_conf")
	defer metric.End(err)

	err = s.seNodeDeploymenttNodeConf(nCtx, token, conf)

	return err
}

// UpdateNodeDeploymentInfo update node deployment info.
func (s *Storage) UpdateNodeDeploymentInfo(nCtx contextx.IContext, token string, info *types.DeploymentInfo) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update_node_deployment_info")
	defer metric.End(err)

	err = s.updateNodeDeploymentInfo(nCtx, token, info)

	return err
}

// CreateNodeDeployment createNodeDeployment a node deployment.
func (s *Storage) CreateNodeDeployment(nCtx contextx.IContext, nodeDeployment *types.NodeDeployment) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("create_node_deployment")
	defer metric.End(err)

	err = s.createNodeDeployment(nCtx, nodeDeployment)

	return err
}

// ListNodeWorkflow lists node workflow by page and conditions.
func (s *Storage) ListNodeWorkflow(nCtx contextx.IContext, page types.Page, conditions ...*types.NodeWorkflowCondition) (
	[]*types.NodeWorkflow, int64, error) {

	var (
		results []*types.NodeWorkflow
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_node_workflow")
	defer metric.End(err)

	results, num, err = s.listNodeWorkflow(nCtx, page, conditions...)

	return results, num, err
}

// CountNodeWorkflow counts node workflow by conditions.
func (s *Storage) CountNodeWorkflow(nCtx contextx.IContext, conditions ...*types.NodeWorkflowCondition) (int64, error) {
	var num int64
	var err error

	// record metric.
	metric := s.metric().Start("count_node_workflow")
	defer metric.End(err)

	num, err = s.countNodeWorkflow(nCtx, conditions...)

	return num, err
}

// DistinctNodeWorkflow distincts node workflow fields.
func (s *Storage) DistinctNodeWorkflow(
	nCtx contextx.IContext, request types.NodeWorkflowDistinctRequest, conditions ...*types.NodeWorkflowCondition) (
	*types.NodeWorkflowDistinctResult, error) {

	var (
		result *types.NodeWorkflowDistinctResult
		err    error
	)

	// record metric.
	metric := s.metric().Start("distinct_node_workflow")
	defer metric.End(err)

	result, err = s.distinctNodeWorkflow(nCtx, request, conditions...)

	return result, err
}

// GetNodeWorkflow gets a node workflow by workflow-id.
func (s *Storage) GetNodeWorkflow(nCtx contextx.IContext, workflowID string) (*types.NodeWorkflow, error) {
	var (
		nodeWorkflow *types.NodeWorkflow
		err          error
	)

	// record metric.
	metric := s.metric().Start("get_node_workflow")
	defer metric.End(err)

	nodeWorkflow, err = s.getNodeWorkflow(nCtx, workflowID)

	return nodeWorkflow, err
}

// CreateNodeWorkflow creates a new node workflow.
func (s *Storage) CreateNodeWorkflow(nCtx contextx.IContext, workflow *types.NodeWorkflow) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("create_node_workflow")
	defer metric.End(err)

	err = s.createNodeWorkflow(nCtx, workflow)

	return err
}

// UpdateNodeWorkflowStatus updates the status of a node workflow.
func (s *Storage) UpdateNodeWorkflowStatus(nCtx contextx.IContext, workflowID string, status types.NodeWorkflowStatus) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update_node_workflow_status")
	defer metric.End(err)

	err = s.updateNodeWorkflowStatus(nCtx, workflowID, status)

	return err
}
