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

// Package operation ...
package operation

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// ListOperation list workflow operation.
// nolint: funlen
func (h *handler) ListOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginWorkflowOperationListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID := req.GetWorkflowID()
	if err := h.authorizedPluginHistoryView(rCtx, workflowID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("workflow-id", workflowID).
			Error("failed to list operation, permission denied")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	workflow, err := h.daoPluginWorkflow.GetPluginWorkflow(rCtx, workflowID)
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

	// if no operations match the filter, return empty result early.
	if len(operations) == 0 {
		resp := new(protoBackend.PluginWorkflowOperationListResp)
		resp.ConvertResultFromTypes(0, nil)

		return resp.GetData(), nil
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

	page, err := req.ConvertPageToTypes()
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

// DistinctOperation defines the plugin workflow operation distinct handler.
func (h *handler) DistinctOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginWorkflowOperationDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct operation, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// get workflow to get trigger id.
	workflow, err := h.daoPluginWorkflow.GetPluginWorkflow(rCtx, req.GetWorkflowId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct operation, failed to get plugin workflow")
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
		operationCond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct operation, failed to distinct operation fields")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PluginWorkflowOperationDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// ListOperationInstance list workflow operation instance.
func (h *handler) ListOperationInstance(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginWorkflowOperationInstanceListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// Get operation to extract trigger ID and check permission
	operationIDs := req.GetOperationId()
	if len(operationIDs) == 0 {
		logger.G.Biz(rCtx).Error("failed to list operation instance, no operation ID provided")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("operation_id is required"))
	}

	operation, err := h.storageWorkflow.GetOperation(rCtx, operationIDs[0])
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to get operation")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// Find workflow by trigger ID to check permission
	// Note: We query by TriggerID since PluginWorkflowExactFields doesn't support TriggerID field
	// This is a workaround - ideally we should add TriggerID to PluginWorkflowExactFields
	workflows, _, err := h.daoPluginWorkflow.ListPluginWorkflow(rCtx, types.UnlimitedPage())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list operation instance, failed to list workflows")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// Find workflow with matching TriggerID
	var targetWorkflow *types.PluginWorkflow
	for _, wf := range workflows {
		if wf.TriggerID == operation.TriggerID {
			targetWorkflow = wf
			break
		}
	}
	if targetWorkflow == nil {
		logger.G.Biz(rCtx).Error("failed to list operation instance, workflow not found for trigger ID")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, fmt.Errorf("workflow not found for trigger_id: %s", operation.TriggerID))
	}

	// Check permission
	if err := h.authorizedPluginHistoryView(rCtx, targetWorkflow.WorkflowID); err != nil {
		return nil, err
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

	// Get operation to extract trigger ID
	operation, err := h.storageWorkflow.GetOperation(rCtx, instance.Metadata.OperationID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get operation instance log, failed to get operation")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// Find workflow by trigger ID to check permission
	// Note: We query all workflows since PluginWorkflowExactFields doesn't support TriggerID field
	workflows, _, err := h.daoPluginWorkflow.ListPluginWorkflow(rCtx, types.UnlimitedPage())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get operation instance log, failed to list workflows")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// Find workflow with matching TriggerID
	var targetWorkflow *types.PluginWorkflow
	for _, wf := range workflows {
		if wf.TriggerID == operation.TriggerID {
			targetWorkflow = wf
			break
		}
	}
	if targetWorkflow == nil {
		logger.G.Biz(rCtx).Error("failed to get operation instance log, workflow not found for trigger ID")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, fmt.Errorf("workflow not found for trigger_id: %s", operation.TriggerID))
	}

	// Check permission
	if err := h.authorizedPluginHistoryView(rCtx, targetWorkflow.WorkflowID); err != nil {
		return nil, err
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
