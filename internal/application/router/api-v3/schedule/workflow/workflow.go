/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package workflow describes the scheduler workflow router.
package workflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	backendHandler backend.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{rg: rg.Group("/workflow"), backendHandler: capability.BackendHandler}
}

// Load registers scheduler APIs; authorization remains in backend.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)
	h.rg.POST("/interval/update", restserver.Handler(h.UpdateInterval))
	h.rg.POST("/list", restserver.Handler(h.List))
	h.rg.POST("/operation/list", restserver.Handler(h.ListOperation))
	h.rg.POST("/operation/instance/list", restserver.Handler(h.ListOperationInstance))
	h.rg.POST("/operation/instance/log/get", restserver.Handler(h.GetOperationInstanceLog))
}

// UpdateInterval saves a schedule expression through the authorized backend endpoint.
func (h *handler) UpdateInterval(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ScheduleWorkflowUpdateIntervalReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update schedule workflow interval, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.backendHandler.UpdateScheduleWorkflowInterval(rCtx, req.GetWorkflowId(), req.GetInterval()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update schedule workflow interval")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	return &protoApplication.ScheduleWorkflowUpdateIntervalResp_Data{}, nil
}

// List reads schedule configurations without modifying scheduler state.
func (h *handler) List(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ScheduleWorkflowListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule workflows, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	condition, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule workflows, invalid conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	resp := new(protoApplication.ScheduleWorkflowListResp)
	if req.GetOnlyCount() {
		total, countErr := h.backendHandler.CountScheduleWorkflow(rCtx, condition)
		if countErr != nil {
			logger.G.Biz(rCtx).WithErr(countErr).Error("failed to count schedule workflows")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, countErr)
		}
		resp.ConvertScheduleWorkflowsFromTypes(total, nil)

		return resp.GetData(), nil
	}
	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule workflows, invalid page")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	items, total, err := h.backendHandler.ListScheduleWorkflow(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule workflows")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}
	resp.ConvertScheduleWorkflowsFromTypes(total, items)

	return resp.GetData(), nil
}

// ListOperation reads root operations or the direct children of one parent instance.
func (h *handler) ListOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ScheduleWorkflowOperationListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule workflow operations, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	resp := new(protoApplication.ScheduleWorkflowOperationListResp)
	if req.GetOnlyCount() {
		total, err := h.backendHandler.CountScheduleWorkflowOperation(rCtx, req.GetWorkflowId(), req.GetParentOperInstId())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count schedule workflow operations")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
		resp.ConvertResultFromTypes(total, nil)

		return resp.GetData(), nil
	}
	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule workflow operations, invalid page")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	items, total, err := h.backendHandler.ListScheduleWorkflowOperation(rCtx, page, req.GetWorkflowId(), req.GetParentOperInstId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule workflow operations")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}
	resp.ConvertResultFromTypes(total, items)

	return resp.GetData(), nil
}

// ListOperationInstance reads retained execution summaries for operation or instance IDs.
func (h *handler) ListOperationInstance(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ScheduleWorkflowOperationInstanceListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule workflow instances, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	condition := req.ConvertConditionsToTypes()
	resp := new(protoApplication.ScheduleWorkflowOperationInstanceListResp)
	if req.GetOnlyCount() {
		total, err := h.backendHandler.CountScheduleWorkflowOperationInstance(rCtx, condition)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count schedule workflow instances")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
		resp.ConvertResultFromTypes(total, nil)

		return resp.GetData(), nil
	}
	items, total, err := h.backendHandler.ListScheduleWorkflowOperationInstance(rCtx, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule workflow instances")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}
	resp.ConvertResultFromTypes(total, items)

	return resp.GetData(), nil
}

// GetOperationInstanceLog reads bilingual step logs and subworkflow references.
func (h *handler) GetOperationInstanceLog(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.ScheduleWorkflowOperationInstanceLogGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get schedule workflow logs, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	instance, err := h.backendHandler.GetScheduleWorkflowOperationInstanceLog(rCtx, req.GetOperInstId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get schedule workflow instance logs")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}
	resp := new(protoApplication.ScheduleWorkflowOperationInstanceLogGetResp)
	resp.ConvertResultFromTypes(instance)

	return resp.GetData(), nil
}
