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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	scheduledworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/scheduled-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// listScheduledWorkflow lists scheduled workflow by page and conditions.
func (s *Storage) listScheduledWorkflow(
	nCtx contextx.IContext, page types.Page, conditions ...*types.ScheduledWorkflowCondition) (
	[]*types.ScheduledWorkflow, int64, error) {

	return s.daoScheduledWorkflow.List(nCtx, page, convertScheduledWorkflowConditionsToOptions(conditions...)...)
}

// listScheduledWorkflowWithoutCount lists scheduled workflow by page and conditions without count.
func (s *Storage) listScheduledWorkflowWithoutCount(
	nCtx contextx.IContext, page types.Page, conditions ...*types.ScheduledWorkflowCondition) (
	[]*types.ScheduledWorkflow, error) {

	return s.daoScheduledWorkflow.ListWithoutCount(nCtx, page, convertScheduledWorkflowConditionsToOptions(conditions...)...)
}

// countScheduledWorkflow counts scheduled workflow by conditions.
func (s *Storage) countScheduledWorkflow(
	nCtx contextx.IContext, conditions ...*types.ScheduledWorkflowCondition) (
	int64, error) {

	return s.daoScheduledWorkflow.Count(nCtx, convertScheduledWorkflowConditionsToOptions(conditions...)...)
}

// getScheduledWorkflow gets a scheduled workflow by workflow id.
func (s *Storage) getScheduledWorkflow(nCtx contextx.IContext, workflowID string) (*types.ScheduledWorkflow, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return nil, errors.New("workflow id cannot be empty")
	}

	return s.daoScheduledWorkflow.Get(nCtx, workflowID)
}

// createScheduledWorkflow creates a new scheduled workflow.
func (s *Storage) createScheduledWorkflow(nCtx contextx.IContext, workflow *types.ScheduledWorkflow) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if workflow == nil {
		return errors.New("scheduled workflow cannot be nil")
	}

	if workflow.WorkflowID == "" {
		return errors.New("scheduled workflow id cannot be empty")
	}

	return s.daoScheduledWorkflow.Create(nCtx, workflow)
}

func (s *Storage) updateScheduledWorkflowTriggerID(nCtx contextx.IContext, workflowID, triggerID string) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return errors.New("workflow id cannot be empty")
	}

	if triggerID == "" {
		return errors.New("trigger id cannot be empty")
	}

	return s.daoScheduledWorkflow.UpdateTriggerID(nCtx, workflowID, triggerID)
}

func (s *Storage) updateScheduledWorkflowPrivateData(nCtx contextx.IContext, workflowID string, data map[string]any) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return errors.New("workflow id cannot be empty")
	}

	return s.daoScheduledWorkflow.UpdatePrivateData(nCtx, workflowID, data)
}

func (s *Storage) switchScheduleWorkflow(nCtx contextx.IContext, workflowID string, enable bool) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return errors.New("workflow id cannot be empty")
	}

	return s.daoScheduledWorkflow.Switch(nCtx, workflowID, enable)
}

// convertScheduledWorkflowConditionsToOptions converts scheduled workflow conditions to options.
func convertScheduledWorkflowConditionsToOptions(
	conditions ...*types.ScheduledWorkflowCondition) []scheduledworkflow.OptFn {

	opts := make([]scheduledworkflow.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.OperateTimeRange != nil {
			opts = append(opts, topoevent.WithOperateTimeRange(*condition.OperateTimeRange))
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				scheduledworkflow.WithWorkflowID(condition.ExactInclude.WorkflowID...),
				scheduledworkflow.WithWorkflowName(condition.ExactInclude.WorkflowName...),
				scheduledworkflow.WithOperator(condition.ExactInclude.Operator...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				scheduledworkflow.WithoutWorkflowID(condition.ExactExclude.WorkflowID...),
				scheduledworkflow.WithoutWorkflowName(condition.ExactExclude.WorkflowName...),
				scheduledworkflow.WithoutOperator(condition.ExactExclude.Operator...),
			)
		}
	}

	return opts
}
