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
	storageOperation "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operation"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operinstdata"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/gin-gonic/gin"
)

const (
	maxNodeWorkflowLimit = 500
	maxOperationLimit    = 500
)

type handler struct {
	rg                  *gin.RouterGroup
	storageNodeWorkflow nodeworkflow.IStorage
	storageOperation    storageOperation.IStorage
	storageOperInstData operinstdata.IStorage
	logger              logger.Logger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                  rg.Group("/workflow"),
		storageNodeWorkflow: capability.StorageNodeWorkflow,
		storageOperation:    capability.StorageOperation,
		storageOperInstData: capability.StorageOperInst,
		logger:              capability.Logger,
	}
}

// Load loads workflow handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", rest.RestHandlerFunc(h.ListNodeWorkflow))
	h.rg.POST("/distinct", rest.RestHandlerFunc(h.DistinctNodeWorkflow))
	h.rg.POST("/operation/list", rest.RestHandlerFunc(h.ListOperation))
	h.rg.POST("/operation/instance/list", rest.RestHandlerFunc(h.ListOperationInstance))
	h.rg.POST("/operation/instance/status/list/", rest.RestHandlerFunc(h.ListOperationInstanceStatus))
	h.rg.POST("/operation/instance/log/get", rest.RestHandlerFunc(h.GetOperationInstanceLog))
}

// List workflows.
func (h *handler) ListNodeWorkflow(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list node workflow, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.NodeWorkflowListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list node workflow, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storageNodeWorkflow.CountNodeWorkflow(sCtx, req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(sCtx, "failed to list node workflow, failed to count workflow. err: %v", err)
			return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.NodeWorkflowListResp)
		resp.ConvertNodeWorkflowsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	workflows, num, err := h.storageNodeWorkflow.ListNodeWorkflow(sCtx,
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

// DistinctNodeWorkflow workflow distinct.
func (h *handler) DistinctNodeWorkflow(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to distinct node workflow, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.NodeWorkflowDistinctReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to distinct node workflow, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	result, err := h.storageNodeWorkflow.DistinctNodeWorkflow(
		sCtx,
		types.NewNodeWorkflowDistinctRequestAllSet(),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to distinct host. failed to distinct host fields: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowDistinctResp)
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

	req := new(protoBackend.NodeWorkflowOperationListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if req.Validate() != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, req.Validate())
	}

	workflow, err := h.storageNodeWorkflow.GetNodeWorkflow(sCtx, req.ConvertConditionsToComm())
	if err != nil {
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	result, num, err := h.storageOperation.ListOperation(sCtx,
		req.ConvertPageToTypes(maxOperationLimit), workflow.TriggerID)
	if err != nil {
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
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

	if req.Validate() != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation instance, operation ID is required")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	opera, err := h.storageOperation.GetOperation(sCtx, req.ConvertConditionsToComm())
	if err != nil {
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	result, num, err := h.storageOperInstData.ListOperationInstanceBriefData(sCtx, types.UnlimitedPage(),
		operation.ListOperationInstanceCondition{TriggerIDs: []string{opera.TriggerID}})

	if err != nil {
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
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

	instance, err := h.storageOperInstData.GetOperationInstanceFullData(sCtx, req.GetOperInstId())
	if err != nil {
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationInstanceLogGetResp)

	resp.ConvertResultFromTypes(instance)

	return resp.GetData(), nil
}

// ListOperationInstanceStatus list workflow operation instance status.
func (h *handler) ListOperationInstanceStatus(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list operation instance status, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.NodeWorkflowOperationInstanceListStatusReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list operation instance status, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	result, _, err := h.storageOperInstData.ListOperationInstanceBriefData(sCtx, types.UnlimitedPage(),
		req.ConvertListStatusConditionsToTypes())
	if err != nil {
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationInstanceListStatusResp)

	resp.ConvertWorkflowOperInstanceStatusFromTypes(result)

	return resp.GetData(), nil
}
