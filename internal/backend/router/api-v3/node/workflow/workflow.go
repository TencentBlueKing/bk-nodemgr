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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	nodeworkflow "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-workflow"
	storageOperation "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operation"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operinstdata"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

const (
	maxNodeWorkflowLimit = 500
	maxOperationLimit    = 500
)

type handler struct {
	rg                  *gin.RouterGroup
	manager             manager.Manager
	storageNodeWorkflow nodeworkflow.IStorage
	storageOperation    storageOperation.IStorage
	storageOperInstData operinstdata.IStorage
	logger              logger.Logger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                  rg.Group("/workflow"),
		manager:             capability.Manager,
		storageNodeWorkflow: capability.StorageNodeWorkflow,
		storageOperation:    capability.StorageOperation,
		storageOperInstData: capability.StorageOperInst,
		logger:              capability.Logger,
	}
}

// Load loads workflow handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListNodeWorkflow))
	h.rg.POST("/distinct", restserver.Handler(h.DistinctNodeWorkflow))
	h.rg.POST("/operation/list", restserver.Handler(h.ListOperation))
	h.rg.POST("/operation/instance/list", restserver.Handler(h.ListOperationInstance))
	h.rg.POST("/operation/instance/status/list", restserver.Handler(h.ListOperationInstanceStatus))
	h.rg.POST("/operation/instance/log/get", restserver.Handler(h.GetOperationInstanceLog))
	h.rg.POST("/operation/retry", restserver.Handler(h.WorkflowOperationRetry))
}

// List workflows.
func (h *handler) ListNodeWorkflow(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list node workflow, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storageNodeWorkflow.CountNodeWorkflow(ctx, req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to list node workflow, failed to count workflow. err: %v", err)
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.NodeWorkflowListResp)
		resp.ConvertNodeWorkflowsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	workflows, num, err := h.storageNodeWorkflow.ListNodeWorkflow(ctx,
		req.ConvertPageToTypes(maxNodeWorkflowLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list node workflow. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowListResp)

	resp.ConvertNodeWorkflowsFromTypes(num, workflows)

	return resp.GetData(), nil
}

// DistinctNodeWorkflow workflow distinct.
func (h *handler) DistinctNodeWorkflow(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowDistinctReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to distinct node workflow, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.storageNodeWorkflow.DistinctNodeWorkflow(
		ctx,
		types.NewNodeWorkflowDistinctRequestAllSet(),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to distinct host. failed to distinct host fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// ListOperation list workflow operation.
func (h *handler) ListOperation(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation, failed to validate request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, req.Validate())
	}

	workflow, err := h.storageNodeWorkflow.GetNodeWorkflow(ctx, req.ConvertConditionsToComm())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	result, num, err := h.storageOperation.ListOperationByCondition(ctx, req.ConvertPageToTypes(maxOperationLimit),
		req.ConvertConditionsToTypes(workflow.TriggerID))
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationListResp)

	// only count.
	if req.GetOnlyCount() {
		resp.ConvertResultFromTypes(num, nil)

		return resp.GetData(), nil
	}

	resp.ConvertResultFromTypes(num, result)

	return resp.GetData(), nil
}

// ListOperationInstance list workflow operation instance.
func (h *handler) ListOperationInstance(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationInstanceListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation instance, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation instance, operation ID is required")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, num, err := h.storageOperInstData.ListOperInstanceBriefByOperation(
		ctx, types.UnlimitedPage(), req.GetOperationId()...)
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationInstanceListResp)

	// only count.
	if req.GetOnlyCount() {
		resp.ConvertResultFromTypes(num, nil)

		return resp.GetData(), nil
	}

	resp.ConvertResultFromTypes(num, result)

	return resp.GetData(), nil
}

// GetOperationInstanceLog get operation instance log.
func (h *handler) GetOperationInstanceLog(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationInstanceLogGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get operation instance log, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	instance, err := h.storageOperInstData.GetOperationInstanceFullData(ctx, req.GetOperInstId())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationInstanceLogGetResp)

	resp.ConvertResultFromTypes(instance)

	return resp.GetData(), nil
}

// ListOperationInstanceStatus list workflow operation instance status.
func (h *handler) ListOperationInstanceStatus(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationInstanceListStatusReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list operation instance status, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, _, err := h.storageOperInstData.ListOperationInstanceBriefData(ctx, types.UnlimitedPage(),
		req.ConvertListStatusConditionsToTypes())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationInstanceListStatusResp)
	resp.ConvertWorkflowOperInstanceStatusFromTypes(result)

	return resp.GetData(), nil
}
