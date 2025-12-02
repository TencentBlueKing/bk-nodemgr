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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
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

// NodeWorkflowStatistics represents the status of operations in a node workflow.
type NodeWorkflowStatistics struct {
	WorkflowID      string
	TotalCount      int
	InitCount       int
	LaunchedCount   int
	RunningCount    int
	SuccessCount    int
	FailedCount     int
	TimeoutCount    int
	TerminatedCount int
}

// Validate check body.
func (x *NodeWorkflowStatisticsReq) Validate() error {
	for _, workflowID := range x.GetWorkflowId() {
		if workflowID == "" {
			return errors.New("workflow_id can not be empty")
		}
	}

	return nil
}

// ConvertNodeWorkflowsFromTypes convert node workflows from types.
func (x *NodeWorkflowListResp) ConvertNodeWorkflowsFromTypes(num int64, workflows []*types.NodeWorkflow,
	businessMap map[int64]string) {

	items := make([]*NodeWorkflowInfo, 0, len(workflows))
	for _, workflow := range workflows {
		item := newEmptyNodeWorkflow()

		*item.WorkflowId = workflow.WorkflowID
		*item.TriggerId = workflow.TriggerID
		item.BkBizId = workflow.BizIDs

		*item.Type = string(workflow.Type)
		*item.Status = string(workflow.Status)
		*item.Operator = workflow.Operator
		*item.OperateTime = workflow.OperateTime.UnixMilli()
		*item.FinishTime = workflow.FinishTime.UnixMilli()
		bizNames := make([]string, 0, len(workflow.BizIDs))
		for _, bizID := range workflow.BizIDs {
			if name, exists := businessMap[bizID]; exists {
				bizNames = append(bizNames, name)
			} else {
				bizNames = append(bizNames, "")
			}
		}
		item.BkBizName = bizNames

		items = append(items, item)
	}

	x.Data = &NodeWorkflowListResp_Data{
		Total: num,
		Items: items,
	}
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

// AutoConvert auto convert.
func (x *NodeWorkflowStatisticsReq) AutoConvert() {
}

// ConvertConditionsToWorkflowConditionTypes convert conditions to types.
func (x *NodeWorkflowStatisticsReq) ConvertConditionsToWorkflowConditionTypes() *types.NodeWorkflowCondition {
	return convertNodeWorkflowConditionsToTypes(
		&NodeWorkflowExactConditions{
			WorkflowId: x.GetWorkflowId(),
		}, nil, nil)
}

// ConvertNodeWorkflowsFromTypes convert node workflows from types.
func (x *NodeWorkflowStatisticsResp) ConvertNodeWorkflowsFromTypes(result []*NodeWorkflowStatistics) {
	items := make([]*NodeWorkflowStatisticsResp_StatisticsInfo, len(result))

	for i, item := range result {
		items[i] = newEmptyNodeWorkflowOperationStatus()
		*items[i].WorkflowId = item.WorkflowID
		*items[i].TotalCount = int64(item.TotalCount)
		*items[i].InitCount = int64(item.InitCount)
		*items[i].LaunchedCount = int64(item.LaunchedCount)
		*items[i].RunningCount = int64(item.RunningCount)
		*items[i].SuccessCount = int64(item.SuccessCount)
		*items[i].FailedCount = int64(item.FailedCount)
		*items[i].TimeoutCount = int64(item.TimeoutCount)
		*items[i].TerminatedCount = int64(item.TerminatedCount)
	}

	x.Data = &NodeWorkflowStatisticsResp_Data{
		Items: items,
	}
}

// Validate check body.
func (x *NodeWorkflowDistinctReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *NodeWorkflowDistinctReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *NodeWorkflowDistinctReq) ConvertConditionsToTypes() *types.NodeWorkflowCondition {
	return convertNodeWorkflowConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(), nil)
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
func (x *NodeWorkflowOperationListReq) ConvertConditionsToTypes() *types.NodeWorkflowOperationCondition {
	return convertWorkflowOperationConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions())
}

// ConvertPageToTypes convert page to types.
func (x *NodeWorkflowOperationListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// AutoConvert auto convert.
func (x *NodeWorkflowOperationListReq) AutoConvert() {
}

// ConvertResultFromTypes convert workflow id to types.
func (x *NodeWorkflowOperationListResp) ConvertResultFromTypes(
	total int64, result []*types.NodeWorkflowListOperationResult, operationsSummary []*types.OperationSummary) {

	items := make([]*NodeWorkflowOperation, len(result))
	for idx, op := range result {
		if op == nil {
			continue
		}

		items[idx] = &NodeWorkflowOperation{
			OperationId: op.OperationID,
			InstanceIds: op.OperInstanceIDs,
			Param: &NodeWorkflowOperationParam{
				Operator:            op.Operator,
				BkNetworkareaId:     op.NetworkAreaID,
				BkNetworkunitId:     op.NetworkUnitID,
				BkBizId:             op.BizID,
				BkHostInnerList:     op.InnerIPList,
				BkHostInneripV6List: op.InnerIPV6List,
				NodeVersion:         op.NodeVersion,
			},
			Status: &NodeWorkflowOperationStatus{
				State:           string(operationsSummary[idx].LastStatus),
				TotalTimeSecond: operationsSummary[idx].TotalDuration,
			},
		}
	}

	x.Data = &NodeWorkflowOperationListResp_Data{
		TotalCount: total,
		Operations: items,
	}
}

// GetCountOnly get count only.
func (x *NodeWorkflowOperationListResp) GetCountOnly() interface{} {
	if x.GetData() == nil {
		return &NodeWorkflowOperationListResp_Data{
			TotalCount: 0,
		}
	}

	x.Data.Operations = nil

	return x.GetData()
}

// Validate check body.
func (x *NodeWorkflowOperationInstanceListReq) Validate() error {
	if x.GetOperationId() == "" {
		return errors.New("operation_id is required")
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

// ConvertResultFromTypes convert node workflows from types.
func (x *NodeWorkflowOperationInstanceListResp) ConvertResultFromTypes(total int64,
	result []*operation.InstanceBriefData) {

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

// Validate check body.
func (x *NodeWorkflowOperationInstanceLogGetReq) Validate() error {
	if x.GetOperInstId() == "" {
		return errors.New("oper_inst_id is required and cannot be empty")
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

// AutoConvert auto convert.
func (x *NodeWorkflowOperationTerminateReq) AutoConvert() {
}

// Validate convert workflow id.
func (x *NodeWorkflowOperationTerminateReq) Validate() error {
	if x.GetWorkflowId() == "" {
		return errors.New("workflow_id is required")
	}

	if len(x.GetOperationIds()) == 0 {
		return errors.New("operation_id is required")
	}

	return nil
}

// ConvertNodeWorkflowOperationTerminateParamToTypes convert terminate node operation param to types.
func (x *NodeWorkflowOperationTerminateReq) ConvertNodeWorkflowOperationTerminateParamToTypes() *types.NodeWorkflowOperationTerminateParam {
	return &types.NodeWorkflowOperationTerminateParam{
		WorkflowID:   x.GetWorkflowId(),
		OperationIDs: x.GetOperationIds(),
	}
}

// AutoConvert auto convert.
func (x *NodeWorkflowOperationRetryReq) AutoConvert() {
}

// Validate convert workflow id.
func (x *NodeWorkflowOperationRetryReq) Validate() error {
	if x.GetWorkflowId() == "" {
		return errors.New("workflow_id is required")
	}

	if len(x.GetOperationIds()) == 0 {
		return errors.New("operation_id is required")
	}

	for _, oper := range x.GetOperationIds() {
		if oper == "" {
			return errors.New("operation_id can not be empty")
		}
	}

	return nil
}

// ConvertNodeWorkflowOperationRetryParamToTypes convert retry node operation param to types.
func (x *NodeWorkflowOperationRetryReq) ConvertNodeWorkflowOperationRetryParamToTypes() *types.NodeWorkflowOperationRetryParam {
	return &types.NodeWorkflowOperationRetryParam{
		WorkflowID:   x.GetWorkflowId(),
		OperationIDs: x.GetOperationIds(),
		RetryMode:    operation.RetryMode(x.GetRetryMod()),
	}
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

func convertNodeWorkConditionsFromTypes(condition *types.NodeWorkflowCondition) (
	*NodeWorkflowExactConditions, *NodeWorkflowFuzzyConditions, *TimeRange, error) {

	if condition == nil {
		return nil, nil, nil, nil
	}

	var timeRange *TimeRange
	var exactCond *NodeWorkflowExactConditions
	var fuzzyCond *NodeWorkflowFuzzyConditions

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
			State:           types.WorkflowOperationStatusListToStringList(condition.ExactInclude.State),
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

func convertWorkflowOperationConditionsToTypes(
	exactCond *NodeWorkflowOperationExactConditions,
	_ *NodeWorkflowOperationFuzzyConditions) *types.NodeWorkflowOperationCondition {

	condition := &types.NodeWorkflowOperationCondition{}
	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.NodeWorkflowOperationExactFields{
			WorkflowID:    exactCond.GetWorkflowId(),
			State:         types.StringListToWorkflowOperationStatusList(exactCond.GetState()),
			InnerIP:       exactCond.GetBkHostInnerip(),
			InnerIPv6:     exactCond.GetBkHostInneripV6(),
			BizID:         exactCond.GetBkBizId(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			NodeVersion:   exactCond.GetNodeVersion(),
		}
	}

	return condition
}

// newEmptyNodeWorkflow creates a new empty NodeWorkflowInfo.
func newEmptyNodeWorkflow() *NodeWorkflowInfo {
	return &NodeWorkflowInfo{
		WorkflowId:  new(string),
		TriggerId:   new(string),
		Type:        new(string),
		BkBizId:     make([]int64, 0),
		Operator:    new(string),
		OperateTime: new(int64),
		FinishTime:  new(int64),
		Status:      new(string),
	}
}

func newEmptyNodeWorkflowOperationStatus() *NodeWorkflowStatisticsResp_StatisticsInfo {
	return &NodeWorkflowStatisticsResp_StatisticsInfo{
		WorkflowId:      new(string),
		TotalCount:      new(int64),
		InitCount:       new(int64),
		LaunchedCount:   new(int64),
		RunningCount:    new(int64),
		SuccessCount:    new(int64),
		FailedCount:     new(int64),
		TimeoutCount:    new(int64),
		TerminatedCount: new(int64),
	}
}
