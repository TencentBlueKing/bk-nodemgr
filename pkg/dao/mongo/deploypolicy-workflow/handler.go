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

package deploypolicyworkflow

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler provides tenant-scoped deploy policy workflow persistence.
type IHandler interface {
	Create(nCtx contextx.IContext, workflow *types.DeployPolicyWorkflow) error
	Get(nCtx contextx.IContext, opts ...OptFn) (*types.DeployPolicyWorkflow, error)
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.DeployPolicyWorkflow, int64, error)
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)
	UpdateStatus(nCtx contextx.IContext, workflowID string, status types.DeployPolicyWorkflowStatus) error
	UpdateFinishTime(nCtx contextx.IContext, workflowID string, finishTime time.Time) error
	RecordChild(nCtx contextx.IContext, workflowIDs []string, child types.DeployPolicyWorkflowChild) error
}

// Handler routes workflow operations to tenant collections.
type Handler struct {
	client *mongo.Database
	daoMap sync.Map
}

// New creates a tenant-aware workflow handler.
func New(client *mongo.Database) *Handler {
	return &Handler{client: client}
}

func (h *Handler) tenantDao(nCtx contextx.IContext) (*dao, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("invalid workflow tenant: %w", err)
	}
	tenantID := nCtx.TenantID()
	if cached, ok := h.daoMap.Load(tenantID); ok {
		return cached.(*dao), nil //nolint:forcetypeassert // Only tenantDao writes this map.
	}
	d := newDao(h.client, tenantID)
	// Idempotency requires the unique index; do not cache an uninitialized DAO.
	if err := d.EnsureIndexes(); err != nil {
		return nil, fmt.Errorf("failed to ensure deploy policy workflow indexes: %w", err)
	}
	cached, _ := h.daoMap.LoadOrStore(tenantID, d)

	return cached.(*dao), nil //nolint:forcetypeassert // Only tenantDao writes this map.
}

// Create persists a tenant-scoped deploy policy workflow.
func (h *Handler) Create(nCtx contextx.IContext, workflow *types.DeployPolicyWorkflow) error {
	if workflow == nil {
		return base.ErrEmptyParamData()
	}
	d, err := h.tenantDao(nCtx)
	if err != nil {
		return fmt.Errorf("failed to resolve workflow tenant: %w", err)
	}
	data := convertWorkflowFromTypes(workflow)
	data.TenantID = nCtx.TenantID()
	if err := d.Create(nCtx, data); err != nil {
		return fmt.Errorf("failed to create deploy policy workflow: %w", err)
	}

	return nil
}

// Get fetches one tenant-scoped workflow matching the options.
func (h *Handler) Get(nCtx contextx.IContext, opts ...OptFn) (*types.DeployPolicyWorkflow, error) {
	d, err := h.tenantDao(nCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workflow tenant: %w", err)
	}
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}
	data, err := d.Get(nCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get deploy policy workflow: %w", err)
	}

	return convertWorkflowToTypes(data), nil
}

// List lists tenant-scoped workflows and their total count.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.DeployPolicyWorkflow, int64, error) {
	d, err := h.tenantDao(nCtx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to resolve workflow tenant: %w", err)
	}
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}
	total, err := d.Count(nCtx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deploy policy workflows: %w", err)
	}
	datas, err := d.List(nCtx, filter, base.ParsePage(page))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list deploy policy workflows: %w", err)
	}
	workflows := make([]*types.DeployPolicyWorkflow, len(datas))
	for i, data := range datas {
		workflows[i] = convertWorkflowToTypes(data)
	}

	return workflows, total, nil
}

// Count counts tenant-scoped workflows matching the options.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	d, err := h.tenantDao(nCtx)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve workflow tenant: %w", err)
	}
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}
	total, err := d.Count(nCtx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count deploy policy workflows: %w", err)
	}

	return total, nil
}

// UpdateStatus updates the status of one tenant-scoped workflow.
func (h *Handler) UpdateStatus(nCtx contextx.IContext, workflowID string, status types.DeployPolicyWorkflowStatus) error {
	if workflowID == "" {
		return base.ErrInvalidParam(errors.New("workflow id is empty"))
	}
	if err := status.Validate(); err != nil {
		return base.ErrInvalidParam(err)
	}
	d, err := h.tenantDao(nCtx)
	if err != nil {
		return fmt.Errorf("failed to resolve workflow tenant: %w", err)
	}
	filter := WithWorkflowID(workflowID)(base.AliveFilter())
	if err := d.UpdateField(nCtx, filter, FieldKeyStatus, string(status)); err != nil {
		return fmt.Errorf("failed to update deploy policy workflow status: %w", err)
	}

	return nil
}

// UpdateFinishTime updates the finish time; zero clears it for a running operation.
func (h *Handler) UpdateFinishTime(nCtx contextx.IContext, workflowID string, finishTime time.Time) error {
	if workflowID == "" {
		return base.ErrInvalidParam(errors.New("workflow id is empty"))
	}
	d, err := h.tenantDao(nCtx)
	if err != nil {
		return fmt.Errorf("failed to resolve workflow tenant: %w", err)
	}
	filter := WithWorkflowID(workflowID)(base.AliveFilter())
	if err := d.UpdateField(nCtx, filter, FieldKeyFinishTime, finishTime); err != nil {
		return fmt.Errorf("failed to update deploy policy workflow finish time: %w", err)
	}

	return nil
}

// RecordChild adds a child to each parent without duplicating its workflow ID and domain.
func (h *Handler) RecordChild(nCtx contextx.IContext, workflowIDs []string, child types.DeployPolicyWorkflowChild) error {
	if len(workflowIDs) == 0 {
		return nil
	}
	if child.WorkflowID == "" {
		return base.ErrInvalidParam(errors.New("child workflow id is empty"))
	}
	if child.WorkflowDomain != types.WorkflowDomainNode && child.WorkflowDomain != types.WorkflowDomainPlugin {
		return base.ErrInvalidParam(fmt.Errorf("invalid child workflow domain: %s", child.WorkflowDomain))
	}
	if err := validateWorkflowIDs(workflowIDs); err != nil {
		return fmt.Errorf("invalid parent workflow ids: %w", err)
	}
	d, err := h.tenantDao(nCtx)
	if err != nil {
		return fmt.Errorf("failed to resolve workflow tenant: %w", err)
	}
	storedChild := Child{WorkflowID: child.WorkflowID, WorkflowDomain: string(child.WorkflowDomain)}
	seen := make(map[string]struct{}, len(workflowIDs))
	for _, workflowID := range workflowIDs {
		if _, ok := seen[workflowID]; ok {
			continue
		}
		seen[workflowID] = struct{}{}
		if err := d.recordChild(nCtx, workflowID, storedChild); err != nil {
			return fmt.Errorf("failed to record deploy policy workflow child: %w", err)
		}
	}

	return nil
}

func validateWorkflowIDs(workflowIDs []string) error {
	for _, workflowID := range workflowIDs {
		if workflowID == "" {
			return base.ErrInvalidParam(errors.New("workflow id is empty"))
		}
	}

	return nil
}

func convertWorkflowToTypes(data *Data) *types.DeployPolicyWorkflow {
	children := make([]types.DeployPolicyWorkflowChild, len(data.Children))
	for i, child := range data.Children {
		children[i] = types.DeployPolicyWorkflowChild{
			WorkflowID: child.WorkflowID, WorkflowDomain: types.WorkflowDomain(child.WorkflowDomain),
		}
	}

	return &types.DeployPolicyWorkflow{
		TenantID: data.TenantID, WorkflowID: data.WorkflowID, OperationID: data.OperationID,
		TriggerID: data.TriggerID, DeployPolicyID: data.DeployPolicyID, Operator: data.Operator,
		OperateTime: data.OperateTime, FinishTime: data.FinishTime, Status: types.DeployPolicyWorkflowStatus(data.Status),
		Children: children,
	}
}

func convertWorkflowFromTypes(workflow *types.DeployPolicyWorkflow) *Data {
	children := make([]Child, len(workflow.Children))
	for i, child := range workflow.Children {
		children[i] = Child{
			WorkflowID: child.WorkflowID, WorkflowDomain: string(child.WorkflowDomain),
		}
	}

	return &Data{
		TenantID: workflow.TenantID, WorkflowID: workflow.WorkflowID, OperationID: workflow.OperationID,
		TriggerID: workflow.TriggerID, DeployPolicyID: workflow.DeployPolicyID, Operator: workflow.Operator,
		OperateTime: workflow.OperateTime, FinishTime: workflow.FinishTime, Status: string(workflow.Status),
		Children: children,
	}
}
