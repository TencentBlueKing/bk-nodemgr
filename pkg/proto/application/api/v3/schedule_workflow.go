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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// Validate checks required update fields; backend validates the runtime schedule syntax.
func (x *ScheduleWorkflowUpdateIntervalReq) Validate() error {
	if strings.TrimSpace(x.GetWorkflowId()) == "" {
		return errors.New("workflow_id is required")
	}
	if strings.TrimSpace(x.GetInterval()) == "" {
		return errors.New("interval is required")
	}

	return nil
}

// AutoConvert trims surrounding whitespace in the schedule expression.
func (x *ScheduleWorkflowUpdateIntervalReq) AutoConvert() {
	x.Interval = strings.TrimSpace(x.GetInterval())
}

// Validate checks pagination and the configuration operation time range.
func (x *ScheduleWorkflowListReq) Validate() error {
	if err := validatePage(x.GetPage()); err != nil {
		return err
	}
	_, err := x.ConvertConditionsToTypes()
	return err
}

// AutoConvert intentionally leaves the request unchanged.
func (x *ScheduleWorkflowListReq) AutoConvert() {}

// ConvertPageToTypes converts pagination without changing backend limits.
func (x *ScheduleWorkflowListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertConditionsToTypes converts schedule configuration filters.
func (x *ScheduleWorkflowListReq) ConvertConditionsToTypes() (*types.ScheduledWorkflowCondition, error) {
	timeRange, err := convertTimeRangeToTypes(x.GetOperateTimeRange())
	if err != nil {
		return nil, fmt.Errorf("invalid operate time range: %w", err)
	}
	condition := &types.ScheduledWorkflowCondition{OperateTimeRange: timeRange}
	if exact := x.GetExactIncludeConditions(); exact != nil {
		condition.ExactInclude = &types.ScheduledWorkflowExactFields{
			WorkflowID: exact.GetWorkflowId(), WorkflowName: exact.GetWorkflowName(), Operator: exact.GetOperator(),
		}
	}
	return condition, nil
}

// ConvertScheduleWorkflowsFromTypes converts schedule configuration results.
func (x *ScheduleWorkflowListResp) ConvertScheduleWorkflowsFromTypes(total int64, workflows []*types.ScheduledWorkflow) {
	items := conv.SliceToSlice(workflows, func(workflow *types.ScheduledWorkflow) *ScheduleWorkflowListResp_ScheduleWorkflow {
		operateTime := timeToUnixMilli(workflow.OperateTime)
		return &ScheduleWorkflowListResp_ScheduleWorkflow{
			WorkflowId: &workflow.WorkflowID, WorkflowName: &workflow.WorkflowName,
			TriggerId: &workflow.TriggerID, Operator: &workflow.Operator,
			OperateTime: &operateTime, Enabled: workflow.Enabled, Interval: workflow.Interval,
		}
	})
	x.Data = &ScheduleWorkflowListResp_Data{Total: total, Items: items}
}

// Validate checks the workflow ID and pagination.
func (x *ScheduleWorkflowOperationListReq) Validate() error {
	if x.GetWorkflowId() == "" {
		return errors.New("workflow_id is required")
	}
	return validatePage(x.GetPage())
}

// AutoConvert intentionally leaves the request unchanged.
func (x *ScheduleWorkflowOperationListReq) AutoConvert() {}

// ConvertPageToTypes converts pagination to the domain contract.
func (x *ScheduleWorkflowOperationListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertResultFromTypes converts operation records and persisted operator metadata.
func (x *ScheduleWorkflowOperationListResp) ConvertResultFromTypes(total int64, result []*operation.Operation) {
	items := conv.SliceToSlice(result, func(op *operation.Operation) *ScheduleWorkflowOperationListResp_ScheduleWorkflowOperation {
		item := &ScheduleWorkflowOperationListResp_ScheduleWorkflowOperation{
			OperationId: op.OperationID, InstanceIds: op.InstanceIDs,
			Operator: conv.ToStringDefault(op.Param.InitContent["operator"], ""), CreateTime: timeToUnixMilli(op.CreateTime),
			ParentOperationId: op.Param.ParentOperationID, ParentOperInstId: op.Param.ParentOperInstID,
		}
		if latest := op.LatestInstBriefData; latest != nil {
			item.LatestOperInstBriefData = &WorkflowOperInstBriefData{
				LifeCycle:                 convertNodeWorkflowOperInstLifeCycleFromTypes(latest.Lifecycle),
				LatestActionInstBriefData: convertNodeWorkflowActionInstBriefDataFromTypes(latest.LatestActionInstBriefData),
			}
		}
		return item
	})
	x.Data = &ScheduleWorkflowOperationListResp_Data{Total: total, Operations: items}
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

// AutoConvert intentionally leaves the request unchanged.
func (x *ScheduleWorkflowOperationInstanceListReq) AutoConvert() {}

// ConvertConditionsToTypes converts operation and instance identifiers.
func (x *ScheduleWorkflowOperationInstanceListReq) ConvertConditionsToTypes() *types.OperInstDataCondition {
	return &types.OperInstDataCondition{ExactInclude: &types.OperInstDataExactFields{
		OperationID: x.GetOperationId(), OperInstID: x.GetOperInstId(),
	}}
}

// ConvertResultFromTypes converts execution instances using existing lifecycle helpers.
func (x *ScheduleWorkflowOperationInstanceListResp) ConvertResultFromTypes(total int64, result []*operation.InstanceBriefData) {
	items := conv.SliceToSlice(result, func(instance *operation.InstanceBriefData) *WorflowOperationInstanceData {
		return &WorflowOperationInstanceData{
			OperationId: instance.Metadata.OperationID, OperInstId: instance.Metadata.OperationInstanceID,
			OperInstStatus: string(instance.Lifecycle.State), OperationDefName: instance.Metadata.OperationDefName,
			ParentOperationId: instance.Metadata.ParentOperationID, ActionNames: instance.Metadata.ActionNames,
			LifeCycle:                 convertNodeWorkflowOperInstLifeCycleFromTypes(instance.Lifecycle),
			LatestActionInstBriefData: convertNodeWorkflowActionInstBriefDataFromTypes(instance.LatestActionInstBriefData),
		}
	})
	x.Data = &ScheduleWorkflowOperationInstanceListResp_Data{Total: total, OperInstData: items}
}

// Validate checks the execution instance identifier.
func (x *ScheduleWorkflowOperationInstanceLogGetReq) Validate() error {
	if x.GetOperInstId() == "" {
		return errors.New("oper_inst_id is required")
	}
	return nil
}

// AutoConvert intentionally leaves the request unchanged.
func (x *ScheduleWorkflowOperationInstanceLogGetReq) AutoConvert() {}

// ConvertResultFromTypes reuses the shared workflow log and child-reference conversion.
func (x *ScheduleWorkflowOperationInstanceLogGetResp) ConvertResultFromTypes(result *operation.InstanceData) {
	resp := new(NodeWorkflowOperationInstanceLogGetResp)
	resp.ConvertResultFromTypes(result)
	x.Data = &ScheduleWorkflowOperationInstanceLogGetResp_Data{
		OperInstLogs: resp.GetData().GetOperInstLogs(), ExtraExecutionLogs: resp.GetData().GetExtraExecutionLogs(),
	}
}
