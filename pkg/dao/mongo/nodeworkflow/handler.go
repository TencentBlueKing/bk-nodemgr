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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/counter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
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
	dao     *dao
	logger  logger.Logger
	counter counter.Handler
}

// New new a Handler.
func New(client *mongo.Database, logger logger.Logger) *Handler {
	h := &Handler{
		dao:     newDao(client, logger),
		logger:  logger,
		counter: counter.New(client, logger),
	}

	if err := h.dao.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure nodedeloyment indexes, err: %v",
			errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	return h
}

// Count counts node workflow by opts.
func (h *Handler) Count(ctx context.Context, opts ...OptFn) (int64, error) {
	if ctx == nil {
		return 0, base.ErrInvalidContext()
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.dao.Count(ctx, filter)
}

// List lists node workflow by page and opts.
func (h *Handler) List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.NodeWorkflow, int64, error) {
	if ctx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	datas, err := h.dao.List(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	workflows := make([]*types.NodeWorkflow, len(datas))
	for idx, data := range datas {
		workflows[idx] = convDataToWorkflow(data)
	}

	return workflows, num, nil
}

// Data is the data of node workflow.
func convDataToWorkflow(data *Data) *types.NodeWorkflow {
	return &types.NodeWorkflow{
		WorkflowID:  data.WorkflowID,
		TriggerID:   data.TriggerID,
		OperType:    types.NodeWorkflowOperType(data.OperType),
		TaskType:    types.NodeWorkflowTaskType(data.TaskType),
		BizIDs:      data.BizIDs,
		ExecuteUser: data.ExecuteUser,
		ExecuteTime: data.ExecuteTime,
		Status:      types.NodeWorkflowStatus(data.Status),
	}
}

// Create creates a new node workflow.
func (h *Handler) Create(ctx context.Context, workflow *types.NodeWorkflow) error {
	if ctx == nil {
		return base.ErrInvalidContext()
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

	var err error
	data := ConvNodeWorkflowToData(workflow)
	data.WorkflowID, err = h.counter.Generate(ctx, TableName)

	if err != nil {
		return err
	}

	if err := h.dao.Create(ctx, data); err != nil {
		return err
	}

	return nil
}

// ConvNodeWorkflowToData convert node workflow to data.
func ConvNodeWorkflowToData(workflow *types.NodeWorkflow) *Data {
	return &Data{
		WorkflowID:  workflow.WorkflowID,
		TriggerID:   workflow.TriggerID,
		OperType:    string(workflow.OperType),
		TaskType:    string(workflow.TaskType),
		BizIDs:      workflow.BizIDs,
		ExecuteUser: workflow.ExecuteUser,
		ExecuteTime: workflow.ExecuteTime,
		Status:      string(workflow.Status),
	}
}

// UpdateStatus updates the status of a node workflow.
func (h *Handler) UpdateStatus(ctx context.Context, workflowID int64, status types.NodeWorkflowStatus) error {
	if ctx == nil {
		return base.ErrInvalidContext()
	}

	if workflowID <= 0 {
		return errors.New("workflow id should not be empty")
	}

	if err := status.Validate(); err != nil {
		return err
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)
	if err := h.dao.UpdateField(ctx, filter, FieldKeyStatus, string(status)); err != nil {
		return err
	}

	return nil
}
