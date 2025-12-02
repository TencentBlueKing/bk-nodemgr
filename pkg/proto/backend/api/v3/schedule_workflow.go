/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *EnableScheduleWorkflowReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *EnableScheduleWorkflowReq) AutoConvert() {
}

// Validate check body.
func (req *DisableScheduleWorkflowReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (req *DisableScheduleWorkflowReq) AutoConvert() {
}

// Validate check body.
func (req *ListScheduleWorkflowReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (req *ListScheduleWorkflowReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *ListScheduleWorkflowReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *ListScheduleWorkflowReq) ConvertConditionsFromTypes(condition *types.ScheduledWorkflowCondition) error {
	exactCond, fuzzyCond, err := convertScheduleWorkConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertConditionsToTypes convert conditions to types.
func (x *ListScheduleWorkflowReq) ConvertConditionsToTypes() *types.ScheduledWorkflowCondition {
	return convertScheduleWorkflowConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions())
}

// convertScheduleWorkflowConditionsToTypes converts ScheduleWorkflowExactConditions and
func convertScheduleWorkflowConditionsToTypes(
	exactCond *ScheduleWorkflowExactConditions,
	_ *ScheduleWorkflowFuzzyConditions) *types.ScheduledWorkflowCondition {

	condition := &types.ScheduledWorkflowCondition{}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.ScheduledWorkflowExactFields{
			WorkflowID:   exactCond.GetWorkflowId(),
			WorkflowName: exactCond.GetWorkflowName(),
			Operator:     exactCond.GetOperator(),
		}
	}

	return condition
}

// convertScheduleWorkConditionsFromTypes converts ScheduleWorkflowCondition to
func convertScheduleWorkConditionsFromTypes(condition *types.ScheduledWorkflowCondition) (
	*ScheduleWorkflowExactConditions, *ScheduleWorkflowFuzzyConditions, error) {

	if condition == nil {
		return nil, nil, nil
	}

	var exactCond *ScheduleWorkflowExactConditions
	var fuzzyCond *ScheduleWorkflowFuzzyConditions

	if condition.ExactInclude != nil {
		exactCond = &ScheduleWorkflowExactConditions{
			WorkflowId:   condition.ExactInclude.WorkflowID,
			WorkflowName: condition.ExactInclude.WorkflowName,
			Operator:     condition.ExactInclude.Operator,
		}
	}

	if condition.FuzzyInclude != nil || condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, nil, errors.New("fuzzy-include, exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, nil
}

func newEmptyScheduleWorkflow() *ListScheduleWorkflowResp_ScheduleWorkflow {
	return &ListScheduleWorkflowResp_ScheduleWorkflow{
		WorkflowId:   new(string),
		WorkflowName: new(string),
		TriggerId:    new(string),
		Operator:     new(string),
		OperateTime:  new(int64),
	}
}

// ConvertScheduleWorkflowsFromTypes convert node workflows from types.
func (x *ListScheduleWorkflowResp) ConvertScheduleWorkflowsFromTypes(num int64, workflows []*types.ScheduledWorkflow) {
	items := make([]*ListScheduleWorkflowResp_ScheduleWorkflow, 0, len(workflows))
	for _, workflow := range workflows {
		item := newEmptyScheduleWorkflow()
		*item.WorkflowId = workflow.WorkflowID
		*item.WorkflowName = workflow.WorkflowName
		*item.TriggerId = workflow.TriggerID
		*item.Operator = workflow.Operator
		*item.OperateTime = workflow.OperateTime.UnixMilli()
		items = append(items, item)
	}

	x.Data = &ListScheduleWorkflowResp_Data{
		Total: num,
		Items: items,
	}
}
