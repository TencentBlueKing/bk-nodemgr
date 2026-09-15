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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler provides tenant-scoped deploy policy workflow persistence.
type IHandler interface {
	Create(nCtx contextx.IContext, workflow *types.DeployPolicyWorkflow) error
	Get(nCtx contextx.IContext, opts ...OptFn) (*types.DeployPolicyWorkflow, error)
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

// RecordChild atomically adds or confirms a child without downgrading acknowledgement.
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
	storedChild := Child{WorkflowID: child.WorkflowID, WorkflowDomain: string(child.WorkflowDomain), Confirmed: child.Confirmed}
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
			WorkflowID: child.WorkflowID, WorkflowDomain: types.WorkflowDomain(child.WorkflowDomain), Confirmed: child.Confirmed,
		}
	}

	return &types.DeployPolicyWorkflow{
		TenantID: data.TenantID, WorkflowID: data.WorkflowID, OperationID: data.OperationID,
		TriggerID: data.TriggerID, DeployPolicyID: data.DeployPolicyID, Operator: data.Operator,
		OperateTime: data.OperateTime,
		Children:    children,
	}
}

func convertWorkflowFromTypes(workflow *types.DeployPolicyWorkflow) *Data {
	children := make([]Child, len(workflow.Children))
	for i, child := range workflow.Children {
		children[i] = Child{
			WorkflowID: child.WorkflowID, WorkflowDomain: string(child.WorkflowDomain), Confirmed: child.Confirmed,
		}
	}

	return &Data{
		TenantID: workflow.TenantID, WorkflowID: workflow.WorkflowID, OperationID: workflow.OperationID,
		TriggerID: workflow.TriggerID, DeployPolicyID: workflow.DeployPolicyID, Operator: workflow.Operator,
		OperateTime: workflow.OperateTime,
		Children:    children,
	}
}
