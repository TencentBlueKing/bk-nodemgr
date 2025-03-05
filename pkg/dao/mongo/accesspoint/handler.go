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
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Handler accesspoint handler interface.
type Handler interface {
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
	client *mongo.Database
	logger logger.Logger
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *handler) tenantDao(ctx context.Context, tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao)
	}

	newDaoClient := newDao(tenantID, h.client, h.logger)
	if err := newDaoClient.ensureIndexes(ctx); err != nil {
		h.logger.Warnf("failed to ensure accesspoint indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New create a new accesspoint handler.
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
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

	return h.tenantDao(ctx, tenantID).count(ctx, filter)
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

	num, err := h.tenantDao(ctx, tenantID).count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := new(options.FindOptions)
	if page.Offset > 0 {
		findOpt.SetSkip(int64(page.Offset))
	}
	if page.Limit > 0 {
		findOpt.SetLimit(int64(page.Limit))
	}

	accessPoints, err := h.tenantDao(ctx, tenantID).list(ctx, filter, findOpt)
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

	filter := base.AliveFilter()
	opt := base.WithInt64Values("data.access_point_id", accessPointID)
	filter = opt(filter)

	data, err := h.tenantDao(ctx, tenantID).get(ctx, filter)
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
		return -1, errors.New("accesspoint is empty")
	}

	if accessPoint.TenantID != tenantID {
		return -1, fmt.Errorf("tenantID not match, ctx-tenantID(%s), accesspoint-tenantID(%s)",
			tenantID, accessPoint.TenantID)
	}

	if accessPoint.Name == "" {
		return -1, errors.New("accesspoint name is empty")
	}

	if accessPoint.NetworkAreaID < 0 {
		return -1, errors.New("accesspoint networkarea-id is invalid")
	}

	return h.tenantDao(ctx, accessPoint.TenantID).create(ctx, convertAccessPointFromTypes(accessPoint))
}

// CreateMany creates many accesspoints and return the generated ids.
func (h *handler) CreateMany(ctx context.Context, accessPoints ...*types.AccessPoint) ([]int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	data := make([]*AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		data[idx] = convertAccessPointFromTypes(accessPoint)

		if data[idx].TenantID != tenantID {
			return nil, fmt.Errorf("tenantID not match, ctx-tenantID(%s), accesspoint-tenantID(%s)", tenantID, accessPoint.TenantID)
		}
	}

	return h.tenantDao(ctx, tenantID).createMany(ctx, data)
}

// UpdateMany updates accesspoint.
func (h *handler) UpdateMany(ctx context.Context, accessPoints ...*types.AccessPoint) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	data := make([]*AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		data[idx] = convertAccessPointFromTypes(accessPoint)

		if data[idx].TenantID != tenantID {
			return fmt.Errorf("tenantID not match, ctx-tenantID(%s), accesspoint-tenantID(%s)", tenantID, accessPoint.TenantID)
		}
	}

	if err := h.tenantDao(ctx, tenantID).updateMany(ctx, data); err != nil {
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
		return errors.New("not accesspoint-id specified")
	}

	if err := h.tenantDao(ctx, tenantID).deleteMany(ctx, accessPointIDs...); err != nil {
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
