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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
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
	items := make([]*NodeWorkflowInfo, len(workflows))
	for idx, workflow := range workflows {
		item := newEmptyNodeWorkflow()
		*item.WorkflowId = workflow.WorkflowID
		item.BkBizId = workflow.BizIDs
		*item.Type = string(workflow.Type)
		*item.Status = string(workflow.Status)
		*item.TriggerId = workflow.TriggerID
		*item.Operator = workflow.Operator
		*item.OperateTime = workflow.OperateTime.UnixMilli()
		*item.FinishTime = workflow.FinishTime.UnixMilli()
		items[idx] = item
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
			TenantID:    item.GetTenantId(),
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

// GetWorkflowID get workflow id.
func (x *NodeWorkflowOperationListReq) GetWorkflowID() string {
	return x.GetExactIncludeConditions().GetWorkflowId()
}

// ConvertConditionsToDeploymentTypes convert conditions to deployment types.
func (x *NodeWorkflowOperationListReq) ConvertConditionsToDeploymentTypes(tokens []string) *types.NodeDeploymentCondition {
	return convertNodeWorkflowOperationConditionsToNodeDeploymentTypes(tokens,
		x.GetExactIncludeConditions(),
	)
}

func convertNodeWorkflowOperationConditionsToNodeDeploymentTypes(tokens []string,
	exactCond *NodeWorkflowOperationExactConditions) *types.NodeDeploymentCondition {

	condition := &types.NodeDeploymentCondition{
		ExactInclude: &types.NodeDeploymentExactFields{},
	}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.NodeDeploymentExactFields{
			InnerIP:       exactCond.GetBkHostInnerip(),
			InnerIPv6:     exactCond.GetBkHostInneripV6(),
			BizID:         exactCond.GetBkBizId(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			NetworkUnitID: exactCond.GetBkNetworkunitId(),
			NodeVersion:   exactCond.GetNodeVersion(),
		}
	}

	// exact conditions token
	condition.ExactInclude.Token = tokens

	return condition
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

	return convertNodeWorkflowOperationConditionsToTypes(
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
func (x *NodeWorkflowOperationListResp) ConvertResultFromTypes(total int64, result []*types.NodeWorkflowListOperationResult) {
	items := make([]*NodeWorkflowOperation, len(result))
	for idx, op := range result {
		item := &NodeWorkflowOperation{
			OperationId:         op.OperationID,
			InstanceIds:         op.OperInstanceIDs,
			BkNetworkareaId:     op.NetworkAreaID,
			BkNetworkunitId:     op.NetworkUnitID,
			BkBizId:             op.BizID,
			BkHostId:            op.HostID,
			BkHostInneripList:   op.InnerIPList,
			BkHostInneripV6List: op.InnerIPV6List,
			NodeVersion:         op.NodeVersion,
			Operator:            op.Operator,
		}
		items[idx] = item
	}

	x.Data = &NodeWorkflowOperationListResp_Data{
		Total:      total,
		Operations: items,
	}
}

// ConvertWorkflowOperationToTypes convert workflow operation to types.
func (x *NodeWorkflowOperationListResp) ConvertWorkflowOperationToTypes() ([]*types.NodeWorkflowListOperationResult, int64) {
	data := x.GetData()
	if data == nil {
		return nil, 0
	}

	items := data.GetOperations()
	result := make([]*types.NodeWorkflowListOperationResult, len(items))

	for idx, item := range items {
		operation := &types.NodeWorkflowListOperationResult{
			OperationID:     item.GetOperationId(),
			Operator:        item.GetOperator(),
			OperInstanceIDs: item.GetInstanceIds(),
			NetworkAreaID:   item.GetBkNetworkareaId(),
			NetworkUnitID:   item.GetBkNetworkunitId(),
			BizID:           item.GetBkBizId(),
			HostID:          item.GetBkHostId(),
			InnerIPList:     item.GetBkHostInneripList(),
			InnerIPV6List:   item.GetBkHostInneripV6List(),
			NodeVersion:     item.GetNodeVersion(),
		}

		result[idx] = operation
	}

	return result, data.GetTotal()
}

// Validate check body.
func (x *NodeWorkflowOperationInstanceListReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *NodeWorkflowOperationInstanceListReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *NodeWorkflowOperationInstanceListReq) ConvertConditionsToTypes() []string {
	return x.GetOperationId()
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *NodeWorkflowOperationInstanceListReq) ConvertConditionsFromTypes(condition *types.OperInstDataCondition) {
	if condition == nil {
		return
	}

	if condition.ExactInclude != nil {
		x.OperationId = condition.ExactInclude.OperationID
		x.OperInstId = condition.ExactInclude.OperInstID
	}
}

// ConvertResultFromTypes ...
func (x *NodeWorkflowOperationInstanceListResp) ConvertResultFromTypes(
	num int64, result []*operation.InstanceBriefData) {

	items := make([]*WorflowOperationInstanceData, len(result))
	for idx, opinstance := range result {
		items[idx] = &WorflowOperationInstanceData{
			OperInstId:                opinstance.Metadata.OperationInstanceID,
			OperationId:               opinstance.Metadata.OperationID,
			OperInstStatus:            string(opinstance.Lifecycle.State),
			OperationDefName:          opinstance.Metadata.OperationDefName,
			ParentOperationId:         opinstance.Metadata.ParentOperationID,
			ActionNames:               opinstance.Metadata.ActionNames,
			LifeCycle:                 convertNodeWorkflowOperInstLifeCycleFromTypes(opinstance.Lifecycle),
			LatestActionInstBriefData: convertNodeWorkflowActionInstBriefDataFromTypes(opinstance.LatestActionInstBriefData),
		}
	}

	x.Data = &NodeWorkflowOperationInstanceListResp_Data{
		OperInstData: items,
		Total:        num,
	}
}

// ConvertOperationInstanceFromTypes convert operation instance from types.
func (x *NodeWorkflowOperationInstanceListResp) ConvertOperationInstanceFromTypes(
	total int64, result *operation.InstanceBriefData) {

	instance := &WorflowOperationInstanceData{
		OperInstId:                result.Metadata.OperationInstanceID,
		OperationId:               result.Metadata.OperationID,
		OperInstStatus:            string(result.Lifecycle.State),
		OperationDefName:          result.Metadata.OperationDefName,
		ParentOperationId:         result.Metadata.ParentOperationID,
		ActionNames:               result.Metadata.ActionNames,
		LifeCycle:                 convertNodeWorkflowOperInstLifeCycleFromTypes(result.Lifecycle),
		LatestActionInstBriefData: convertNodeWorkflowActionInstBriefDataFromTypes(result.LatestActionInstBriefData),
	}

	x.Data = &NodeWorkflowOperationInstanceListResp_Data{
		OperInstData: []*WorflowOperationInstanceData{instance},
		Total:        total,
	}
}

// ConvertWorkflowOperationInstanceToTypes convert workflow operation instance to types.
func (x *NodeWorkflowOperationInstanceListResp) ConvertWorkflowOperationInstanceToTypes() (
	[]*operation.InstanceBriefData, int64) {

	data := x.GetData()
	if data == nil {
		return nil, 0
	}

	items := data.GetOperInstData()
	result := make([]*operation.InstanceBriefData, len(items))

	for idx, item := range items {
		inst := &operation.InstanceBriefData{
			Metadata: &operation.InstanceMetadata{
				OperationInstanceID: item.GetOperInstId(),
				OperationID:         item.GetOperationId(),
				OperationDefName:    item.GetOperationDefName(),
				ParentOperationID:   item.GetParentOperationId(),
				ActionNames:         item.GetActionNames(),
			},
			Lifecycle:                 convertNodeWorkflowOperInstLifeCycleToTypes(item.GetLifeCycle()),
			LatestActionInstBriefData: convertNodeWorkflowActionInstBriefDataToTypes(item.GetLatestActionInstBriefData()),
		}

		result[idx] = inst
	}

	return result, data.GetTotal()
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

	operInstLogs := make(map[string]*WorkflowActionData)
	for actionID, v := range result.ActionInstanceDataMap {
		lifecycle := &WorkflowLifeCycle{
			State:      string(v.Lifecycle.State),
			CreateTime: v.Lifecycle.CreatedAt.Unix(),
			StartTime:  v.Lifecycle.StartedAt.Unix(),
			EndTime:    v.Lifecycle.EndedAt.Unix(),
		}

		messages := make([]*WorkflowActionMessage_Message, len(v.Messages))
		for idx, msg := range v.Messages {
			messages[idx] = &WorkflowActionMessage_Message{
				Time:  msg.Time.Unix(),
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
			instance.Messages = make([]common.Message, 0, len(actionData.GetMessage().GetLogs()))
			for _, msg := range actionData.GetMessage().GetLogs() {
				if msg == nil {
					continue
				}
				instance.Messages = append(instance.Messages, common.Message{
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

// ConvertOperationTerminateParamFromTypes convert operation terminate param from types.
func (x *NodeWorkflowOperationTerminateReq) ConvertOperationTerminateParamFromTypes(
	terminateParam *types.NodeWorkflowOperationTerminateParam) {

	x.WorkflowId = terminateParam.WorkflowID
	x.OperationIds = terminateParam.OperationIDs
}

// ConvertOperationRetryParamFromTypes convert operation retry param from types.
func (x *NodeWorkflowOperationRetryReq) ConvertOperationRetryParamFromTypes(retryParm *types.NodeWorkflowOperationRetryParam) {
	x.WorkflowId = retryParm.WorkflowID
	x.OperationIds = retryParm.OperationIDs
	x.RetryMod = string(retryParm.RetryMode)
}

// AutoConvert auto convert.
func (x *NodeWorkflowOperationManualInfoGetReq) AutoConvert() {
}

// Validate convert workflow id.
func (x *NodeWorkflowOperationManualInfoGetReq) Validate() error {
	if x.GetWorkflowId() == "" {
		return errors.New("workflow_id is required")
	}

	if x.GetOperationId() == "" {
		return errors.New("operation_id is required")
	}

	return nil
}

// ConvertManualInfoFromTypes convert manual info from types.
func (x *NodeWorkflowOperationManualInfoGetResp) ConvertManualInfoFromTypes(manualInfo *types.NodeWorkflowOperationManualInfo) {
	commands := make([]*ManualCommand, 0)
	for _, cmd := range manualInfo.Commands {
		commands = append(commands, &ManualCommand{
			Type:    string(cmd.Type),
			Command: cmd.Command,
		})
	}

	networkPolicies := make([]*NetworkPolicy, 0)
	for _, policy := range manualInfo.NetworkPolicies {
		networkPolicies = append(networkPolicies, &NetworkPolicy{
			Name:          policy.Name,
			DescriptionEn: policy.DescriptionEN,
			DescriptionZh: policy.DescriptionZH,
			Source: &NetworkPolicyEndpoint{
				Name:          policy.Source.Name,
				Type:          string(policy.Source.Type),
				Values:        policy.Source.Values,
				DescriptionEn: policy.Source.DescriptionEN,
				DescriptionZh: policy.Source.DescriptionZH,
			},
			Target: &NetworkPolicyEndpoint{
				Name:          policy.Target.Name,
				Type:          string(policy.Target.Type),
				Values:        policy.Target.Values,
				DescriptionEn: policy.Target.DescriptionEN,
				DescriptionZh: policy.Target.DescriptionZH,
			},
			Service: &NetworkPolicyService{
				Protocol:      string(policy.Service.Protocol),
				Ports:         policy.Service.Ports,
				DescriptionEn: policy.Service.DescriptionEN,
				DescriptionZh: policy.Service.DescriptionZH,
			},
		})
	}

	x.Data = &ManualInfo{
		Commands:        commands,
		NetworkPolicies: networkPolicies,
	}
}

// ConvertManualInfoToTypes convert manual info to types.
func (x *NodeWorkflowOperationManualInfoGetResp) ConvertManualInfoToTypes() *types.NodeWorkflowOperationManualInfo {
	commands := make([]*types.NodeWorkflowOperationManualCommand, 0)
	for _, cmd := range x.GetData().GetCommands() {
		commands = append(commands, &types.NodeWorkflowOperationManualCommand{
			Type:    types.NodeWorkflowOperationManualCommandType(cmd.GetType()),
			Command: cmd.GetCommand(),
		})
	}

	networkPolicies := make([]*types.NetworkPolicy, 0)
	for _, policy := range x.GetData().GetNetworkPolicies() {
		networkPolicies = append(networkPolicies, &types.NetworkPolicy{
			Name:          policy.GetName(),
			DescriptionEN: policy.GetDescriptionEn(),
			DescriptionZH: policy.GetDescriptionZh(),
			Source: types.NetworkPolicyEndpoint{
				Name:          policy.GetSource().GetName(),
				Type:          criteria.NetEndpointType(policy.GetSource().GetType()),
				Values:        policy.GetSource().GetValues(),
				DescriptionEN: policy.GetSource().GetDescriptionEn(),
				DescriptionZH: policy.GetSource().GetDescriptionZh(),
			},
			Target: types.NetworkPolicyEndpoint{
				Name:          policy.GetTarget().GetName(),
				Type:          criteria.NetEndpointType(policy.GetTarget().GetType()),
				Values:        policy.GetTarget().GetValues(),
				DescriptionEN: policy.GetTarget().GetDescriptionEn(),
				DescriptionZH: policy.GetTarget().GetDescriptionZh(),
			},
			Service: types.NetworkPolicyService{
				Protocol:      criteria.NetType(policy.GetService().GetProtocol()),
				Ports:         policy.GetService().GetPorts(),
				DescriptionEN: policy.GetService().GetDescriptionEn(),
				DescriptionZH: policy.GetService().GetDescriptionZh(),
			},
		})
	}

	return &types.NodeWorkflowOperationManualInfo{
		Commands:        commands,
		NetworkPolicies: networkPolicies,
	}
}

// Validate validates the request.
func (x *NodeWorkflowOperationInstanceStatusDistributionListReq) Validate() error {
	return nil
}

// AutoConvert auto converts the request.
func (x *NodeWorkflowOperationInstanceStatusDistributionListReq) AutoConvert() {
}

// ConvertDistributionFromTypes converts distribution from types.
func (x *NodeWorkflowOperationInstanceStatusDistributionListResp) ConvertDistributionFromTypes(
	distribution map[string]map[string]int64) {

	if distribution == nil {
		return
	}

	items := make(map[string]*NodeWorkflowOperationInstanceStatusDistributionListResp_StatusDistribution)
	for triggerID, stateMap := range distribution {
		items[triggerID] = &NodeWorkflowOperationInstanceStatusDistributionListResp_StatusDistribution{
			StateCounts: stateMap,
		}
	}

	x.Data = &NodeWorkflowOperationInstanceStatusDistributionListResp_Data{
		Items: items,
	}
}

// ConvertDistributionToTypes converts distribution to types.
func (x *NodeWorkflowOperationInstanceStatusDistributionListResp) ConvertDistributionToTypes() map[string]map[string]int64 {
	data := x.GetData()
	if data == nil {
		return make(map[string]map[string]int64)
	}

	result := make(map[string]map[string]int64)
	for triggerID, statusDist := range data.GetItems() {
		if statusDist == nil {
			continue
		}
		result[triggerID] = statusDist.GetStateCounts()
	}

	return result
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

func convertNodeWorkflowOperationConditionsToTypes(
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

func convertNodeWorkflowActionInstBriefDataFromTypes(data *action.InstanceBriefData) *WorkflowActionInstBriefData {
	if data == nil {
		return &WorkflowActionInstBriefData{}
	}

	return &WorkflowActionInstBriefData{
		Name: data.Name,
		Tags: conv.SliceToSlice(data.Tags, func(tag action.Tag) string { return string(tag) }),
	}
}

func convertNodeWorkflowActionInstBriefDataToTypes(data *WorkflowActionInstBriefData) *action.InstanceBriefData {
	if data == nil {
		return &action.InstanceBriefData{}
	}

	return &action.InstanceBriefData{
		Name: data.GetName(),
		Tags: conv.SliceToSlice(data.GetTags(), func(tag string) action.Tag { return action.Tag(tag) }),
	}
}

func convertNodeWorkflowOperInstLifeCycleFromTypes(data *operation.Lifecycle) *WorkflowLifeCycle {
	if data == nil {
		return &WorkflowLifeCycle{}
	}

	return &WorkflowLifeCycle{
		State:      string(data.State),
		CreateTime: data.CreatedAt.Unix(),
		StartTime:  data.StartedAt.Unix(),
		EndTime:    data.EndedAt.Unix(),
	}
}

func convertNodeWorkflowOperInstLifeCycleToTypes(data *WorkflowLifeCycle) *operation.Lifecycle {
	if data == nil {
		return &operation.Lifecycle{}
	}

	return &operation.Lifecycle{
		State:     operation.State(data.GetState()),
		CreatedAt: time.Unix(data.GetCreateTime(), 0),
		StartedAt: time.Unix(data.GetStartTime(), 0),
		EndedAt:   time.Unix(data.GetEndTime(), 0),
	}
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
