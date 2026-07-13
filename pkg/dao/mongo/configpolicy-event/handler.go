/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package configpolicyevent

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler policy event handler interface.
type IHandler interface {
	// List list policy events by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ConfigPolicyEvent, int64, error)

	// ListWithoutCount list policy events by page and conditions without count.
	ListWithoutCount(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ConfigPolicyEvent, error)

	// Count count policy events by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// CreateMany create multiple policy events.
	CreateMany(nCtx contextx.IContext, events ...*types.ConfigPolicyEvent) error

	IDistinctor
}

// IDistinctor defines the interface for distinctor.
type IDistinctor interface {
	// DistinctBizID distincts with field business id.
	DistinctBizID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error)

	// DistinctEventType distincts with field event type.
	DistinctEventType(nCtx contextx.IContext, opts ...OptFn) ([]types.ConfigPolicyEventType, error)

	// DistinctVersion distincts with field version.
	DistinctVersion(nCtx contextx.IContext, opts ...OptFn) ([]int64, error)

	// DistinctOperator distincts with field operator.
	DistinctOperator(nCtx contextx.IContext, opts ...OptFn) ([]string, error)

	// DistinctConfigPolicyType distincts with field config policy type.
	DistinctConfigPolicyType(nCtx contextx.IContext, opts ...OptFn) ([]types.ConfigPolicyType, error)

	// DistinctConfigPolicyID distincts with field config policy id.
	DistinctConfigPolicyID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error)

	// DistinctConfigPolicyName distincts with field config policy name.
	DistinctConfigPolicyName(nCtx contextx.IContext, opts ...OptFn) ([]string, error)
}

var _ IHandler = &Handler{}

// Handler policy event handler.
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

	newDaoClient := newDao(tenantID, h.client)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure config policy indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new accesspoint handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// List list policy events by page and conditions.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ConfigPolicyEvent, int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(nCtx.TenantID()).Count(nCtx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count policy events: %w", err)
	}

	findOpt := base.ParsePage(page)

	events, err := h.tenantDao(nCtx.TenantID()).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list policy events: %w", err)
	}

	data := make([]*types.ConfigPolicyEvent, len(events))
	for idx, event := range events {
		data[idx] = convertConfigPolicyEventToTypes(event)
	}

	return data, num, nil
}

// ListWithoutCount list policy events by page and conditions without count.
func (h *Handler) ListWithoutCount(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ConfigPolicyEvent, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	findOpt := base.ParsePage(page)

	events, err := h.tenantDao(nCtx.TenantID()).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to list policy events: %w", err)
	}

	data := make([]*types.ConfigPolicyEvent, len(events))
	for idx, event := range events {
		data[idx] = convertConfigPolicyEventToTypes(event)
	}

	return data, nil
}

// Count count the number of policy events by conditions.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(nCtx.TenantID()).Count(nCtx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count policy events: %w", err)
	}

	return num, nil
}

// CreateMany create a new policy event.
func (h *Handler) CreateMany(nCtx contextx.IContext, events ...*types.ConfigPolicyEvent) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	if len(events) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*ConfigPolicyEvent, len(events))
	for idx, event := range events {
		if event == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertConfigPolicyEventFromTypes(event)

		// generate sequence
		sequence, err := h.tenantDao(nCtx.TenantID()).counter.Generate(nCtx, tableNamePrefix)
		if err != nil {
			return fmt.Errorf("failed to generate sequence: %w", err)
		}

		data[idx].EventID = sequence
	}

	if err := h.tenantDao(nCtx.TenantID()).CreateMany(nCtx, data); err != nil {
		return fmt.Errorf("failed to create policy events: %w", err)
	}

	return nil
}

// DistinctBizID returns distinct values of business id field.
func (h *Handler) DistinctBizID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error) {
	return h.distinctInt64(nCtx, FieldKeyBizID, opts...)
}

// DistinctEventType returns distinct values of event type field.
func (h *Handler) DistinctEventType(nCtx contextx.IContext, opts ...OptFn) ([]types.ConfigPolicyEventType, error) {
	result, err := h.distinctString(nCtx, FieldKeyType, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to distinct config policy event type: %w", err)
	}

	eventTypeList, err := types.StringListToConfigPolicyEventTypeList(result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert string list to config policy event type list: %w", err)
	}

	return eventTypeList, nil
}

// DistinctConfigPolicyType returns distinct values of config policy type field.
func (h *Handler) DistinctConfigPolicyType(nCtx contextx.IContext, opts ...OptFn) ([]types.ConfigPolicyType, error) {
	result, err := h.distinctString(nCtx, FieldKeyConfigPolicyType, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to distinct config policy type: %w", err)
	}

	policyTypeList, err := types.StringListToConfigPolicyTypeList(result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert string list to config policy type list: %w", err)
	}

	return policyTypeList, nil
}

// DistinctOperator returns distinct values of operator field.
func (h *Handler) DistinctOperator(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	return h.distinctString(nCtx, FieldKeyOperator, opts...)
}

// DistinctVersion returns distinct values of version field.
func (h *Handler) DistinctVersion(nCtx contextx.IContext, opts ...OptFn) ([]int64, error) {
	return h.distinctInt64(nCtx, FieldKeyVersion, opts...)
}

// DistinctConfigPolicyID returns distinct values of config policy id field.
func (h *Handler) DistinctConfigPolicyID(nCtx contextx.IContext, opts ...OptFn) ([]int64, error) {
	return h.distinctInt64(nCtx, FieldKeyConfigPolicyID, opts...)
}

// DistinctConfigPolicyName returns distinct values of config policy name field.
func (h *Handler) DistinctConfigPolicyName(nCtx contextx.IContext, opts ...OptFn) ([]string, error) {
	return h.distinctString(nCtx, FieldKeyConfigPolicyName, opts...)
}

func (h *Handler) distinctString(nCtx contextx.IContext, key string, opts ...OptFn) ([]string, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	result, err := h.tenantDao(nCtx.TenantID()).DistinctString(nCtx, key, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to distinct string: %w", err)
	}

	return result, nil
}

// distinctInt64 returns distinct values of specified field.
func (h *Handler) distinctInt64(nCtx contextx.IContext, key string, opts ...OptFn) ([]int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	result, err := h.tenantDao(nCtx.TenantID()).DistinctInt64(nCtx, key, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to distinct int64: %w", err)
	}

	return result, nil
}

func convertConfigPolicyEventToTypes(event *ConfigPolicyEvent) *types.ConfigPolicyEvent {
	return &types.ConfigPolicyEvent{
		TenantID:         event.TenantID,
		BizID:            event.BizID,
		Type:             types.ConfigPolicyEventType(event.Type),
		ConfigPolicyID:   event.ConfigpolicyID,
		ConfigPolicyName: event.ConfigpolicyName,
		ConfigPolicyType: types.ConfigPolicyType(event.ConfigpolicyType),
		Version:          event.Version,
		Operator:         event.Operator,
		OperateTime:      event.OperateTime,
	}
}

func convertConfigPolicyEventFromTypes(event *types.ConfigPolicyEvent) *ConfigPolicyEvent {
	return &ConfigPolicyEvent{
		TenantID:         event.TenantID,
		BizID:            event.BizID,
		Type:             string(event.Type),
		ConfigpolicyType: string(event.ConfigPolicyType),
		Version:          event.Version,
		ConfigpolicyID:   event.ConfigPolicyID,
		ConfigpolicyName: event.ConfigPolicyName,
		Operator:         event.Operator,
		OperateTime:      event.OperateTime,
	}
}
