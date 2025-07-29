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
	"sort"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
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
	backendHandler backend.Handler
	logger         logger.Logger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/workflow"),
		backendHandler: capability.BackendHandler,
		logger:         capability.Logger,
	}
}

// Load loads workflow handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", rest.RestHandlerFunc(h.List))
	h.rg.POST("/statistics", rest.RestHandlerFunc(h.Statistics))
	h.rg.POST("/distinct", rest.RestHandlerFunc(h.Distinct))
	h.rg.POST("/operation/list", rest.RestHandlerFunc(h.ListOperation))
	h.rg.POST("/operation/instance/list", rest.RestHandlerFunc(h.ListOperationInstance))
	h.rg.POST("/operation/instance/log/get", rest.RestHandlerFunc(h.GetOperationInstanceLog))
	h.rg.POST("/operation/retry", rest.RestHandlerFunc(h.OperationRetry))

}

// List workflows.
func (h *handler) List(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list workflow, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := new(protoApplication.NodeWorkflowListResp)
	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountNodeWorkflow(
			ctx,
			req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to list workflow, failed to count workflow. err: %v", err)
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}
		resp.ConvertNodeWorkflowsFromTypes(num, nil, nil)
		// only count, no data.
		return resp.GetData(), nil
	}

	workflows, num, err := h.backendHandler.ListNodeWorkflow(ctx,
		req.ConvertPageToTypes(maxWorkflowLimit), req.ConvertConditionsToTypes())

	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list workflow, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	businessMap, err := h.listAllBusiness(ctx)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list business, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertNodeWorkflowsFromTypes(num, workflows, businessMap)

	return resp.GetData(), nil
}

// TODO：如果数据量过大可能要分页或者拆协程查询
// Statistics workflow statistics.
func (h *handler) Statistics(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowStatisticsReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to statistics workflow, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	workflows, _, err := h.backendHandler.ListNodeWorkflow(
		ctx, types.UnlimitedPage(), req.ConvertConditionsToWorkflowConditionTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to statistics workflow, failed to list workflow. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	instanceStatus, err := h.backendHandler.ListNodeWorkflowOperationInstanceStatus(
		ctx, convertWorkflowToTriggerID(workflows))
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list workflow instance status, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
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
				State:       types.OperationState(instatus.State),
			}
		}
	}

	result := calculateStats(workflows, workflowStatusMap, req.GetWorkflowId())

	resp := new(protoApplication.NodeWorkflowStatisticsResp)
	resp.ConvertNodeWorkflowsFromTypes(result)

	return resp.GetData(), nil
}

// Distinct workflow distinct.
func (h *handler) Distinct(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowDistinctReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to distinct workflow, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	result, err := h.backendHandler.DistinctNodeWorkflow(
		ctx,
		types.NewNodeWorkflowDistinctRequestAllSet(),
		req.ConvertConditionsToTypes())
	resp := new(protoApplication.NodeWorkflowDistinctResp)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to distinct workflow, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// ListOperation list workflow operation.
func (h *handler) ListOperation(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, req.Validate())
	}

	var targetStates []string
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		targetStates = exactCond.GetState()
	}

	result, total, err := h.backendHandler.ListNodeWorkflowOperation(
		ctx, req.ConvertPageToTypes(maxOperationLimit), req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	if len(result) == 0 {
		resp := new(protoApplication.NodeWorkflowOperationListResp)
		resp.ConvertResultFromTypes(total, nil, nil)

		if req.GetOnlyCount() {
			return resp.GetCountOnly(), nil
		}

		return resp.GetData(), nil
	}

	operationIDs := make([]string, 0, len(result))
	for _, operation := range result {
		operationIDs = append(operationIDs, operation.OperationID)
	}

	allInstances, _, err := h.backendHandler.ListNodeWorkflowOperationInstance(ctx, operationIDs...)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation instance, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	instancesByOpID := groupInstancesByOperationID(allInstances)
	summaries := calculateOperationSummaries(operationIDs, instancesByOpID)

	filteredResults, filteredSummaries := filterOperationsByStates(
		result,
		summaries,
		targetStates,
	)

	resp := new(protoApplication.NodeWorkflowOperationListResp)
	resp.ConvertResultFromTypes(int64(len(filteredResults)), filteredResults, filteredSummaries)

	if req.GetOnlyCount() {
		return resp.GetCountOnly(), nil
	}

	return resp.GetData(), nil
}

// ListOperationInstance list workflow operation instance.
func (h *handler) ListOperationInstance(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationInstanceListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation instance, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation instance, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, req.Validate())
	}

	resp := new(protoApplication.NodeWorkflowOperationInstanceListResp)

	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountNodeWorkflowOperationInstance(
			ctx, req.ConvertConditionsToComm())
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to list operation instance, failed to count operation instance. err: %v", err)
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}
		resp.ConvertResultFromTypes(num, nil)
		// only count, no data.
		return resp.GetData(), nil
	}

	instances, num, err := h.backendHandler.ListNodeWorkflowOperationInstance(
		ctx, req.ConvertConditionsToComm())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation instance, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertResultFromTypes(num, instances)

	return resp.GetData(), nil
}

// GetOperationInstanceLog get operation instance log.
func (h *handler) GetOperationInstanceLog(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationInstanceLogGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get operation instance log, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	logs, err := h.backendHandler.GetNodeWorkflowOperationInstanceLog(
		ctx,
		req.GetOperInstId())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get operation instance log, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.NodeWorkflowOperationInstanceLogGetResp)

	resp.ConvertResultFromTypes(logs)

	return resp.GetData(), nil
}

func (h *handler) OperationRetry(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationRetryReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to retry operation, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to retry operation, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	InstanceIDs, err := h.backendHandler.OperationRetry(ctx, req.ConvertRetryParamToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to retry operation: %v", err)
		return nil, err
	}
	resp := new(protoApplication.NodeWorkflowOperationRetryResp)

	resp.ConvertOperInstanceID(InstanceIDs)

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
				case types.StateInit:
					statusList.InitCount++
				case types.StateRunning:
					statusList.RunningCount++
				case types.StateLaunched:
					statusList.LaunchedCount++
				case types.StateSuccess:
					statusList.SuccessCount++
				case types.StateFailed:
					statusList.FailedCount++
				case types.StateTimeout:
					statusList.TimeoutCount++
				case types.StateTerminated:
					statusList.TerminatedCount++
				}
			}
		}
	}

	return result
}

func (h *handler) listAllBusiness(ctx *rest.Context) (map[int64]string, error) {
	businesses, num, err := h.backendHandler.ListBusiness(ctx, types.UnlimitedPage(), nil)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list business, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
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

		if _, exists := grouped[opID]; !exists {
			grouped[opID] = make([]*operation.InstanceBriefData, 0)
		}

		grouped[opID] = append(grouped[opID], instance)
	}

	return grouped
}

func filterOperationsByStates(
	ops []*operation.Operation,
	summaries []*types.OperationSummary,
	targetStates []string,
) ([]*operation.Operation, []*types.OperationSummary) {

	if len(targetStates) == 0 {
		return ops, summaries
	}
	stateLookup := make(map[string]bool)
	for _, state := range targetStates {
		if state != "" {
			stateLookup[strings.ToLower(state)] = true
		}
	}

	filteredOps := make([]*operation.Operation, 0, len(ops))
	filteredSummaries := make([]*types.OperationSummary, 0, len(summaries))

	for i, op := range ops {
		summary := summaries[i]
		if _, exists := stateLookup[summary.LastStatus]; exists {
			filteredOps = append(filteredOps, op)
			filteredSummaries = append(filteredSummaries, summary)
		}
	}

	return filteredOps, filteredSummaries
}

func calculateOperationSummaries(operationIDs []string,
	instancesByOpID map[string][]*operation.InstanceBriefData) []*types.OperationSummary {

	summaries := make([]*types.OperationSummary, len(operationIDs))

	for idx, opID := range operationIDs {
		instances, exists := instancesByOpID[opID]

		if !exists || len(instances) == 0 {
			summaries[idx] = &types.OperationSummary{
				TotalDuration: 0,
				LastStatus:    "empty_instances",
			}

			continue
		}

		sort.Slice(instances, func(i, j int) bool {
			return instances[i].Lifecycle.CreatedAt.Before(instances[j].Lifecycle.CreatedAt)
		})

		var totalSeconds int64
		for _, inst := range instances {
			if !inst.Lifecycle.CreatedAt.IsZero() && !inst.Lifecycle.EndedAt.IsZero() {
				durationSec := inst.Lifecycle.EndedAt.Unix() - inst.Lifecycle.CreatedAt.Unix()

				totalSeconds += durationSec
			}
		}
		lastInstance := instances[len(instances)-1]
		lastStatus := string(lastInstance.Lifecycle.State)

		summaries[idx] = &types.OperationSummary{
			TotalDuration: totalSeconds,
			LastStatus:    lastStatus,
		}
	}

	return summaries
}
