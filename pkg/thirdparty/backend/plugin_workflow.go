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

// IHandlerPluginWorkflow defines the plugin workflow Handler.
// nolint: interfacebloat
type IHandlerPluginWorkflow interface {
	// ListPluginWorkflow list plugin workflow within specified tenant in contextx.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the plugin-workflow list with page and the total count with filter.
	ListPluginWorkflow(ctx contextx.IContext, page types.Page, condition *types.PluginWorkflowCondition) (
		[]*types.PluginWorkflow, int64, error)

	// CountPluginWorkflow count plugin workflow by conditions.
	//	@param ctx contextx, contains tenant-id.
	//	@param condition the filter conditions.
	//	@return the plugin-workflow count with filter.
	CountPluginWorkflow(ctx contextx.IContext, condition *types.PluginWorkflowCondition) (int64, error)

	// DistinctPluginWorkflow distinct plugin workflow by conditions.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param request the plugin workflow distinct request.
	// @param conditions the filter conditions.
	// @return the plugin-workflow distinct result.
	DistinctPluginWorkflow(ctx contextx.IContext, request types.PluginWorkflowDistinctRequest,
		condition *types.PluginWorkflowCondition) (*types.PluginWorkflowDistinctResult, error)

	// ListPluginWorkflowOperation list plugin workflow operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param workflowID the workflow id.
	// @return the operation list with page and the total count with filter.
	ListPluginWorkflowOperation(ctx contextx.IContext, page types.Page, condition *types.PluginWorkflowOperationCondition) (
		[]*types.PluginWorkflowListOperationResult, int64, error)

	// CountPluginWorkflowOperation count plugin workflow operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param workflowID the workflow id.
	// @return the operation count with filter.
	CountPluginWorkflowOperation(ctx contextx.IContext, condition *types.PluginWorkflowOperationCondition) (int64, error)

	// ListPluginWorkflowOperationInstance list plugin workflow operation instance.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param operationID the operation id.
	// @return the operation instance list with page and the total count with filter.
	ListPluginWorkflowOperationInstance(ctx contextx.IContext, operationID ...string) (
		[]*operation.InstanceBriefData, int64, error)

	// CountPluginWorkflowOperationInstance count plugin workflow operation instance.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param operationID the operation id.
	// @return the operation instance count with filter.
	CountPluginWorkflowOperationInstance(ctx contextx.IContext, operationID ...string) (int64, error)

	// GetPluginWorkflowOperationInstanceLog distinct plugin workflow by conditions.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param request the plugin workflow distinct request.
	// @param condition the filter conditions.
	// @return the plugin-workflow distinct result.
	GetPluginWorkflowOperationInstanceLog(ctx contextx.IContext, instanceID string) (*operation.InstanceData, error)

	// ListPluginWorkflowOperationInstanceStatus list plugin workflow operation instance status.
	// @param ctx contextx, contains tenant-id.
	// @param triggerID the trigger id.
	// @return the operation instance status list.
	ListPluginWorkflowOperationInstanceStatus(ctx contextx.IContext,
		condition *types.PluginWorkflowOperInstanceStatusCondition) ([]*operation.InstanceStatus, error)

	// RetryPluginWorkflowOperation retry plugin operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param retryParam the retry param.
	// @return the error.
	RetryPluginWorkflowOperation(ctx contextx.IContext, retryParam *types.PluginWorkflowOperationRetryParam) error

	// TerminatePluginWorkflowOperation terminate plugin operation.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param param the terminate param.
	// @return the error.
	TerminatePluginWorkflowOperation(ctx contextx.IContext, terminateParam *types.PluginWorkflowOperationTerminateParam) error
}

// ListPluginWorkflow list node workflow within specified tenant in contextx.
func (h *Handler) ListPluginWorkflow(ctx contextx.IContext, page types.Page, condition *types.PluginWorkflowCondition) (
	[]*types.PluginWorkflow, int64, error) {

	req := &protoBackend.PluginWorkflowListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listPluginWorkflows(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	result, num := resp.ConvertPluginWorkflowsToTypes()

	return result, num, nil
}

// CountPluginWorkflow count host within specified tenant in contextx.
func (h *Handler) CountPluginWorkflow(ctx contextx.IContext, condition *types.PluginWorkflowCondition) (int64, error) {
	req := &protoBackend.PluginWorkflowListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listPluginWorkflows(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctPluginWorkflow distinct node workflow by conditions.
func (h *Handler) DistinctPluginWorkflow(ctx contextx.IContext, _ types.PluginWorkflowDistinctRequest,
	conditions *types.PluginWorkflowCondition) (*types.PluginWorkflowDistinctResult, error) {

	req := &protoBackend.PluginWorkflowDistinctReq{}
	if err := req.ConvertConditionsFromTypes(conditions); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctPluginWorkflows(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertWorkflowDistinctToTypes(), nil
}

// ListPluginWorkflowOperation list workflow  operation.
func (h *Handler) ListPluginWorkflowOperation(ctx contextx.IContext,
	page types.Page, condition *types.PluginWorkflowOperationCondition) (
	[]*types.PluginWorkflowListOperationResult, int64, error) {

	req := &protoBackend.PluginWorkflowOperationListReq{
		Page:      convertPage(page),
		OnlyCount: false,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listPluginWorkflowOperation(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	operations, total := resp.ConvertWorkflowOperationToTypes()

	return operations, total, nil
}

// CountPluginWorkflowOperation count workflow  operation.
func (h *Handler) CountPluginWorkflowOperation(ctx contextx.IContext,
	condition *types.PluginWorkflowOperationCondition) (int64, error) {

	req := &protoBackend.PluginWorkflowOperationListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}
	resp, err := h.cli.listPluginWorkflowOperation(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotalCount(), nil
}

// ListPluginWorkflowOperationInstance list workflow operation instance.
func (h *Handler) ListPluginWorkflowOperationInstance(
	ctx contextx.IContext, operationID ...string) ([]*operation.InstanceBriefData, int64, error) {

	req := &protoBackend.PluginWorkflowOperationInstanceListReq{
		OnlyCount:   false,
		OperationId: operationID,
	}

	resp, err := h.cli.listPluginWorkflowOperationInstance(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	operations, total := resp.ConvertWorkflowOperationInstanceToTypes()

	return operations, total, nil
}

// CountPluginWorkflowOperationInstance count workflow operation instance.
func (h *Handler) CountPluginWorkflowOperationInstance(ctx contextx.IContext, operationID ...string) (int64, error) {
	req := &protoBackend.PluginWorkflowOperationInstanceListReq{
		OnlyCount:   true,
		OperationId: operationID,
	}

	resp, err := h.cli.listPluginWorkflowOperationInstance(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// GetPluginWorkflowOperationInstanceLog get workflow operation instance log.
func (h *Handler) GetPluginWorkflowOperationInstanceLog(ctx contextx.IContext, instanceID string) (
	*operation.InstanceData, error) {

	req := &protoBackend.PluginWorkflowOperationInstanceLogGetReq{
		OperInstId: instanceID,
	}

	resp, err := h.cli.getPluginWorkflowOperationInstanceLog(ctx, req)
	if err != nil {
		return nil, err
	}

	operations := resp.ConvertWorkflowOperationInstanceLogToTypes()

	return operations, nil
}

// ListPluginWorkflowOperationInstanceStatus list workflow operation instance status.
func (h *Handler) ListPluginWorkflowOperationInstanceStatus(ctx contextx.IContext,
	conditions *types.PluginWorkflowOperInstanceStatusCondition) ([]*operation.InstanceStatus, error) {

	req := &protoBackend.PluginWorkflowOperationInstanceListStatusReq{
		Page: &protoBackend.Page{},
	}

	if err := req.ConvertConditionsFromTypes(conditions); err != nil {
		return nil, err
	}

	resp, err := h.cli.listPluginWorkflowOperationInstanceStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	instanceStatus := resp.ConvertWorkflowOperationInstanceStatusToTypes()

	return instanceStatus, nil
}

// RetryPluginWorkflowOperation retry node workflow operation.
func (h *Handler) RetryPluginWorkflowOperation(ctx contextx.IContext, retryParam *types.PluginWorkflowOperationRetryParam) error {
	req := &protoBackend.PluginWorkflowOperationRetryReq{}

	req.ConvertOperationRetryParamFromTypes(retryParam)

	err := h.cli.retryPluginWorkflowOperation(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// TerminatePluginWorkflowOperation terminate node operation.
func (h *Handler) TerminatePluginWorkflowOperation(ctx contextx.IContext, terminateParam *types.PluginWorkflowOperationTerminateParam) error {
	req := &protoBackend.PluginWorkflowOperationTerminateReq{}

	req.ConvertOperationTerminateParamFromTypes(terminateParam)

	err := h.cli.terminatePluginWorkflowOperation(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
