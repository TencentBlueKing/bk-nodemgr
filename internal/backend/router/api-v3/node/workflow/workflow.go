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
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
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
	rg              *gin.RouterGroup
	manager         manager.IManager
	daoNodeWorkflow nodeStg.IDaoNodeWorkflow
	storageWorkflow workflow.IStorage
	logger          logger.ILogger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:              rg.Group("/workflow"),
		manager:         capability.Manager,
		daoNodeWorkflow: capability.StorageNode,
		storageWorkflow: capability.StorageWorkflow,
		logger:          capability.Logger,
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
func (h *handler) ListNodeWorkflow(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowListReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to list node workflow, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.daoNodeWorkflow.CountNodeWorkflow(rCtx, req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(rCtx, "failed to list node workflow, failed to count workflow. err: %v", err)
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.NodeWorkflowListResp)
		resp.ConvertNodeWorkflowsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	workflows, num, err := h.daoNodeWorkflow.ListNodeWorkflow(rCtx,
		req.ConvertPageToTypes(maxNodeWorkflowLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to list node workflow. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowListResp)

	resp.ConvertNodeWorkflowsFromTypes(num, workflows)

	return resp.GetData(), nil
}

// DistinctNodeWorkflow workflow distinct.
func (h *handler) DistinctNodeWorkflow(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to distinct node workflow, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.daoNodeWorkflow.DistinctNodeWorkflow(
		rCtx,
		types.NewNodeWorkflowDistinctRequestAllSet(),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to distinct host. failed to distinct host fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// ListOperation list workflow operation.
func (h *handler) ListOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationListReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to list operation, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to list operation, failed to validate request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, req.Validate())
	}

	workflow, err := h.daoNodeWorkflow.GetNodeWorkflow(rCtx, req.ConvertConditionsToComm())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	result, num, err := h.storageWorkflow.ListOperationByNodeWorkflowOperationCondition(
		rCtx, req.ConvertPageToTypes(maxOperationLimit), req.ConvertConditionsToTypes(workflow.TriggerID))
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
func (h *handler) ListOperationInstance(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationInstanceListReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to list operation instance, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to list operation instance, operation ID is required")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, num, err := h.storageWorkflow.ListOperInstanceBriefWithoutActionInstByOperationID(
		rCtx, types.UnlimitedPage(), req.GetOperationId()...)
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
func (h *handler) GetOperationInstanceLog(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationInstanceLogGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to get operation instance log, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	instance, err := h.storageWorkflow.GetOperationInstanceFullData(rCtx, req.GetOperInstId())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationInstanceLogGetResp)

	resp.ConvertResultFromTypes(instance)

	return resp.GetData(), nil
}

// ListOperationInstanceStatus list workflow operation instance status.
func (h *handler) ListOperationInstanceStatus(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationInstanceListStatusReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to list operation instance status, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, _, err := h.storageWorkflow.ListOperationInstanceBriefDataWithoutActionInst(rCtx, types.UnlimitedPage(),
		req.ConvertListStatusConditionsToTypes())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationInstanceListStatusResp)
	resp.ConvertWorkflowOperInstanceStatusFromTypes(result)

	return resp.GetData(), nil
}
