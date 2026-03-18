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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// plugin workflow list max limit
	maxPluginWorkflowLimit = 500
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

// PageTimeout return page timeout.
func (x *PluginWorkflowListReq) PageTimeout() time.Duration {
	return backendPagingListTimeout
}

// PageLimit return page limit.
func (x *PluginWorkflowListReq) PageLimit() int {
	return maxPluginWorkflowLimit
}

// ConvertPageToTypes convert page to types.
func (x *PluginWorkflowListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
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
		item.BkHostId = workflow.HostIDs
		*item.Type = string(workflow.Type)
		*item.Status = string(workflow.Status)
		*item.TriggerId = workflow.TriggerID
		*item.Operator = workflow.Operator
		*item.OperateTime = workflow.OperateTime.UnixMilli()
		*item.FinishTime = workflow.FinishTime.UnixMilli()
		items = append(items, item)
	}

	x.Data = &PluginWorkflowListResp_Data{
		Total: num,
		Items: items,
	}
}

// ConvertPluginWorkflowsToTypes convert node workflows to types.
func (x *PluginWorkflowListResp) ConvertPluginWorkflowsToTypes() ([]*types.PluginWorkflow, int64) {
	data := x.GetData()
	if data == nil {
		return nil, 0
	}

	items := data.GetItems()
	result := make([]*types.PluginWorkflow, len(items))

	for idx, item := range items {
		workflow := &types.PluginWorkflow{
			TenantID:    item.GetTenantId(),
			WorkflowID:  item.GetWorkflowId(),
			TriggerID:   item.GetTriggerId(),
			Type:        types.PluginWorkflowType(item.GetType()),
			Status:      types.PluginWorkflowStatus(item.GetStatus()),
			HostIDs:     item.GetBkHostId(),
			Operator:    item.GetOperator(),
			OperateTime: time.UnixMilli(item.GetOperateTime()),
			FinishTime:  time.UnixMilli(item.GetFinishTime()),
		}

		result[idx] = workflow
	}

	return result, data.GetTotal()
}

func newEmptyPluginWorkflow() *PluginWorkflowInfo {
	return &PluginWorkflowInfo{
		WorkflowId:  new(string),
		Type:        new(string),
		TriggerId:   new(string),
		BkHostId:    make([]int64, 0),
		Operator:    new(string),
		OperateTime: new(int64),
		FinishTime:  new(int64),
		Status:      new(string),
	}
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

// ConvertConditionsFromTypes convert conditions from types.
func (x *PluginWorkflowDistinctReq) ConvertConditionsFromTypes(condition *types.PluginWorkflowCondition) error {
	exactCond, fuzzyCond, _, err := convertPluginWorkConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PluginWorkflowDistinctReq) ConvertConditionsToTypes() (*types.PluginWorkflowCondition, error) {
	return convertPluginWorkflowConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(),
		nil)
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

// ConvertWorkflowDistinctToTypes ...
func (x *PluginWorkflowDistinctResp) ConvertWorkflowDistinctToTypes() *types.PluginWorkflowDistinctResult {
	data := x.GetData()
	if data == nil {
		return nil
	}

	result := &types.PluginWorkflowDistinctResult{
		Type:     types.StringListToPluginWorkflowTypeList(data.GetType()),
		HostID:   data.GetBkHostId(),
		Status:   types.StringListToPluginWorkflowStatusList(data.GetStatus()),
		Operator: data.GetOperator(),
	}

	return result
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

// GetWorkflowID get workflow id.
func (x *PluginWorkflowOperationListReq) GetWorkflowID() string {
	return x.GetWorkflowId()
}

// PageLimit return page limit.
func (x *PluginWorkflowOperationListReq) PageLimit() int {
	return maxPluginWorkflowLimit
}

// ConvertPageToTypes convert page to types.
func (x *PluginWorkflowOperationListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
}

// ConvertConditionsToOperationTypes convert conditions to operation types.
func (x *PluginWorkflowOperationListReq) ConvertConditionsToOperationTypes(triggerID string) *types.OperationCondition {
	return convertPluginWorkflowOperationConditionsToOperationTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), triggerID)
}

func convertPluginWorkflowOperationConditionsToOperationTypes(exactCond *PluginWorkflowOperationListReq_ExactConditions,
	_ *PluginWorkflowOperationListReq_FuzzyConditions, triggerID string) *types.OperationCondition {

	condition := &types.OperationCondition{
		ExactInclude: &types.OperationExactFields{
			TriggerID: []string{triggerID},
		},
	}

	if exactCond != nil {
		condition.ExactInclude.State = operation.StringListToStateList(exactCond.GetState())
	}

	return condition
}

// ConvertConditionsToDeploymentTypes convert conditions to deployment types.
func (x *PluginWorkflowOperationListReq) ConvertConditionsToDeploymentTypes(tokens []string) *types.PluginDeploymentCondition {
	return convertPluginWorkflowOperationConditionsToPluginDeploymentTypes(tokens, x.GetExactIncludeConditions())
}

func convertPluginWorkflowOperationConditionsToPluginDeploymentTypes(tokens []string,
	exactCond *PluginWorkflowOperationListReq_ExactConditions) *types.PluginDeploymentCondition {

	condition := &types.PluginDeploymentCondition{
		ExactInclude: &types.PluginDeploymentExactFields{},
	}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.PluginDeploymentExactFields{
			HostID:        exactCond.GetBkHostId(),
			PluginName:    exactCond.GetPluginName(),
			PluginVersion: exactCond.GetPluginVersion(),
		}
	}

	// exact conditions token
	condition.ExactInclude.Token = tokens

	return condition
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
	*PluginWorkflowOperationListReq_ExactConditions, *PluginWorkflowOperationListReq_FuzzyConditions, error) {

	if condition == nil {
		return nil, nil, nil
	}

	var exactCond *PluginWorkflowOperationListReq_ExactConditions
	var fuzzyCond *PluginWorkflowOperationListReq_FuzzyConditions

	if condition.ExactInclude != nil {
		exactCond = &PluginWorkflowOperationListReq_ExactConditions{
			BkHostId:      condition.ExactInclude.HostID,
			PluginName:    condition.ExactInclude.PluginName,
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
	return convertPluginWorkflowOperationConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions())
}

func convertPluginWorkflowOperationConditionsToTypes(exactCond *PluginWorkflowOperationListReq_ExactConditions,
	_ *PluginWorkflowOperationListReq_FuzzyConditions) *types.ApplicationPluginOperationListCondition {

	condition := &types.ApplicationPluginOperationListCondition{}
	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.ApplicationPluginOperationListExactFields{
			HostID:        exactCond.GetBkHostId(),
			PluginName:    exactCond.GetPluginName(),
			PluginVersion: exactCond.GetPluginVersion(),
		}
	}

	return condition
}

// ConvertResultFromTypes convert workflow id to types.
func (x *PluginWorkflowOperationListResp) ConvertResultFromTypes(total int64, result []*types.PluginWorkflowListOperationResult) {
	items := make([]*PluginWorkflowOperationListResp_PluginWorkflowOperation, len(result))
	for idx, op := range result {
		item := &PluginWorkflowOperationListResp_PluginWorkflowOperation{
			OperationId: op.OperationID,
			InstanceIds: op.OperInstanceIDs,
			Operator:    op.Operator,
			CreateTime:  op.CreateTime.UnixMilli(),

			PluginDeploymentInfo: &PluginWorkflowOperationListResp_PluginDeploymentInfo{
				BkHostId:            op.HostID,
				BkBizId:             op.BizID,
				BkNetworkareaId:     op.NetworkAreaID,
				BkNetworkunitId:     op.NetworkUnitID,
				BkHostInneripList:   op.InnerIPList,
				BkHostInneripV6List: op.InnerIPV6List,
				PluginName:          op.PluginName,
				PluginVersion:       op.PluginVersion,
			},
		}

		if op.LastInstanceBriefData != nil {
			item.LatestOperInstBriefData = &PluginWorkflowOperationListResp_LatestOperInstBriefData{
				LifeCycle:                 convertPluginWorkflowOperInstLifeCycleFromTypes(op.LastInstanceBriefData.Lifecycle),
				LatestActionInstBriefData: convertPluginWorkflowActionInstBriefDataFromTypes(op.LastInstanceBriefData.LatestActionInstBriefData),
			}
		}

		items[idx] = item
	}

	x.Data = &PluginWorkflowOperationListResp_Data{
		Total:      total,
		Operations: items,
	}
}

func convertPluginWorkflowOperInstLifeCycleFromTypes(data *operation.Lifecycle) *WorkflowLifeCycle {
	if data == nil {
		return &WorkflowLifeCycle{}
	}

	return &WorkflowLifeCycle{
		State:      string(data.State),
		CreateTime: data.CreatedAt.UnixMilli(),
		StartTime:  data.StartedAt.UnixMilli(),
		EndTime:    data.EndedAt.UnixMilli(),
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

// ConvertWorkflowOperationToTypes convert workflow operation to types.
func (x *PluginWorkflowOperationListResp) ConvertWorkflowOperationToTypes() ([]*types.PluginWorkflowListOperationResult, int64) {
	data := x.GetData()
	if data == nil {
		return nil, 0
	}

	items := data.GetOperations()
	result := make([]*types.PluginWorkflowListOperationResult, len(items))

	for idx, item := range items {
		operResult := &types.PluginWorkflowListOperationResult{
			OperationID:     item.GetOperationId(),
			OperInstanceIDs: item.GetInstanceIds(),
			Operator:        item.GetOperator(),
			CreateTime:      time.UnixMilli(item.GetCreateTime()),
		}

		deployInfo := item.GetPluginDeploymentInfo()
		if deployInfo != nil {
			operResult.HostID = deployInfo.GetBkHostId()
			operResult.BizID = deployInfo.GetBkBizId()
			operResult.NetworkAreaID = deployInfo.GetBkNetworkareaId()
			operResult.NetworkUnitID = deployInfo.GetBkNetworkunitId()
			operResult.InnerIPList = deployInfo.GetBkHostInneripList()
			operResult.InnerIPV6List = deployInfo.GetBkHostInneripV6List()
			operResult.PluginName = deployInfo.GetPluginName()
			operResult.PluginVersion = deployInfo.GetPluginVersion()
		}

		latestOperInstBriefData := item.GetLatestOperInstBriefData()
		if latestOperInstBriefData != nil {
			operResult.LastInstanceBriefData = &operation.InstanceBriefData{
				Lifecycle:                 convertPluginWorkflowOperInstLifeCycleToTypes(latestOperInstBriefData.GetLifeCycle()),
				LatestActionInstBriefData: convertPluginWorkflowActionInstBriefDataToTypes(latestOperInstBriefData.GetLatestActionInstBriefData()),
			}
		}

		result[idx] = operResult
	}

	return result, data.GetTotal()
}

func convertPluginWorkflowOperInstLifeCycleToTypes(data *WorkflowLifeCycle) *operation.Lifecycle {
	if data == nil {
		return &operation.Lifecycle{}
	}

	return &operation.Lifecycle{
		State:     operation.State(data.GetState()),
		CreatedAt: time.UnixMilli(data.GetCreateTime()),
		StartedAt: time.UnixMilli(data.GetStartTime()),
		EndedAt:   time.UnixMilli(data.GetEndTime()),
	}
}

func convertPluginWorkflowActionInstBriefDataToTypes(data *WorkflowActionInstBriefData) *action.InstanceBriefData {
	if data == nil {
		return &action.InstanceBriefData{}
	}

	return &action.InstanceBriefData{
		Name: data.GetName(),
		Tags: conv.SliceToSlice(data.GetTags(), func(tag string) action.Tag { return action.Tag(tag) }),
	}
}

// ===============================================================================
// PluginWorkflowOperationInstanceList Related Interfaces
// ===============================================================================

// Validate check body.
func (x *PluginWorkflowOperationInstanceListReq) Validate() error {
	if len(x.GetOperationId()) == 0 || x.GetOperationId()[0] == "" {
		return errors.New("operation id is required")
	}

	return nil
}

// ConvertConditionsToComm ...
func (x *PluginWorkflowOperationInstanceListReq) ConvertConditionsToComm() []string {
	return x.GetOperationId()
}

// AutoConvert auto convert.
func (x *PluginWorkflowOperationInstanceListReq) AutoConvert() {
}

// ConvertResultFromTypes ...
func (x *PluginWorkflowOperationInstanceListResp) ConvertResultFromTypes(
	num int64, result []*operation.InstanceBriefData) {

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
				CreateTime: opinstance.Lifecycle.CreatedAt.UnixMilli(),
				StartTime:  opinstance.Lifecycle.StartedAt.UnixMilli(),
				EndTime:    opinstance.Lifecycle.EndedAt.UnixMilli(),
				StopTime:   opinstance.Lifecycle.StoppedAt.UnixMilli(),
			},
		}
		items = append(items, oper)
	}

	x.Data = &PluginWorkflowOperationInstanceListResp_Data{
		OperInstData: items,
		Total:        num,
	}
}

// ConvertOperationInstanceFromTypes  ...
func (x *PluginWorkflowOperationInstanceListResp) ConvertOperationInstanceFromTypes(
	total int64, result *operation.InstanceBriefData) {

	items := make([]*WorflowOperationInstanceData, 1)

	oper := &WorflowOperationInstanceData{
		OperInstId:        result.Metadata.OperationInstanceID,
		OperationId:       result.Metadata.OperationID,
		OperInstStatus:    string(result.Lifecycle.State),
		OperationDefName:  result.Metadata.OperationDefName,
		ParentOperationId: result.Metadata.ParentOperationID,
		ActionNames:       result.Metadata.ActionNames,
		LifeCycle: &WorkflowLifeCycle{
			State:      string(result.Lifecycle.State),
			CreateTime: result.Lifecycle.CreatedAt.UnixMilli(),
			StartTime:  result.Lifecycle.StartedAt.UnixMilli(),
			EndTime:    result.Lifecycle.EndedAt.UnixMilli(),
			StopTime:   result.Lifecycle.StoppedAt.UnixMilli(),
		},
	}

	items[0] = oper
	x.Data = &PluginWorkflowOperationInstanceListResp_Data{
		OperInstData: items,
		Total:        total,
	}
}

// ConvertWorkflowOperationInstanceToTypes ...
func (x *PluginWorkflowOperationInstanceListResp) ConvertWorkflowOperationInstanceToTypes() (
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
				CreatedAt: time.UnixMilli(item.GetLifeCycle().GetCreateTime()),
				StartedAt: time.UnixMilli(item.GetLifeCycle().GetStartTime()),
				EndedAt:   time.UnixMilli(item.GetLifeCycle().GetEndTime()),
			},
		}

		result[idx] = inst
	}

	return result, total
}

// ===============================================================================
// PluginWorkflowOperationInstanceLogGet Related Interfaces
// ===============================================================================

// Validate check body.
func (x *PluginWorkflowOperationInstanceLogGetReq) Validate() error {
	if x.GetOperInstId() == "" {
		return errors.New("workflow_id is required and cannot be empty")
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
			CreateTime: v.Lifecycle.CreatedAt.UnixMilli(),
			StartTime:  v.Lifecycle.StartedAt.UnixMilli(),
			EndTime:    v.Lifecycle.EndedAt.UnixMilli(),
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
			DisplayNameZh: v.DisplayNameZh,
			DisplayNameEn: v.DisplayNameEn,
			LifeCycle:     lifecycle,
			Message:       &WorkflowActionMessage{Logs: messages},
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

// ConvertWorkflowOperationInstanceLogToTypes convert workflow operation instance log to types.
func (x *PluginWorkflowOperationInstanceLogGetResp) ConvertWorkflowOperationInstanceLogToTypes() *operation.InstanceData {
	data := x.GetData()
	if data == nil {
		return nil
	}

	result := &operation.InstanceData{
		ActionInstanceDataMap: make(map[string]*action.InstanceData),
		InstanceBriefData: operation.InstanceBriefData{
			Metadata: &operation.InstanceMetadata{
				ExtraExecutionMessages: make([]common.Message, 0),
			},
		},
	}

	for actionID, actionData := range x.GetData().GetOperInstLogs() {
		if actionData == nil {
			continue
		}

		instance := &action.InstanceData{
			DisplayNameZh: actionData.GetDisplayNameZh(),
			DisplayNameEn: actionData.GetDisplayNameEn(),
			Lifecycle: &action.Lifecycle{
				State:     action.State(actionData.GetLifeCycle().GetState()),
				CreatedAt: time.UnixMilli(actionData.GetLifeCycle().GetCreateTime()),
				StartedAt: time.UnixMilli(actionData.GetLifeCycle().GetStartTime()),
				EndedAt:   time.UnixMilli(actionData.GetLifeCycle().GetEndTime()),
			},
		}

		if actionData.GetMessage() != nil {
			instance.Messages = make([]common.Message, 0, len(actionData.GetMessage().GetLogs()))
			for _, msg := range actionData.GetMessage().GetLogs() {
				if msg == nil {
					continue
				}
				instance.Messages = append(instance.Messages, common.Message{
					Time:   time.UnixMilli(msg.GetTime()),
					TextZh: msg.GetTextZh(),
					TextEn: msg.GetTextEn(),
					Level:  msg.GetLevel(),
				})
			}
		}

		result.ActionInstanceDataMap[actionID] = instance
	}

	for _, msg := range x.GetData().GetExtraExecutionLogs().GetLogs() {
		if msg == nil {
			continue
		}

		result.InstanceBriefData.Metadata.ExtraExecutionMessages = append(result.InstanceBriefData.Metadata.ExtraExecutionMessages, common.Message{
			Time:   time.UnixMilli(msg.GetTime()),
			TextZh: msg.GetTextZh(),
			TextEn: msg.GetTextEn(),
			Level:  msg.GetLevel(),
		})
	}

	return result
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
func (x *PluginWorkflowOperationRetryReq) ConvertOperationRetryParamFromTypes(retryParm *types.PluginWorkflowOperationRetryParam) {
	x.WorkflowId = retryParm.WorkflowID
	x.OperationIds = retryParm.OperationIDs
	x.RetryMod = string(retryParm.RetryMode)
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
func (x *PluginWorkflowOperationTerminateReq) ConvertOperationTerminateParamFromTypes(terminateParam *types.PluginWorkflowOperationTerminateParam) {
	x.WorkflowId = terminateParam.WorkflowID
	x.OperationIds = terminateParam.OperationIDs
}

// ===============================================================================
// PluginWorkflowOperationInstanceStatusDistributionList Related Interfaces
// ===============================================================================

// Validate validates the request.
func (x *PluginWorkflowOperationInstanceStatusDistributionListReq) Validate() error {
	return nil
}

// AutoConvert auto converts the request.
func (x *PluginWorkflowOperationInstanceStatusDistributionListReq) AutoConvert() {
}

// ConvertDistributionFromTypes converts distribution from types.
func (x *PluginWorkflowOperationInstanceStatusDistributionListResp) ConvertDistributionFromTypes(
	distribution map[string]*operation.InstanceStatusDistribution) {

	if distribution == nil {
		return
	}

	items := make(map[string]*PluginWorkflowOperationInstanceStatusDistributionListResp_StatusDistribution)
	for triggerID, dist := range distribution {
		items[triggerID] = &PluginWorkflowOperationInstanceStatusDistributionListResp_StatusDistribution{
			NotInitedCount: dist.NotInitedCount,
			StateCounts:    dist.StatusMap,
		}
	}

	x.Data = &PluginWorkflowOperationInstanceStatusDistributionListResp_Data{
		Items: items,
	}
}

// ConvertDistributionToTypes converts distribution to types.
func (x *PluginWorkflowOperationInstanceStatusDistributionListResp) ConvertDistributionToTypes() map[string]*operation.InstanceStatusDistribution {
	data := x.GetData()
	if data == nil {
		return make(map[string]*operation.InstanceStatusDistribution)
	}

	result := make(map[string]*operation.InstanceStatusDistribution)
	for triggerID, statusDist := range data.GetItems() {
		if statusDist == nil {
			continue
		}
		result[triggerID] = &operation.InstanceStatusDistribution{
			NotInitedCount: statusDist.GetNotInitedCount(),
			StatusMap:      statusDist.GetStateCounts(),
		}
	}

	return result
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
			HostID:     exactCond.GetBkHostId(),
			Type:       types.StringListToPluginWorkflowTypeList(exactCond.GetType()),
			Status:     types.StringListToPluginWorkflowStatusList(exactCond.GetStatus()),
			WorkflowID: exactCond.GetWorkflowId(),
			Operator:   exactCond.GetOperator(),
		}
	}

	return condition, nil
}

func convertPluginWorkConditionsFromTypes(condition *types.PluginWorkflowCondition) (
	*PluginWorkflowExactConditions, *PluginWorkflowFuzzyConditions, *TimeRange, error) {

	if condition == nil {
		return nil, nil, nil, nil
	}

	var exactCond *PluginWorkflowExactConditions
	var fuzzyCond *PluginWorkflowFuzzyConditions
	var timeRange *TimeRange

	if condition.OperateTimeRange != nil {
		timeRange = &TimeRange{
			StartTimestampSec: condition.OperateTimeRange.StartTime.Unix(),
			EndTimestampSec:   condition.OperateTimeRange.EndTime.Unix(),
		}
	}

	if condition.ExactInclude != nil {
		exactCond = &PluginWorkflowExactConditions{
			BkHostId:   condition.ExactInclude.HostID,
			WorkflowId: condition.ExactInclude.WorkflowID,
			Type:       types.PluginWorkflowTypeListToStringList(condition.ExactInclude.Type),
			Status:     types.PluginWorkflowStatusListToStringList(condition.ExactInclude.Status),
			Operator:   condition.ExactInclude.Operator,
		}
	}

	if condition.FuzzyInclude != nil || condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, nil, nil, errors.New("fuzzy-include, exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, timeRange, nil
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

// ConvertSelectorToTypes convert selector to types.
func (x *PluginWorkflowOperationDistinctReq) ConvertSelectorToTypes() types.WorkflowOperationDistinctSelector {
	if x.GetSelector() == nil {
		return types.WorkflowOperationDistinctSelector{}
	}

	selector := x.GetSelector()
	return types.WorkflowOperationDistinctSelector{
		State: selector.GetState(),
	}
}

// ConvertSelectorFromTypes convert selector from types.
func (x *PluginWorkflowOperationDistinctReq) ConvertSelectorFromTypes(selector types.WorkflowOperationDistinctSelector) {
	x.Selector = &PluginWorkflowOperationDistinctSelector{
		State: selector.State,
	}
}

// ConvertResultFromTypes convert result from types.
func (x *PluginWorkflowOperationDistinctResp) ConvertResultFromTypes(result *types.WorkflowOperationDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &PluginWorkflowOperationDistinctResp_Data{
		State: conv.SliceToSlice(result.State, func(s operation.State) string {
			return string(s)
		}),
	}
}

// ConvertOperationDistinctToTypes convert operation distinct to types.
func (x *PluginWorkflowOperationDistinctResp) ConvertOperationDistinctToTypes() (*types.WorkflowOperationDistinctResult, error) {
	if x.GetData() == nil {
		return nil, fmt.Errorf("distinct operation failed, get empty data")
	}

	var err error
	result := new(types.WorkflowOperationDistinctResult)
	result.State, err = conv.SliceToSliceWithError(
		x.GetData().GetState(),
		func(s string) (operation.State, error) {
			state := operation.State(s)
			if err := state.Validate(); err != nil {
				return "", err
			}

			return state, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	return result, nil
}
