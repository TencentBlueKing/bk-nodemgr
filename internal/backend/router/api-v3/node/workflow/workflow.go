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
	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

const (
	maxNodeWorkflowLimit = 500
)

type handler struct {
	rg           *gin.RouterGroup
	nodeMgrIface managerIface.INodeManager

	daoNodeWorkflow   nodeStg.IDaoNodeWorkflow
	daoNodeDeployment nodeStg.IDaoNodeDeployment

	storageWorkflow workflow.IStorage
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                rg.Group("/workflow"),
		nodeMgrIface:      capability.Manager,
		daoNodeWorkflow:   capability.StorageNode,
		daoNodeDeployment: capability.StorageNode,
		storageWorkflow:   capability.StorageWorkflow,
	}
}

// Load loads workflow handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListNodeWorkflow))
	h.rg.POST("/distinct", restserver.Handler(h.DistinctNodeWorkflow))
	h.rg.POST("/operation/list", restserver.Handler(h.ListOperation))
	h.rg.POST("/operation/retry", restserver.Handler(h.RetryOperation))
	h.rg.POST("/operation/terminate", restserver.Handler(h.TerminateOperation))
	h.rg.POST("/operation/instance/list", restserver.Handler(h.ListOperationInstance))
	h.rg.POST("/operation/instance/status/list", restserver.Handler(h.ListOperationInstanceStatus))
	h.rg.POST("/operation/instance/log/get", restserver.Handler(h.GetOperationInstanceLog))
}

// List workflows.
func (h *handler) ListNodeWorkflow(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.daoNodeWorkflow.CountNodeWorkflow(rCtx, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list node workflow, failed to count workflow")
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
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node workflow")
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
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct node workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.daoNodeWorkflow.DistinctNodeWorkflow(
		rCtx,
		types.NewNodeWorkflowDistinctRequestAllSet(),
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct host. failed to distinct host fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// ListOperation list workflow operation.
// nolint: funlen
func (h *handler) ListOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflow, err := h.daoNodeWorkflow.GetNodeWorkflow(rCtx, req.GetWorkflowID())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get node workflow")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// list all operations by trigger id.
	operations, _, err := h.storageWorkflow.ListOperationByTriggerID(rCtx, types.UnlimitedPage(), workflow.TriggerID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	tokens := make([]string, len(operations))
	operationMaps := make(map[string]*struct {
		operationID     string
		operator        string
		operInstanceIDs []string
	}, len(operations))

	for idx, op := range operations {
		param := new(utils.NodeActionStandardParam)

		if err := conv.MapToStruct(op.Param.InitContent, param); err != nil {
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		tokens[idx] = param.Token
		operationMaps[param.Token] = &struct {
			operationID     string
			operator        string
			operInstanceIDs []string
		}{
			operationID:     op.OperationID,
			operator:        param.Operator,
			operInstanceIDs: op.InstanceIDs,
		}
	}

	// list all deployments by condition.
	deployments, num, err := h.daoNodeDeployment.ListNodeDeployment(rCtx, types.UnlimitedPage(), req.ConvertConditionsToDeploymentTypes(tokens))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node deployment")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	result := make([]*types.NodeWorkflowListOperationResult, len(deployments))
	for idx, dep := range deployments {
		op, exists := operationMaps[dep.Token]
		if !exists {
			continue
		}

		result[idx] = &types.NodeWorkflowListOperationResult{
			Operator:        op.operator,
			NetworkAreaID:   dep.Info.Host.Static.NetworkAreaID,
			NetworkUnitID:   dep.Info.Host.Dynamic.NetworkUnitID,
			InnerIPList:     dep.Info.Host.Static.InnerIPList,
			InnerIPV6List:   dep.Info.Host.Static.InnerIPV6List,
			BizID:           dep.Info.Host.Static.BizID,
			OperationID:     op.operationID,
			OperInstanceIDs: op.operInstanceIDs,
			NodeVersion:     dep.Info.Host.Dynamic.NodeVersion,
		}
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
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, operation ID is required")
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
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get operation instance log, failed to decode request body")
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
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance status, failed to decode request body")
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
