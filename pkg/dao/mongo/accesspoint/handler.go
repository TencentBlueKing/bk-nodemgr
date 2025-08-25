/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package accesspoint

import (
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler accesspoint handler interface.
type IHandler interface {
	// Count counts accesspoint by conditions.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// List lists accesspoint by page and conditions.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.AccessPoint, int64, error)

	// Get gets accesspoint by id.
	Get(ctx context.Context, accessPointID int64) (*types.AccessPoint, error)

	// Create creates accesspoint.
	Create(ctx context.Context, accessPoint *types.AccessPoint) (int64, error)

	// CreateMany creates many accesspoint.
	CreateMany(ctx context.Context, accessPoints ...*types.AccessPoint) ([]int64, error)

	// UpdateMany updates accesspoint.
	UpdateMany(ctx context.Context, accessPoints ...*types.AccessPoint) error

	// DeleteMany deletes accesspoint.
	DeleteMany(ctx context.Context, accessPointIDs ...int64) error
}

type handler struct {
	dao    *dao
	logger logger.ILogger
}

// New create a new accesspoint handler.
func New(client *mongo.Database, logger logger.ILogger) IHandler {
	h := &handler{
		dao:    newDao(client, logger),
		logger: logger,
	}

	if err := h.dao.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure accesspoint indexes, err: %v",
			errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	return h
}

// Count counts networkunit by conditions.
func (h *handler) Count(ctx context.Context, opts ...OptFn) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}
	filter = append(filter, tenantFilter(tenantID))

	return h.dao.Count(ctx, filter)
}

// List lists accesspoint by page and conditions.
func (h *handler) List(ctx context.Context, page types.Page, opts ...OptFn) (
	[]*types.AccessPoint, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}
	filter = append(filter, tenantFilter(tenantID))

	num, err := h.dao.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	accessPoints, err := h.dao.List(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		data[idx] = convertAccessPointToTypes(accessPoint)
	}

	return data, num, nil
}

// Get gets accesspoint by id.
func (h *handler) Get(ctx context.Context, accessPointID int64) (*types.AccessPoint, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	if accessPointID < 0 {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithInt64Values(FieldKeyAccessPointID, accessPointID)
	filter = opt(filter)
	filter = append(filter, tenantFilter(tenantID))

	data, err := h.dao.Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return convertAccessPointToTypes(data), nil
}

// Create creates a new accesspoint and return the generated id.
func (h *handler) Create(ctx context.Context, accessPoint *types.AccessPoint) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return -1, err
	}

	if accessPoint == nil {
		return -1, base.ErrEmptyParamData()
	}

	if err = base.CheckTenantIDMatched(tenantID, accessPoint.TenantID); err != nil {
		return -1, err
	}

	if accessPoint.Name == "" {
		return -1, errors.New("accesspoint name is empty")
	}

	if accessPoint.NetworkAreaID < 0 {
		return -1, errors.New("accesspoint networkarea-id is invalid")
	}

	return h.dao.create(ctx, convertAccessPointFromTypes(accessPoint))
}

// CreateMany creates many accesspoints and return the generated ids.
func (h *handler) CreateMany(ctx context.Context, accessPoints ...*types.AccessPoint) ([]int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	if len(accessPoints) == 0 {
		return nil, errors.New("empty accesspoint")
	}

	data := make([]*AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		if accessPoint == nil {
			return nil, base.ErrInvalidItemInParamList()
		}

		data[idx] = convertAccessPointFromTypes(accessPoint)

		if err = base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return nil, err
		}
	}

	return h.dao.createMany(ctx, data)
}

// UpdateMany updates accesspoint.
func (h *handler) UpdateMany(ctx context.Context, accessPoints ...*types.AccessPoint) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	data := make([]*AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		if accessPoint == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertAccessPointFromTypes(accessPoint)

		if err = base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.dao.updateMany(ctx, tenantID, data); err != nil {
		return err
	}

	return nil
}

// DeleteMany deletes accesspoint by ids.
func (h *handler) DeleteMany(ctx context.Context, accessPointIDs ...int64) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(accessPointIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	if err := h.dao.deleteMany(ctx, tenantID, accessPointIDs...); err != nil {
		return err
	}

	return nil
}

func convertAccessPointFromTypes(accessPoint *types.AccessPoint) *AccessPoint {
	return &AccessPoint{
		TenantID:        accessPoint.TenantID,
		AccessPointID:   accessPoint.ID,
		AccessPointName: accessPoint.Name,
		NetworkAreaID:   accessPoint.NetworkAreaID,
		Endpoints: &Endpoints{
			Cluster: accessPoint.Endpoints.Cluster,
			File:    accessPoint.Endpoints.File,
			Data:    accessPoint.Endpoints.Data,
		},
	}
}

func convertAccessPointToTypes(accessPoint *AccessPoint) *types.AccessPoint {
	endpoints := types.Endpoints{}

	if accessPoint.Endpoints != nil {
		endpoints.Cluster = accessPoint.Endpoints.Cluster
		endpoints.File = accessPoint.Endpoints.File
		endpoints.Data = accessPoint.Endpoints.Data
	}

	return &types.AccessPoint{
		ID:            accessPoint.AccessPointID,
		Name:          accessPoint.AccessPointName,
		TenantID:      accessPoint.TenantID,
		NetworkAreaID: accessPoint.NetworkAreaID,
		Endpoints:     endpoints,
	}
}
