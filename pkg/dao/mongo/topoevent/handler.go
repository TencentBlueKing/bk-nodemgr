/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topoevent

import (
	"context"
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler topo event handler interface.
type IHandler interface {
	// List list topo events by page and conditions.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.TopoEvent, int64, error)

	// Count count topo events by conditions.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// CreateMany create multiple topo events.
	CreateMany(ctx context.Context, events ...*types.TopoEvent) error

	IDistinctor
}

// IDistinctor defines the interface for distinctor.
type IDistinctor interface {
	// DistinctType distincts with field type.
	DistinctType(ctx context.Context, opts ...OptFn) ([]types.TopoEventType, error)

	// DistinctNetworkAreaID distincts with field networkarea-id.
	DistinctNetworkAreaID(ctx context.Context, opts ...OptFn) ([]int64, error)

	// DistinctNetworkUnitID distincts with field networkunit-id.
	DistinctNetworkUnitID(ctx context.Context, opts ...OptFn) ([]int64, error)

	// DistinctAccessPointID distincts with field accesspoint-id.
	DistinctAccessPointID(ctx context.Context, opts ...OptFn) ([]int64, error)

	// DistinctOperator distincts with field operator.
	DistinctOperator(ctx context.Context, opts ...OptFn) ([]string, error)
}

type handler struct {
	client *mongo.Database
	logger logger.ILogger
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao)
	}

	newDaoClient := newDao(tenantID, h.client, h.logger)
	if err := newDaoClient.ensureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure topoevent indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New create a new topo event handler.
func New(client *mongo.Database, logger logger.ILogger) IHandler {
	return &handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// List list topo events by page and conditions.
func (h *handler) List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.TopoEvent, int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	// descending sort by operate time.
	findOpt.SetSort(bson.D{{FieldKeyOperateTime, -1}})

	events, err := h.tenantDao(tenantID).list(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.TopoEvent, len(events))
	for idx, event := range events {
		data[idx] = convertTopoEventToTypes(event)
	}

	return data, num, nil
}

// Count count the number of topo events by conditions.
func (h *handler) Count(ctx context.Context, opts ...OptFn) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).count(ctx, filter)
	if err != nil {
		return 0, err
	}

	return num, nil
}

// Create create a new topo event.
func (h *handler) CreateMany(ctx context.Context, events ...*types.TopoEvent) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*TopoEvent, len(events))
	for idx, event := range events {
		if event == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertTopoEventFromTypes(event)

		if err = base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.tenantDao(tenantID).createMany(ctx, data); err != nil {
		return err
	}

	return nil
}

// DistinctType returns distinct values of type field.
func (h *handler) DistinctType(ctx context.Context, opts ...OptFn) ([]types.TopoEventType, error) {
	result, err := h.distinctString(ctx, FieldKeyType, opts...)
	if err != nil {
		return nil, err
	}

	return types.StringListToTopoEventTypeList(result), nil
}

// DistinctNetworkAreaID returns distinct values of networkarea-id field.
func (h *handler) DistinctNetworkAreaID(ctx context.Context, opts ...OptFn) ([]int64, error) {
	return h.distinctInt64(ctx, FieldKeyNetworkAreaID, opts...)
}

// DistinctNetworkUnitID returns distinct values of networkunit-id field.
func (h *handler) DistinctNetworkUnitID(ctx context.Context, opts ...OptFn) ([]int64, error) {
	return h.distinctInt64(ctx, FieldKeyNetworkUnitID, opts...)
}

// DistinctAccessPointID returns distinct values of accesspoint-id field.
func (h *handler) DistinctAccessPointID(ctx context.Context, opts ...OptFn) ([]int64, error) {
	return h.distinctInt64(ctx, FieldKeyAccessPointID, opts...)
}

// DistinctOperator returns distinct values of operator field.
func (h *handler) DistinctOperator(ctx context.Context, opts ...OptFn) ([]string, error) {
	return h.distinctString(ctx, FieldKeyOperator, opts...)
}

// distinctInt64 returns distinct values of specified field.
func (h *handler) distinctInt64(ctx context.Context, key string, opts ...OptFn) ([]int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).distinctInt64(ctx, key, filter, nil)
}

func (h *handler) distinctString(ctx context.Context, key string, opts ...OptFn) ([]string, error) {
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

func convertTopoEventFromTypes(event *types.TopoEvent) *TopoEvent {
	return &TopoEvent{
		TenantID:        event.TenantID,
		Type:            string(event.Type),
		NetworkAreaID:   event.NetworkAreaID,
		NetworkAreaName: event.NetworkAreaName,
		NetworkUnitID:   event.NetworkUnitID,
		NetworkUnitName: event.NetworkUnitName,
		AccessPointID:   event.AccessPointID,
		AccessPointName: event.AccessPointName,
		OperateTime:     event.OperateTime,
		Operator:        event.Operator,
	}
}

func convertTopoEventToTypes(event *TopoEvent) *types.TopoEvent {
	return &types.TopoEvent{
		TenantID:        event.TenantID,
		Type:            types.TopoEventType(event.Type),
		NetworkAreaID:   event.NetworkAreaID,
		NetworkAreaName: event.NetworkAreaName,
		NetworkUnitID:   event.NetworkUnitID,
		NetworkUnitName: event.NetworkUnitName,
		AccessPointID:   event.AccessPointID,
		AccessPointName: event.AccessPointName,
		OperateTime:     event.OperateTime,
		Operator:        event.Operator,
	}
}
