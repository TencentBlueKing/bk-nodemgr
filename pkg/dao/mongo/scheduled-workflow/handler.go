/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package scheduledworkflow provides storage for schedule workflow.
package scheduledworkflow

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler schedule workflow handler interface.
type IHandler interface {
	// Get gets schedule workflow by id.
	Get(nCtx contextx.IContext, workflowID string) (*types.ScheduledWorkflow, error)

	// Count counts schedule workflow by opts.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists schedule workflow by page and opts.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ScheduledWorkflow, int64, error)

	// Create creates a new schedule workflow.
	Create(nCtx contextx.IContext, workflow *types.ScheduledWorkflow) error

	// UpdateTriggerID updates schedule workflow's trigger id.
	UpdateTriggerID(nCtx contextx.IContext, workflowID, triggerID string) error

	// UpdatePrivateData updates schedule workflow's private data.
	UpdatePrivateData(nCtx contextx.IContext, workflowID string, privateData map[string]any) error

	// Switch enables or disables a scheduled workflow.
	Switch(nCtx contextx.IContext, workflowID string, enable bool) error

	// IDistinctor distincts schedule workflow fields.
	IDistinctor
}

// IDistinctor schedule workflow distinctor interface.
type IDistinctor interface {
	// DistinctScheduledWorkflowName distincts with field type.
	DistinctScheduledWorkflowName(nCtx contextx.IContext, opts ...OptFn) ([]string, error)

	// DistinctScheduledWorkflowOperator distincts with field operator.
	DistinctScheduledWorkflowOperator(nCtx contextx.IContext, opts ...OptFn) ([]string, error)
}

// Handler this is a Handler to operate schedule workflow table.
type Handler struct {
	dao *dao
}

// New create a new schedule workflow handler.
func New(client *mongo.Database) *Handler {
	h := &Handler{
		dao: newDao(client),
	}

	if err := h.dao.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure scheduled workflow indexes")
	}

	return h
}

// Count counts schedule workflow by opts.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.dao.Count(nCtx, filter)
}

// List lists schedule workflow by page and opts.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ScheduledWorkflow, int64, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	datas, err := h.dao.List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	workflows := make([]*types.ScheduledWorkflow, len(datas))
	for idx, data := range datas {
		workflows[idx] = convertScheduledWorkflowToTypes(data)
	}

	return workflows, num, nil
}

// Create creates a new schedule workflow.
func (h *Handler) Create(nCtx contextx.IContext, workflow *types.ScheduledWorkflow) error {
	if workflow == nil {
		return base.ErrEmptyParamData()
	}

	if err := h.dao.Create(nCtx, convertScheduledWorkflowFromTypes(workflow)); err != nil {
		return err
	}

	return nil
}

// UpdateTriggerID updates schedule workflow's trigger id.
func (h *Handler) UpdateTriggerID(nCtx contextx.IContext, workflowID, triggerID string) error {
	if workflowID == "" {
		return errors.New("workflow id should not be empty")
	}

	if triggerID == "" {
		return errors.New("trigger id should not be empty")
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)

	if err := h.dao.UpdateField(nCtx, filter, FieldKeyTriggerID, triggerID); err != nil {
		return err
	}

	return nil
}

// UpdatePrivateData updates schedule workflow's private data.
func (h *Handler) UpdatePrivateData(nCtx contextx.IContext, workflowID string, privateData map[string]any) error {
	if workflowID == "" {
		return errors.New("workflow id should not be empty")
	}

	if privateData == nil {
		return errors.New("private data should not be nil")
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)

	if err := h.dao.UpdateField(nCtx, filter, FieldKeyPrivateData, privateData); err != nil {
		return err
	}

	return nil
}

// Get gets schedule workflow by id.
func (h *Handler) Get(nCtx contextx.IContext, workflowID string) (*types.ScheduledWorkflow, error) {
	if workflowID == "" {
		return nil, errors.New("workflow id should not be empty")
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)
	data, err := h.dao.Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertScheduledWorkflowToTypes(data), nil
}

// Switch enables or disables a scheduled workflow.
func (h *Handler) Switch(nCtx contextx.IContext, workflowID string, enable bool) error {
	if workflowID == "" {
		return errors.New("workflow id should not be empty")
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)

	if err := h.dao.UpdateField(nCtx, filter, FieldKeyEnabled, enable); err != nil {
		return err
	}

	return nil
}

// DistinctScheduledWorkflowName distincts with field type.
func (h *Handler) DistinctScheduledWorkflowName(nCtx contextx.IContext, opts ...OptFn) (
	[]string, error) {

	result, err := h.distinctString(nCtx, FieldKeyWorkflowName, opts...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DistinctScheduledWorkflowOperator distincts with field operator.
func (h *Handler) DistinctScheduledWorkflowOperator(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	return h.distinctString(nCtx, FieldKeyOperator, opts...)
}

// distinctString returns distinct values of specified field.
func (h *Handler) distinctString(nCtx contextx.IContext, key string, opts ...OptFn) ([]string, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.dao.distinctString(nCtx, key, filter, nil)
}

// convertScheduledWorkflowToTypes convert schedule workflow to types.
func convertScheduledWorkflowToTypes(data *ScheduledWorkflow) *types.ScheduledWorkflow {
	return &types.ScheduledWorkflow{
		WorkflowID:   data.WorkflowID,
		WorkflowName: data.WorkflowName,
		TenantID:     data.TenantID,
		TriggerID:    data.TriggerID,
		Enabled:      data.Enabled,
		Interval:     data.Interval,
		PrivateData:  data.PrivateData,
		Operator:     data.Operator,
		OperateTime:  data.OperateTime,
	}
}

// convertScheduledWorkflowFromTypes convert schedule workflow from types.
func convertScheduledWorkflowFromTypes(workflow *types.ScheduledWorkflow) *ScheduledWorkflow {
	return &ScheduledWorkflow{
		WorkflowID:   workflow.WorkflowID,
		WorkflowName: workflow.WorkflowName,
		TenantID:     workflow.TenantID,
		TriggerID:    workflow.TriggerID,
		Enabled:      workflow.Enabled,
		Interval:     workflow.Interval,
		PrivateData:  workflow.PrivateData,
		Operator:     workflow.Operator,
		OperateTime:  workflow.OperateTime,
	}
}
