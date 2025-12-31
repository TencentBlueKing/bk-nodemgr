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
	"errors"
	"fmt"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/gin-gonic/gin"
)

const (
	maxPluginWorkflowLimit = 500
)

type handler struct {
	rg             *gin.RouterGroup
	pluginMgrIface managerIface.IPluginManager

	daoPluginWorkflow   pluginStg.IDaoPluginWorkflow
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoHost             topoStg.IStorageHost

	storageWorkflow workflow.IStorage
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                  rg.Group("/workflow"),
		pluginMgrIface:      capability.Manager,
		daoPluginWorkflow:   capability.StoragePlugin,
		daoPluginDeployment: capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
		storageWorkflow:     capability.StorageWorkflow,
	}
}

// Load loads workflow handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListPluginWorkflow))
	h.rg.POST("/distinct", restserver.Handler(h.DistinctPluginWorkflow))
	h.rg.POST("/operation/list", restserver.Handler(h.ListOperation))
	h.rg.POST("/operation/retry", restserver.Handler(h.RetryOperation))
	h.rg.POST("/operation/terminate", restserver.Handler(h.TerminateOperation))
	h.rg.POST("/operation/instance/list", restserver.Handler(h.ListOperationInstance))
	h.rg.POST("/operation/instance/status_distribution/list", restserver.Handler(h.ListOperationInstanceStatusDistribution))
	h.rg.POST("/operation/instance/log/get", restserver.Handler(h.GetOperationInstanceLog))
}

// List workflows.
func (h *handler) ListPluginWorkflow(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginWorkflowListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		cond, err := req.ConvertConditionsToTypes()
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to convert conditions")
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}
		num, err := h.daoPluginWorkflow.CountPluginWorkflow(rCtx, cond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to count workflow")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PluginWorkflowListResp)
		resp.ConvertPluginWorkflowsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes(maxPluginWorkflowLimit)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	cond, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflows, total, err := h.daoPluginWorkflow.ListPluginWorkflow(rCtx, page, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PluginWorkflowListResp)

	resp.ConvertPluginWorkflowsFromTypes(total, workflows)

	return resp.GetData(), nil
}

// DistinctPluginWorkflow workflow distinct.
func (h *handler) DistinctPluginWorkflow(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginWorkflowDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	cond, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin workflow, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	result, err := h.daoPluginWorkflow.DistinctPluginWorkflow(
		rCtx,
		types.NewPluginWorkflowDistinctRequestAllSet(),
		cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin workflow. failed to distinct plugin workflow fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PluginWorkflowDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// ListOperation list workflow operation.
// nolint: funlen
func (h *handler) ListOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginWorkflowOperationListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflow, err := h.daoPluginWorkflow.GetPluginWorkflow(rCtx, req.GetWorkflowID())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get plugin workflow")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// list all operations by trigger id.
	operations, _, err := h.storageWorkflow.ListOperation(rCtx, types.UnlimitedPage(), req.ConvertConditionsToOperationTypes(workflow.TriggerID))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	tokens := make([]string, len(operations))
	operationMaps := make(map[string]*struct {
		operation *operation.Operation
		operator  string
	}, len(operations))

	for idx, op := range operations {
		param := new(utils.PluginActionStandardParam)

		if err := conv.MapToStruct(op.Param.InitContent, param); err != nil {
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		tokens[idx] = param.Token
		operationMaps[param.Token] = &struct {
			operation *operation.Operation
			operator  string
		}{
			operation: op,
			operator:  param.Operator,
		}
	}

	page, err := req.ConvertPageToTypes(maxPluginWorkflowLimit)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// list all deployments by condition.
	deployments, num, err := h.daoPluginDeployment.ListPluginDeployment(rCtx, page, req.ConvertConditionsToDeploymentTypes(tokens))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin deployment")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PluginWorkflowOperationListResp)

	// only count.
	if req.GetOnlyCount() {
		resp.ConvertResultFromTypes(num, nil)

		return resp.GetData(), nil
	}

	hostIDList := conv.SliceToSlice[*types.PluginDeployment, int64](deployments, func(dep *types.PluginDeployment) int64 {
		return dep.Info.Process.HostID
	})

	hosts, _, err := h.daoHost.ListHost(rCtx, types.UnlimitedPage(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{HostID: hostIDList},
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if len(hosts) == 0 {
		logger.G.Biz(rCtx).Error("failed to list operation, no host found")
		return nil, resterrf.ErrWrap(resterrf.Aborted, errors.New("no host found"))
	}

	hostIDMap, err := conv.SliceToMap[int64, *types.Host](hosts, func(host *types.Host) int64 {
		return host.HostID
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to convert host slice to map")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	result := make([]*types.PluginWorkflowListOperationResult, len(deployments))
	for idx, deployment := range deployments {
		op, exist := operationMaps[deployment.Token]
		if !exist {
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("invalid token. token(%s)", deployment.Token))
		}

		result[idx] = &types.PluginWorkflowListOperationResult{
			HostID:                deployment.Info.Process.HostID,
			BizID:                 hostIDMap[deployment.Info.Process.HostID].Static.BizID,
			NetworkAreaID:         hostIDMap[deployment.Info.Process.HostID].Static.NetworkAreaID,
			NetworkUnitID:         hostIDMap[deployment.Info.Process.HostID].Dynamic.NetworkUnitID,
			InnerIPList:           hostIDMap[deployment.Info.Process.HostID].Static.InnerIPList,
			InnerIPV6List:         hostIDMap[deployment.Info.Process.HostID].Static.InnerIPV6List,
			PluginName:            deployment.Info.Process.PluginName,
			PluginVersion:         deployment.Info.InstallOptions.Version,
			OperationID:           op.operation.OperationID,
			OperInstanceIDs:       op.operation.InstanceIDs,
			Operator:              op.operator,
			CreateTime:            op.operation.CreateTime,
			LastInstanceBriefData: op.operation.LatestInstBriefData,
		}
	}

	resp.ConvertResultFromTypes(num, result)

	return resp.GetData(), nil
}

// ListOperationInstance list workflow operation instance.
func (h *handler) ListOperationInstance(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginWorkflowOperationInstanceListReq)
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

	resp := new(protoBackend.PluginWorkflowOperationInstanceListResp)

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
	req := new(protoBackend.PluginWorkflowOperationInstanceLogGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get operation instance log, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	instance, err := h.storageWorkflow.GetOperationInstanceFullData(rCtx, req.GetOperInstId())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PluginWorkflowOperationInstanceLogGetResp)

	resp.ConvertResultFromTypes(instance)

	return resp.GetData(), nil
}

// ListOperationInstanceStatusDistribution lists the latest operation instance status distribution by trigger id.
func (h *handler) ListOperationInstanceStatusDistribution(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginWorkflowOperationInstanceStatusDistributionListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance status distribution, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.storageWorkflow.GetLatestOperationInstanceStatusDistributionByTriggerID(rCtx, req.GetTriggerId()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance status distribution")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PluginWorkflowOperationInstanceStatusDistributionListResp)
	resp.ConvertDistributionFromTypes(result)

	return resp.GetData(), nil
}
