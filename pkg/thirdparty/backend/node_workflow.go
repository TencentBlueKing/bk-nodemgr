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

package backend

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// IHandlerNodeWorkflow defines the node workflow Handler.
// nolint: interfacebloat
type IHandlerNodeWorkflow interface {
	// ListNodeWorkflow list node workflow within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the node-workflow list with page and the total count with filter and error.
	ListNodeWorkflow(nCtx contextx.IContext, page types.Page, condition *types.NodeWorkflowCondition) ([]*types.NodeWorkflow, int64, error)

	// CountNodeWorkflow count node workflow by conditions.
	//	@param nCtx contextx, contains tenant-id.
	//	@param condition the filter conditions.
	//	@return the node-workflow count with filter and error.
	CountNodeWorkflow(nCtx contextx.IContext, condition *types.NodeWorkflowCondition) (int64, error)

	// DistinctNodeWorkflow distinct node workflow by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param request the node workflow distinct request.
	// @param conditions the filter conditions.
	// @return the node-workflow distinct result and error.
	DistinctNodeWorkflow(nCtx contextx.IContext, request types.NodeWorkflowDistinctRequest, condition *types.NodeWorkflowCondition) (
		*types.NodeWorkflowDistinctResult, error)

	// ListNodeWorkflowOperation list node workflow operation.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param workflowID the workflow id.
	// @return the operation list with page and the total count with filter and error.
	ListNodeWorkflowOperation(nCtx contextx.IContext, page types.Page, workflowID string, condition *types.ApplicationNodeOperationListCondition) (
		[]*types.NodeWorkflowListOperationResult, int64, error)

	// CountNodeWorkflowOperation count node workflow operation.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param workflowID the workflow id.
	// @return the operation count with filter and error.
	CountNodeWorkflowOperation(nCtx contextx.IContext, workflowID string, condition *types.ApplicationNodeOperationListCondition) (int64, error)

	// DistinctNodeWorkflowOperation distinct node workflow operation by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param workflowID the workflow id.
	// @param selector the node workflow operation distinct selector.
	// @return the node-workflow operation distinct result and error.
	DistinctNodeWorkflowOperation(nCtx contextx.IContext, workflowID string,
		selector types.WorkflowOperationDistinctSelector) (*types.WorkflowOperationDistinctResult, error)

	// ListNodeWorkflowOperationInstance list node workflow operation instance.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the operation instance list with page and the total count with filter and error.
	ListNodeWorkflowOperationInstance(nCtx contextx.IContext, condition *types.OperInstDataCondition) (
		[]*operation.InstanceBriefData, int64, error)

	// CountNodeWorkflowOperationInstance count node workflow operation instance.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param operationID the operation id.
	// @return the operation instance count with filter and error.
	CountNodeWorkflowOperationInstance(nCtx contextx.IContext, condition *types.OperInstDataCondition) (int64, error)

	// GetNodeWorkflowOperationInstanceLog distinct node workflow by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param request the node workflow distinct request.
	// @param condition the filter conditions.
	// @return the node-workflow distinct result and error.
	GetNodeWorkflowOperationInstanceLog(nCtx contextx.IContext, instanceID string) (*operation.InstanceData, error)

	// TerminateNodeWorkflowOperation terminate node operation.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param param the terminate param.
	// @return the error.
	TerminateNodeWorkflowOperation(nCtx contextx.IContext, terminateParam *types.NodeWorkflowOperationTerminateParam) error

	// RetryNodeWorkflowOperation retry node operation.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param retryParam the retry param.
	// @return the error.
	RetryNodeWorkflowOperation(nCtx contextx.IContext, retryParam *types.NodeWorkflowOperationRetryParam) error

	// GetNodeWorkflowOperationManualInfo get workflow operation manual info.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param workflowID the workflow id.
	// @param operationID the operation id.
	// @return the manual info and error.
	GetNodeWorkflowOperationManualInfo(nCtx contextx.IContext, workflowID, operationID string) (*types.NodeWorkflowOperationManualInfo, error)

	// ListNodeWorkflowOperationInstanceStatusDistribution lists the latest operation instance status distribution by trigger id.
	// @param nCtx contextx.IContext, contains tenant-id.
	// @param triggerIDs the trigger ids.
	// @return the latest operation instance status distribution: map[trigger-id]*InstanceStatusDistribution and the error.
	ListNodeWorkflowOperationInstanceStatusDistribution(nCtx contextx.IContext, triggerIDs []string) (
		map[string]*operation.InstanceStatusDistribution, error)

	// GetNodeWorkflowOperationOfflineInstallInfo gets offline install info for an operation.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param operationID the operation id.
	// @return the offline install info data and error.
	GetNodeWorkflowOperationOfflineInstallInfo(nCtx contextx.IContext, operationID string) (
		*protoBackend.NodeWorkflowOperationOfflineInstallInfoGetResp_Data, error)

	// SubmitNodeWorkflowOperationOfflineInstallResult submits the offline install result for an operation.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param operationID the operation id.
	// @param resultData the result data (JSON string of installer.data.json).
	// @return error.
	SubmitNodeWorkflowOperationOfflineInstallResult(nCtx contextx.IContext, operationID, resultData string) error
}

// ListNodeWorkflow list node workflow within specified tenant in contextx.
func (h *Handler) ListNodeWorkflow(nCtx contextx.IContext, page types.Page, condition *types.NodeWorkflowCondition) (
	[]*types.NodeWorkflow, int64, error) {

	req := new(protoBackend.NodeWorkflowListReq)
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	var (
		total     int64
		workflows []*types.NodeWorkflow
	)
	executor := pageexecutor.NewPageExecutor[*types.NodeWorkflow](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.NodeWorkflow, error) {
		req.Page = convertPage(page)
		resp, err := h.cli.listNodeWorkflow(nCtx, req)
		if err != nil {
			return nil, err
		}

		workflows, total = resp.ConvertNodeWorkflowsToTypes()

		return workflows, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, total, nil
}

// CountNodeWorkflow count host within specified tenant in contextx.
func (h *Handler) CountNodeWorkflow(nCtx contextx.IContext, condition *types.NodeWorkflowCondition) (int64, error) {
	req := &protoBackend.NodeWorkflowListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listNodeWorkflow(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctNodeWorkflow distinct node workflow by conditions.
func (h *Handler) DistinctNodeWorkflow(nCtx contextx.IContext, _ types.NodeWorkflowDistinctRequest, conditions *types.NodeWorkflowCondition) (
	*types.NodeWorkflowDistinctResult, error) {

	req := &protoBackend.NodeWorkflowDistinctReq{}
	if err := req.ConvertConditionsFromTypes(conditions); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctNodeWorkflow(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertWorkflowDistinctToTypes(), nil
}

// ListNodeWorkflowOperation list workflow  operation.
func (h *Handler) ListNodeWorkflowOperation(
	nCtx contextx.IContext, page types.Page, workflowID string, condition *types.ApplicationNodeOperationListCondition) (
	[]*types.NodeWorkflowListOperationResult, int64, error) {

	req := &protoBackend.NodeWorkflowOperationListReq{
		Page:       convertPage(page),
		OnlyCount:  false,
		WorkflowId: workflowID,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listNodeWorkflowOperation(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	operations, total := resp.ConvertWorkflowOperationToTypes()

	return operations, total, nil
}

// CountNodeWorkflowOperation count workflow  operation.
func (h *Handler) CountNodeWorkflowOperation(nCtx contextx.IContext, workflowID string, condition *types.ApplicationNodeOperationListCondition) (
	int64, error) {

	req := &protoBackend.NodeWorkflowOperationListReq{
		OnlyCount:  true,
		WorkflowId: workflowID,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}
	resp, err := h.cli.listNodeWorkflowOperation(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctNodeWorkflowOperation distinct workflow operation by conditions.
func (h *Handler) DistinctNodeWorkflowOperation(nCtx contextx.IContext, workflowID string,
	selector types.WorkflowOperationDistinctSelector) (*types.WorkflowOperationDistinctResult, error) {

	req := &protoBackend.NodeWorkflowOperationDistinctReq{
		WorkflowId: workflowID,
	}
	req.ConvertSelectorFromTypes(selector)

	resp, err := h.cli.distinctNodeWorkflowOperation(nCtx, req)
	if err != nil {
		return nil, err
	}

	result, err := resp.ConvertOperationDistinctToTypes()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListNodeWorkflowOperationInstance list workflow operation instance.
func (h *Handler) ListNodeWorkflowOperationInstance(nCtx contextx.IContext, condition *types.OperInstDataCondition) (
	[]*operation.InstanceBriefData, int64, error) {

	req := &protoBackend.NodeWorkflowOperationInstanceListReq{
		OnlyCount: false,
	}
	req.ConvertConditionsFromTypes(condition)

	resp, err := h.cli.listNodeWorkflowOperationInstance(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	operations, total := resp.ConvertWorkflowOperationInstanceToTypes()

	return operations, total, nil
}

// CountNodeWorkflowOperationInstance count workflow operation instance.
func (h *Handler) CountNodeWorkflowOperationInstance(nCtx contextx.IContext, condition *types.OperInstDataCondition) (int64, error) {
	req := &protoBackend.NodeWorkflowOperationInstanceListReq{
		OnlyCount: true,
	}
	req.ConvertConditionsFromTypes(condition)

	resp, err := h.cli.listNodeWorkflowOperationInstance(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// GetNodeWorkflowOperationInstanceLog get workflow operation instance log.
func (h *Handler) GetNodeWorkflowOperationInstanceLog(nCtx contextx.IContext, instanceID string) (*operation.InstanceData, error) {
	req := &protoBackend.NodeWorkflowOperationInstanceLogGetReq{
		OperInstId: instanceID,
	}

	resp, err := h.cli.getNodeWorkflowOperationInstanceLog(nCtx, req)
	if err != nil {
		return nil, err
	}

	operations := resp.ConvertWorkflowOperationInstanceLogToTypes()

	return operations, nil
}

// RetryNodeWorkflowOperation retry node workflow operation.
func (h *Handler) RetryNodeWorkflowOperation(nCtx contextx.IContext, retryParam *types.NodeWorkflowOperationRetryParam) error {
	req := &protoBackend.NodeWorkflowOperationRetryReq{}

	req.ConvertOperationRetryParamFromTypes(retryParam)

	_, err := h.cli.retryNodeWorkflowOperation(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// TerminateNodeWorkflowOperation terminate node operation.
func (h *Handler) TerminateNodeWorkflowOperation(nCtx contextx.IContext, terminateParam *types.NodeWorkflowOperationTerminateParam) error {
	req := &protoBackend.NodeWorkflowOperationTerminateReq{}

	req.ConvertOperationTerminateParamFromTypes(terminateParam)

	_, err := h.cli.terminateNodeWorkflowOperation(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// GetNodeWorkflowOperationManualInfo get workflow operation manual info.
func (h *Handler) GetNodeWorkflowOperationManualInfo(nCtx contextx.IContext, workflowID, operationID string) (
	*types.NodeWorkflowOperationManualInfo, error) {

	req := &protoBackend.NodeWorkflowOperationManualInfoGetReq{
		WorkflowId:  workflowID,
		OperationId: operationID,
	}

	resp, err := h.cli.getNodeWorkflowManualInfo(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertManualInfoToTypes(), nil
}

// GetNodeWorkflowOperationOfflineInstallInfo gets offline install info for an operation.
func (h *Handler) GetNodeWorkflowOperationOfflineInstallInfo(nCtx contextx.IContext, operationID string) (
	*protoBackend.NodeWorkflowOperationOfflineInstallInfoGetResp_Data, error) {

	req := &protoBackend.NodeWorkflowOperationOfflineInstallInfoGetReq{
		OperationId: operationID,
	}

	resp, err := h.cli.getNodeWorkflowOfflineInstallInfo(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetData(), nil
}

// SubmitNodeWorkflowOperationOfflineInstallResult submits the offline install result for an operation.
func (h *Handler) SubmitNodeWorkflowOperationOfflineInstallResult(nCtx contextx.IContext, operationID, resultData string) error {
	req := &protoBackend.NodeWorkflowOperationOfflineInstallResultSubmitReq{
		OperationId: operationID,
		ResultData:  resultData,
	}

	_, err := h.cli.submitNodeWorkflowOfflineInstallResult(nCtx, req)

	return err
}

// ListNodeWorkflowOperationInstanceStatusDistribution lists the latest operation instance status distribution by trigger id.
func (h *Handler) ListNodeWorkflowOperationInstanceStatusDistribution(nCtx contextx.IContext, triggerIDs []string) (
	map[string]*operation.InstanceStatusDistribution, error) {

	req := &protoBackend.NodeWorkflowOperationInstanceStatusDistributionListReq{
		TriggerId: triggerIDs,
	}

	resp, err := h.cli.listNodeWorkflowOperationInstanceStatusDistribution(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertDistributionToTypes(), nil
}
