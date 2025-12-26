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
	h.rg.POST("/operation/manual/solution/get", restserver.Handler(h.GetManualSolution))
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

	// if no workflow, return empty result.
	if len(workflows) == 0 {
		resp := new(protoApplication.NodeWorkflowStatisticsResp)
		if err := resp.ConvertNodeWorkflowsFromDistribution(req.GetWorkflowId(), nil, nil); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics workflow, failed to convert distribution")

			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		return resp.GetData(), nil
	}

	triggerIDs := make([]string, len(workflows))
	triggerToWorkflowMap := make(map[string]string, len(workflows))
	for idx, wf := range workflows {
		triggerIDs[idx] = wf.TriggerID
		triggerToWorkflowMap[wf.TriggerID] = wf.WorkflowID
	}

	// get latest operation instance status distribution by trigger id.
	distribution, err := h.backendHandler.ListNodeWorkflowOperationInstanceStatusDistribution(
		rCtx, triggerIDs)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics workflow, failed to get latest operation instance status distribution")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	// convert distribution to response.
	resp := new(protoApplication.NodeWorkflowStatisticsResp)
	if err := resp.ConvertNodeWorkflowsFromDistribution(req.GetWorkflowId(), distribution, triggerToWorkflowMap); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics workflow,failed to connvert distribution")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

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

	if req.GetOnlyCount() {
		cnt, err := h.backendHandler.CountNodeWorkflowOperation(rCtx, req.GetWorkflowID(), req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation, failed to count operation.")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.NodeWorkflowOperationListResp)
		resp.ConvertResultFromTypes(cnt, nil)

		return resp.GetData(), nil
	}

	operation, cnt, err := h.backendHandler.ListNodeWorkflowOperation(rCtx, req.ConvertPageToTypes(maxOperationLimit), req.WorkflowId, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation, failed to list operation")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.NodeWorkflowOperationListResp)
	resp.ConvertResultFromTypes(cnt, operation)

	return resp.GetData(), nil
}

// ListOperationInstance list workflow operation instance.
func (h *handler) ListOperationInstance(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationInstanceListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to decode request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountNodeWorkflowOperationInstance(
			rCtx, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to count operation instance: %v", err)
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.NodeWorkflowOperationInstanceListResp)
		resp.ConvertResultFromTypes(num, nil)
		// only count, no data.
		return resp.GetData(), nil
	}

	instances, num, err := h.backendHandler.ListNodeWorkflowOperationInstance(
		rCtx, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}
	resp := new(protoApplication.NodeWorkflowOperationInstanceListResp)

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

	err := h.backendHandler.RetryNodeWorkflowOperation(rCtx, req.ConvertNodeWorkflowOperationRetryParamToTypes())
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

	err := h.backendHandler.TerminateNodeWorkflowOperation(rCtx, req.ConvertNodeWorkflowOperationTerminateParamToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to terminate operation")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.NodeWorkflowOperationTerminateResp)

	return resp.GetData(), nil
}

// GetManualSolution get manual solution.
func (h *handler) GetManualSolution(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeWorkflowOperationManualSolutionGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get manual solution, failed to decode request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	manualInfo, err := h.backendHandler.GetNodeWorkflowOperationManualInfo(
		rCtx,
		req.GetWorkflowId(),
		req.GetOperationId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get manual info: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.NodeWorkflowOperationManualSolutionGetResp)

	// convert ManualInfo to ManualSolution
	resp.ConvertResultFromTypes(manualInfo)

	return resp.GetData(), nil
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
