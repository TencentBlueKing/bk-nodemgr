/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package networkunit

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler networkunit handler interface.
type IHandler interface {
	// Count counts networkunit by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists networkunit by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.NetworkUnit, int64, error)

	// Get gets networkunit by id.
	Get(nCtx contextx.IContext, networkUnitID int64) (*types.NetworkUnit, error)

	// Create creates a networkunit.
	Create(nCtx contextx.IContext, networkUnit *types.NetworkUnit) (int64, error)

	// UpdateMany updates networkunit.
	UpdateMany(nCtx contextx.IContext, networkUnits ...*types.NetworkUnit) error

	// DeleteMany deletes networkunit.
	DeleteMany(nCtx contextx.IContext, networkUnitIDs ...int64) error
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
		logger.G.Sys().WithErr(err).Warn("failed to ensure networkunit indexes")
	}

	return h
}

// Count counts networkunit by conditions.
func (h *handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	tenantID, err := tenant.GetID(nCtx)
	if err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}
	filter = append(filter, tenantFilter(tenantID))

	return h.dao.Count(nCtx, filter)
}

// List lists networkunit by page and conditions.
func (h *handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) (
	[]*types.NetworkUnit, int64, error) {

	tenantID, err := tenant.GetID(nCtx)
	if err != nil {
		return nil, 0, err
	}

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

	networkUnits, err := h.dao.List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.NetworkUnit, len(networkUnits))
	for idx, networkunit := range networkUnits {
		data[idx] = convertNetworkUnitToTypes(networkunit)
	}

	return data, num, nil
}

// Get gets networkunit by id.
func (h *handler) Get(nCtx contextx.IContext, networkUnitID int64) (*types.NetworkUnit, error) {
	tenantID, err := tenant.GetID(nCtx)
	if err != nil {
		return nil, err
	}

	if networkUnitID < 0 {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithInt64Values(FieldKeyNetworkUnitID, networkUnitID)
	filter = opt(filter)
	filter = append(filter, tenantFilter(tenantID))

	data, err := h.dao.Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertNetworkUnitToTypes(data), nil
}

// Create creates a new networkunit and return the generated id.
func (h *handler) Create(nCtx contextx.IContext, networkUnit *types.NetworkUnit) (int64, error) {
	tenantID, err := tenant.GetID(nCtx)
	if err != nil {
		return -1, err
	}

	if networkUnit == nil {
		return -1, base.ErrEmptyParamData()
	}

	if err = base.CheckTenantIDMatched(tenantID, networkUnit.TenantID); err != nil {
		return -1, err
	}

	if networkUnit.Name == "" {
		return -1, errors.New("networkunit name is empty")
	}

	if networkUnit.NetworkAreaID < 0 {
		return -1, errors.New("accesspoint networkarea-id is invalid")
	}

	return h.dao.create(nCtx, convertNetworkUnitFromTypes(networkUnit))
}

// UpdateMany updates networkunit.
func (h *handler) UpdateMany(nCtx contextx.IContext, networkUnits ...*types.NetworkUnit) error {
	tenantID, err := tenant.GetID(nCtx)
	if err != nil {
		return err
	}

	if len(networkUnits) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*NetworkUnit, len(networkUnits))
	for idx, networkUnit := range networkUnits {
		if networkUnit == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = convertNetworkUnitFromTypes(networkUnit)

		if err = base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.dao.updateMany(nCtx, tenantID, data); err != nil {
		return err
	}

	return nil
}

// DeleteMany deletes networkunit by ids.
func (h *handler) DeleteMany(nCtx contextx.IContext, networkUnitIDs ...int64) error {
	tenantID, err := tenant.GetID(nCtx)
	if err != nil {
		return err
	}

	if len(networkUnitIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	if err := h.dao.deleteMany(nCtx, tenantID, networkUnitIDs...); err != nil {
		return err
	}

	return nil
}

func convertNetworkUnitFromTypes(networkUnit *types.NetworkUnit) *NetworkUnit {
	return &NetworkUnit{
		TenantID:        networkUnit.TenantID,
		NetworkUnitID:   networkUnit.ID,
		NetworkUnitName: networkUnit.Name,
		NetworkAreaID:   networkUnit.NetworkAreaID,
		AccessPoints:    networkUnit.AccessPoints,
		Links: &Links{
			Cluster: convertLinksFromTypes(networkUnit.Links.Cluster),
			File:    convertLinksFromTypes(networkUnit.Links.File),
			Data:    convertLinksFromTypes(networkUnit.Links.Data),
		},
		IsDirect:        networkUnit.IsDirect,
		DirectEndpoints: convertEndpointsFromTypes(networkUnit.DirectEndpoints),
	}
}

func convertNetworkUnitToTypes(networkArea *NetworkUnit) *types.NetworkUnit {
	var links types.Links
	if networkArea.Links != nil {
		links = types.Links{
			Cluster: convertLinksToTypes(networkArea.Links.Cluster),
			File:    convertLinksToTypes(networkArea.Links.File),
			Data:    convertLinksToTypes(networkArea.Links.Data),
		}
	}

	return &types.NetworkUnit{
		TenantID:        networkArea.TenantID,
		ID:              networkArea.NetworkUnitID,
		Name:            networkArea.NetworkUnitName,
		NetworkAreaID:   networkArea.NetworkAreaID,
		AccessPoints:    networkArea.AccessPoints,
		Links:           links,
		IsDirect:        networkArea.IsDirect,
		DirectEndpoints: convertEndpointsToTypes(networkArea.DirectEndpoints),
	}
}

func convertLinksFromTypes(link *types.Link) *Link {
	if link == nil {
		return nil
	}

	return &Link{
		NetworkAreaID: link.NetworkAreaID,
		NetworkUnitID: link.NetworkUnitID,
		AccessPointID: link.AccessPointID,
	}
}

func convertLinksToTypes(link *Link) *types.Link {
	if link == nil {
		return nil
	}

	return &types.Link{
		NetworkAreaID: link.NetworkAreaID,
		NetworkUnitID: link.NetworkUnitID,
		AccessPointID: link.AccessPointID,
	}
}

func convertEndpointsFromTypes(endpoint *types.Endpoints) *Endpoints {
	if endpoint == nil {
		return nil
	}

	return &Endpoints{
		Cluster: endpoint.Cluster,
		File:    endpoint.File,
		Data:    endpoint.Data,
	}
}

func convertEndpointsToTypes(endpoint *Endpoints) *types.Endpoints {
	if endpoint == nil {
		return nil
	}

	return &types.Endpoints{
		Cluster: endpoint.Cluster,
		File:    endpoint.File,
		Data:    endpoint.Data,
	}
}
