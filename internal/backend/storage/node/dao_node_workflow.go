/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// listNodeWorkflow lists node workflow by page and conditions.
func (s *Storage) listNodeWorkflow(ctx context.Context, page types.Page, conditions ...*types.NodeWorkflowCondition) (
	[]*types.NodeWorkflow, int64, error) {

	var results []*types.NodeWorkflow
	var num int64
	var err error

	page.Sort = types.WithSortFields(page.Sort,
		types.WithFieldDesc(nodeworkflow.FieldKeyOperateTime))
	opts := convertNodeWorkflowConditionsToOptions(conditions...)
	if results, num, err = s.daoNodeWorkflow.List(ctx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// countNodeWorkflow counts node workflow by conditions.
func (s *Storage) countNodeWorkflow(ctx context.Context, conditions ...*types.NodeWorkflowCondition) (int64, error) {
	var num int64
	var err error

	opts := convertNodeWorkflowConditionsToOptions(conditions...)
	if num, err = s.daoNodeWorkflow.Count(ctx, opts...); err != nil {
		return 0, err
	}

	return num, nil
}

// distinctNodeWorkflow distincts node workflow fields.
func (s *Storage) distinctNodeWorkflow(
	ctx context.Context, request types.NodeWorkflowDistinctRequest, conditions ...*types.NodeWorkflowCondition) (
	*types.NodeWorkflowDistinctResult, error) {

	var err error

	result := new(types.NodeWorkflowDistinctResult)
	opts := convertNodeWorkflowConditionsToOptions(conditions...)

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

	if err = gp.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

// getNodeWorkflow gets a node workflow by workflow-id.
func (s *Storage) getNodeWorkflow(ctx context.Context, workflowID string) (*types.NodeWorkflow, error) {
	if ctx == nil {
		return nil, basestorage.ErrNilContent()
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

// createNodeWorkflow creates a new node workflow.
func (s *Storage) createNodeWorkflow(ctx context.Context, workflow *types.NodeWorkflow) error {
	if ctx == nil {
		return basestorage.ErrNilContent()
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
		return fmt.Errorf("failed to createNodeDeployment workflow: %w", err)
	}

	return nil
}

// updateNodeWorkflowStatus updates the status of a node workflow.
func (s *Storage) updateNodeWorkflowStatus(ctx context.Context, workflowID string, status types.NodeWorkflowStatus) error {
	if ctx == nil {
		return basestorage.ErrNilContent()
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
