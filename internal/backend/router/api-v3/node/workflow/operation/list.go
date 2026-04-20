/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package operation

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// ListOperation list workflow operation.
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

	operations, _, err := h.storageWorkflow.ListOperation(rCtx, types.UnlimitedPage(),
		req.ConvertConditionsToOperationTypes(workflow.TriggerID))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// if no operations match the filter, return empty result early.
	if len(operations) == 0 {
		resp := new(protoBackend.NodeWorkflowOperationListResp)
		resp.ConvertResultFromTypes(0, nil)

		return resp.GetData(), nil
	}

	tokens := make([]string, len(operations))
	operationMap := make(map[string]*struct {
		operation *operation.Operation
		operator  string
	}, len(operations))

	for idx, op := range operations {
		param := new(utils.NodeActionStandardParam)
		if err := conv.MapToStruct(op.Param.InitContent, param); err != nil {
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		tokens[idx] = param.Token
		operationMap[param.Token] = &struct {
			operation *operation.Operation
			operator  string
		}{
			operation: op,
			operator:  param.Operator,
		}
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node deployment, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	deployments, num, err := h.daoNodeDeployment.ListNodeDeployment(rCtx, page, req.ConvertConditionsToDeploymentTypes(tokens))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node deployment")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if req.GetOnlyCount() {
		resp := new(protoBackend.NodeWorkflowOperationListResp)
		resp.ConvertResultFromTypes(num, nil)

		return resp.GetData(), nil
	}

	result := make([]*types.NodeWorkflowListOperationResult, len(deployments))
	for idx, deployment := range deployments {
		op, exist := operationMap[deployment.Token]
		if !exist {
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("invalid token. token(%s)", deployment.Token))
		}

		result[idx] = &types.NodeWorkflowListOperationResult{
			OperationID:           op.operation.OperationID,
			Operator:              op.operator,
			OperInstanceIDs:       op.operation.InstanceIDs,
			HostID:                deployment.Info.Host.HostID,
			BizID:                 deployment.Info.Host.Static.BizID,
			InnerIPList:           deployment.Info.Host.Static.InnerIPList,
			InnerIPV6List:         deployment.Info.Host.Static.InnerIPV6List,
			NetworkAreaID:         deployment.Info.Host.Static.NetworkAreaID,
			NetworkUnitID:         deployment.Info.Host.Dynamic.NetworkUnitID,
			NodeVersion:           deployment.Info.Host.Dynamic.NodeVersion,
			CreateTime:            op.operation.CreateTime,
			LastInstanceBriefData: op.operation.LatestInstBriefData,
		}
	}

	resp := new(protoBackend.NodeWorkflowOperationListResp)
	resp.ConvertResultFromTypes(num, result)

	return resp.GetData(), nil
}

// DistinctOperation defines the node workflow operation distinct handler.
func (h *handler) DistinctOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct operation, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// get workflow to get trigger id.
	workflow, err := h.daoNodeWorkflow.GetNodeWorkflow(rCtx, req.GetWorkflowID())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct operation, failed to get node workflow")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	operationCond := &types.OperationCondition{
		ExactInclude: &types.OperationExactFields{
			TriggerID: []string{workflow.TriggerID},
		},
	}

	result, err := h.storageWorkflow.DistinctOperation(
		rCtx,
		req.ConvertSelectorToTypes(),
		operationCond,
	)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct operation, failed to distinct operation fields")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// ListOperationInstance list workflow operation instance.
func (h *handler) ListOperationInstance(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationInstanceListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, num, err := h.storageWorkflow.ListOperationInstanceBriefDataWithoutActionInst(
		rCtx, types.UnlimitedPage(), &types.OperInstDataCondition{ExactInclude: &types.OperInstDataExactFields{
			OperationID: req.GetOperationId(),
			OperInstID:  req.GetOperInstId(),
		}})
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

// ListOperationInstanceStatusDistribution lists the latest operation instance status distribution by trigger id.
func (h *handler) ListOperationInstanceStatusDistribution(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationInstanceStatusDistributionListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance status distribution, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.storageWorkflow.GetLatestOperationInstanceStatusDistributionByTriggerID(rCtx, req.GetTriggerId()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance status distribution")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationInstanceStatusDistributionListResp)
	resp.ConvertDistributionFromTypes(result)

	return resp.GetData(), nil
}
