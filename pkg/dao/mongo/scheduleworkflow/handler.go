/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package scheduleworkflow provides storage for schedule workflow.
package scheduleworkflow

import (
	"context"
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/schedule"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler schedule workflow handler interface.
type IHandler interface {
	// Get gets schedule workflow by id.
	Get(ctx context.Context, workflowID string) (*schedule.Schedule, error)

	// Count counts schedule workflow by opts.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// List lists schedule workflow by page and opts.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*schedule.Schedule, int64, error)

	// Create creates a new schedule workflow.
	Create(ctx context.Context, workflow *schedule.Schedule) error

	// IDistinctor distincts schedule workflow fields.
	IDistinctor
}

// IDistinctor schedule workflow distinctor interface.
type IDistinctor interface {
	// DistinctScheduleWorkflowName distincts with field type.
	DistinctScheduleWorkflowName(ctx context.Context, opts ...OptFn) ([]string, error)

	// DistinctScheduleWorkflowOperator distincts with field operator.
	DistinctScheduleWorkflowOperator(ctx context.Context, opts ...OptFn) ([]string, error)
}

// Handler this is a Handler to operate schedule workflow table.
type Handler struct {
	client *mongo.Database
	logger logger.ILogger

	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(h.client, h.logger)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure schedule workflow indexes: %v",
			errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new schedule workflow handler.
func New(client *mongo.Database, logger logger.ILogger) *Handler {
	return &Handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// Count counts schedule workflow by opts.
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

// List lists schedule workflow by page and opts.
func (h *Handler) List(ctx context.Context, page types.Page, opts ...OptFn) ([]*schedule.Schedule, int64, error) {
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

	workflows := make([]*schedule.Schedule, len(datas))
	for idx, data := range datas {
		workflows[idx] = convertScheduleWorkflowToTypes(data)
	}

	return workflows, num, nil
}

// Create creates a new schedule workflow.
func (h *Handler) Create(ctx context.Context, workflow *schedule.Schedule) error {
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

	if err := h.tenantDao(tenantID).Create(ctx, convertScheduleWorkflowFromTypes(workflow)); err != nil {
		return err
	}

	return nil
}

// Get gets schedule workflow by id.
func (h *Handler) Get(ctx context.Context, workflowID string) (*schedule.Schedule, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	if workflowID == "" {
		return nil, errors.New("workflow id should not be empty")
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)
	data, err := h.tenantDao(tenantID).Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return convertScheduleWorkflowToTypes(data), nil
}

// DistinctScheduleWorkflowName distincts with field type.
func (h *Handler) DistinctScheduleWorkflowName(ctx context.Context, opts ...OptFn) (
	[]string, error) {

	result, err := h.distinctString(ctx, FieldKeyWorkflowName, opts...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DistinctScheduleWorkflowOperator distincts with field operator.
func (h *Handler) DistinctScheduleWorkflowOperator(ctx context.Context, opts ...OptFn) ([]string, error) {
	return h.distinctString(ctx, FieldKeyOperator, opts...)
}

// distinctString returns distinct values of specified field.
func (h *Handler) distinctString(ctx context.Context, key string, opts ...OptFn) ([]string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).distinctString(ctx, key, filter, nil)
}

// convertScheduleWorkflowToTypes convert schedule workflow to types.
func convertScheduleWorkflowToTypes(data *ScheduleWorkflow) *schedule.Schedule {
	return &schedule.Schedule{
		WorkflowID:   data.WorkflowID,
		WorkflowName: data.WorkflowName,
		TriggerID:    data.TriggerID,
		Operator:     data.Operator,
		OperateTime:  data.OperateTime,
	}
}

// convertScheduleWorkflowFromTypes convert schedule workflow from types.
func convertScheduleWorkflowFromTypes(workflow *schedule.Schedule) *ScheduleWorkflow {
	return &ScheduleWorkflow{
		WorkflowID:   workflow.WorkflowID,
		WorkflowName: workflow.WorkflowName,
		TriggerID:    workflow.TriggerID,
		Operator:     workflow.Operator,
		OperateTime:  workflow.OperateTime,
	}
}
