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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoNodeWorkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// listNodeWorkflow lists node workflow by page and conditions.
func (s *Storage) listNodeWorkflow(nCtx contextx.IContext, page types.Page, conditions ...*types.NodeWorkflowCondition) (
	[]*types.NodeWorkflow, int64, error) {

	var results []*types.NodeWorkflow
	var num int64
	var err error

	page.Sort = types.WithSortFields(page.Sort,
		types.WithFieldDesc(daoNodeWorkflow.FieldKeyOperateTime))
	opts := convertNodeWorkflowConditionsToOptions(conditions...)
	if results, num, err = s.daoNodeWorkflow.List(nCtx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// countNodeWorkflow counts node workflow by conditions.
func (s *Storage) countNodeWorkflow(nCtx contextx.IContext, conditions ...*types.NodeWorkflowCondition) (int64, error) {
	var num int64
	var err error

	opts := convertNodeWorkflowConditionsToOptions(conditions...)
	if num, err = s.daoNodeWorkflow.Count(nCtx, opts...); err != nil {
		return 0, err
	}

	return num, nil
}

// distinctNodeWorkflow distincts node workflow fields.
func (s *Storage) distinctNodeWorkflow(
	nCtx contextx.IContext, request types.NodeWorkflowDistinctRequest, conditions ...*types.NodeWorkflowCondition) (
	*types.NodeWorkflowDistinctResult, error) {

	var err error

	result := new(types.NodeWorkflowDistinctResult)
	opts := convertNodeWorkflowConditionsToOptions(conditions...)

	gp := gopool.NewPool()
	if request.BizID {
		gp.Go(func() error {
			var err error
			result.BizID, err = s.daoNodeWorkflow.DistinctNodeWorkflowBkBizID(nCtx, opts...)

			return err
		})
	}
	if request.Operator {
		gp.Go(func() error {
			var err error
			result.Operator, err = s.daoNodeWorkflow.DistinctNodeWorkflowOperator(nCtx, opts...)

			return err
		})
	}
	if request.Status {
		gp.Go(func() error {
			var err error
			result.Status, err = s.daoNodeWorkflow.DistinctNodeWorkflowStatus(nCtx, opts...)

			return err
		})
	}
	if request.Type {
		gp.Go(func() error {
			var err error
			result.Type, err = s.daoNodeWorkflow.DistinctNodeWorkflowType(nCtx, opts...)

			return err
		})
	}

	if err = gp.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

// getNodeWorkflow gets a node workflow by workflow-id.
func (s *Storage) getNodeWorkflow(nCtx contextx.IContext, workflowID string) (*types.NodeWorkflow, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return nil, errors.New("workflowID cannot be empty")
	}

	workflow, err := s.daoNodeWorkflow.Get(nCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow by id: %w", err)
	}

	if workflow == nil {
		return nil, fmt.Errorf("workflow not found results, workflow id: %s", workflowID)
	}

	return workflow, nil
}

// createNodeWorkflow creates a new node workflow.
func (s *Storage) createNodeWorkflow(nCtx contextx.IContext, workflow *types.NodeWorkflow) error {
	if nCtx == nil {
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

	if err := s.daoNodeWorkflow.Create(nCtx, workflow); err != nil {
		return fmt.Errorf("failed to createNodeDeployment workflow: %w", err)
	}

	return nil
}

// updateNodeWorkflowStatus updates the status of a node workflow.
func (s *Storage) updateNodeWorkflowStatus(nCtx contextx.IContext, workflowID string, status types.NodeWorkflowStatus) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return errors.New("workflowID cannot be empty")
	}

	if err := status.Validate(); err != nil {
		return fmt.Errorf("invalid workflow status: %s", status)
	}

	if err := s.daoNodeWorkflow.UpdateStatus(nCtx, workflowID, status); err != nil {
		return err
	}

	return nil
}

// convertNodeWorkflowConditionsToOptions converts node workflow conditions to options.
func convertNodeWorkflowConditionsToOptions(conditions ...*types.NodeWorkflowCondition) []daoNodeWorkflow.OptFn {
	opts := make([]daoNodeWorkflow.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.OperateTimeRange != nil {
			opts = append(opts, topoevent.WithOperateTimeRange(*condition.OperateTimeRange))
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				daoNodeWorkflow.WithWorkflowID(condition.ExactInclude.WorkflowID...),
				daoNodeWorkflow.WithBizID(condition.ExactInclude.BizID...),
				daoNodeWorkflow.WithType(condition.ExactInclude.Type...),
				daoNodeWorkflow.WithOperator(condition.ExactInclude.Operator...),
				daoNodeWorkflow.WithStatus(condition.ExactInclude.Status...))
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				daoNodeWorkflow.WithoutWorkflowID(condition.ExactExclude.WorkflowID...),
				daoNodeWorkflow.WithoutBizID(condition.ExactExclude.BizID...),
				daoNodeWorkflow.WithoutType(condition.ExactExclude.Type...),
				daoNodeWorkflow.WithoutOperator(condition.ExactExclude.Operator...),
				daoNodeWorkflow.WithoutStatus(condition.ExactExclude.Status...))
		}
	}

	return opts
}
