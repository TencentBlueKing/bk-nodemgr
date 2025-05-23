/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeworkflow

import (
	"context"
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler node workflow Handler interface.
type IHandler interface {
	// Count counts node workflow by opts.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// List lists node workflow by page and opts.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.NodeWorkflow, int64, error)

	// Create creates a new node workflow.
	Create(ctx context.Context, workflow *types.NodeWorkflow) error

	// UpdateStatus updates the status of a node workflow.
	UpdateStatus(ctx context.Context, workflowID int64, status types.NodeWorkflowStatus) error
}

// Handler this is a Handler to operate node workflow table.
type Handler struct {
	client *mongo.Database
	logger logger.Logger

	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao)
	}

	newDaoClient := newDao(tenantID, h.client, h.logger)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure node workflow indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New new a Handler.
func New(client *mongo.Database, logger logger.Logger) *Handler {
	return &Handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// Count counts node workflow by opts.
func (h *Handler) Count(ctx context.Context, opts ...OptFn) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).Count(ctx, filter)
}

// List lists node workflow by page and opts.
func (h *Handler) List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.NodeWorkflow, int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	datas, err := h.tenantDao(tenantID).List(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	workflows := make([]*types.NodeWorkflow, len(datas))
	for idx, data := range datas {
		workflows[idx] = convertNodeWorkflowToTypes(data)
	}

	return workflows, num, nil
}

// Create creates a new node workflow.
func (h *Handler) Create(ctx context.Context, workflow *types.NodeWorkflow) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if workflow == nil {
		return base.ErrEmptyParamData()
	}

	if workflow.TriggerID == "" {
		return errors.New("trigger id should not be empty")
	}

	if workflow.Status != types.NodeWorkflowStatusRunning {
		return errors.New("status should be running")
	}

	if err := h.tenantDao(tenantID).Create(ctx, convertNodeWorkflowFromTypes(workflow)); err != nil {
		return err
	}

	return nil
}

// UpdateStatus updates the status of a node workflow.
func (h *Handler) UpdateStatus(ctx context.Context, workflowID int64, status types.NodeWorkflowStatus) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if workflowID <= 0 {
		return errors.New("workflow id should not be empty")
	}

	if err := status.Validate(); err != nil {
		return err
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)
	if err := h.tenantDao(tenantID).UpdateField(ctx, filter, FieldKeyStatus, string(status)); err != nil {
		return err
	}

	return nil
}

// convertNodeWorkflowToTypes convert node workflow to types.
func convertNodeWorkflowToTypes(data *Data) *types.NodeWorkflow {
	return &types.NodeWorkflow{
		WorkflowID:  data.WorkflowID,
		TriggerID:   data.TriggerID,
		Type:        types.NodeWorkflowType(data.Type),
		BizIDs:      data.BizIDs,
		Operator:    data.Operator,
		OperateTime: data.OperateTime,
		Status:      types.NodeWorkflowStatus(data.Status),
	}
}

// convertNodeWorkflowFromTypes convert node workflow from types.
func convertNodeWorkflowFromTypes(workflow *types.NodeWorkflow) *Data {
	return &Data{
		WorkflowID:  workflow.WorkflowID,
		TriggerID:   workflow.TriggerID,
		Type:        string(workflow.Type),
		BizIDs:      workflow.BizIDs,
		Operator:    workflow.Operator,
		OperateTime: workflow.OperateTime,
		Status:      string(workflow.Status),
	}
}
