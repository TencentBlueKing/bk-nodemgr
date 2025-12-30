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

	items := make([]*NodeWorkflowInfo, len(workflows))
	for idx, workflow := range workflows {
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

		items[idx] = item
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
func (x *NodeWorkflowListReq) ConvertConditionsToTypes() (*types.NodeWorkflowCondition, error) {
	return convertNodeWorkflowConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(),
		x.GetOperateTimeRange())
}

// AutoConvert auto convert.
func (x *NodeWorkflowStatisticsReq) AutoConvert() {
}

// ConvertConditionsToWorkflowConditionTypes convert conditions to types.
func (x *NodeWorkflowStatisticsReq) ConvertConditionsToWorkflowConditionTypes() (*types.NodeWorkflowCondition, error) {
	return convertNodeWorkflowConditionsToTypes(
		&NodeWorkflowExactConditions{
			WorkflowId: x.GetWorkflowId(),
		}, nil, nil)
}

// ConvertNodeWorkflowsFromTypes convert node workflows from types.
func (x *NodeWorkflowStatisticsResp) ConvertNodeWorkflowsFromTypes(result []*NodeWorkflowStatistics) {
	items := make([]*WorkflowStatisticsInfo, len(result))

	for i, item := range result {
		items[i] = newEmptyNodeWorkflowStatistics()
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

// ConvertNodeWorkflowsFromDistribution converts node workflows from distribution and trigger mapping.
func (x *NodeWorkflowStatisticsResp) ConvertNodeWorkflowsFromDistribution(
	workflowIDs []string,
	distribution map[string]*operation.InstanceStatusDistribution,
	triggerToWorkflowMap map[string]string) error {

	result := make(map[string]*WorkflowStatisticsInfo, len(workflowIDs))
	for _, workflowID := range workflowIDs {
		info := newEmptyNodeWorkflowStatistics()
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

	x.Data = &NodeWorkflowStatisticsResp_Data{
		Items: conv.MapValueToSlice(result),
	}
	return nil
}

// Validate check body.
func (x *NodeWorkflowDistinctReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *NodeWorkflowDistinctReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *NodeWorkflowDistinctReq) ConvertConditionsToTypes() (*types.NodeWorkflowCondition, error) {
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
	if x.GetWorkflowId() == "" {
		return errors.New("workflow_id can not be empty")
	}

	return validatePage(x.GetPage())
}

// GetWorkflowID get workflow id.
func (x *NodeWorkflowOperationListReq) GetWorkflowID() string {
	return x.GetWorkflowId()
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *NodeWorkflowOperationListReq) ConvertConditionsFromTypes(condition *types.ApplicationNodeOperationListCondition) error {
	exactCond, err := convertNodeWorkOperConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

// ConvertConditionsToTypes convert conditions to types.
func (x *NodeWorkflowOperationListReq) ConvertConditionsToTypes() *types.ApplicationNodeOperationListCondition {
	return convertNodeWorkflowOperationConditionsToTypes(
		x.GetExactIncludeConditions())
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
	total int64, result []*types.NodeWorkflowListOperationResult) {

	items := make([]*NodeWorkflowOperation, len(result))
	for idx, op := range result {
		if op == nil {
			continue
		}

		items[idx] = &NodeWorkflowOperation{
			OperationId: op.OperationID,
			InstanceIds: op.OperInstanceIDs,
			Operator:    op.Operator,
			CreateTime:  op.CreateTime.UnixMilli(),
			NodeDeploymentInfo: &NodeDeploymentInfo{
				BkHostId:            op.HostID,
				BkBizId:             op.BizID,
				BkHostInneripList:   op.InnerIPList,
				BkHostInneripV6List: op.InnerIPV6List,
				BkNetworkareaId:     op.NetworkAreaID,
				BkNetworkunitId:     op.NetworkUnitID,
				NodeVersion:         op.NodeVersion,
			},
			LatestOperInstBriefData: convertNodeWorkflowOperInstBriefDataFromTypes(op),
		}
	}

	x.Data = &NodeWorkflowOperationListResp_Data{
		Total:      total,
		Operations: items,
	}
}

// Validate check body.
func (x *NodeWorkflowOperationInstanceListReq) Validate() error {
	return nil
}

// ConvertConditionsToTypes convert conditions to types.
func (x *NodeWorkflowOperationInstanceListReq) ConvertConditionsToTypes() *types.OperInstDataCondition {
	return &types.OperInstDataCondition{
		ExactInclude: &types.OperInstDataExactFields{
			OperationID: []string{x.GetOperationId()},
		},
	}
}

// AutoConvert auto convert.
func (x *NodeWorkflowOperationInstanceListReq) AutoConvert() {
}

// ConvertResultFromTypes convert node workflows from types.
func (x *NodeWorkflowOperationInstanceListResp) ConvertResultFromTypes(total int64,
	result []*operation.InstanceBriefData) {

	items := make([]*WorflowOperationInstanceData, len(result))
	for idx, opinstance := range result {
		items[idx] = &WorflowOperationInstanceData{
			OperInstId:        opinstance.Metadata.OperationInstanceID,
			OperationId:       opinstance.Metadata.OperationID,
			OperInstStatus:    string(opinstance.Lifecycle.State),
			OperationDefName:  opinstance.Metadata.OperationDefName,
			ParentOperationId: opinstance.Metadata.ParentOperationID,
			ActionNames:       opinstance.Metadata.ActionNames,
			LifeCycle: &WorkflowLifeCycle{
				State:      string(opinstance.Lifecycle.State),
				CreateTime: opinstance.Lifecycle.CreatedAt.UnixMilli(),
				StartTime:  opinstance.Lifecycle.StartedAt.UnixMilli(),
				EndTime:    opinstance.Lifecycle.EndedAt.UnixMilli(),
				StopTime:   opinstance.Lifecycle.StoppedAt.UnixMilli(),
			},
		}
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

	operInstLogs := make(map[string]*WorkflowActionData)
	for actionID, v := range result.ActionInstanceDataMap {
		lifecycle := &WorkflowLifeCycle{
			State:      string(v.Lifecycle.State),
			CreateTime: v.Lifecycle.CreatedAt.UnixMilli(),
			StartTime:  v.Lifecycle.StartedAt.UnixMilli(),
			EndTime:    v.Lifecycle.EndedAt.UnixMilli(),
		}

		messages := make([]*WorkflowActionMessage_Message, len(v.Messages))
		for idx, msg := range v.Messages {
			messages[idx] = &WorkflowActionMessage_Message{
				Time:  msg.Time.UnixMilli(),
				Text:  msg.Text,
				Level: msg.Level,
			}
		}

		operInstLogs[actionID] = &WorkflowActionData{
			LifeCycle: lifecycle,
			Message:   &WorkflowActionMessage{Logs: messages},
		}
	}

	x.Data = &NodeWorkflowOperationInstanceLogGetResp_Data{
		OperInstLogs: operInstLogs,
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

// AutoConvert auto convert.
func (x *NodeWorkflowOperationManualSolutionGetReq) AutoConvert() {
}

// Validate convert workflow id.
func (x *NodeWorkflowOperationManualSolutionGetReq) Validate() error {
	if x.GetWorkflowId() == "" {
		return errors.New("workflow_id is required")
	}

	if len(x.GetOperationId()) == 0 {
		return errors.New("operation_id is required")
	}

	return nil
}

const (
	manualStepTypeCommand       = "command"
	manualStepTypeNetworkPolicy = "network_policy"
	manualStepTypeDescription   = "description"
	manualStepTypeDownload      = "download"
)

// ConvertResultFromTypes convert ManualInfo to ManualSolution.
func (x *NodeWorkflowOperationManualSolutionGetResp) ConvertResultFromTypes(manualInfo *types.NodeWorkflowOperationManualInfo) {
	if manualInfo == nil {
		x.Data = []*ManualSolution{}
		return
	}

	solutions := make([]*ManualSolution, 0)
	for _, cmd := range manualInfo.Commands {
		switch cmd.Type {
		case types.NodeWorkflowOperationManualCommandTypeBash:
			solutions = append(solutions, &ManualSolution{
				Type:          string(cmd.Type),
				DescriptionEn: "Use bash to manually install node.",
				DescriptionZh: "使用 bash 手动安装节点.",
				Steps: []*ManualSolutionStep{
					{
						Type:      manualStepTypeCommand,
						NameEn:    "Execute bash command",
						NameZh:    "执行 bash 命令",
						ContentEn: cmd.Command,
						ContentZh: cmd.Command,
					},
				},
			})
		case types.NodeWorkflowOperationManualCommandTypeBat:
			solutions = append(solutions, &ManualSolution{
				Type:          string(cmd.Type),
				DescriptionEn: "Use bat to manually install node.",
				DescriptionZh: "使用 bat 手动安装节点.",
				Steps: []*ManualSolutionStep{
					{
						Type:      manualStepTypeCommand,
						NameEn:    "Execute bat command",
						NameZh:    "执行 bat 命令",
						ContentEn: cmd.Command,
						ContentZh: cmd.Command,
					},
				},
			})
		}
	}

	x.Data = solutions
}

func convertNodeWorkflowConditionsToTypes(
	exactCond *NodeWorkflowExactConditions,
	_ *NodeWorkflowFuzzyConditions, timeRange *TimeRange) (*types.NodeWorkflowCondition, error) {

	condition := &types.NodeWorkflowCondition{}

	if timeRange != nil {
		operateTimeRange, err := convertTimeRangeToTypes(timeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to convert operate time range: %w", err)
		}
		condition.OperateTimeRange = operateTimeRange
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

	return condition, nil
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

func convertNodeWorkOperConditionsFromTypes(condition *types.ApplicationNodeOperationListCondition) (
	*NodeWorkflowOperationExactConditions, error) {

	var exactCond *NodeWorkflowOperationExactConditions

	if condition.ExactInclude != nil {
		exactCond = &NodeWorkflowOperationExactConditions{
			BkBizId: condition.ExactInclude.BizID,
			State: conv.SliceToSlice(condition.ExactInclude.State, func(s operation.State) string {
				return string(s)
			}),
			BkNetworkareaId: condition.ExactInclude.NetworkAreaID,
			BkNetworkunitId: condition.ExactInclude.NetworkUnitID,
			BkHostInnerip:   condition.ExactInclude.HostInnerIP,
			BkHostInneripV6: condition.ExactInclude.HostInnerIPV6,
			NodeVersion:     condition.ExactInclude.NodeVersion,
		}
	}

	if condition.FuzzyInclude != nil || condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, errors.New("fuzzy-include, exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, nil
}

func convertNodeWorkflowOperationConditionsToTypes(
	exactCond *NodeWorkflowOperationExactConditions) *types.ApplicationNodeOperationListCondition {

	condition := &types.ApplicationNodeOperationListCondition{}
	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.ApplicationNodeOperationListExactFields{
			BizID:         exactCond.GetBkBizId(),
			HostInnerIP:   exactCond.GetBkHostInnerip(),
			HostInnerIPV6: exactCond.GetBkHostInneripV6(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			NetworkUnitID: exactCond.GetBkNetworkunitId(),
			State:         operation.StringListToStateList(exactCond.GetState()),
		}
	}

	return condition
}

func convertNodeWorkflowOperInstBriefDataFromTypes(data *types.NodeWorkflowListOperationResult) *WorkflowOperInstBriefData {
	if data == nil || data.LastInstanceBriefData == nil {
		return nil
	}

	return &WorkflowOperInstBriefData{
		LifeCycle:                 convertNodeWorkflowOperInstLifeCycleFromTypes(data.LastInstanceBriefData.Lifecycle),
		LatestActionInstBriefData: convertNodeWorkflowActionInstBriefDataFromTypes(data.LastInstanceBriefData.LatestActionInstBriefData),
	}
}

func convertNodeWorkflowActionInstBriefDataFromTypes(data *action.InstanceBriefData) *WorkflowActionInstBriefData {
	if data == nil {
		return &WorkflowActionInstBriefData{}
	}

	return &WorkflowActionInstBriefData{
		Name: data.Name,
		Tags: conv.SliceToSlice(data.Tags, func(tag action.Tag) string { return string(tag) }),
	}
}

func convertNodeWorkflowOperInstLifeCycleFromTypes(data *operation.Lifecycle) *WorkflowLifeCycle {
	if data == nil {
		return &WorkflowLifeCycle{}
	}

	return &WorkflowLifeCycle{
		State:      string(data.State),
		CreateTime: data.CreatedAt.UnixMilli(),
		StartTime:  data.StartedAt.UnixMilli(),
		EndTime:    data.EndedAt.UnixMilli(),
		StopTime:   data.StoppedAt.UnixMilli(),
	}
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

func newEmptyNodeWorkflowStatistics() *WorkflowStatisticsInfo {
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
