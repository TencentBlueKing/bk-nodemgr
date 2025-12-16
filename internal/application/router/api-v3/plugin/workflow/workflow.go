/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflow describes the workflow router.
package workflow

import (
	"fmt"
	"sort"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/gin-gonic/gin"
)

const (
	maxWorkflowLimit  = 500
	maxOperationLimit = 500
)

type handler struct {
	rg             *gin.RouterGroup
	backendHandler backend.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/workflow"),
		backendHandler: capability.BackendHandler,
	}
}

// Load loads workflow handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.List))
	h.rg.POST("/statistics", restserver.Handler(h.Statistics))
	h.rg.POST("/distinct", restserver.Handler(h.Distinct))
	h.rg.POST("/operation/list", restserver.Handler(h.ListOperation))
	h.rg.POST("/operation/retry", restserver.Handler(h.RetryOperation))
	h.rg.POST("/operation/terminate", restserver.Handler(h.TerminateOperation))
	h.rg.POST("/operation/instance/list", restserver.Handler(h.ListOperationInstance))
	h.rg.POST("/operation/instance/log/get", restserver.Handler(h.GetOperationInstanceLog))
}

// List plugin workflows.
func (h *handler) List(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PluginWorkflowListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	resp := new(protoApplication.PluginWorkflowListResp)
	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountPluginWorkflow(
			rCtx,
			req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to count plugin workflow")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
		resp.ConvertPluginWorkflowsFromTypes(num, nil)
		// only count, no data.
		return resp.GetData(), nil
	}

	workflows, num, err := h.backendHandler.ListPluginWorkflow(rCtx, req.ConvertPageToTypes(maxWorkflowLimit), req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertPluginWorkflowsFromTypes(num, workflows)

	return resp.GetData(), nil
}

// TODO：如果数据量过大可能要分页或者拆协程查询
// Statistics plugin workflow statistics.
func (h *handler) Statistics(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PluginWorkflowStatisticsReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics plugin workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflows, _, err := h.backendHandler.ListPluginWorkflow(
		rCtx, types.UnlimitedPage(), req.ConvertConditionsToWorkflowConditionTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics plugin workflow, failed to list plugin workflow")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	var instanceStatus []*operation.InstanceStatus
	if len(workflows) > 0 {
		instanceStatus, err = h.backendHandler.ListPluginWorkflowOperationInstanceStatus(rCtx, convertWorkflowToTriggerID(workflows))
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow instance status: %v", err)
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
	}

	workflowStatusMap := make(map[string]map[string]*types.PluginWorkflowOperationStatus)
	for _, instatus := range instanceStatus {
		innerMap, exists := workflowStatusMap[instatus.TriggerID]
		if !exists {
			innerMap = make(map[string]*types.PluginWorkflowOperationStatus)
			workflowStatusMap[instatus.TriggerID] = innerMap
		}

		current, exists := innerMap[instatus.OperationID]
		if !exists || current.Index < instatus.Index {
			innerMap[instatus.OperationID] = &types.PluginWorkflowOperationStatus{
				OperationID: instatus.OperationID,
				Index:       instatus.Index,
				TriggerID:   instatus.TriggerID,
				State:       types.PluginWorkflowOperationState(instatus.State),
			}
		}
	}

	result := calculateStats(workflows, workflowStatusMap, req.GetWorkflowId())

	resp := new(protoApplication.PluginWorkflowStatisticsResp)
	resp.ConvertPluginWorkflowsFromTypes(result)

	return resp.GetData(), nil
}

// Distinct plugin workflow distinct.
func (h *handler) Distinct(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PluginWorkflowDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.backendHandler.DistinctPluginWorkflow(
		rCtx,
		types.NewPluginWorkflowDistinctRequestAllSet(),
		req.ConvertConditionsToTypes())
	resp := new(protoApplication.PluginWorkflowDistinctResp)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin workflow: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// ListOperation list plugin workflow operation.
func (h *handler) ListOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PluginWorkflowOperationListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation, failed to decode request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation, failed to validate request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, req.Validate())
	}

	var targetStates []string
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		targetStates = exactCond.GetState()
	}

	result, num, err := h.backendHandler.ListPluginWorkflowOperation(
		rCtx, req.ConvertPageToTypes(maxOperationLimit), req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	if len(result) == 0 {
		resp := new(protoApplication.PluginWorkflowOperationListResp)
		resp.ConvertResultFromTypes(num, nil, nil)

		if req.GetOnlyCount() {
			return resp.GetCountOnly(), nil
		}

		return resp.GetData(), nil
	}

	operationIDs := make([]string, 0, len(result))
	for _, operation := range result {
		operationIDs = append(operationIDs, operation.OperationID)
	}

	// get all operation instances.
	// need to grouby operation id. than use last instance status and calculate total time.
	allInstances, _, err := h.backendHandler.ListPluginWorkflowOperationInstance(rCtx, operationIDs...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	// group by operation id.
	instancesByOpID := groupInstancesByOperationID(allInstances)

	// calculate each operation total time and last state.
	summaries, err := calculateOperationSummaries(operationIDs, instancesByOpID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to calculate operation summary: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	// filter by states.
	filteredResults, filteredSummaries := filterOperationsByState(
		result,
		summaries,
		targetStates,
	)

	resp := new(protoApplication.PluginWorkflowOperationListResp)
	resp.ConvertResultFromTypes(num, filteredResults, filteredSummaries)

	if req.GetOnlyCount() {
		return resp.GetCountOnly(), nil
	}

	return resp.GetData(), nil
}

// ListOperationInstance list plugin workflow operation instance.
func (h *handler) ListOperationInstance(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PluginWorkflowOperationInstanceListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to decode request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to validate request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, req.Validate())
	}

	resp := new(protoApplication.PluginWorkflowOperationInstanceListResp)

	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountPluginWorkflowOperationInstance(rCtx, req.ConvertConditionsToComm()...)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to count operation instance: %v", err)
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
		resp.ConvertResultFromTypes(num, nil)
		// only count, no data.
		return resp.GetData(), nil
	}

	instances, num, err := h.backendHandler.ListPluginWorkflowOperationInstance(rCtx, req.ConvertConditionsToComm()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertResultFromTypes(num, instances)

	return resp.GetData(), nil
}

// GetOperationInstanceLog get operation instance log.
func (h *handler) GetOperationInstanceLog(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PluginWorkflowOperationInstanceLogGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get operation instance log, failed to decode request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	logs, err := h.backendHandler.GetPluginWorkflowOperationInstanceLog(
		rCtx,
		req.GetOperInstId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get operation instance log: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PluginWorkflowOperationInstanceLogGetResp)

	resp.ConvertResultFromTypes(logs)

	return resp.GetData(), nil
}

// RetryOperation retry operation.
func (h *handler) RetryOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PluginWorkflowOperationRetryReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to retry plugin workflow operation, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	err := h.backendHandler.RetryPluginWorkflowOperation(rCtx, req.ConvertPluginWorkflowOperationRetryParamToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to retry plugin workflow operation")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}
	resp := new(protoApplication.PluginWorkflowOperationRetryResp)

	return resp.GetData(), nil
}

// TerminateOperation terminate operation.
func (h *handler) TerminateOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PluginWorkflowOperationTerminateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to terminate plugin workflow operation, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	err := h.backendHandler.TerminatePluginWorkflowOperation(rCtx, req.ConvertPluginWorkflowOperationTerminateParamToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to terminate plugin workflow operation")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PluginWorkflowOperationTerminateResp)

	return resp.GetData(), nil
}

func convertWorkflowToTriggerID(workflows []*types.PluginWorkflow) *types.PluginWorkflowOperInstanceStatusCondition {
	triggerIDs := make([]string, len(workflows))
	for i, workflow := range workflows {
		triggerIDs[i] = workflow.TriggerID
	}

	return &types.PluginWorkflowOperInstanceStatusCondition{
		ExactInclude: &types.PluginWorkflowOperInstanceStatusExactFields{
			TriggerID: triggerIDs,
		},
	}
}

func calculateStats(workflows []*types.PluginWorkflow, statusMap map[string]map[string]*types.PluginWorkflowOperationStatus,
	reqIDs []string) []*protoApplication.PluginWorkflowStatistics {

	idIndexMap := make(map[string]int)
	for i, id := range reqIDs {
		idIndexMap[id] = i
	}

	result := make([]*protoApplication.PluginWorkflowStatistics, len(reqIDs))
	for i, id := range reqIDs {
		result[i] = &protoApplication.PluginWorkflowStatistics{WorkflowID: id}
	}

	for _, workflow := range workflows {
		idx, exists := idIndexMap[workflow.WorkflowID]
		if !exists {
			continue
		}

		if innerMap, exists := statusMap[workflow.TriggerID]; exists {
			statusList := result[idx]
			for _, inst := range innerMap {
				statusList.TotalCount++
				switch inst.State {
				case types.PluginWorkflowOperationStateInit:
					statusList.InitCount++
				case types.PluginWorkflowOperationStateRunning:
					statusList.RunningCount++
				case types.PluginWorkflowOperationStateLaunched:
					statusList.LaunchedCount++
				case types.PluginWorkflowOperationStateSuccess:
					statusList.SuccessCount++
				case types.PluginWorkflowOperationStateFailed:
					statusList.FailedCount++
				case types.PluginWorkflowOperationStateTimeout:
					statusList.TimeoutCount++
				case types.PluginWorkflowOperationStateTerminated:
					statusList.TerminatedCount++
				}
			}
		}
	}

	return result
}

func groupInstancesByOperationID(
	instances []*operation.InstanceBriefData) map[string][]*operation.InstanceBriefData {

	grouped := make(map[string][]*operation.InstanceBriefData)
	for _, instance := range instances {
		opID := instance.Metadata.OperationID
		grouped[opID] = append(grouped[opID], instance)
	}

	return grouped
}

func filterOperationsByState(
	ops []*types.PluginWorkflowListOperationResult,
	summaries []*types.PluginWorkflowOperationSummary,
	targetStates []string) (
	[]*types.PluginWorkflowListOperationResult, []*types.PluginWorkflowOperationSummary) {

	// no state filter required. directly return.
	if len(targetStates) == 0 {
		return ops, summaries
	}

	// record the states that need to be filtered.
	targetStateSet := make(map[types.PluginWorkflowOperationState]struct{}, len(targetStates))
	for _, state := range targetStates {
		targetStateSet[types.PluginWorkflowOperationState(state)] = struct{}{}
	}

	matchedOperations := make([]*types.PluginWorkflowListOperationResult, len(ops))
	matchedSummaries := make([]*types.PluginWorkflowOperationSummary, len(ops))

	for i, op := range ops {
		summary := summaries[i]
		if _, ok := targetStateSet[summary.LastStatus]; !ok {
			continue
		}

		matchedOperations[i] = op
		matchedSummaries[i] = summary
	}

	return matchedOperations, matchedSummaries
}

func calculateOperationSummaries(operationIDs []string,
	instancesByOpID map[string][]*operation.InstanceBriefData) ([]*types.PluginWorkflowOperationSummary, error) {

	summaries := make([]*types.PluginWorkflowOperationSummary, len(operationIDs))

	for idx, opID := range operationIDs {
		instances, exists := instancesByOpID[opID]

		// no instances found.
		if !exists || len(instances) == 0 {
			summaries[idx] = &types.PluginWorkflowOperationSummary{
				TotalDuration: 0,
				LastStatus:    types.PluginWorkflowOperationStateInit,
			}

			continue
		}

		sort.Slice(instances, func(i, j int) bool {
			return instances[i].Lifecycle.CreatedAt.Before(instances[j].Lifecycle.CreatedAt)
		})

		// calculate total duration.
		var totalSeconds int64
		for _, inst := range instances {
			if inst.Lifecycle.CreatedAt.IsZero() || inst.Lifecycle.EndedAt.IsZero() {
				continue
			}

			totalSeconds += inst.Lifecycle.EndedAt.Unix() - inst.Lifecycle.CreatedAt.Unix()
		}

		// calculate last status.
		lastInstance := instances[len(instances)-1]
		lastStatus, err := types.InstanceStatusToPluginWorkflowOperationState(lastInstance.Lifecycle.State)
		if err != nil {
			return nil, fmt.Errorf("failed to convert instance status to operation state: %w", err)
		}

		summaries[idx] = &types.PluginWorkflowOperationSummary{
			TotalDuration: totalSeconds,
			LastStatus:    lastStatus,
		}
	}

	return summaries, nil
}
