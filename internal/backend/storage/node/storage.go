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
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoBase "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/nodedeployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/nodeworkflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
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
func NewStorage(client *mongo.Client, database string, logger logger.ILogger) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
		monitoredWorkflows:      make(map[string]*types.NodeWorkflow),
		monitoredWorkflowsMutex: sync.RWMutex{},
	}

	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	err = s.registerScheduler()
	if err != nil {
		s.Logger.Errorf("register scheduler failed, err: %v", err)
		return nil, fmt.Errorf("register scheduler failed, err: %w", err)
	}

	return s, nil
}

// Storage provides a node workflow storage handler.
type Storage struct {
	basestorage.Storage

	// dao
	daoNodeDeployment nodedeployment.IHandler
	daoNodeWorkflow   nodeworkflow.IHandler
	daoOperInstData   operinstdata.IHandler

	monitoredWorkflows      map[string]*types.NodeWorkflow
	monitoredWorkflowsMutex sync.RWMutex
}

func (s *Storage) initDao() error {
	s.daoNodeWorkflow = nodeworkflow.New(s.Database, s.Logger)
	s.daoOperInstData = operinstdata.New(s.Database, s.Logger)
	s.daoNodeDeployment = nodedeployment.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) registerScheduler() error {
	s.Scheduler = scheduler.NewScheduler(scheduler.WithLogger(s.Logger))
	err := s.Scheduler.RegisterTask(scheduler.NewTask(
		"obtain monitored workflows",
		5*time.Second,  // nolint: mnd
		20*time.Second, // nolint: mnd
		s.obtainMonitoredWorkflows,
	))
	if err != nil {
		s.Logger.Errorf("register obtain monitored workflows task failed, err: %v", err)
		return fmt.Errorf("register obtain monitored workflows task failed, err: %w", err)
	}

	err = s.Scheduler.RegisterTask(scheduler.NewTask(
		"monitor workflow status",
		1*time.Second,  // nolint: mnd
		10*time.Second, // nolint: mnd
		s.monitorWorkflowStatus,
	))
	if err != nil {
		s.Logger.Errorf("register monitor workflow status task failed, err: %v", err)
		return fmt.Errorf("register monitor workflow status task failed, err: %w", err)
	}

	return nil
}

// obtainMonitoredWorkflows Obtain a list of workflows that need to be listened to.
func (s *Storage) obtainMonitoredWorkflows(ctx contextx.IContext) error {
	runningWorkflows, _, err := s.daoNodeWorkflow.List(
		ctx,
		types.UnlimitedPage(),
		nodeworkflow.WithStatus(types.NodeWorkflowStatusRunning))
	if err != nil {
		return fmt.Errorf("query running workflows failed, err: %w", err)
	}

	recentFinishedWorkflows, _, err := s.daoNodeWorkflow.List(
		ctx,
		types.UnlimitedPage(),
		nodeworkflow.WithStatus(types.GetFinishedNodeWorkflowStatus()...),
		daoBase.WithUpdateAtTimeRange(types.RecentTimeRange(recentMonitoredTime)),
	)
	if err != nil {
		return fmt.Errorf("query recent finished workflows failed, err: %w", err)
	}

	convFn := func(workflows []*types.NodeWorkflow) (map[string]*types.NodeWorkflow, error) {
		return conv.SliceToMap(workflows, func(workflow *types.NodeWorkflow) string {
			return workflow.WorkflowID
		})
	}

	runningWorkflowMap, err := convFn(runningWorkflows)
	if err != nil {
		return fmt.Errorf("convert running workflows to map failed, err: %w", err)
	}

	recentFinishedWorkflowMap, err := convFn(recentFinishedWorkflows)
	if err != nil {
		return fmt.Errorf("convert recent finished workflows to map failed, err: %w", err)
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

func (s *Storage) monitorWorkflowStatus(ctx contextx.IContext) error {
	s.monitoredWorkflowsMutex.RLock()
	defer s.monitoredWorkflowsMutex.RUnlock()

	if len(s.monitoredWorkflows) == 0 {
		return nil
	}

	operInst, err := s.daoOperInstData.ListAllLastOperInst(ctx,
		operinstdata.WithTriggerID(conv.MapKeyToSlice(s.monitoredWorkflows)...))
	if err != nil {
		return fmt.Errorf("query last operation instance failed, err: %w", err)
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

		err = s.daoNodeWorkflow.UpdateStatus(ctx, s.monitoredWorkflows[triggerID].WorkflowID, status)
		if err != nil {
			return fmt.Errorf("update node workflow status failed, err: %w", err)
		}

		err = s.daoNodeWorkflow.UpdateFinishTime(
			ctx, s.monitoredWorkflows[triggerID].WorkflowID, finishTime)
		if err != nil {
			return fmt.Errorf("update node workflow finish time failed, err: %w", err)
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
func (s *Storage) GetNodeDeploymentNodeConf(ctx context.Context, token string) (*types.NodeConf, error) {
	var (
		nodeConf *types.NodeConf
		err      error
	)

	// record metric.
	metric := s.metric().Start("get_node_deployment_node_conf")
	defer metric.End(err)

	nodeConf, err = s.getNodeDeploymentNodeConf(ctx, token)

	return nodeConf, err
}

// GetNodeDeploymentInfo get node deployment info.
func (s *Storage) GetNodeDeploymentInfo(ctx context.Context, token string) (*types.DeploymentInfo, error) {
	var (
		deployInfo *types.DeploymentInfo
		err        error
	)

	// record metric.
	metric := s.metric().Start("get_node_deployment_info")
	defer metric.End(err)

	deployInfo, err = s.getNodeDeploymentInfo(ctx, token)

	return deployInfo, err
}

// SetNodeDeploymentNodeConf set gse node conf.
func (s *Storage) SetNodeDeploymentNodeConf(ctx context.Context, token string, conf *types.NodeConf) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("set_node_deployment_node_conf")
	defer metric.End(err)

	err = s.seNodeDeploymenttNodeConf(ctx, token, conf)

	return err
}

// UpdateNodeDeploymentInfo update node deployment info.
func (s *Storage) UpdateNodeDeploymentInfo(ctx context.Context, token string, info *types.DeploymentInfo) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update_node_deployment_info")
	defer metric.End(err)

	err = s.updateNodeDeploymentInfo(ctx, token, info)

	return err
}

// CreateNodeDeployment createNodeDeployment a node deployment.
func (s *Storage) CreateNodeDeployment(ctx context.Context, nodeDeployment *types.NodeDeployment) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("create_node_deployment")
	defer metric.End(err)

	err = s.createNodeDeployment(ctx, nodeDeployment)

	return err
}

// ListNodeWorkflow lists node workflow by page and conditions.
func (s *Storage) ListNodeWorkflow(ctx context.Context, page types.Page, conditions ...*types.NodeWorkflowCondition) (
	[]*types.NodeWorkflow, int64, error) {

	var (
		results []*types.NodeWorkflow
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_node_workflow")
	defer metric.End(err)

	results, num, err = s.listNodeWorkflow(ctx, page, conditions...)

	return results, num, err
}

// CountNodeWorkflow counts node workflow by conditions.
func (s *Storage) CountNodeWorkflow(ctx context.Context, conditions ...*types.NodeWorkflowCondition) (int64, error) {
	var num int64
	var err error

	// record metric.
	metric := s.metric().Start("count_node_workflow")
	defer metric.End(err)

	num, err = s.countNodeWorkflow(ctx, conditions...)

	return num, err
}

// DistinctNodeWorkflow distincts node workflow fields.
func (s *Storage) DistinctNodeWorkflow(
	ctx context.Context, request types.NodeWorkflowDistinctRequest, conditions ...*types.NodeWorkflowCondition) (
	*types.NodeWorkflowDistinctResult, error) {

	var (
		result *types.NodeWorkflowDistinctResult
		err    error
	)

	// record metric.
	metric := s.metric().Start("distinct_node_workflow")
	defer metric.End(err)

	result, err = s.distinctNodeWorkflow(ctx, request, conditions...)

	return result, err
}

// GetNodeWorkflow gets a node workflow by workflow-id.
func (s *Storage) GetNodeWorkflow(ctx context.Context, workflowID string) (*types.NodeWorkflow, error) {
	var (
		nodeWorkflow *types.NodeWorkflow
		err          error
	)

	// record metric.
	metric := s.metric().Start("get_node_workflow")
	defer metric.End(err)

	nodeWorkflow, err = s.getNodeWorkflow(ctx, workflowID)

	return nodeWorkflow, err
}

// CreateNodeWorkflow creates a new node workflow.
func (s *Storage) CreateNodeWorkflow(ctx context.Context, workflow *types.NodeWorkflow) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("create_node_workflow")
	defer metric.End(err)

	err = s.createNodeWorkflow(ctx, workflow)

	return err
}

// UpdateNodeWorkflowStatus updates the status of a node workflow.
func (s *Storage) UpdateNodeWorkflowStatus(ctx context.Context, workflowID string, status types.NodeWorkflowStatus) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update_node_workflow_status")
	defer metric.End(err)

	err = s.updateNodeWorkflowStatus(ctx, workflowID, status)

	return err
}
