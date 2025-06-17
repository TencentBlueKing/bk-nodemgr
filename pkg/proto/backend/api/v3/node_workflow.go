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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *NodeWorkflowListReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert auto convert.
func (x *NodeWorkflowListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *NodeWorkflowListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *NodeWorkflowListReq) ConvertConditionsFromTypes(condition *types.NodeWorkflowCondition) error {
	exactCond, fuzzyCond, err := convertNodeWorkConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertConditionsToTypes convert conditions to types.
func (x *NodeWorkflowListReq) ConvertConditionsToTypes() *types.NodeWorkflowCondition {
	return convertNodeWorkflowConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions())
}

// ConvertNodeWorkflowsFromTypes convert node workflows from types.
func (x *NodeWorkflowListResp) ConvertNodeWorkflowsFromTypes(num int64, workflows []*types.NodeWorkflow) {
	items := make([]*NodeWorkflowInfo, 0, len(workflows))
	for _, workflow := range workflows {
		item := newEmptyNodeWorkflow()
		*item.WorkflowId = workflow.WorkflowID
		item.BkBizId = workflow.BizIDs
		*item.Type = string(workflow.Type)
		*item.Status = string(workflow.Status)
		*item.TriggerId = workflow.TriggerID
		*item.Operator = workflow.Operator
		*item.OperateTime = workflow.OperateTime.UnixMilli()

		items = append(items, item)
	}

	x.Data = &NodeWorkflowListResp_Data{
		Total: num,
		Items: items,
	}
}

// ConvertNodeWorkflowsToTypes convert node workflows to types.
func (x *NodeWorkflowListResp) ConvertNodeWorkflowsToTypes() ([]*types.NodeWorkflow, int64) {
	data := x.GetData()
	if data == nil {
		return nil, 0
	}

	items := data.GetItems()
	result := make([]*types.NodeWorkflow, len(items))

	for idx, item := range items {
		workflow := &types.NodeWorkflow{
			WorkflowID:  item.GetWorkflowId(),
			TriggerID:   item.GetTriggerId(),
			Type:        types.NodeWorkflowType(item.GetType()),
			Status:      types.NodeWorkflowStatus(item.GetStatus()),
			BizIDs:      item.GetBkBizId(),
			Operator:    item.GetOperator(),
			OperateTime: time.UnixMilli(item.GetOperateTime()),
		}

		result[idx] = workflow
	}

	return result, data.GetTotal()
}

// Validate check body.
func (x *NodeWorkflowDistinctReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *NodeWorkflowDistinctReq) AutoConvert() {
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *NodeWorkflowDistinctReq) ConvertConditionsFromTypes(condition *types.NodeWorkflowCondition) error {
	exactCond, fuzzyCond, err := convertNodeWorkConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertConditionsToTypes convert conditions to types.
func (x *NodeWorkflowDistinctReq) ConvertConditionsToTypes() *types.NodeWorkflowCondition {
	return convertNodeWorkflowConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions())
}

// ConvertResultFromTypes convert result from types.
func (x *NodeWorkflowDistinctResp) ConvertResultFromTypes(result *types.NodeWorkflowDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &NodeWorkflowDistinctResp_Data{
		Type:     formatRespSlice(types.NodeWorkflowTypeListToStringList(result.Type)),
		BkBizId:  formatRespSlice(result.BizID),
		Operator: formatRespSlice(result.Operator),
		Status:   formatRespSlice(types.NodeWorkflowStatusListToStringList(result.Status)),
	}
}

// ConvertWorkflowDistinctToTypes ...
func (x *NodeWorkflowDistinctResp) ConvertWorkflowDistinctToTypes() *types.NodeWorkflowDistinctResult {
	data := x.GetData()
	if data == nil {
		return nil
	}

	result := &types.NodeWorkflowDistinctResult{
		Type:     types.StringListToNodeWorkflowTypeList(data.GetType()),
		BizID:    data.GetBkBizId(),
		Status:   types.StringListToNodeWorkflowStatusList(data.GetStatus()),
		Operator: data.GetOperator(),
	}

	return result
}

// Validate check body.
func (x *NodeWorkflowOperationListReq) Validate() error {
	return validatePage(x.GetPage())
}

// ConvertConditionsToComm convert conditions to comm.
func (x *NodeWorkflowOperationListReq) ConvertConditionsToComm() string {
	return x.GetWorkflowId()
}

// AutoConvert auto convert.
func (x *NodeWorkflowOperationListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *NodeWorkflowOperationListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertResultFromTypes convert workflow id to types.
func (x *NodeWorkflowOperationListResp) ConvertResultFromTypes(total int64, result []*operation.Operation) {
	items := make([]*NodeWorkflowOperation, 0, total)
	for _, op := range result {
		item := &NodeWorkflowOperation{
			OperationId: op.OperationID,
			Definition: &OperationDefinition{
				OpertionName: op.Definition.Name(),
				ActionNames:  op.Definition.ActionDefNames(),
			},
			InstanceIds: op.InstanceIDs,
			Param: &OperationParam{
				TimeoutSecond: int64(op.Definition.DefaultParameters().Timeout.Seconds()),
			},
		}
		items = append(items, item)
	}

	x.Data = &NodeWorkflowOperationListResp_Data{
		TotalCount: total,
		Operations: items,
	}
}

// ConvertWorkflowOperationToTypes convert workflow operation to types.
func (x *NodeWorkflowOperationListResp) ConvertWorkflowOperationToTypes() ([]*operation.Operation, int64) {
	data := x.GetData()
	if data == nil {
		return nil, 0
	}

	items := data.GetOperations()
	result := make([]*operation.Operation, len(items))

	for idx, item := range items {
		operation := &operation.Operation{
			OperationID: item.GetOperationId(),
			Definition: &operation.DefinitionSnapshot{
				SnapshotName:           item.GetDefinition().GetOpertionName(),
				SnapshotActionDefNames: item.GetDefinition().GetActionNames(),
			},
			InstanceIDs: item.GetInstanceIds(),
			Param: operation.Param{
				Timeout: time.Duration(item.GetParam().GetTimeoutSecond()),
			},
		}

		result[idx] = operation
	}

	return result, data.GetTotalCount()
}

// Validate check body.
func (x *NodeWorkflowOperationInstanceListReq) Validate() error {
	if x.GetOperationId() == "" {
		return errors.New("operation id is required")
	}
	return nil
}

// ConvertConditionsToComm ...
func (x *NodeWorkflowOperationInstanceListReq) ConvertConditionsToComm() string {
	return x.GetOperationId()
}

// AutoConvert auto convert.
func (x *NodeWorkflowOperationInstanceListReq) AutoConvert() {
}

// ConvertResultFromTypes ...
func (x *NodeWorkflowOperationInstanceListResp) ConvertResultFromTypes(
	total int64, result []*operation.InstanceBriefData) {

	items := make([]*NodeWorflowOperationInstanceData, 0, len(result))
	for _, opinstance := range result {
		oper := &NodeWorflowOperationInstanceData{
			OperInstId:        opinstance.Metadata.OperationInstanceID,
			OperationId:       opinstance.Metadata.OperationID,
			OperInstStatus:    string(opinstance.Lifecycle.State),
			OperationDefName:  opinstance.Metadata.OperationDefName,
			ParentOperationId: opinstance.Metadata.ParentOperationID,
			ActionNames:       opinstance.Metadata.ActionNames,
		}

		items = append(items, oper)
	}

	x.Data = &NodeWorkflowOperationInstanceListResp_Data{
		OperInstData: items,
		Total:        total,
	}
}

// ConvertOperationInstanceFromTypes  ...
func (x *NodeWorkflowOperationInstanceListResp) ConvertOperationInstanceFromTypes(
	total int64, result *operation.InstanceBriefData) {

	items := make([]*NodeWorflowOperationInstanceData, 1)

	oper := &NodeWorflowOperationInstanceData{
		OperInstId:        result.Metadata.OperationInstanceID,
		OperationId:       result.Metadata.OperationID,
		OperInstStatus:    string(result.Lifecycle.State),
		OperationDefName:  result.Metadata.OperationDefName,
		ParentOperationId: result.Metadata.ParentOperationID,
		ActionNames:       result.Metadata.ActionNames,
	}

	items[0] = oper
	x.Data = &NodeWorkflowOperationInstanceListResp_Data{
		OperInstData: items,
		Total:        total,
	}
}

// ConvertWorkflowOperationInstanceToTypes ...
func (x *NodeWorkflowOperationInstanceListResp) ConvertWorkflowOperationInstanceToTypes() (
	[]*operation.InstanceBriefData, int64) {

	data := x.GetData()
	if data == nil {
		return nil, 0
	}

	total := data.GetTotal()
	items := data.GetOperInstData()
	result := make([]*operation.InstanceBriefData, total)

	for idx, item := range items {
		inst := &operation.InstanceBriefData{
			Metadata: &operation.InstanceMetadata{
				OperationInstanceID: item.GetOperInstId(),
				OperationID:         item.GetOperationId(),
				OperationDefName:    item.GetOperationDefName(),
				ParentOperationID:   item.GetParentOperationId(),
				ActionNames:         item.GetActionNames(),
			},
			Lifecycle: &operation.Lifecycle{
				State: operation.State(item.GetOperInstStatus()),
			},
		}

		result[idx] = inst
	}

	return result, total
}

// Validate check body.
func (x *NodeWorkflowOperationInstanceListStatusReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *NodeWorkflowOperationInstanceListStatusReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *NodeWorkflowOperationInstanceListStatusReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertListStatusConditionsToTypes ...
func (x *NodeWorkflowOperationInstanceListStatusReq) ConvertListStatusConditionsToTypes() operation.ListOperationInstanceCondition {
	return operation.ListOperationInstanceCondition{
		TriggerIDs: x.GetExactIncludeConditions().GetTriggerId(),
		States:     operation.GetAllStates(),
	}
}

// ConvertWorkflowOperInstanceStatusFromTypes ...
func (x *NodeWorkflowOperationInstanceListStatusResp) ConvertWorkflowOperInstanceStatusFromTypes(
	result []*operation.InstanceBriefData) {

	items := make([]*NodeWorkflowOperationInstanceStatus, 0)
	for _, opinstance := range result {
		item := &NodeWorkflowOperationInstanceStatus{
			Index:       int64(opinstance.Metadata.Index),
			Status:      string(opinstance.Lifecycle.State),
			OperationId: opinstance.Metadata.OperationID,
			TriggerId:   opinstance.Metadata.TriggerID,
		}

		items = append(items, item)
	}

	x.Data = &NodeWorkflowOperationInstanceListStatusResp_Data{
		Items: items,
	}
}

// ConvertWorkflowOperationInstanceStatusToTypes convert workflow operation instance status to types.
func (x *NodeWorkflowOperationInstanceListStatusResp) ConvertWorkflowOperationInstanceStatusToTypes() []*operation.InstanceStatus {

	data := x.GetData()
	if data == nil {
		return nil
	}

	items := data.GetItems()

	result := make([]*operation.InstanceStatus, 0)
	for _, value := range items {
		result = append(result, &operation.InstanceStatus{
			TriggerID:   value.GetTriggerId(),
			State:       operation.State(value.GetStatus()),
			Index:       int(value.GetIndex()),
			OperationID: value.GetOperationId(),
		})
	}

	return result
}

// ConvertConditionsFromTypes ...
func (x *NodeWorkflowOperationInstanceListStatusReq) ConvertConditionsFromTypes(
	condition *types.NodeWorkflowOperInstanceStatusCondition) error {

	exactCond, fuzzyCond, err := convertNodeWorkOperaInstanceStatusConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// Validate check body.
func (x *NodeWorkflowOperationInstanceLogGetReq) Validate() error {
	if x.GetOperInstId() == "" {
		return errors.New("workflow_id is required and cannot be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeWorkflowOperationInstanceLogGetReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *NodeWorkflowOperationInstanceLogGetResp) ConvertResultFromTypes(result *operation.InstanceData) {
	if result == nil {
		return
	}

	operInstLogs := make(map[string]*ActionMessage)
	for actionID, v := range result.ActionInstanceDataMap {
		actionMsg := &ActionMessage{
			Logs: make([]*ActionMessage_Message, 0, len(v.Messages)),
		}
		for _, msg := range v.Messages {
			actionMsg.Logs = append(actionMsg.Logs, &ActionMessage_Message{
				Time: msg.Time.Unix(),
				Text: msg.Text,
			})
		}
		operInstLogs[actionID] = actionMsg
	}

	x.Data = &NodeWorkflowOperationInstanceLogGetResp_Data{
		OperInstLogs: operInstLogs,
	}
}

// ConvertWorkflowOperationInstanceLogToTypes convert workflow operation instance log to types.
func (x *NodeWorkflowOperationInstanceLogGetResp) ConvertWorkflowOperationInstanceLogToTypes() *operation.InstanceData {
	data := x.GetData()
	if data == nil {
		return nil
	}

	result := &operation.InstanceData{
		ActionInstanceDataMap: make(map[string]*action.InstanceData),
	}

	for actionID, actionMsg := range data.GetOperInstLogs() {
		if actionMsg == nil {
			continue
		}

		actionData := &action.InstanceData{
			Messages: make([]action.Message, 0, len(actionMsg.GetLogs())),
		}

		for _, msg := range actionMsg.GetLogs() {
			if msg == nil {
				continue
			}

			actionData.Messages = append(actionData.Messages, action.Message{
				Time: time.Unix(msg.GetTime(), 0),
				Text: msg.GetText(),
			})
		}

		result.ActionInstanceDataMap[actionID] = actionData
	}

	return result
}

func convertNodeWorkflowConditionsToTypes(
	exactCond *NodeWorkflowExactConditions,
	_ *NodeWorkflowFuzzyConditions) *types.NodeWorkflowCondition {

	condition := &types.NodeWorkflowCondition{}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.NodeWorkflowExactFields{
			BizID:      exactCond.GetBkBizId(),
			Type:       types.StringListToNodeWorkflowTypeList(exactCond.GetType()),
			Status:     types.StringListToNodeWorkflowStatusList(exactCond.GetStatus()),
			WorkflowID: exactCond.GetWorkflowId(),
			Operator:   exactCond.GetOperator(),
		}
	}

	return condition
}

func convertNodeWorkConditionsFromTypes(condition *types.NodeWorkflowCondition) (
	*NodeWorkflowExactConditions, *NodeWorkflowFuzzyConditions, error) {

	if condition == nil {
		return nil, nil, nil
	}

	var exactCond *NodeWorkflowExactConditions
	var fuzzyCond *NodeWorkflowFuzzyConditions

	if condition.ExactInclude != nil {
		exactCond = &NodeWorkflowExactConditions{
			BkBizId:    condition.ExactInclude.BizID,
			WorkflowId: condition.ExactInclude.WorkflowID,
			Type:       types.NodeWorkflowTypeListToStringList(condition.ExactInclude.Type),
			Status:     types.NodeWorkflowStatusListToStringList(condition.ExactInclude.Status),
			Operator:   condition.ExactInclude.Operator,
		}
	}

	if condition.FuzzyInclude != nil || condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, nil, errors.New("fuzzy-include, exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, nil
}

func convertNodeWorkOperaInstanceStatusConditionsFromTypes(condition *types.NodeWorkflowOperInstanceStatusCondition) (
	*NodeWorkflowInstanceStatusExactConditions, *NodeWorkflowInstanceStatusFuzzyConditions, error) {

	if condition == nil {
		return nil, nil, nil
	}

	var exactCond *NodeWorkflowInstanceStatusExactConditions
	var fuzzyCond *NodeWorkflowInstanceStatusFuzzyConditions

	if condition.ExactInclude != nil {
		exactCond = &NodeWorkflowInstanceStatusExactConditions{
			TriggerId: condition.ExactInclude.TriggerID,
		}
	}

	if condition.FuzzyInclude != nil || condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, nil, errors.New("fuzzy-include, exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, nil
}

func newEmptyNodeWorkflow() *NodeWorkflowInfo {
	return &NodeWorkflowInfo{
		WorkflowId:  new(string),
		Type:        new(string),
		TriggerId:   new(string),
		BkBizId:     make([]int64, 0),
		Operator:    new(string),
		OperateTime: new(int64),
		Status:      new(string),
	}
}
