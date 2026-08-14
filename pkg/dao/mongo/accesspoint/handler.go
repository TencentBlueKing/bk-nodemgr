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
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler accesspoint handler interface.
type IHandler interface {
	// Count counts accesspoint by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists accesspoint by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.AccessPoint, int64, error)

	// Get gets accesspoint by id.
	Get(nCtx contextx.IContext, accessPointID int64) (*types.AccessPoint, error)

	// Create creates accesspoint.
	Create(nCtx contextx.IContext, accessPoint *types.AccessPoint) (int64, error)

	// CreateMany creates many accesspoint.
	CreateMany(nCtx contextx.IContext, accessPoints ...*types.AccessPoint) ([]int64, error)

	// UpdateMany updates accesspoint.
	UpdateMany(nCtx contextx.IContext, accessPoints ...*types.AccessPoint) error

	// DeleteMany deletes accesspoint.
	DeleteMany(nCtx contextx.IContext, accessPointIDs ...int64) error
}

type handler struct {
	client *mongo.Database
	daoMap sync.Map
}

func (h *handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(tenantID, h.client)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure accesspoint indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// tenantDao is the sole writer of daoMap, so stored values are always *dao.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new accesspoint handler.
func New(client *mongo.Database) IHandler {
	return &handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Count counts accesspoint by conditions.
func (h *handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

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

// List lists accesspoint by page and conditions.
func (h *handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) (
	[]*types.AccessPoint, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	tenantDao := h.tenantDao(tenantID)
	num, err := tenantDao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	accessPoints, err := tenantDao.List(nCtx, filter, findOpt)
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
func (h *handler) Get(nCtx contextx.IContext, accessPointID int64) (*types.AccessPoint, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	if accessPointID < 0 {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithInt64Values(FieldKeyAccessPointID, accessPointID)
	filter = opt(filter)

	data, err := h.tenantDao(tenantID).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertAccessPointToTypes(data), nil
}

// Create creates a new accesspoint and return the generated id.
func (h *handler) Create(nCtx contextx.IContext, accessPoint *types.AccessPoint) (int64, error) {
	if nCtx == nil {
		return -1, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return -1, err
	}

	tenantID := nCtx.TenantID()

	if accessPoint == nil {
		return -1, base.ErrEmptyParamData()
	}

	if err := base.CheckTenantIDMatched(tenantID, accessPoint.TenantID); err != nil {
		return -1, err
	}

	if accessPoint.Name == "" {
		return -1, errors.New("accesspoint name is empty")
	}

	if accessPoint.NetworkAreaID < 0 {
		return -1, errors.New("accesspoint networkarea-id is invalid")
	}

	return h.tenantDao(tenantID).create(nCtx, convertAccessPointFromTypes(accessPoint))
}

// CreateMany creates many accesspoints and return the generated ids.
func (h *handler) CreateMany(nCtx contextx.IContext, accessPoints ...*types.AccessPoint) ([]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	if len(accessPoints) == 0 {
		return nil, errors.New("empty accesspoint")
	}

	data := make([]*AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		if accessPoint == nil {
			return nil, base.ErrInvalidItemInParamList()
		}

		data[idx] = convertAccessPointFromTypes(accessPoint)

		if err := base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return nil, err
		}
	}

	return h.tenantDao(tenantID).createMany(nCtx, data)
}

// UpdateMany updates accesspoint.
func (h *handler) UpdateMany(nCtx contextx.IContext, accessPoints ...*types.AccessPoint) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(accessPoints) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		if accessPoint == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertAccessPointFromTypes(accessPoint)

		if err := base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.tenantDao(tenantID).updateMany(nCtx, data); err != nil {
		return err
	}

	return nil
}

// DeleteMany deletes accesspoint by ids.
func (h *handler) DeleteMany(nCtx contextx.IContext, accessPointIDs ...int64) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	if len(accessPointIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	if err := h.tenantDao(nCtx.TenantID()).deleteMany(nCtx, accessPointIDs...); err != nil {
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
