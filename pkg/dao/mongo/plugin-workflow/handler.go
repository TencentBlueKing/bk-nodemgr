/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pluginworkflow

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

// IHandler plugin workflow Handler interface.
type IHandler interface {
	// Get gets plugin workflow by id.
	Get(nCtx contextx.IContext, workflowID string) (*types.PluginWorkflow, error)

	// Count counts plugin workflow by opts.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists plugin workflow by page and opts.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PluginWorkflow, int64, error)

	// Create creates a new plugin workflow.
	Create(nCtx contextx.IContext, workflow *types.PluginWorkflow) error

	// UpdateStatus updates the status of a plugin workflow.
	UpdateStatus(nCtx contextx.IContext, workflowID string, status types.PluginWorkflowStatus) error

	// UpdateFinishTime updates the finish time of a plugin workflow.
	UpdateFinishTime(nCtx contextx.IContext, workflowID string, finishTime time.Time) error

	IDistinctor
}

// IDistinctor plugin workflow distinctor interface.
type IDistinctor interface {
	// DistinctPluginWorkflowType distincts with field type.
	DistinctPluginWorkflowType(nCtx contextx.IContext, opts ...OptFn) ([]types.PluginWorkflowType, error)

	// DistinctPluginWorkflowBkHostID distincts with field bk-host-id.
	DistinctPluginWorkflowBkHostID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error)

	// DistinctPluginWorkflowOperator distincts with field operator.
	DistinctPluginWorkflowOperator(nCtx contextx.IContext, opts ...OptFn) ([]string, error)

	// DistinctPluginWorkflowStatus distincts with field status.
	DistinctPluginWorkflowStatus(nCtx contextx.IContext, opts ...OptFn) ([]types.PluginWorkflowStatus, error)
}

// Handler this is a Handler to operate plugin workflow table.
type Handler struct {
	client *mongo.Database

	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(h.client, tenantID)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure plugin workflow indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New new a Handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Count counts plugin workflow by opts.
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

// List lists plugin workflow by page and opts.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PluginWorkflow, int64, error) {
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

	workflows := make([]*types.PluginWorkflow, len(datas))
	for idx, data := range datas {
		workflows[idx] = convertPluginWorkflowToTypes(data)
	}

	return workflows, num, nil
}

// Create creates a new plugin workflow.
func (h *Handler) Create(nCtx contextx.IContext, workflow *types.PluginWorkflow) error {
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

	if err := h.tenantDao(tenantID).Create(nCtx, convertPluginWorkflowFromTypes(workflow)); err != nil {
		return err
	}

	return nil
}

// Get gets plugin workflow by id.
func (h *Handler) Get(nCtx contextx.IContext, workflowID string) (*types.PluginWorkflow, error) {
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

	return convertPluginWorkflowToTypes(data), nil
}

// UpdateStatus updates the status of a plugin workflow.
func (h *Handler) UpdateStatus(nCtx contextx.IContext, workflowID string, status types.PluginWorkflowStatus) error {
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

// UpdateFinishTime updates the finish time of a plugin workflow.
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

// DistinctPluginWorkflowType distincts with field type.
func (h *Handler) DistinctPluginWorkflowType(nCtx contextx.IContext, opts ...OptFn) ([]types.PluginWorkflowType, error) {
	result, err := h.distinctString(nCtx, FieldKeyType, opts...)
	if err != nil {
		return nil, err
	}

	pluginWorkflowTypes := make([]types.PluginWorkflowType, 0, len(result))
	for _, v := range result {
		pluginWorkflowTypes = append(pluginWorkflowTypes, types.PluginWorkflowType(v))
	}

	return pluginWorkflowTypes, nil
}

// DistinctPluginWorkflowBkHostID distincts with field bk-biz-id.
func (h *Handler) DistinctPluginWorkflowBkHostID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error) {
	return h.distinctInt64(nCtx, FieldKeyHostIDs, opts...)
}

// DistinctPluginWorkflowOperator distincts with field operator.
func (h *Handler) DistinctPluginWorkflowOperator(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	return h.distinctString(nCtx, FieldKeyOperator, opts...)
}

// DistinctPluginWorkflowStatus distincts with field plugin-version.
func (h *Handler) DistinctPluginWorkflowStatus(nCtx contextx.IContext, opts ...OptFn) ([]types.PluginWorkflowStatus, error) {
	result, err := h.distinctString(nCtx, FieldKeyStatus, opts...)
	if err != nil {
		return nil, err
	}

	pluginWorkflowStatus := make([]types.PluginWorkflowStatus, 0, len(result))
	for _, v := range result {
		pluginWorkflowStatus = append(pluginWorkflowStatus, types.PluginWorkflowStatus(v))
	}

	return pluginWorkflowStatus, nil
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

// convertPluginWorkflowToTypes convert plugin workflow to types.
func convertPluginWorkflowToTypes(data *Data) *types.PluginWorkflow {
	return &types.PluginWorkflow{
		WorkflowID:  data.WorkflowID,
		TriggerID:   data.TriggerID,
		Type:        types.PluginWorkflowType(data.Type),
		HostIDs:     data.HostIDs,
		Operator:    data.Operator,
		OperateTime: data.OperateTime,
		FinishTime:  data.FinishTime,
		Status:      types.PluginWorkflowStatus(data.Status),
	}
}

// convertPluginWorkflowFromTypes convert plugin workflow from types.
func convertPluginWorkflowFromTypes(workflow *types.PluginWorkflow) *Data {
	return &Data{
		WorkflowID:  workflow.WorkflowID,
		TriggerID:   workflow.TriggerID,
		Type:        string(workflow.Type),
		HostIDs:     workflow.HostIDs,
		Operator:    workflow.Operator,
		OperateTime: workflow.OperateTime,
		Status:      string(workflow.Status),
	}
}
