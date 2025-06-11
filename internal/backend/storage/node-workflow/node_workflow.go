/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeworkflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/base"
	daoBase "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/nodeworkflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName defines the storage name.
const StorageName = "node_workflow"

const (
	recentMonitoredTime = 5 * time.Minute
)

// NewStorage creates a new node workflow storage handler.
func NewStorage(client *mongo.Client, database string, logger logger.Logger) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: base.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}

	err := base.InitStorage(&s.Storage,
		base.WithStartFunc(s.initDao),
		base.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	s.registerScheduler()

	return s, nil
}

// Storage provides a node workflow storage handler.
type Storage struct {
	base.Storage

	daoNodeWorkflow nodeworkflow.IHandler
	daoOperInstData operinstdata.IHandler

	monitoredWorkflows      map[string]*types.NodeWorkflow
	monitoredWorkflowsMutex sync.RWMutex
}

func (s *Storage) initDao() error {
	s.daoNodeWorkflow = nodeworkflow.New(s.Database, s.Logger)
	s.daoOperInstData = operinstdata.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) registerScheduler() {
	s.Scheduler = scheduler.NewScheduler(scheduler.WithLogger(s.Logger), scheduler.WithInterval(time.Second*5)) // nolint: mnd
	s.Scheduler.RegisterTask(&scheduler.Task{
		ID:       "obtain monitored workflows",
		Interval: 5 * time.Second,  // nolint: mnd
		Timeout:  20 * time.Second, // nolint: mnd
		Fn:       s.obtainMonitoredWorkflows,
	})

	s.Scheduler.RegisterTask(&scheduler.Task{
		ID:       "monitor workflow status",
		Interval: 1 * time.Second,  // nolint: mnd
		Timeout:  10 * time.Second, // nolint: mnd
		Fn:       s.monitorWorkflowStatus,
	})
}

// obtainMonitoredWorkflows Obtain a list of workflows that need to be listened to.
func (s *Storage) obtainMonitoredWorkflows(ctx context.Context) error {
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

func (s *Storage) monitorWorkflowStatus(ctx context.Context) error {
	s.monitoredWorkflowsMutex.RLock()
	defer s.monitoredWorkflowsMutex.RUnlock()

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
		status := calWorkflowStatus(operInsts)
		if err := s.daoNodeWorkflow.UpdateStatus(ctx, s.monitoredWorkflows[triggerID].WorkflowID, status); err != nil {
			return fmt.Errorf("update node workflow status failed, err: %w", err)
		}

		delete(s.monitoredWorkflows, triggerID)
	}

	return nil
}

func calWorkflowStatus(operationInsts []*operation.InstanceBriefData) types.NodeWorkflowStatus {
	successCount := 0
	failedCount := 0

	for _, inst := range operationInsts {
		switch inst.Lifecycle.State {
		case operation.StateSuccess:
			successCount++
		case operation.StateFailed:
			failedCount++
		}
	}

	total := len(operationInsts)

	switch {
	case successCount == total:
		return types.NodeWorkflowStatusSuccess
	case failedCount == total:
		return types.NodeWorkflowStatusFailed
	default:
		return types.NodeWorkflowStatusPartialFailed
	}
}

func (s *Storage) check() error {
	if s.daoNodeWorkflow == nil {
		return errors.New("dao node workflow is nil")
	}

	return nil
}

// ListNodeWorkflow lists node workflow by page and conditions.
func (s *Storage) ListNodeWorkflow(ctx context.Context, page types.Page, conditions ...*types.NodeWorkflowCondition) (
	[]*types.NodeWorkflow, int64, error) {

	opts := convertNodeWorkflowConditionsToOptions(conditions...)

	return s.daoNodeWorkflow.List(ctx, page, opts...)
}

// CountNodeWorkflow counts node workflow by conditions.
func (s *Storage) CountNodeWorkflow(ctx context.Context, conditions ...*types.NodeWorkflowCondition) (int64, error) {

	opts := convertNodeWorkflowConditionsToOptions(conditions...)

	return s.daoNodeWorkflow.Count(ctx, opts...)
}

// DistinctNodeWorkflow distincts node workflow fields.
func (s *Storage) DistinctNodeWorkflow(
	ctx context.Context, request types.NodeWorkflowDistinctRequest, conditions ...*types.NodeWorkflowCondition) (
	*types.NodeWorkflowDistinctResult, error) {

	opts := convertNodeWorkflowConditionsToOptions(conditions...)

	result := new(types.NodeWorkflowDistinctResult)

	gp := gopool.NewPool()
	if request.BizID {
		gp.Go(func() error {
			var err error
			result.BizID, err = s.daoNodeWorkflow.DistinctNodeWorkflowBkBizID(ctx, opts...)

			return err
		})
	}
	if request.Operator {
		gp.Go(func() error {
			var err error
			result.Operator, err = s.daoNodeWorkflow.DistinctNodeWorkflowOperator(ctx, opts...)

			return err
		})
	}
	if request.Status {
		gp.Go(func() error {
			var err error
			result.Status, err = s.daoNodeWorkflow.DistinctNodeWorkflowStatus(ctx, opts...)

			return err
		})
	}
	if request.Type {
		gp.Go(func() error {
			var err error
			result.Type, err = s.daoNodeWorkflow.DistinctNodeWorkflowType(ctx, opts...)

			return err
		})
	}

	if err := gp.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetNodeWorkflow gets a node workflow by workflow-id.
func (s *Storage) GetNodeWorkflow(ctx context.Context, workflowID string) (*types.NodeWorkflow, error) {
	if ctx == nil {
		return nil, base.ErrNilContent()
	}

	if workflowID == "" {
		return nil, errors.New("workflowID cannot be empty")
	}

	workflow, err := s.daoNodeWorkflow.Get(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow by id: %w", err)
	}

	if workflow == nil {
		return nil, fmt.Errorf("workflow not found results, workflow id: %s", workflowID)
	}

	return workflow, nil
}

// CreateNodeWorkflow creates a new node workflow.
func (s *Storage) CreateNodeWorkflow(ctx context.Context, workflow *types.NodeWorkflow) error {
	if ctx == nil {
		return base.ErrNilContent()
	}

	if workflow == nil {
		return errors.New("workflow cannot be nil")
	}

	if err := workflow.Type.Validate(); err != nil {
		return fmt.Errorf("invalid workflow data.type: %w", err)
	}

	if err := workflow.Status.Validate(); err != nil {
		return fmt.Errorf("invalid workflow data.status: %w", err)
	}

	if err := s.daoNodeWorkflow.Create(ctx, workflow); err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	return nil
}

// UpdateNodeWorkflowStatus updates the status of a node workflow.
func (s *Storage) UpdateNodeWorkflowStatus(
	ctx context.Context, workflowID string, status types.NodeWorkflowStatus) error {

	if ctx == nil {
		return base.ErrNilContent()
	}

	if workflowID == "" {
		return errors.New("workflowID cannot be empty")
	}

	if err := status.Validate(); err != nil {
		return fmt.Errorf("invalid workflow status: %s", status)
	}

	if err := s.daoNodeWorkflow.UpdateStatus(ctx, workflowID, status); err != nil {
		return err
	}

	return nil
}

// convertNodeWorkflowConditionsToOptions converts node workflow conditions to options.
func convertNodeWorkflowConditionsToOptions(conditions ...*types.NodeWorkflowCondition) []nodeworkflow.OptFn {
	opts := make([]nodeworkflow.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.OperateTimeRange != nil {
			opts = append(opts, topoevent.WithOperateTimeRange(*condition.OperateTimeRange))
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				nodeworkflow.WithWorkflowID(condition.ExactInclude.WorkflowID...),
				nodeworkflow.WithBizID(condition.ExactInclude.BizID...),
				nodeworkflow.WithType(condition.ExactInclude.Type...),
				nodeworkflow.WithOperator(condition.ExactInclude.Operator...),
				nodeworkflow.WithStatus(condition.ExactInclude.Status...))
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				nodeworkflow.WithoutWorkflowID(condition.ExactExclude.WorkflowID...),
				nodeworkflow.WithoutBizID(condition.ExactExclude.BizID...),
				nodeworkflow.WithoutType(condition.ExactExclude.Type...),
				nodeworkflow.WithoutOperator(condition.ExactExclude.Operator...),
				nodeworkflow.WithoutStatus(condition.ExactExclude.Status...))
		}
	}

	return opts
}
