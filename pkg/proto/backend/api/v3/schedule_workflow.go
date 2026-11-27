/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package v3

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
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
	if err := validatePage(req.GetPage()); err != nil {
		return err
	}

	if _, err := convertTimeRangeToTypes(req.GetOperateTimeRange()); err != nil {
		return fmt.Errorf("invalid operate time range: %w", err)
	}

	return nil
}

// AutoConvert auto convert.
func (req *ListScheduleWorkflowReq) AutoConvert() {
}

const (
	// schedule workflow list max limit
	maxScheduleWorkflowLimit = 500
)

// PageLimit return page limit.
func (x *ListScheduleWorkflowReq) PageLimit() int {
	return maxScheduleWorkflowLimit
}

// ConvertPageToTypes convert page to types.
func (x *ListScheduleWorkflowReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *ListScheduleWorkflowReq) ConvertConditionsFromTypes(condition *types.ScheduledWorkflowCondition) error {
	exactCond, fuzzyCond, err := convertScheduleWorkConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond
	x.OperateTimeRange = nil
	if condition != nil && condition.OperateTimeRange != nil {
		x.OperateTimeRange = &TimeRange{
			StartTimestampSec: condition.OperateTimeRange.StartTime.Unix(),
			EndTimestampSec:   condition.OperateTimeRange.EndTime.Unix(),
		}
	}

	return nil
}

// ConvertConditionsToTypes converts conditions after Validate has checked the time range.
func (x *ListScheduleWorkflowReq) ConvertConditionsToTypes() *types.ScheduledWorkflowCondition {
	condition := convertScheduleWorkflowConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions())
	if timeRange := x.GetOperateTimeRange(); timeRange != nil {
		condition.OperateTimeRange = &types.TimeRange{
			StartTime: time.Unix(timeRange.GetStartTimestampSec(), 0),
			EndTime:   time.Unix(timeRange.GetEndTimestampSec(), 0),
		}
	}

	return condition
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
		item.Enabled = workflow.Enabled
		item.Interval = workflow.Interval
		items = append(items, item)
	}

	x.Data = &ListScheduleWorkflowResp_Data{
		Total: num,
		Items: items,
	}
}

// Validate checks the workflow ID and pagination.
func (x *ScheduleWorkflowOperationListReq) Validate() error {
	if x.GetWorkflowId() == "" {
		return errors.New("workflow_id is required")
	}

	return validatePage(x.GetPage())
}

// AutoConvert auto converts the request.
func (x *ScheduleWorkflowOperationListReq) AutoConvert() {
}

// PageLimit returns the operation list page limit.
func (x *ScheduleWorkflowOperationListReq) PageLimit() int {
	return maxScheduleWorkflowLimit
}

// ConvertPageToTypes converts pagination to types.
func (x *ScheduleWorkflowOperationListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
}

// ConvertResultFromTypes converts operations and their persisted operator metadata.
func (x *ScheduleWorkflowOperationListResp) ConvertResultFromTypes(
	total int64, result []*operation.Operation) {

	items := make([]*ScheduleWorkflowOperationListResp_ScheduleWorkflowOperation, len(result))
	for idx, op := range result {
		// Missing operator metadata intentionally produces an empty operator.
		operator := conv.ToStringDefault(op.Param.InitContent["operator"], "")
		item := &ScheduleWorkflowOperationListResp_ScheduleWorkflowOperation{
			OperationId:       op.OperationID,
			InstanceIds:       op.InstanceIDs,
			Operator:          operator,
			CreateTime:        timeToUnixMilli(op.CreateTime),
			ParentOperationId: op.Param.ParentOperationID,
			ParentOperInstId:  op.Param.ParentOperInstID,
		}
		if op.LatestInstBriefData != nil {
			item.LatestOperInstBriefData = &ScheduleWorkflowOperationListResp_LatestOperInstBriefData{
				LifeCycle:                 convertNodeWorkflowOperInstLifeCycleFromTypes(op.LatestInstBriefData.Lifecycle),
				LatestActionInstBriefData: convertNodeWorkflowActionInstBriefDataFromTypes(op.LatestInstBriefData.LatestActionInstBriefData),
			}
		}
		items[idx] = item
	}

	x.Data = &ScheduleWorkflowOperationListResp_Data{
		Operations: items,
		Total:      total,
	}
}

// Validate requires operation or instance IDs and rejects empty identifiers.
func (x *ScheduleWorkflowOperationInstanceListReq) Validate() error {
	if len(x.GetOperationId()) == 0 && len(x.GetOperInstId()) == 0 {
		return errors.New("operation_id or oper_inst_id is required")
	}

	for _, id := range x.GetOperationId() {
		if id == "" {
			return errors.New("operation_id cannot be empty")
		}
	}
	for _, id := range x.GetOperInstId() {
		if id == "" {
			return errors.New("oper_inst_id cannot be empty")
		}
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ScheduleWorkflowOperationInstanceListReq) AutoConvert() {
}

// ConvertConditionsToTypes converts operation and instance ID filters to types.
func (x *ScheduleWorkflowOperationInstanceListReq) ConvertConditionsToTypes() *types.OperInstDataCondition {
	return &types.OperInstDataCondition{
		ExactInclude: &types.OperInstDataExactFields{
			OperationID: x.GetOperationId(),
			OperInstID:  x.GetOperInstId(),
		},
	}
}

// ConvertResultFromTypes converts operation instance summaries to the response.
func (x *ScheduleWorkflowOperationInstanceListResp) ConvertResultFromTypes(
	total int64, result []*operation.InstanceBriefData) {

	items := make([]*WorflowOperationInstanceData, len(result))
	for idx, instance := range result {
		lifecycle := convertNodeWorkflowOperInstLifeCycleFromTypes(instance.Lifecycle)
		items[idx] = &WorflowOperationInstanceData{
			OperationId:               instance.Metadata.OperationID,
			OperInstId:                instance.Metadata.OperationInstanceID,
			OperInstStatus:            lifecycle.GetState(),
			OperationDefName:          instance.Metadata.OperationDefName,
			ParentOperationId:         instance.Metadata.ParentOperationID,
			ActionNames:               instance.Metadata.ActionNames,
			LifeCycle:                 lifecycle,
			LatestActionInstBriefData: convertNodeWorkflowActionInstBriefDataFromTypes(instance.LatestActionInstBriefData),
		}
	}

	x.Data = &ScheduleWorkflowOperationInstanceListResp_Data{
		Total:        total,
		OperInstData: items,
	}
}

// Validate checks the operation instance ID.
func (x *ScheduleWorkflowOperationInstanceLogGetReq) Validate() error {
	if x.GetOperInstId() == "" {
		return errors.New("oper_inst_id is required")
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ScheduleWorkflowOperationInstanceLogGetReq) AutoConvert() {
}

// ConvertResultFromTypes converts action and extra execution logs to the response.
func (x *ScheduleWorkflowOperationInstanceLogGetResp) ConvertResultFromTypes(result *operation.InstanceData) {
	if result == nil {
		return
	}

	x.Data = &ScheduleWorkflowOperationInstanceLogGetResp_Data{
		OperInstLogs:       convertActionInstanceLogs(result),
		ExtraExecutionLogs: convertExtraExecutionLogs(result),
	}
}

// ConvertScheduleWorkflowsToTypes converts schedule configurations to domain models.
func (x *ListScheduleWorkflowResp) ConvertScheduleWorkflowsToTypes() ([]*types.ScheduledWorkflow, int64) {
	items := conv.SliceToSlice(x.GetData().GetItems(), func(item *ListScheduleWorkflowResp_ScheduleWorkflow) *types.ScheduledWorkflow {
		return &types.ScheduledWorkflow{
			WorkflowID: item.GetWorkflowId(), WorkflowName: item.GetWorkflowName(),
			TriggerID: item.GetTriggerId(), Operator: item.GetOperator(),
			OperateTime: time.UnixMilli(item.GetOperateTime()), Enabled: item.GetEnabled(), Interval: item.GetInterval(),
		}
	})
	return items, x.GetData().GetTotal()
}

// ConvertResultToTypes converts operation records to domain models.
func (x *ScheduleWorkflowOperationListResp) ConvertResultToTypes() ([]*operation.Operation, int64) {
	items := conv.SliceToSlice(x.GetData().GetOperations(), func(item *ScheduleWorkflowOperationListResp_ScheduleWorkflowOperation) *operation.Operation {
		op := &operation.Operation{
			OperationID: item.GetOperationId(), InstanceIDs: item.GetInstanceIds(),
			CreateTime: time.UnixMilli(item.GetCreateTime()),
			Param: operation.Param{
				ParentOperationID: item.GetParentOperationId(), ParentOperInstID: item.GetParentOperInstId(),
				InitContent: map[string]any{"operator": item.GetOperator()},
			},
		}
		if latest := item.GetLatestOperInstBriefData(); latest != nil {
			op.LatestInstBriefData = &operation.InstanceBriefData{
				Lifecycle:                 convertNodeWorkflowOperInstLifeCycleToTypes(latest.GetLifeCycle()),
				LatestActionInstBriefData: convertNodeWorkflowActionInstBriefDataToTypes(latest.GetLatestActionInstBriefData()),
			}
			op.LatestInstBriefData.Lifecycle.StoppedAt = time.UnixMilli(latest.GetLifeCycle().GetStopTime())
		}
		return op
	})
	return items, x.GetData().GetTotal()
}

// ConvertConditionsFromTypes converts the supported operation and instance filters.
func (x *ScheduleWorkflowOperationInstanceListReq) ConvertConditionsFromTypes(condition *types.OperInstDataCondition) {
	if condition == nil || condition.ExactInclude == nil {
		return
	}
	x.OperationId = condition.ExactInclude.OperationID
	x.OperInstId = condition.ExactInclude.OperInstID
}

// ConvertResultToTypes reuses the shared workflow instance conversion.
func (x *ScheduleWorkflowOperationInstanceListResp) ConvertResultToTypes() ([]*operation.InstanceBriefData, int64) {
	resp := &NodeWorkflowOperationInstanceListResp{Data: &NodeWorkflowOperationInstanceListResp_Data{
		Total: x.GetData().GetTotal(), OperInstData: x.GetData().GetOperInstData(),
	}}
	items, total := resp.ConvertWorkflowOperationInstanceToTypes()
	for idx, item := range x.GetData().GetOperInstData() {
		items[idx].Lifecycle.StoppedAt = time.UnixMilli(item.GetLifeCycle().GetStopTime())
	}
	return items, total
}

// ConvertResultToTypes reuses the shared workflow log conversion.
func (x *ScheduleWorkflowOperationInstanceLogGetResp) ConvertResultToTypes() *operation.InstanceData {
	resp := &NodeWorkflowOperationInstanceLogGetResp{Data: &NodeWorkflowOperationInstanceLogGetResp_Data{
		OperInstLogs: x.GetData().GetOperInstLogs(), ExtraExecutionLogs: x.GetData().GetExtraExecutionLogs(),
	}}
	return resp.ConvertWorkflowOperationInstanceLogToTypes()
}
