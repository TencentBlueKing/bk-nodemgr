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
	"errors"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"

	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler node workflow Handler interface.
type IHandler interface {
	// Get gets node workflow by id.
	Get(nCtx contextx.IContext, workflowID string) (*types.NodeWorkflow, error)

	// Count counts node workflow by opts.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists node workflow by page and opts.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.NodeWorkflow, int64, error)

	// Create creates a new node workflow.
	Create(nCtx contextx.IContext, workflow *types.NodeWorkflow) error

	// UpdateStatus updates the status of a node workflow.
	UpdateStatus(nCtx contextx.IContext, workflowID string, status types.NodeWorkflowStatus) error

	// UpdateFinishTime updates the finish time of a node workflow.
	UpdateFinishTime(nCtx contextx.IContext, workflowID string, finishTime time.Time) error

	// IDistinctor distincts node workflow fields.
	IDistinctor
}

// IDistinctor node workflow distinctor interface.
type IDistinctor interface {
	// DistinctNodeWorkflowType distincts with field type.
	DistinctNodeWorkflowType(nCtx contextx.IContext, opts ...OptFn) ([]types.NodeWorkflowType, error)

	// DistinctNodeWorkflowBkBizID distincts with field bk-biz-id.
	DistinctNodeWorkflowBkBizID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error)

	// DistinctNodeWorkflowOperator distincts with field operator.
	DistinctNodeWorkflowOperator(nCtx contextx.IContext, opts ...OptFn) ([]string, error)

	// DistinctNodeWorkflowStatus distincts with field node-version.
	DistinctNodeWorkflowStatus(nCtx contextx.IContext, opts ...OptFn) ([]types.NodeWorkflowStatus, error)
}

// Handler this is a Handler to operate node workflow table.
type Handler struct {
	client *mongo.Database

	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		// nolint: forcetypeassert
		return d.(*dao)
	}

	newDaoClient := newDao(h.client, tenantID)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure node workflow indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	// nolint: forcetypeassert
	return d.(*dao)
}

// New new a Handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Count counts node workflow by opts.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).Count(nCtx, filter)
}

// List lists node workflow by page and opts.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.NodeWorkflow, int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	datas, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt)
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
func (h *Handler) Create(nCtx contextx.IContext, workflow *types.NodeWorkflow) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if workflow == nil {
		return base.ErrEmptyParamData()
	}

	if workflow.TriggerID == "" {
		return errors.New("trigger id should not be empty")
	}

	if workflow.Status != types.NodeWorkflowStatusRunning {
		return errors.New("status should be running")
	}

	if err := h.tenantDao(tenantID).Create(nCtx, convertNodeWorkflowFromTypes(workflow)); err != nil {
		return err
	}

	return nil
}

// Get gets node workflow by id.
func (h *Handler) Get(nCtx contextx.IContext, workflowID string) (*types.NodeWorkflow, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	if workflowID == "" {
		return nil, errors.New("workflow id should not be empty")
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)
	data, err := h.tenantDao(tenantID).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertNodeWorkflowToTypes(data), nil
}

// UpdateStatus updates the status of a node workflow.
func (h *Handler) UpdateStatus(nCtx contextx.IContext, workflowID string, status types.NodeWorkflowStatus) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if workflowID == "" {
		return errors.New("workflow id should not be empty")
	}

	if err := status.Validate(); err != nil {
		return err
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)
	if err := h.tenantDao(tenantID).UpdateField(nCtx, filter, FieldKeyStatus, string(status)); err != nil {
		return err
	}

	return nil
}

// UpdateFinishTime updates the finish time of a node workflow.
func (h *Handler) UpdateFinishTime(nCtx contextx.IContext, workflowID string, finishTime time.Time) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if workflowID == "" {
		return errors.New("workflow id should not be empty")
	}

	if finishTime.IsZero() {
		return errors.New("finish time should not be zero")
	}

	filter := base.AliveFilter()
	filter = WithWorkflowID(workflowID)(filter)
	if err := h.tenantDao(tenantID).UpdateField(nCtx, filter, FieldKeyFinishTime, finishTime); err != nil {
		return err
	}

	return nil
}

// DistinctNodeWorkflowType distincts with field type.
func (h *Handler) DistinctNodeWorkflowType(nCtx contextx.IContext, opts ...OptFn) ([]types.NodeWorkflowType, error) {
	result, err := h.distinctString(nCtx, FieldKeyType, opts...)
	if err != nil {
		return nil, err
	}

	return types.StringListToNodeWorkflowTypeList(result), nil
}

// DistinctNodeWorkflowBkBizID distincts with field bk-biz-id.
func (h *Handler) DistinctNodeWorkflowBkBizID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error) {
	return h.distinctInt64(nCtx, FieldKeyBizID, opts...)
}

// DistinctNodeWorkflowOperator distincts with field operator.
func (h *Handler) DistinctNodeWorkflowOperator(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	return h.distinctString(nCtx, FieldKeyOperator, opts...)
}

// DistinctNodeWorkflowStatus distincts with field node-version.
func (h *Handler) DistinctNodeWorkflowStatus(nCtx contextx.IContext, opts ...OptFn) ([]types.NodeWorkflowStatus, error) {
	result, err := h.distinctString(nCtx, FieldKeyStatus, opts...)
	if err != nil {
		return nil, err
	}

	return types.StringListToNodeWorkflowStatusList(result), nil
}

// distinctInt64 returns distinct values of specified field.
func (h *Handler) distinctInt64(nCtx contextx.IContext, key string, opts ...OptFn) ([]int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).distinctInt64(nCtx, key, filter, nil)
}

// distinctString returns distinct values of specified field.
func (h *Handler) distinctString(nCtx contextx.IContext, key string, opts ...OptFn) ([]string, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).distinctString(nCtx, key, filter, nil)
}

// convertNodeWorkflowToTypes convert node workflow to types.
func convertNodeWorkflowToTypes(data *Data) *types.NodeWorkflow {
	return &types.NodeWorkflow{
		TenantID:    data.TenantID,
		WorkflowID:  data.WorkflowID,
		TriggerID:   data.TriggerID,
		Type:        types.NodeWorkflowType(data.Type),
		BizIDs:      data.BizIDs,
		Operator:    data.Operator,
		OperateTime: data.OperateTime,
		FinishTime:  data.FinishTime,
		Status:      types.NodeWorkflowStatus(data.Status),
	}
}

// convertNodeWorkflowFromTypes convert node workflow from types.
func convertNodeWorkflowFromTypes(workflow *types.NodeWorkflow) *Data {
	return &Data{
		TenantID:    workflow.TenantID,
		WorkflowID:  workflow.WorkflowID,
		TriggerID:   workflow.TriggerID,
		Type:        string(workflow.Type),
		BizIDs:      workflow.BizIDs,
		Operator:    workflow.Operator,
		OperateTime: workflow.OperateTime,
		Status:      string(workflow.Status),
	}
}
