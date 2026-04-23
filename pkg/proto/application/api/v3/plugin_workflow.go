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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// ===============================================================================
// PluginWorkflowList Related Interfaces
// ===============================================================================

// Validate check body.
func (x *PluginWorkflowListReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert auto convert.
func (x *PluginWorkflowListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *PluginWorkflowListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PluginWorkflowListReq) ConvertConditionsFromTypes(condition *types.PluginWorkflowCondition) error {
	exactCond, fuzzyCond, timeRange, err := convertPluginWorkConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond
	x.OperateTimeRange = timeRange

	return nil
}

func convertPluginWorkConditionsFromTypes(condition *types.PluginWorkflowCondition) (
	*PluginWorkflowExactConditions, *PluginWorkflowFuzzyConditions, *TimeRange, error) {

	if condition == nil {
		return nil, nil, nil, nil
	}

	var timeRange *TimeRange
	var exactCond *PluginWorkflowExactConditions
	var fuzzyCond *PluginWorkflowFuzzyConditions

	if condition.OperateTimeRange != nil {
		timeRange = &TimeRange{
			StartTimestampSec: condition.OperateTimeRange.StartTime.Unix(),
			EndTimestampSec:   condition.OperateTimeRange.EndTime.Unix(),
		}
	}

	if condition.ExactInclude != nil {
		exactCond = &PluginWorkflowExactConditions{
			BkHostId:        condition.ExactInclude.HostID,
			BkBizId:         condition.ExactInclude.BizID,
			WorkflowId:      condition.ExactInclude.WorkflowID,
			Type:            types.PluginWorkflowTypeListToStringList(condition.ExactInclude.Type),
			Status:          types.PluginWorkflowStatusListToStringList(condition.ExactInclude.Status),
			Operator:        condition.ExactInclude.Operator,
			BkHostInnerip:   condition.ExactInclude.HostInnerIP,
			BkHostInneripV6: condition.ExactInclude.HostInnerIPV6,
		}
	}

	if condition.FuzzyInclude != nil || condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, nil, nil, errors.New("fuzzy-include, exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, timeRange, nil
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PluginWorkflowListReq) ConvertConditionsToTypes() (*types.PluginWorkflowCondition, error) {
	return convertPluginWorkflowConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(),
		x.GetOperateTimeRange())
}

// ConvertPluginWorkflowsFromTypes convert node workflows from types.
func (x *PluginWorkflowListResp) ConvertPluginWorkflowsFromTypes(num int64, workflows []*types.PluginWorkflow) {
	items := make([]*PluginWorkflowInfo, 0, len(workflows))
	for _, workflow := range workflows {
		item := newEmptyPluginWorkflow()

		*item.WorkflowId = workflow.WorkflowID
		*item.TriggerId = workflow.TriggerID
		item.BkHostId = workflow.HostIDs
		item.BkBizId = workflow.BizIDs

		*item.Type = string(workflow.Type)
		*item.Status = string(workflow.Status)
		*item.Operator = workflow.Operator
		*item.OperateTime = timeToUnixMilli(workflow.OperateTime)
		*item.FinishTime = timeToUnixMilli(workflow.FinishTime)

		items = append(items, item)
	}

	x.Data = &PluginWorkflowListResp_Data{
		Total: num,
		Items: items,
	}
}

// newEmptyPluginWorkflow creates a new empty PluginWorkflowInfo.
func newEmptyPluginWorkflow() *PluginWorkflowInfo {
	return &PluginWorkflowInfo{
		WorkflowId:  new(string),
		TriggerId:   new(string),
		Type:        new(string),
		BkHostId:    make([]int64, 0),
		BkBizId:     make([]int64, 0),
		Operator:    new(string),
		OperateTime: new(int64),
		FinishTime:  new(int64),
		Status:      new(string),
	}
}

// ===============================================================================
// PluginWorkflowStatistics Related Interfaces
// ===============================================================================

// PluginWorkflowStatistics represents the status of operations in a node workflow.
type PluginWorkflowStatistics struct {
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
func (x *PluginWorkflowStatisticsReq) Validate() error {
	for _, workflowID := range x.GetWorkflowId() {
		if workflowID == "" {
			return errors.New("workflow_id can not be empty")
		}
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginWorkflowStatisticsReq) AutoConvert() {
}

// ConvertConditionsToWorkflowConditionTypes convert conditions to types.
func (x *PluginWorkflowStatisticsReq) ConvertConditionsToWorkflowConditionTypes() (*types.PluginWorkflowCondition, error) {
	return convertPluginWorkflowConditionsToTypes(
		&PluginWorkflowExactConditions{
			WorkflowId: x.GetWorkflowId(),
		}, nil, nil)
}

// ConvertPluginWorkflowsFromTypes convert node workflows from types.
func (x *PluginWorkflowStatisticsResp) ConvertPluginWorkflowsFromTypes(result []*PluginWorkflowStatistics) {
	items := make([]*WorkflowStatisticsInfo, len(result))

	for i, item := range result {
		items[i] = newEmptyPluginWorkflowStatistics()
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

	x.Data = &PluginWorkflowStatisticsResp_Data{
		Items: items,
	}
}

// ConvertPluginWorkflowsFromDistribution converts plugin workflows from distribution and trigger mapping.
func (x *PluginWorkflowStatisticsResp) ConvertPluginWorkflowsFromDistribution(
	workflowIDs []string,
	distribution map[string]*operation.InstanceStatusDistribution,
	triggerToWorkflowMap map[string]string) error {

	result := make(map[string]*WorkflowStatisticsInfo, len(workflowIDs))
	for _, workflowID := range workflowIDs {
		info := newEmptyPluginWorkflowStatistics()
		*info.WorkflowId = workflowID
		result[workflowID] = info
	}

	for triggerID, dist := range distribution {
		workflowID, exists := triggerToWorkflowMap[triggerID]
		if !exists {
			continue
		}

		info, exists := result[workflowID]
		if !exists {
			continue
		}

		// add not inited count
		*info.TotalCount += dist.NotInitedCount

		for stateStr, count := range dist.StatusMap {
			state := operation.State(stateStr)
			if err := state.Validate(); err != nil {
				return err
			}

			*info.TotalCount += count
			switch state {
			case operation.StateInit:
				*info.InitCount += count
			case operation.StateRunning:
				*info.RunningCount += count
			case operation.StateLaunched:
				*info.LaunchedCount += count
			case operation.StateSuccess:
				*info.SuccessCount += count
			case operation.StateFailed:
				*info.FailedCount += count
			case operation.StateTimeout:
				*info.TimeoutCount += count
			case operation.StateTerminated:
				*info.TerminatedCount += count
			}
		}
	}

	x.Data = &PluginWorkflowStatisticsResp_Data{
		Items: conv.MapValueToSlice(result),
	}

	return nil
}

// ===============================================================================
// PluginWorkflowDistinct Related Interfaces
// ===============================================================================

// Validate check body.
func (x *PluginWorkflowDistinctReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PluginWorkflowDistinctReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PluginWorkflowDistinctReq) ConvertConditionsToTypes() (*types.PluginWorkflowCondition, error) {
	return convertPluginWorkflowConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(), nil)
}

// ConvertResultFromTypes convert result from types.
func (x *PluginWorkflowDistinctResp) ConvertResultFromTypes(result *types.PluginWorkflowDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &PluginWorkflowDistinctResp_Data{
		Type:     formatRespSlice(types.PluginWorkflowTypeListToStringList(result.Type)),
		BkHostId: formatRespSlice(result.HostID),
		Operator: formatRespSlice(result.Operator),
		Status:   formatRespSlice(types.PluginWorkflowStatusListToStringList(result.Status)),
	}
}

// ===============================================================================
// PluginWorkflowOperationList Related Interfaces
// ===============================================================================

// Validate check body.
func (x *PluginWorkflowOperationListReq) Validate() error {
	if x.GetWorkflowId() == "" {
		return errors.New("workflow_id can not be empty")
	}

	return validatePage(x.GetPage())
}

// AutoConvert auto convert.
func (x *PluginWorkflowOperationListReq) AutoConvert() {
}

// ConvertConditionsToComm convert conditions to comm.
func (x *PluginWorkflowOperationListReq) ConvertConditionsToComm() string {
	return x.GetWorkflowId()
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PluginWorkflowOperationListReq) ConvertConditionsFromTypes(condition *types.ApplicationPluginOperationListCondition) error {
	exactCond, fuzzyCond, err := convertPluginWorkOperConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

func convertPluginWorkOperConditionsFromTypes(condition *types.ApplicationPluginOperationListCondition) (
	*PluginWorkflowOperationExactConditions, *PluginWorkflowOperationFuzzyConditions, error) {

	if condition == nil {
		return nil, nil, nil
	}

	var exactCond *PluginWorkflowOperationExactConditions
	var fuzzyCond *PluginWorkflowOperationFuzzyConditions

	if condition.ExactInclude != nil {
		exactCond = &PluginWorkflowOperationExactConditions{
			BkHostId:      condition.ExactInclude.HostID,
			PluginName:    condition.ExactExclude.PluginName,
			PluginVersion: condition.ExactInclude.PluginVersion,
			State: conv.SliceToSlice(condition.ExactInclude.State, func(s operation.State) string {
				return string(s)
			}),
		}
	}

	if condition.FuzzyInclude != nil || condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, nil, errors.New("fuzzy-include, exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, nil
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PluginWorkflowOperationListReq) ConvertConditionsToTypes() *types.ApplicationPluginOperationListCondition {
	return convertPluginWorkflowOperationConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions())
}

func convertPluginWorkflowOperationConditionsToTypes(exactCond *PluginWorkflowOperationExactConditions,
	_ *PluginWorkflowOperationFuzzyConditions) *types.ApplicationPluginOperationListCondition {

	condition := &types.ApplicationPluginOperationListCondition{}
	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.ApplicationPluginOperationListExactFields{
			HostID:        exactCond.GetBkHostId(),
			PluginName:    exactCond.GetPluginName(),
			PluginVersion: exactCond.GetPluginVersion(),
			State:         operation.StringListToStateList(exactCond.GetState()),
		}
	}

	return condition
}

// ConvertPageToTypes convert page to types.
func (x *PluginWorkflowOperationListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertResultFromTypes convert workflow id to types.
func (x *PluginWorkflowOperationListResp) ConvertResultFromTypes(total int64, result []*types.PluginWorkflowListOperationResult) {
	items := make([]*PluginWorkflowOperation, len(result))
	for idx, op := range result {
		if op == nil {
			continue
		}

		items[idx] = &PluginWorkflowOperation{
			OperationId: op.OperationID,
			InstanceIds: op.OperInstanceIDs,
			Operator:    op.Operator,
			CreateTime:  timeToUnixMilli(op.CreateTime),
			PluginDeploymentInfo: &PluginDeploymentInfo{
				BkHostId:            op.HostID,
				BkBizId:             op.BizID,
				BkNetworkareaId:     op.NetworkAreaID,
				BkNetworkunitId:     op.NetworkUnitID,
				BkHostInneripList:   op.InnerIPList,
				BkHostInneripV6List: op.InnerIPV6List,
				PluginName:          op.PluginName,
				PluginVersion:       op.PluginVersion,
			},
			LatestOperInstBriefData: convertPluginWorkflowOperInstBriefDataFromTypes(op),
		}
	}

	x.Data = &PluginWorkflowOperationListResp_Data{
		Total:      total,
		Operations: items,
	}
}

func convertPluginWorkflowOperInstBriefDataFromTypes(data *types.PluginWorkflowListOperationResult) *WorkflowOperInstBriefData {
	if data == nil || data.LastInstanceBriefData == nil {
		return nil
	}

	return &WorkflowOperInstBriefData{
		LifeCycle:                 convertPluginWorkflowOperInstLifeCycleFromTypes(data.LastInstanceBriefData.Lifecycle),
		LatestActionInstBriefData: convertPluginWorkflowActionInstBriefDataFromTypes(data.LastInstanceBriefData.LatestActionInstBriefData),
	}
}

func convertPluginWorkflowActionInstBriefDataFromTypes(data *action.InstanceBriefData) *WorkflowActionInstBriefData {
	if data == nil {
		return &WorkflowActionInstBriefData{}
	}

	return &WorkflowActionInstBriefData{
		Name: data.Name,
		Tags: conv.SliceToSlice(data.Tags, func(tag action.Tag) string { return string(tag) }),
	}
}

func convertPluginWorkflowOperInstLifeCycleFromTypes(data *operation.Lifecycle) *WorkflowLifeCycle {
	if data == nil {
		return &WorkflowLifeCycle{}
	}

	return &WorkflowLifeCycle{
		State:      string(data.State),
		CreateTime: timeToUnixMilli(data.CreatedAt),
		StartTime:  timeToUnixMilli(data.StartedAt),
		EndTime:    timeToUnixMilli(data.EndedAt),
		StopTime:   timeToUnixMilli(data.StoppedAt),
	}
}

// GetCountOnly get count only.
func (x *PluginWorkflowOperationListResp) GetCountOnly() interface{} {
	if x.GetData() == nil {
		return &PluginWorkflowOperationListResp_Data{
			Total: 0,
		}
	}

	x.Data.Operations = nil

	return x.GetData()
}

// ===============================================================================
// PluginWorkflowOperationInstanceList Related Interfaces
// ===============================================================================

// Validate check body.
func (x *PluginWorkflowOperationInstanceListReq) Validate() error {
	if len(x.GetOperationId()) == 0 {
		return errors.New("operation_id is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginWorkflowOperationInstanceListReq) AutoConvert() {
}

// ConvertConditionsToComm ...
func (x *PluginWorkflowOperationInstanceListReq) ConvertConditionsToComm() []string {
	return x.GetOperationId()
}

// ConvertResultFromTypes convert node workflows from types.
func (x *PluginWorkflowOperationInstanceListResp) ConvertResultFromTypes(total int64,
	result []*operation.InstanceBriefData) {

	items := make([]*WorflowOperationInstanceData, 0, len(result))
	for _, opinstance := range result {
		oper := &WorflowOperationInstanceData{
			OperInstId:        opinstance.Metadata.OperationInstanceID,
			OperationId:       opinstance.Metadata.OperationID,
			OperInstStatus:    string(opinstance.Lifecycle.State),
			OperationDefName:  opinstance.Metadata.OperationDefName,
			ParentOperationId: opinstance.Metadata.ParentOperationID,
			ActionNames:       opinstance.Metadata.ActionNames,
			LifeCycle: &WorkflowLifeCycle{
				State:      string(opinstance.Lifecycle.State),
				CreateTime: timeToUnixMilli(opinstance.Lifecycle.CreatedAt),
				StartTime:  timeToUnixMilli(opinstance.Lifecycle.StartedAt),
				EndTime:    timeToUnixMilli(opinstance.Lifecycle.EndedAt),
				StopTime:   timeToUnixMilli(opinstance.Lifecycle.StoppedAt),
			},
		}

		items = append(items, oper)
	}

	x.Data = &PluginWorkflowOperationInstanceListResp_Data{
		OperInstData: items,
		Total:        total,
	}
}

// ===============================================================================
// PluginWorkflowOperationInstanceLogGet Related Interfaces
// ===============================================================================

// Validate check body.
func (x *PluginWorkflowOperationInstanceLogGetReq) Validate() error {
	if x.GetOperInstId() == "" {
		return errors.New("oper_inst_id is required and cannot be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginWorkflowOperationInstanceLogGetReq) AutoConvert() {

}

// ConvertResultFromTypes convert result from types.
func (x *PluginWorkflowOperationInstanceLogGetResp) ConvertResultFromTypes(result *operation.InstanceData) {
	if result == nil {
		return
	}

	operInstLogs := make(map[string]*WorkflowActionData)
	for actionID, v := range result.ActionInstanceDataMap {
		lifecycle := &WorkflowLifeCycle{
			State:      string(v.Lifecycle.State),
			CreateTime: timeToUnixMilli(v.Lifecycle.CreatedAt),
			StartTime:  timeToUnixMilli(v.Lifecycle.StartedAt),
			EndTime:    timeToUnixMilli(v.Lifecycle.EndedAt),
		}

		messages := make([]*WorkflowActionMessage_Message, 0, len(v.Messages))
		for _, msg := range v.Messages {
			messages = append(messages, &WorkflowActionMessage_Message{
				Time:   msg.Time.UnixMilli(),
				TextZh: msg.TextZh,
				TextEn: msg.TextEn,
				Level:  msg.Level,
			})
		}

		operInstLogs[actionID] = &WorkflowActionData{
			DisplayNameZh:   v.DisplayNameZh,
			DisplayNameEn:   v.DisplayNameEn,
			LifeCycle:       lifecycle,
			Message:         &WorkflowActionMessage{Logs: messages},
			SubWorkflowRefs: subWorkflowRefsFromPrivateData(v.PrivateData),
		}
	}

	extraExecutionLogs := &WorkflowActionMessage{
		Logs: make([]*WorkflowActionMessage_Message, 0),
	}
	for _, msg := range result.Metadata.ExtraExecutionMessages {
		extraExecutionLogs.Logs = append(extraExecutionLogs.Logs, &WorkflowActionMessage_Message{
			Time:   msg.Time.UnixMilli(),
			TextZh: msg.TextZh,
			TextEn: msg.TextEn,
			Level:  msg.Level,
		})
	}

	x.Data = &PluginWorkflowOperationInstanceLogGetResp_Data{
		OperInstLogs:       operInstLogs,
		ExtraExecutionLogs: extraExecutionLogs,
	}
}

// ===============================================================================
// PluginWorkflowOperationRetry Related Interfaces
// ===============================================================================

// AutoConvert auto convert.
func (x *PluginWorkflowOperationRetryReq) AutoConvert() {
}

// Validate convert workflow id.
func (x *PluginWorkflowOperationRetryReq) Validate() error {
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

// ConvertOperationRetryParamFromTypes convert operation retry param from types.
func (x *PluginWorkflowOperationRetryReq) ConvertPluginWorkflowOperationRetryParamFromTypes(retryParm *types.PluginWorkflowOperationRetryParam) {
	x.WorkflowId = retryParm.WorkflowID
	x.OperationIds = retryParm.OperationIDs
	x.RetryMod = string(retryParm.RetryMode)
}

// ConvertPluginWorkflowOperationRetryParamToTypes convert retry node operation param to types.
func (x *PluginWorkflowOperationRetryReq) ConvertPluginWorkflowOperationRetryParamToTypes() *types.PluginWorkflowOperationRetryParam {
	return &types.PluginWorkflowOperationRetryParam{
		WorkflowID:   x.GetWorkflowId(),
		OperationIDs: x.GetOperationIds(),
		RetryMode:    operation.RetryMode(x.GetRetryMod()),
	}
}

// ===============================================================================
// PluginWorkflowOperationTerminate Related Interfaces
// ===============================================================================

// AutoConvert auto convert.
func (x *PluginWorkflowOperationTerminateReq) AutoConvert() {
}

// Validate convert workflow id.
func (x *PluginWorkflowOperationTerminateReq) Validate() error {
	if x.GetWorkflowId() == "" {
		return errors.New("workflow_id is required")
	}

	if len(x.GetOperationIds()) == 0 {
		return errors.New("operation_id is required")
	}

	return nil
}

// ConvertOperationTerminateParamFromTypes convert operation terminate param from types.
func (x *PluginWorkflowOperationTerminateReq) ConvertPluginWorkflowOperationTerminateParamFromTypes(
	terminateParam *types.PluginWorkflowOperationTerminateParam) {

	x.WorkflowId = terminateParam.WorkflowID
	x.OperationIds = terminateParam.OperationIDs
}

// ConvertPluginWorkflowOperationTerminateParamToTypes convert terminate node operation param to types.
func (x *PluginWorkflowOperationTerminateReq) ConvertPluginWorkflowOperationTerminateParamToTypes() *types.PluginWorkflowOperationTerminateParam {
	return &types.PluginWorkflowOperationTerminateParam{
		WorkflowID:   x.GetWorkflowId(),
		OperationIDs: x.GetOperationIds(),
	}
}

// ===============================================================================
// Public Child Functions
// ===============================================================================

func convertPluginWorkflowConditionsToTypes(
	exactCond *PluginWorkflowExactConditions,
	_ *PluginWorkflowFuzzyConditions, timeRange *TimeRange) (*types.PluginWorkflowCondition, error) {

	condition := &types.PluginWorkflowCondition{}

	if timeRange != nil {
		operateTimeRange, err := convertTimeRangeToTypes(timeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to convert operate time range: %w", err)
		}
		condition.OperateTimeRange = operateTimeRange
	}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.PluginWorkflowExactFields{
			HostID:        exactCond.GetBkHostId(),
			BizID:         exactCond.GetBkBizId(),
			Type:          types.StringListToPluginWorkflowTypeList(exactCond.GetType()),
			Status:        types.StringListToPluginWorkflowStatusList(exactCond.GetStatus()),
			WorkflowID:    exactCond.GetWorkflowId(),
			Operator:      exactCond.GetOperator(),
			HostInnerIP:   exactCond.GetBkHostInnerip(),
			HostInnerIPV6: exactCond.GetBkHostInneripV6(),
		}
	}

	return condition, nil
}

func newEmptyPluginWorkflowStatistics() *WorkflowStatisticsInfo {
	return &WorkflowStatisticsInfo{
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

// ===============================================================================
// PluginWorkflowOperationDistinct Related Interfaces
// ===============================================================================

// Validate check body.
func (x *PluginWorkflowOperationDistinctReq) Validate() error {
	if x.GetWorkflowId() == "" {
		return errors.New("workflow_id can not be empty")
	}
	return nil
}

// AutoConvert auto convert.
func (x *PluginWorkflowOperationDistinctReq) AutoConvert() {
}

// GetWorkflowID get workflow id.
func (x *PluginWorkflowOperationDistinctReq) GetWorkflowID() string {
	return x.GetWorkflowId()
}

// ConvertResultFromTypes convert result from types.
func (x *PluginWorkflowOperationDistinctResp) ConvertResultFromTypes(result *types.WorkflowOperationDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &PluginWorkflowOperationDistinctResp_Data{
		State: formatRespSlice(conv.SliceToSlice(result.State, func(s operation.State) string {
			return string(s)
		})),
	}
}
