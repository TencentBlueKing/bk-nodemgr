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
	"github.com/gin-gonic/gin"
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
		cond, err := req.ConvertConditionsToTypes()
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to convert conditions")
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}
		num, err := h.backendHandler.CountPluginWorkflow(
			rCtx,
			cond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to count plugin workflow")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
		resp.ConvertPluginWorkflowsFromTypes(num, nil)
		// only count, no data.
		return resp.GetData(), nil
	}

	cond, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflows, num, err := h.backendHandler.ListPluginWorkflow(rCtx, page, cond)
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

	cond, err := req.ConvertConditionsToWorkflowConditionTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics plugin workflow, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	workflows, _, err := h.backendHandler.ListPluginWorkflow(rCtx, types.UnlimitedPage(), cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics plugin workflow, failed to list plugin workflow")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	if len(workflows) == 0 {
		resp := new(protoApplication.PluginWorkflowStatisticsResp)
		if err := resp.ConvertPluginWorkflowsFromDistribution(req.GetWorkflowId(), nil, nil); err != nil {
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
	distribution, err := h.backendHandler.ListPluginWorkflowOperationInstanceStatusDistribution(
		rCtx, triggerIDs)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics workflow, failed to get latest operation instance status distribution")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	// convert distribution to response.
	resp := new(protoApplication.PluginWorkflowStatisticsResp)
	if err := resp.ConvertPluginWorkflowsFromDistribution(req.GetWorkflowId(), distribution, triggerToWorkflowMap); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics workflow, failed to convert distribution")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	return resp.GetData(), nil
}

// Distinct plugin workflow distinct.
func (h *handler) Distinct(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PluginWorkflowDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	cond, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin workflow, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	result, err := h.backendHandler.DistinctPluginWorkflow(
		rCtx,
		types.NewPluginWorkflowDistinctRequestAllSet(),
		cond)
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

	if req.GetOnlyCount() {
		cnt, err := h.backendHandler.CountPluginWorkflowOperation(rCtx, req.GetWorkflowId(), req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation, failed to count operation: %v", err)
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.PluginWorkflowOperationListResp)
		resp.ConvertResultFromTypes(cnt, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation, invalid page info: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	operation, cnt, err := h.backendHandler.ListPluginWorkflowOperation(rCtx, page, req.GetWorkflowId(), req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PluginWorkflowOperationListResp)
	resp.ConvertResultFromTypes(cnt, operation)

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
