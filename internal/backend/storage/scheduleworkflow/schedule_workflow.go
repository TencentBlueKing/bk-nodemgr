/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package scheduleworkflow provides storage for schedule workflow.
package scheduleworkflow

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/scheduleworkflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "schedule_workflow"
)

// NewStorage creates a new schedule workflow storage handler.
func NewStorage(client *mongo.Client, database string, logger logger.Logger) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}

	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

// Storage provides a schedule workflow storage handler.
type Storage struct {
	basestorage.Storage

	daoScheduleWorkflow scheduleworkflow.IHandler
}

func (s *Storage) initDao() error {
	s.daoScheduleWorkflow = scheduleworkflow.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.daoScheduleWorkflow == nil {
		return errors.New("dao schedule workflow is nil")
	}

	return nil
}

// ListScheduleWorkflow lists schedule workflow by page and conditions.
func (s *Storage) ListScheduleWorkflow(ctx context.Context, page types.Page,
	conditions ...*types.ScheduleWorkflowCondition) ([]*types.ScheduleWorkflow, int64, error) {

	opts := convertScheduleWorkflowConditionsToOptions(conditions...)

	return s.daoScheduleWorkflow.List(ctx, page, opts...)
}

// CountScheduleWorkflow counts schedule workflow by conditions.
func (s *Storage) CountScheduleWorkflow(ctx context.Context, conditions ...*types.ScheduleWorkflowCondition) (
	int64, error) {

	opts := convertScheduleWorkflowConditionsToOptions(conditions...)

	return s.daoScheduleWorkflow.Count(ctx, opts...)
}

// DistinctScheduleWorkflow distincts schedule workflow fields.
func (s *Storage) DistinctScheduleWorkflow(
	ctx context.Context, request types.ScheduleWorkflowDistinctRequest, conditions ...*types.ScheduleWorkflowCondition) (
	*types.ScheduleWorkflowDistinctResult, error) {

	opts := convertScheduleWorkflowConditionsToOptions(conditions...)

	result := new(types.ScheduleWorkflowDistinctResult)

	gp := gopool.NewPool()
	if request.Operator {
		gp.Go(func() error {
			var err error
			result.Operator, err = s.daoScheduleWorkflow.DistinctScheduleWorkflowOperator(ctx, opts...)

			return err
		})
	}
	if request.WorkflowName {
		gp.Go(func() error {
			var err error
			result.Type, err = s.daoScheduleWorkflow.DistinctScheduleWorkflowName(ctx, opts...)

			return err
		})
	}

	if err := gp.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetScheduleWorkflow gets a schedule workflow by workflow-id.
func (s *Storage) GetScheduleWorkflow(ctx context.Context, workflowID string) (*types.ScheduleWorkflow, error) {
	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return nil, errors.New("workflow id cannot be empty")
	}

	workflow, err := s.daoScheduleWorkflow.Get(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow by workflow-id(%s), err %w", workflowID, err)
	}

	if workflow == nil {
		return nil, fmt.Errorf("workflow not found results, workflow-id(%s)", workflowID)
	}

	return workflow, nil
}

// CreateScheduleWorkflow creates a new schedule workflow.
func (s *Storage) CreateScheduleWorkflow(ctx context.Context, workflow *types.ScheduleWorkflow) error {
	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if workflow == nil {
		return errors.New("schedule workflow cannot be nil")
	}

	if workflow.WorkflowID == "" {
		return errors.New("schedule workflow id cannot be empty")
	}

	if err := s.daoScheduleWorkflow.Create(ctx, workflow); err != nil {
		return fmt.Errorf("failed to create schedule workflow, workflow-id(%s), err: %w", workflow.WorkflowID, err)
	}

	return nil
}

// convertScheduleWorkflowConditionsToOptions converts schedule workflow conditions to options.
func convertScheduleWorkflowConditionsToOptions(
	conditions ...*types.ScheduleWorkflowCondition) []scheduleworkflow.OptFn {

	opts := make([]scheduleworkflow.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.OperateTimeRange != nil {
			opts = append(opts, topoevent.WithOperateTimeRange(*condition.OperateTimeRange))
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				scheduleworkflow.WithWorkflowID(condition.ExactInclude.WorkflowID...),
				scheduleworkflow.WithWorkflowName(condition.ExactInclude.WorkflowName...),
				scheduleworkflow.WithOperator(condition.ExactInclude.Operator...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				scheduleworkflow.WithoutWorkflowID(condition.ExactExclude.WorkflowID...),
				scheduleworkflow.WithoutWorkflowName(condition.ExactExclude.WorkflowName...),
				scheduleworkflow.WithoutOperator(condition.ExactExclude.Operator...),
			)
		}
	}

	return opts
}
