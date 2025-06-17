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
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
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
}

// List workflows.
func (h *handler) List(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list workflow, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.NodeWorkflowListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list workflow, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := new(protoApplication.NodeWorkflowListResp)
	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountNodeWorkflow(
			sCtx,
			req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(sCtx, "failed to list workflow, failed to count workflow. err: %v", err)
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}
		resp.ConvertNodeWorkflowsFromTypes(num, nil, nil)
		// only count, no data.
		return resp.GetData(), nil
	}

	workflows, num, err := h.backendHandler.ListNodeWorkflow(sCtx,
		req.ConvertPageToTypes(maxWorkflowLimit), req.ConvertConditionsToTypes())

	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list workflow, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	businessMap, err := h.listAllBusiness(ctx)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list business, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertNodeWorkflowsFromTypes(num, workflows, businessMap)

	return resp.GetData(), nil
}

// TODO：如果数据量过大可能要分页或者拆协程查询
// Statistics workflow statistics.
func (h *handler) Statistics(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to statistics workflow, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.NodeWorkflowStatisticsReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to statistics workflow, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	workflows, _, err := h.backendHandler.ListNodeWorkflow(
		sCtx, types.UnlimitedPage(), req.ConvertConditionsToWorkflowConditionTypes())
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to statistics workflow, failed to list workflow. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	instanceStatus, err := h.backendHandler.ListNodeWorkflowOperationInstanceStatus(
		sCtx, convertWorkflowToTriggerID(workflows))
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list workflow instance status, err: %v", err)
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
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to distinct workflow, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.NodeWorkflowDistinctReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to distinct workflow, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	result, err := h.backendHandler.DistinctNodeWorkflow(
		sCtx,
		types.NewNodeWorkflowDistinctRequestAllSet(),
		req.ConvertConditionsToTypes())
	resp := new(protoApplication.NodeWorkflowDistinctResp)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to distinct workflow, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// ListOperation list workflow operation.
func (h *handler) ListOperation(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list operation, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.NodeWorkflowOperationListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if req.Validate() != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, req.Validate())
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountNodeWorkflowOperation(
			sCtx, req.ConvertConditionsToComm())
		if err != nil {
			h.logger.ErrorCtxf(sCtx, "failed to list operation, failed to count operation. err: %v", err)
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}
		resp := new(protoApplication.NodeWorkflowOperationListResp)

		resp.ConvertResultFromTypes(num, nil)
		// only count, no data.
		return resp.GetData(), nil
	}

	result, num, err := h.backendHandler.ListNodeWorkflowOperation(
		sCtx, req.ConvertPageToTypes(maxOperationLimit), req.ConvertConditionsToComm())
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}
	resp := new(protoApplication.NodeWorkflowOperationListResp)

	resp.ConvertResultFromTypes(num, result)

	return resp.GetData(), nil
}

// ListOperationInstance list workflow operation instance.
func (h *handler) ListOperationInstance(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list operation instance, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.NodeWorkflowOperationInstanceListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation instance, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if req.Validate() != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation instance, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, req.Validate())
	}

	resp := new(protoApplication.NodeWorkflowOperationInstanceListResp)

	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountNodeWorkflowOperationInstance(
			sCtx, req.ConvertConditionsToComm())
		if err != nil {
			h.logger.ErrorCtxf(sCtx, "failed to list operation instance, failed to count operation instance. err: %v", err)
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}
		resp.ConvertResultFromTypes(num, nil)
		// only count, no data.
		return resp.GetData(), nil
	}

	instances, num, err := h.backendHandler.ListNodeWorkflowOperationInstance(
		sCtx, req.ConvertConditionsToComm())
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation instance, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp.ConvertResultFromTypes(num, instances)

	return resp.GetData(), nil
}

// GetOperationInstanceLog get operation instance log.
func (h *handler) GetOperationInstanceLog(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to get operation instance log, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.NodeWorkflowOperationInstanceLogGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to get operation instance log, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	logs, err := h.backendHandler.GetNodeWorkflowOperationInstanceLog(
		sCtx,
		req.GetOperInstId())
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to get operation instance log, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.NodeWorkflowOperationInstanceLogGetResp)

	resp.ConvertResultFromTypes(logs)

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
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list business, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	businesses, num, err := h.backendHandler.ListBusiness(sCtx, types.UnlimitedPage(), nil)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list business, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	bizNameMap := make(map[int64]string, num)
	for _, biz := range businesses {
		bizNameMap[biz.BizID] = biz.BizName
	}

	return bizNameMap, nil
}
