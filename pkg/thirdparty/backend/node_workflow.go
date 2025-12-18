/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// IHandlerNodeWorkflow defines the node workflow Handler.
// nolint: interfacebloat
type IHandlerNodeWorkflow interface {
	// ListNodeWorkflow list node workflow within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the node-workflow list with page and the total count with filter.
	ListNodeWorkflow(ctx contextx.IContext, page types.Page, condition *types.NodeWorkflowCondition) (
		[]*types.NodeWorkflow, int64, error)

	// CountNodeWorkflow count node workflow by conditions.
	//	@param ctx contextx, contains tenant-id.
	//	@param condition the filter conditions.
	//	@return the node-workflow count with filter.
	CountNodeWorkflow(ctx contextx.IContext, condition *types.NodeWorkflowCondition) (int64, error)

	// DistinctNodeWorkflow distinct node workflow by conditions.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param request the node workflow distinct request.
	// @param conditions the filter conditions.
	// @return the node-workflow distinct result.
	DistinctNodeWorkflow(ctx contextx.IContext, request types.NodeWorkflowDistinctRequest,
		condition *types.NodeWorkflowCondition) (*types.NodeWorkflowDistinctResult, error)

	// ListNodeWorkflowOperation list node workflow operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param workflowID the workflow id.
	// @return the operation list with page and the total count with filter.
	ListNodeWorkflowOperation(ctx contextx.IContext, page types.Page, condition *types.NodeWorkflowOperationCondition) (
		[]*types.NodeWorkflowListOperationResult, int64, error)

	// CountNodeWorkflowOperation count node workflow operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param workflowID the workflow id.
	// @return the operation count with filter.
	CountNodeWorkflowOperation(ctx contextx.IContext, condition *types.NodeWorkflowOperationCondition) (int64, error)

	// ListNodeWorkflowOperationInstance list node workflow operation instance.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the operation instance list with page and the total count with filter.
	ListNodeWorkflowOperationInstance(ctx contextx.IContext, condition *types.OperInstDataCondition) (
		[]*operation.InstanceBriefData, int64, error)

	// CountNodeWorkflowOperationInstance count node workflow operation instance.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param operationID the operation id.
	// @return the operation instance count with filter.
	CountNodeWorkflowOperationInstance(ctx contextx.IContext, condition *types.OperInstDataCondition) (int64, error)

	// GetNodeWorkflowOperationInstanceLog distinct node workflow by conditions.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param request the node workflow distinct request.
	// @param condition the filter conditions.
	// @return the node-workflow distinct result.
	GetNodeWorkflowOperationInstanceLog(ctx contextx.IContext, instanceID string) (*operation.InstanceData, error)

	// ListNodeWorkflowOperationInstanceStatus list node workflow operation instance status.
	// @param ctx contextx, contains tenant-id.
	// @param triggerID the trigger id.
	// @return the operation instance status list.
	ListNodeWorkflowOperationInstanceStatus(ctx contextx.IContext,
		condition *types.NodeWorkflowOperInstanceStatusCondition) ([]*operation.InstanceStatus, error)

	// TerminateNodeWorkflowOperation terminate node operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param param the terminate param.
	// @return the error.
	TerminateNodeWorkflowOperation(ctx contextx.IContext, terminateParam *types.NodeWorkflowOperationTerminateParam) error

	// RetryNodeWorkflowOperation retry node operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param retryParam the retry param.
	// @return the error.
	RetryNodeWorkflowOperation(ctx contextx.IContext, retryParam *types.NodeWorkflowOperationRetryParam) error

	// GetNodeWorkflowOperationManualInfo get workflow operation manual info.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param workflowID the workflow id.
	// @param operationID the operation id.
	// @return the manual info.
	GetNodeWorkflowOperationManualInfo(ctx contextx.IContext, workflowID, operationID string) (*types.NodeWorkflowOperationManualInfo, error)
}

// ListNodeWorkflow list node workflow within specified tenant in contextx.
func (h *Handler) ListNodeWorkflow(ctx contextx.IContext, page types.Page, condition *types.NodeWorkflowCondition) (
	[]*types.NodeWorkflow, int64, error) {

	req := &protoBackend.NodeWorkflowListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listNodeWorkflow(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	result, num := resp.ConvertNodeWorkflowsToTypes()

	return result, num, nil
}

// CountNodeWorkflow count host within specified tenant in contextx.
func (h *Handler) CountNodeWorkflow(ctx contextx.IContext, condition *types.NodeWorkflowCondition) (int64, error) {
	req := &protoBackend.NodeWorkflowListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listNodeWorkflow(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctNodeWorkflow distinct node workflow by conditions.
func (h *Handler) DistinctNodeWorkflow(ctx contextx.IContext, _ types.NodeWorkflowDistinctRequest,
	conditions *types.NodeWorkflowCondition) (*types.NodeWorkflowDistinctResult, error) {

	req := &protoBackend.NodeWorkflowDistinctReq{}
	if err := req.ConvertConditionsFromTypes(conditions); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctNodeWorkflow(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertWorkflowDistinctToTypes(), nil
}

// ListNodeWorkflowOperation list workflow  operation.
func (h *Handler) ListNodeWorkflowOperation(ctx contextx.IContext,
	page types.Page, condition *types.NodeWorkflowOperationCondition) (
	[]*types.NodeWorkflowListOperationResult, int64, error) {

	req := &protoBackend.NodeWorkflowOperationListReq{
		Page:      convertPage(page),
		OnlyCount: false,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listNodeWorkflowOperation(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	operations, total := resp.ConvertWorkflowOperationToTypes()

	return operations, total, nil
}

// CountNodeWorkflowOperation count workflow  operation.
func (h *Handler) CountNodeWorkflowOperation(ctx contextx.IContext,
	condition *types.NodeWorkflowOperationCondition) (int64, error) {

	req := &protoBackend.NodeWorkflowOperationListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}
	resp, err := h.cli.listNodeWorkflowOperation(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// ListNodeWorkflowOperationInstance list workflow operation instance.
func (h *Handler) ListNodeWorkflowOperationInstance(
	ctx contextx.IContext, condition *types.OperInstDataCondition) ([]*operation.InstanceBriefData, int64, error) {

	req := &protoBackend.NodeWorkflowOperationInstanceListReq{
		OnlyCount: false,
	}
	req.ConvertConditionsFromTypes(condition)

	resp, err := h.cli.listNodeWorkflowOperationInstance(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	operations, total := resp.ConvertWorkflowOperationInstanceToTypes()

	return operations, total, nil
}

// CountNodeWorkflowOperationInstance count workflow operation instance.
func (h *Handler) CountNodeWorkflowOperationInstance(ctx contextx.IContext, condition *types.OperInstDataCondition) (int64, error) {
	req := &protoBackend.NodeWorkflowOperationInstanceListReq{
		OnlyCount: true,
	}
	req.ConvertConditionsFromTypes(condition)

	resp, err := h.cli.listNodeWorkflowOperationInstance(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// GetNodeWorkflowOperationInstanceLog get workflow operation instance log.
func (h *Handler) GetNodeWorkflowOperationInstanceLog(ctx contextx.IContext, instanceID string) (
	*operation.InstanceData, error) {

	req := &protoBackend.NodeWorkflowOperationInstanceLogGetReq{
		OperInstId: instanceID,
	}

	resp, err := h.cli.getNodeWorkflowOperationInstanceLog(ctx, req)
	if err != nil {
		return nil, err
	}

	operations := resp.ConvertWorkflowOperationInstanceLogToTypes()

	return operations, nil
}

// ListNodeWorkflowOperationInstanceStatus list workflow operation instance status.
func (h *Handler) ListNodeWorkflowOperationInstanceStatus(ctx contextx.IContext,
	conditions *types.NodeWorkflowOperInstanceStatusCondition) ([]*operation.InstanceStatus, error) {

	req := &protoBackend.NodeWorkflowOperationInstanceListStatusReq{
		Page: &protoBackend.Page{},
	}

	if err := req.ConvertConditionsFromTypes(conditions); err != nil {
		return nil, err
	}

	resp, err := h.cli.listNodeWorkflowOperationInstanceStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	instanceStatus := resp.ConvertWorkflowOperationInstanceStatusToTypes()

	return instanceStatus, nil
}

// RetryNodeWorkflowOperation retry node workflow operation.
func (h *Handler) RetryNodeWorkflowOperation(ctx contextx.IContext, retryParam *types.NodeWorkflowOperationRetryParam) error {
	req := &protoBackend.NodeWorkflowOperationRetryReq{}

	req.ConvertOperationRetryParamFromTypes(retryParam)

	_, err := h.cli.retryNodeWorkflowOperation(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// TerminateNodeWorkflowOperation terminate node operation.
func (h *Handler) TerminateNodeWorkflowOperation(ctx contextx.IContext, terminateParam *types.NodeWorkflowOperationTerminateParam) error {
	req := &protoBackend.NodeWorkflowOperationTerminateReq{}

	req.ConvertOperationTerminateParamFromTypes(terminateParam)

	_, err := h.cli.terminateNodeWorkflowOperation(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// GetNodeWorkflowOperationManualInfo get workflow operation manual info.
func (h *Handler) GetNodeWorkflowOperationManualInfo(ctx contextx.IContext, workflowID, operationID string) (
	*types.NodeWorkflowOperationManualInfo, error) {

	req := &protoBackend.NodeWorkflowOperationManualInfoGetReq{
		WorkflowId:  workflowID,
		OperationId: operationID,
	}

	resp, err := h.cli.getNodeWorkflowManualInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertManualInfoToTypes(), nil
}
