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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// IHandlerScheduleWorkflow provides tenant-scoped scheduler configuration and queries.
type IHandlerScheduleWorkflow interface {
	UpdateScheduleWorkflowInterval(ctx contextx.IContext, workflowID, interval string) error
	ListScheduleWorkflow(ctx contextx.IContext, page types.Page, condition *types.ScheduledWorkflowCondition) (
		[]*types.ScheduledWorkflow, int64, error)
	CountScheduleWorkflow(ctx contextx.IContext, condition *types.ScheduledWorkflowCondition) (int64, error)
	ListScheduleWorkflowOperation(ctx contextx.IContext, page types.Page, workflowID, parentInstanceID string) (
		[]*operation.Operation, int64, error)
	CountScheduleWorkflowOperation(ctx contextx.IContext, workflowID, parentInstanceID string) (int64, error)
	ListScheduleWorkflowOperationInstance(ctx contextx.IContext, condition *types.OperInstDataCondition) (
		[]*operation.InstanceBriefData, int64, error)
	CountScheduleWorkflowOperationInstance(ctx contextx.IContext, condition *types.OperInstDataCondition) (int64, error)
	GetScheduleWorkflowOperationInstanceLog(ctx contextx.IContext, instanceID string) (*operation.InstanceData, error)
}

// UpdateScheduleWorkflowInterval saves a schedule expression for asynchronous reconciliation.
func (h *Handler) UpdateScheduleWorkflowInterval(ctx contextx.IContext, workflowID, interval string) error {
	_, err := h.cli.updateScheduleWorkflowInterval(ctx, &protoBackend.UpdateScheduleWorkflowIntervalReq{
		WorkflowId: workflowID,
		Interval:   interval,
	})
	if err != nil {
		return fmt.Errorf("failed to update schedule workflow interval: %w", err)
	}

	return nil
}

// ListScheduleWorkflow reads one page of schedule configurations.
func (h *Handler) ListScheduleWorkflow(ctx contextx.IContext, page types.Page, condition *types.ScheduledWorkflowCondition) (
	[]*types.ScheduledWorkflow, int64, error) {

	req := &protoBackend.ListScheduleWorkflowReq{Page: convertPage(page)}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, fmt.Errorf("failed to convert schedule workflow conditions: %w", err)
	}
	resp, err := h.cli.listScheduleWorkflow(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list schedule workflows: %w", err)
	}
	items, total := resp.ConvertScheduleWorkflowsToTypes()

	return items, total, nil
}

// CountScheduleWorkflow counts configurations matching the supplied conditions.
func (h *Handler) CountScheduleWorkflow(ctx contextx.IContext, condition *types.ScheduledWorkflowCondition) (int64, error) {
	req := &protoBackend.ListScheduleWorkflowReq{OnlyCount: true}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, fmt.Errorf("failed to convert schedule workflow conditions: %w", err)
	}
	resp, err := h.cli.listScheduleWorkflow(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to count schedule workflows: %w", err)
	}

	return resp.GetData().GetTotal(), nil
}

// ListScheduleWorkflowOperation reads root operations or direct children of a parent instance.
func (h *Handler) ListScheduleWorkflowOperation(ctx contextx.IContext, page types.Page, workflowID, parentInstanceID string) (
	[]*operation.Operation, int64, error) {

	req := &protoBackend.ScheduleWorkflowOperationListReq{Page: convertPage(page), WorkflowId: workflowID, ParentOperInstId: parentInstanceID}
	resp, err := h.cli.listScheduleWorkflowOperation(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list schedule workflow operations: %w", err)
	}
	items, total := resp.ConvertResultToTypes()

	return items, total, nil
}

// CountScheduleWorkflowOperation counts root operations or direct children of a parent instance.
func (h *Handler) CountScheduleWorkflowOperation(ctx contextx.IContext, workflowID, parentInstanceID string) (int64, error) {
	req := &protoBackend.ScheduleWorkflowOperationListReq{OnlyCount: true, WorkflowId: workflowID, ParentOperInstId: parentInstanceID}
	resp, err := h.cli.listScheduleWorkflowOperation(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to count schedule workflow operations: %w", err)
	}

	return resp.GetData().GetTotal(), nil
}

// ListScheduleWorkflowOperationInstance reads execution summaries matching operation or instance IDs.
func (h *Handler) ListScheduleWorkflowOperationInstance(ctx contextx.IContext, condition *types.OperInstDataCondition) (
	[]*operation.InstanceBriefData, int64, error) {

	req := new(protoBackend.ScheduleWorkflowOperationInstanceListReq)
	req.ConvertConditionsFromTypes(condition)
	resp, err := h.cli.listScheduleWorkflowOperationInstance(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list schedule workflow instances: %w", err)
	}
	items, total := resp.ConvertResultToTypes()

	return items, total, nil
}

// CountScheduleWorkflowOperationInstance counts matching execution instances.
func (h *Handler) CountScheduleWorkflowOperationInstance(ctx contextx.IContext, condition *types.OperInstDataCondition) (int64, error) {
	req := &protoBackend.ScheduleWorkflowOperationInstanceListReq{OnlyCount: true}
	req.ConvertConditionsFromTypes(condition)
	resp, err := h.cli.listScheduleWorkflowOperationInstance(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to count schedule workflow instances: %w", err)
	}

	return resp.GetData().GetTotal(), nil
}

// GetScheduleWorkflowOperationInstanceLog reads bilingual step logs and child workflow references.
func (h *Handler) GetScheduleWorkflowOperationInstanceLog(ctx contextx.IContext, instanceID string) (*operation.InstanceData, error) {
	resp, err := h.cli.getScheduleWorkflowOperationInstanceLog(ctx, &protoBackend.ScheduleWorkflowOperationInstanceLogGetReq{OperInstId: instanceID})
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule workflow instance logs: %w", err)
	}

	return resp.ConvertResultToTypes(), nil
}
