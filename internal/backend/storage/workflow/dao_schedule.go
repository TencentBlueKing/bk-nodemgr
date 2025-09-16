/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/scheduleworkflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/schedule"
)

// listScheduleWorkflow lists schedule workflow by page and conditions.
func (s *Storage) listScheduleWorkflow(
	ctx contextx.IContext, page types.Page, conditions ...*types.ScheduleWorkflowCondition) (
	[]*schedule.Schedule, int64, error) {

	return s.daoScheduleWorkflow.List(ctx, page, convertScheduleWorkflowConditionsToOptions(conditions...)...)
}

// countScheduleWorkflow counts schedule workflow by conditions.
func (s *Storage) countScheduleWorkflow(
	ctx contextx.IContext, conditions ...*types.ScheduleWorkflowCondition) (
	int64, error) {

	return s.daoScheduleWorkflow.Count(ctx, convertScheduleWorkflowConditionsToOptions(conditions...)...)
}

// getScheduleWorkflow gets a schedule workflow by workflow id.
func (s *Storage) getScheduleWorkflow(ctx contextx.IContext, workflowID string) (*schedule.Schedule, error) {
	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return nil, errors.New("workflow id cannot be empty")
	}

	return s.daoScheduleWorkflow.Get(ctx, workflowID)
}

// createScheduleWorkflow creates a new schedule workflow.
func (s *Storage) createScheduleWorkflow(ctx contextx.IContext, workflow *schedule.Schedule) error {
	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if workflow == nil {
		return errors.New("schedule workflow cannot be nil")
	}

	if workflow.WorkflowID == "" {
		return errors.New("schedule workflow id cannot be empty")
	}

	return s.daoScheduleWorkflow.Create(ctx, workflow)
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
