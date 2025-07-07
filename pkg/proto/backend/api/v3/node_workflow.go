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
	exactCond, fuzzyCond, timeRange, err := convertNodeWorkConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond
	x.OperateTimeRange = timeRange

	return nil
}

// ConvertConditionsToTypes convert conditions to types.
func (x *NodeWorkflowListReq) ConvertConditionsToTypes() *types.NodeWorkflowCondition {
	return convertNodeWorkflowConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(),
		x.GetOperateTimeRange())
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
		*item.FinishTime = workflow.FinishTime.UnixMilli()
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
			FinishTime:  time.UnixMilli(item.GetFinishTime()),
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
	exactCond, fuzzyCond, _, err := convertNodeWorkConditionsFromTypes(condition)
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
		x.GetFuzzyIncludeConditions(),
		nil)
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
	if x.GetExactIncludeConditions() == nil || x.GetExactIncludeConditions().GetWorkflowId() == "" {
		return errors.New("workflow_id is required")
	}
	return validatePage(x.GetPage())
}

// ConvertConditionsToComm convert conditions to comm.
func (x *NodeWorkflowOperationListReq) ConvertConditionsToComm() string {
	return x.GetExactIncludeConditions().GetWorkflowId()
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *NodeWorkflowOperationListReq) ConvertConditionsFromTypes(
	condition *types.NodeWorkflowOperationCondition) error {

	exactCond, fuzzyCond, err := convertNodeWorkOperConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertConditionsToTypes convert conditions to types.
func (x *NodeWorkflowOperationListReq) ConvertConditionsToTypes(
	triggerID string) *types.NodeWorkflowOperationCondition {

	return convertWorkflowOperationConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(), triggerID)
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
			Definition: &NodeWorkflowOperationDefinition{
				OpertionName: op.Definition.Name(),
				ActionNames:  op.Definition.ActionDefNames(),
			},
			InstanceIds: op.InstanceIDs,
			Param: &NodeWorkflowOperationParam{
				TimeoutSecond:   int64(op.Definition.DefaultParameters().Timeout.Seconds()),
				BkNetworkareaId: safeGetInt64(op.Param.ExtraContent, "networkarea_id", -1),
				BkBizId:         safeGetInt64(op.Param.ExtraContent, "biz_id", -1),
				BkHostInnerip:   safeGetString(op.Param.ExtraContent, "inner_ip", ""),
				BkHostInneripV6: safeGetString(op.Param.ExtraContent, "inner_ipv6", ""),
				NodeVersion:     safeGetString(op.Param.ExtraContent, "node_version", ""),
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
				ExtraContent: map[string]interface{}{
					"networkarea_id": item.GetParam().GetBkNetworkareaId(),
					"biz_id":         item.GetParam().GetBkBizId(),
					"inner_ip":       item.GetParam().GetBkHostInnerip(),
					"inner_ipv6":     item.GetParam().GetBkHostInneripV6(),
					"node_version":   item.GetParam().GetNodeVersion(),
				},
			},
		}

		result[idx] = operation
	}

	return result, data.GetTotalCount()
}

// Validate check body.
func (x *NodeWorkflowOperationInstanceListReq) Validate() error {
	if len(x.GetOperationId()) == 0 || x.GetOperationId()[0] == "" {
		return errors.New("operation id is required")
	}

	return nil
}

// ConvertConditionsToComm ...
func (x *NodeWorkflowOperationInstanceListReq) ConvertConditionsToComm() []string {
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
			LifeCycle: &LifeCycle{
				State:      string(opinstance.Lifecycle.State),
				CreateTime: opinstance.Lifecycle.CreatedAt.Unix(),
				StartTime:  opinstance.Lifecycle.StartedAt.Unix(),
				EndTime:    opinstance.Lifecycle.EndedAt.Unix(),
			},
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
				State:     operation.State(item.GetOperInstStatus()),
				CreatedAt: time.Unix(item.GetLifeCycle().GetCreateTime(), 0),
				StartedAt: time.Unix(item.GetLifeCycle().GetStartTime(), 0),
				EndedAt:   time.Unix(item.GetLifeCycle().GetEndTime(), 0),
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

	operInstLogs := make(map[string]*NodeWorkflowActionData)
	for actionID, v := range result.ActionInstanceDataMap {

		lifecycle := &LifeCycle{
			State:      string(v.Lifecycle.State),
			CreateTime: v.Lifecycle.CreatedAt.Unix(),
			StartTime:  v.Lifecycle.StartedAt.Unix(),
			EndTime:    v.Lifecycle.EndedAt.Unix(),
		}

		messages := make([]*NodeWorkflowActionMessage_Message, 0, len(v.Messages))
		for _, msg := range v.Messages {
			messages = append(messages, &NodeWorkflowActionMessage_Message{
				Time:  msg.Time.Unix(),
				Text:  msg.Text,
				Level: msg.Level,
			})
		}

		operInstLogs[actionID] = &NodeWorkflowActionData{
			LifeCycle: lifecycle,
			Message:   &NodeWorkflowActionMessage{Logs: messages},
		}
	}

	x.Data = &NodeWorkflowOperationInstanceLogGetResp_Data{
		OperInstLogs: operInstLogs,
		Total:        int64(len(result.ActionInstanceDataMap)),
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

	for actionID, actionData := range x.GetData().GetOperInstLogs() {
		if actionData == nil {
			continue
		}

		instance := &action.InstanceData{
			Lifecycle: &action.Lifecycle{
				State:     action.State(actionData.GetLifeCycle().GetState()),
				CreatedAt: time.Unix(actionData.GetLifeCycle().GetCreateTime(), 0),
				StartedAt: time.Unix(actionData.GetLifeCycle().GetStartTime(), 0),
				EndedAt:   time.Unix(actionData.GetLifeCycle().GetEndTime(), 0),
			},
		}

		if actionData.GetMessage() != nil {
			instance.Messages = make([]action.Message, 0, len(actionData.GetMessage().GetLogs()))
			for _, msg := range actionData.GetMessage().GetLogs() {
				if msg == nil {
					continue
				}
				instance.Messages = append(instance.Messages, action.Message{
					Time:  time.Unix(msg.GetTime(), 0),
					Text:  msg.GetText(),
					Level: msg.GetLevel(),
				})
			}
		}

		result.ActionInstanceDataMap[actionID] = instance
	}

	return result
}

// AutoConvert auto convert.
func (x *NodeWorkflowOperationRetryReq) AutoConvert() {
}

// Validate convert workflow id.
func (x *NodeWorkflowOperationRetryReq) Validate() error {
	if x.GetWorkflowId() == "" {
		return errors.New("workflow_id is required")
	}

	for _, oper := range x.GetOperationId() {
		if oper == "" {
			return errors.New("operation_id can not be empty")
		}
	}

	return nil
}

// ConvertOperationRetryParamFromTypes ...
func (x *NodeWorkflowOperationRetryReq) ConvertOperationRetryParamFromTypes(retryParm types.NodeOperationRetryParam) {
	x.WorkflowId = retryParm.WorkflowID
	x.OperationId = retryParm.OperationIDs
	x.RetryMod = string(retryParm.RetryMode)
}

// ConvertOperInstanceID convert OperationInstance id.
func (x *NodeWorkflowOperationRetryResp) ConvertOperInstanceID(operInstance []string) {
	x.Data = &NodeWorkflowOperationRetryResp_Data{InstanceIds: operInstance}
}

// ConvertResultToComm convert result from types.
func (x *NodeWorkflowOperationRetryResp) ConvertResultToComm() []string {
	return x.GetData().GetInstanceIds()
}

func convertNodeWorkflowConditionsToTypes(
	exactCond *NodeWorkflowExactConditions,
	_ *NodeWorkflowFuzzyConditions, timeRange *TimeRange) *types.NodeWorkflowCondition {

	condition := &types.NodeWorkflowCondition{}

	if timeRange != nil {
		condition.OperateTimeRange = &types.TimeRange{
			StartTime: time.Unix(timeRange.GetStartTimestampSec(), 0),
			EndTime:   time.Unix(timeRange.GetEndTimestampSec(), 0),
		}
	}

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

func convertWorkflowOperationConditionsToTypes(
	exactCond *NodeWorkflowOperationExactConditions,
	_ *NodeWorkflowOperationFuzzyConditions, triggerID string) *types.NodeWorkflowOperationCondition {

	condition := &types.NodeWorkflowOperationCondition{}
	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.NodeWorkflowOperationExactFields{
			TriggerID:     triggerID,
			InnerIP:       exactCond.GetBkHostInnerip(),
			InnerIPv6:     exactCond.GetBkHostInneripV6(),
			BizID:         exactCond.GetBkBizId(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			NodeVersion:   exactCond.GetNodeVersion(),
		}
	}

	return condition
}

func convertNodeWorkConditionsFromTypes(condition *types.NodeWorkflowCondition) (
	*NodeWorkflowExactConditions, *NodeWorkflowFuzzyConditions, *TimeRange, error) {

	if condition == nil {
		return nil, nil, nil, nil
	}

	var exactCond *NodeWorkflowExactConditions
	var fuzzyCond *NodeWorkflowFuzzyConditions
	var timeRange *TimeRange

	if condition.OperateTimeRange != nil {
		timeRange = &TimeRange{
			StartTimestampSec: condition.OperateTimeRange.StartTime.Unix(),
			EndTimestampSec:   condition.OperateTimeRange.EndTime.Unix(),
		}
	}

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
		return nil, nil, nil, errors.New("fuzzy-include, exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, timeRange, nil
}

func convertNodeWorkOperConditionsFromTypes(condition *types.NodeWorkflowOperationCondition) (
	*NodeWorkflowOperationExactConditions, *NodeWorkflowOperationFuzzyConditions, error) {

	if condition == nil {
		return nil, nil, nil
	}

	var exactCond *NodeWorkflowOperationExactConditions
	var fuzzyCond *NodeWorkflowOperationFuzzyConditions

	if condition.ExactInclude != nil {
		exactCond = &NodeWorkflowOperationExactConditions{
			BkBizId:         condition.ExactInclude.BizID,
			WorkflowId:      condition.ExactInclude.WorkflowID,
			BkNetworkareaId: condition.ExactInclude.NetworkAreaID,
			BkHostInnerip:   condition.ExactInclude.InnerIP,
			BkHostInneripV6: condition.ExactInclude.InnerIPv6,
			NodeVersion:     condition.ExactInclude.NodeVersion,
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
		FinishTime:  new(int64),
		Status:      new(string),
	}
}

func safeGetInt64(m map[string]interface{}, key string, def int64) int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	if val, ok := v.(int64); ok {
		return val
	}

	return def
}

func safeGetString(m map[string]interface{}, key string, def string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	if val, ok := v.(string); ok {
		return val
	}

	return def
}
