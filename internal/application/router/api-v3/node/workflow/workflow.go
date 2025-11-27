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

// List workflows.
func (h *handler) List(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	resp := new(protoApplication.NodeWorkflowListResp)
	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountNodeWorkflow(
			rCtx,
			req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list workflow, failed to count workflow")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
		resp.ConvertNodeWorkflowsFromTypes(num, nil, nil)
		// only count, no data.
		return resp.GetData(), nil
	}

	workflows, num, err := h.backendHandler.ListNodeWorkflow(rCtx,
		req.ConvertPageToTypes(maxWorkflowLimit), req.ConvertConditionsToTypes())

	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list workflow")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	businessMap, err := h.listAllBusiness(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list business")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertNodeWorkflowsFromTypes(num, workflows, businessMap)

	return resp.GetData(), nil
}

// TODO：如果数据量过大可能要分页或者拆协程查询
// Statistics workflow statistics.
func (h *handler) Statistics(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowStatisticsReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflows, _, err := h.backendHandler.ListNodeWorkflow(
		rCtx, types.UnlimitedPage(), req.ConvertConditionsToWorkflowConditionTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics workflow, failed to list workflow")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	var instanceStatus []*operation.InstanceStatus
	if len(workflows) > 0 {
		instanceStatus, err = h.backendHandler.ListNodeWorkflowOperationInstanceStatus(
			rCtx, convertWorkflowToTriggerID(workflows))
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list workflow instance status: %v", err)
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
	}

	workflowStatusMap := make(map[string]map[string]*types.NodeWorkflowOperationStatus)
	for _, instatus := range instanceStatus {
		innerMap, exists := workflowStatusMap[instatus.TriggerID]
		if !exists {
			innerMap = make(map[string]*types.NodeWorkflowOperationStatus)
			workflowStatusMap[instatus.TriggerID] = innerMap
		}

		current, exists := innerMap[instatus.OperationID]
		if !exists || current.Index < instatus.Index {
			innerMap[instatus.OperationID] = &types.NodeWorkflowOperationStatus{
				OperationID: instatus.OperationID,
				Index:       instatus.Index,
				TriggerID:   instatus.TriggerID,
				State:       types.NodeWorkflowOperationState(instatus.State),
			}
		}
	}

	result := calculateStats(workflows, workflowStatusMap, req.GetWorkflowId())

	resp := new(protoApplication.NodeWorkflowStatisticsResp)
	resp.ConvertNodeWorkflowsFromTypes(result)

	return resp.GetData(), nil
}

// Distinct workflow distinct.
func (h *handler) Distinct(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.backendHandler.DistinctNodeWorkflow(
		rCtx,
		types.NewNodeWorkflowDistinctRequestAllSet(),
		req.ConvertConditionsToTypes())
	resp := new(protoApplication.NodeWorkflowDistinctResp)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct workflow: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// ListOperation list workflow operation.
func (h *handler) ListOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationListReq)
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

	result, num, err := h.backendHandler.ListNodeWorkflowOperation(
		rCtx, req.ConvertPageToTypes(maxOperationLimit), req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	if len(result) == 0 {
		resp := new(protoApplication.NodeWorkflowOperationListResp)
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
	allInstances, _, err := h.backendHandler.ListNodeWorkflowOperationInstance(rCtx, operationIDs...)
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

	resp := new(protoApplication.NodeWorkflowOperationListResp)
	resp.ConvertResultFromTypes(num, filteredResults, filteredSummaries)

	if req.GetOnlyCount() {
		return resp.GetCountOnly(), nil
	}

	return resp.GetData(), nil
}

// ListOperationInstance list workflow operation instance.
func (h *handler) ListOperationInstance(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationInstanceListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to decode request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to validate request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, req.Validate())
	}

	resp := new(protoApplication.NodeWorkflowOperationInstanceListResp)

	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountNodeWorkflowOperationInstance(
			rCtx, req.ConvertConditionsToComm())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to count operation instance: %v", err)
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
		resp.ConvertResultFromTypes(num, nil)
		// only count, no data.
		return resp.GetData(), nil
	}

	instances, num, err := h.backendHandler.ListNodeWorkflowOperationInstance(
		rCtx, req.ConvertConditionsToComm())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertResultFromTypes(num, instances)

	return resp.GetData(), nil
}

// GetOperationInstanceLog get operation instance log.
func (h *handler) GetOperationInstanceLog(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationInstanceLogGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get operation instance log, failed to decode request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	logs, err := h.backendHandler.GetNodeWorkflowOperationInstanceLog(
		rCtx,
		req.GetOperInstId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get operation instance log: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.NodeWorkflowOperationInstanceLogGetResp)

	resp.ConvertResultFromTypes(logs)

	return resp.GetData(), nil
}

// RetryOperation retry operation.
func (h *handler) RetryOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationRetryReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to retry operation, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	err := h.backendHandler.RetryNodeOperation(rCtx, req.ConvertNodeOperationRetryParamToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to retry operation")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}
	resp := new(protoApplication.NodeWorkflowOperationRetryResp)

	return resp.GetData(), nil
}

// TerminateOperation terminate operation.
func (h *handler) TerminateOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationTerminateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to terminate operation, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	err := h.backendHandler.TerminateNodeOperation(rCtx, req.ConvertNodeOperationTerminateParamToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to terminate operation")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.NodeWorkflowOperationTerminateResp)

	return resp.GetData(), nil
}

func convertWorkflowToTriggerID(workflows []*types.NodeWorkflow) *types.NodeWorkflowOperInstanceStatusCondition {
	triggerIDs := make([]string, len(workflows))
	for i, workflow := range workflows {
		triggerIDs[i] = workflow.TriggerID
	}

	return &types.NodeWorkflowOperInstanceStatusCondition{
		ExactInclude: &types.NodeWorkflowOperInstanceStatusExactFields{
			TriggerID: triggerIDs,
		},
	}
}

func calculateStats(workflows []*types.NodeWorkflow, statusMap map[string]map[string]*types.NodeWorkflowOperationStatus,
	reqIDs []string) []*protoApplication.NodeWorkflowStatistics {

	idIndexMap := make(map[string]int)
	for i, id := range reqIDs {
		idIndexMap[id] = i
	}

	result := make([]*protoApplication.NodeWorkflowStatistics, len(reqIDs))
	for i, id := range reqIDs {
		result[i] = &protoApplication.NodeWorkflowStatistics{WorkflowID: id}
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
				case types.NodeWorkflowOperationStateInit:
					statusList.InitCount++
				case types.NodeWorkflowOperationStateRunning:
					statusList.RunningCount++
				case types.NodeWorkflowOperationStateLaunched:
					statusList.LaunchedCount++
				case types.NodeWorkflowOperationStateSuccess:
					statusList.SuccessCount++
				case types.NodeWorkflowOperationStateFailed:
					statusList.FailedCount++
				case types.NodeWorkflowOperationStateTimeout:
					statusList.TimeoutCount++
				case types.NodeWorkflowOperationStateTerminated:
					statusList.TerminatedCount++
				}
			}
		}
	}

	return result
}

func (h *handler) listAllBusiness(rCtx restserver.IContext) (map[int64]string, error) {
	businesses, num, err := h.backendHandler.ListBusiness(rCtx, types.UnlimitedPage(), nil)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list business: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	bizNameMap := make(map[int64]string, num)
	for _, biz := range businesses {
		bizNameMap[biz.BizID] = biz.BizName
	}

	return bizNameMap, nil
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
	ops []*types.NodeWorkflowListOperationResult,
	summaries []*types.OperationSummary,
	targetStates []string) (
	[]*types.NodeWorkflowListOperationResult, []*types.OperationSummary) {

	// no state filter required. directly return.
	if len(targetStates) == 0 {
		return ops, summaries
	}

	// record the states that need to be filtered.
	targetStateSet := make(map[types.NodeWorkflowOperationState]struct{}, len(targetStates))
	for _, state := range targetStates {
		targetStateSet[types.NodeWorkflowOperationState(state)] = struct{}{}
	}

	matchedOperations := make([]*types.NodeWorkflowListOperationResult, len(ops))
	matchedSummaries := make([]*types.OperationSummary, len(ops))

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
	instancesByOpID map[string][]*operation.InstanceBriefData) ([]*types.OperationSummary, error) {

	summaries := make([]*types.OperationSummary, len(operationIDs))

	for idx, opID := range operationIDs {
		instances, exists := instancesByOpID[opID]

		// no instances found.
		if !exists || len(instances) == 0 {
			summaries[idx] = &types.OperationSummary{
				TotalDuration: 0,
				LastStatus:    types.NodeWorkflowOperationStateInit,
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
		lastStatus, err := types.InstanceStatusToNodeWorkflowOperationState(lastInstance.Lifecycle.State)
		if err != nil {
			return nil, fmt.Errorf("failed to convert instance status to operation state: %w", err)
		}

		summaries[idx] = &types.OperationSummary{
			TotalDuration: totalSeconds,
			LastStatus:    lastStatus,
		}
	}

	return summaries, nil
}
