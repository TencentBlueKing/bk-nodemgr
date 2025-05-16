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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	nodeworkflow "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-workflow"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/gin-gonic/gin"
)

const (
	maxNodeWorkflowLimit = 500
)

type handler struct {
	rg                  *gin.RouterGroup
	nodeWorkflowStorage nodeworkflow.IStorage
	logger              logger.Logger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                  rg.Group("/workflow"),
		nodeWorkflowStorage: capability.NodeWorkflowStorage,
		logger:              capability.Logger,
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

	req := new(protoBackend.NodeWorkflowListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list workflow, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {

	}

	workflows, num, err := h.nodeWorkflowStorage.ListWorkflow(sCtx,
		req.ConvertPageToTypes(maxNodeWorkflowLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list node workflow. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowListResp)
	resp.ConvertNodeWorkflowsFromTypes(num, workflows)

	return resp.GetData(), nil
}

// Statistics workflow statistics.
func (h *handler) Statistics(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to statistics workflow, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.NodeWorkflowStatisticsReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to statistics workflow, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// TODO: call thirdparty backend.

	resp := new(protoBackend.NodeWorkflowStatisticsResp)
	// TODO: convert data from types.

	return resp.GetData(), nil
}

// Distinct workflow distinct.
func (h *handler) Distinct(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to distinct workflow, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.NodeWorkflowDistinctReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to distinct workflow, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// TODO: call thirdparty backend.

	resp := new(protoBackend.NodeWorkflowDistinctResp)
	// TODO: convert data from types.

	return resp.GetData(), nil
}

// ListOperation list workflow operation.
func (h *handler) ListOperation(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list operation, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.NodeWorkflowOperationListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// TODO: call thirdparty backend.

	resp := new(protoBackend.NodeWorkflowOperationListResp)
	// TODO: convert data from types.

	return resp.GetData(), nil
}

// ListOperationInstance list workflow operation instance.
func (h *handler) ListOperationInstance(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list operation instance, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.NodeWorkflowOperationInstanceListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation instance, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// TODO: call thirdparty backend.

	resp := new(protoBackend.NodeWorkflowOperationInstanceListResp)
	// TODO: convert data from types.

	return resp.GetData(), nil
}

// GetOperationInstanceLog get operation instance log.
func (h *handler) GetOperationInstanceLog(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to get operation instance log, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.NodeWorkflowOperationInstanceLogGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to get operation instance log, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// TODO: call thirdparty backend.

	resp := new(protoBackend.NodeWorkflowOperationInstanceLogGetResp)
	// TODO: convert data from types.

	return resp.GetData(), nil
}
