/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package networkarea

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler networkarea handler interface.
type IHandler interface {
	// Count counts networkarea by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists networkarea by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.NetworkArea, int64, error)

	// Get gets networkarea by id.
	Get(nCtx contextx.IContext, networkAreaID int64) (*types.NetworkArea, error)

	// UpsertMany updates or inserts networkarea.
	UpsertMany(nCtx contextx.IContext, networkAreas ...*types.NetworkArea) error

	// UpdateMany updates networkarea.
	UpdateMany(nCtx contextx.IContext, networkAreas ...*types.NetworkArea) error

	// DeleteMany deletes networkarea by ids.
	DeleteMany(nCtx contextx.IContext, networkAreaIDs ...int64) error
}

type handler struct {
	dao *dao
}

// New create a new networkarea handler.
func New(client *mongo.Database) IHandler {
	h := &handler{
		dao: newDao(client),
	}

	if err := h.dao.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure networkarea indexes")
	}

	return h
}

// Count counts networkarea by conditions.
func (h *handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}
	filter = append(filter, tenantFilter(tenantID))

	return h.dao.Count(nCtx, filter)
}

// List lists networkarea by page and conditions.
func (h *handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) (
	[]*types.NetworkArea, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}
	filter = append(filter, tenantFilter(tenantID))

	num, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)
	if findOpt.Sort == nil {
		findOpt.SetSort(bson.D{bson.E{Key: FieldKeyNetworkAreaID, Value: -1}})
	}

	networkAreas, err := h.dao.List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.NetworkArea, len(networkAreas))
	for idx, networkarea := range networkAreas {
		data[idx] = convertNetworkAreaToTypes(networkarea)
	}

	return data, num, nil
}

// Get gets networkarea by id.
func (h *handler) Get(nCtx contextx.IContext, networkAreaID int64) (*types.NetworkArea, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	if networkAreaID < 0 {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithInt64Values(FieldKeyNetworkAreaID, networkAreaID)
	filter = opt(filter)
	filter = append(filter, tenantFilter(tenantID))

	data, err := h.dao.Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertNetworkAreaToTypes(data), nil
}

// UpsertMany updates or inserts networkarea.
func (h *handler) UpsertMany(nCtx contextx.IContext, networkAreas ...*types.NetworkArea) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(networkAreas) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*NetworkArea, len(networkAreas))
	for idx, networkArea := range networkAreas {
		if networkArea == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertNetworkAreaFromTypes(networkArea)

		if err := base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.dao.upsertMany(nCtx, tenantID, data); err != nil {
		return err
	}

	return nil
}

// UpdateMany updates networkarea.
func (h *handler) UpdateMany(nCtx contextx.IContext, networkAreas ...*types.NetworkArea) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(networkAreas) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*NetworkArea, len(networkAreas))
	for idx, networkArea := range networkAreas {
		if networkArea == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertNetworkAreaFromTypes(networkArea)

		if err := base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.dao.updateMany(nCtx, tenantID, data); err != nil {
		return err
	}

	return nil
}

// DeleteMany deletes networkarea by ids.
func (h *handler) DeleteMany(nCtx contextx.IContext, networkAreaIDs ...int64) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(networkAreaIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	if err := h.dao.deleteMany(nCtx, tenantID, networkAreaIDs...); err != nil {
		return err
	}

	return nil
}

func convertNetworkAreaFromTypes(networkArea *types.NetworkArea) *NetworkArea {
	return &NetworkArea{
		TenantID:        networkArea.TenantID,
		NetworkAreaID:   networkArea.ID,
		NetworkAreaName: networkArea.Name,
		CloudVendor:     networkArea.CloudVendor,
	}
}

func convertNetworkAreaToTypes(networkArea *NetworkArea) *types.NetworkArea {
	return &types.NetworkArea{
		TenantID:    networkArea.TenantID,
		ID:          networkArea.NetworkAreaID,
		Name:        networkArea.NetworkAreaName,
		CloudVendor: networkArea.CloudVendor,
	}
}
