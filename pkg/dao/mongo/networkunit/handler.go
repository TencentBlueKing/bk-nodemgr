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

// Handler networkunit handler interface.
type Handler interface {
	// Count counts networkunit by conditions.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// List lists networkunit by page and conditions.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.NetworkUnit, int64, error)

	// Get gets networkunit by id.
	Get(ctx context.Context, networkUnitID int64) (*types.NetworkUnit, error)

	// Create creates a networkunit.
	Create(ctx context.Context, networkUnit *types.NetworkUnit) (int64, error)

	// UpdateMany updates networkunit.
	UpdateMany(ctx context.Context, networkUnits ...*types.NetworkUnit) error

	// DeleteMany deletes networkunit.
	DeleteMany(ctx context.Context, networkUnitIDs ...int64) error
}

type handler struct {
	client *mongo.Database
	logger logger.Logger
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
		h.logger.Warnf("failed to ensure networkunit indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New create a new networkunit handler.
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

	return h.tenantDao(tenantID).count(ctx, filter)
}

// List lists networkunit by page and conditions.
func (h *handler) List(ctx context.Context, page types.Page, opts ...OptFn) (
	[]*types.NetworkUnit, int64, error) {

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

	findOpt := new(options.FindOptions)
	if page.Offset > 0 {
		findOpt.SetSkip(int64(page.Offset))
	}
	if page.Limit > 0 {
		findOpt.SetLimit(int64(page.Limit))
	}

	networkUnits, err := h.tenantDao(tenantID).list(ctx, filter, findOpt)
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
func (h *handler) Get(ctx context.Context, networkUnitID int64) (*types.NetworkUnit, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	opt := base.WithInt64Values("data.networkunit_id", networkUnitID)
	filter = opt(filter)

	data, err := h.tenantDao(tenantID).get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return convertNetworkUnitToTypes(data), nil
}

// Create creates a new networkunit and return the generated id.
func (h *handler) Create(ctx context.Context, networkUnit *types.NetworkUnit) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return -1, err
	}

	if networkUnit == nil {
		return -1, errors.New("networkunit is empty")
	}

	if networkUnit.TenantID != tenantID {
		return -1, fmt.Errorf("tenantID not match, ctx-tenantID(%s), networkunit-tenantID(%s)",
			tenantID, networkUnit.TenantID)
	}

	if networkUnit.Name == "" {
		return -1, errors.New("networkunit name is empty")
	}

	if networkUnit.NetworkAreaID < 0 {
		return -1, errors.New("accesspoint networkarea-id is invalid")
	}

	return h.tenantDao(networkUnit.TenantID).create(ctx, convertNetworkUnitFromTypes(networkUnit))
}

// UpdateMany updates networkunit.
func (h *handler) UpdateMany(ctx context.Context, networkUnits ...*types.NetworkUnit) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	data := make([]*NetworkUnit, len(networkUnits))
	for idx, networkUnit := range networkUnits {
		data[idx] = convertNetworkUnitFromTypes(networkUnit)

		if data[idx].TenantID != tenantID {
			return fmt.Errorf("tenantID not match, ctx-tenantID(%s), networkunit-tenantID(%s)", tenantID, networkUnit.TenantID)
		}
	}

	if err := h.tenantDao(tenantID).updateMany(ctx, data); err != nil {
		return err
	}

	return nil
}

// DeleteMany deletes networkunit by ids.
func (h *handler) DeleteMany(ctx context.Context, networkUnitIDs ...int64) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(networkUnitIDs) == 0 {
		return errors.New("not networkunit-id specified")
	}

	if err := h.tenantDao(tenantID).deleteMany(ctx, networkUnitIDs...); err != nil {
		return err
	}

	return nil
}

func convertNetworkUnitFromTypes(networkUnit *types.NetworkUnit) *NetworkUnit {
	links := &Links{}

	if networkUnit.Links.Cluster != nil {
		links.Cluster = &Link{
			AccessPointID: networkUnit.Links.Cluster.AccessPointID,
		}
	}

	if networkUnit.Links.File != nil {
		links.File = &Link{
			AccessPointID: networkUnit.Links.File.AccessPointID,
		}
	}

	if networkUnit.Links.Data != nil {
		links.Data = &Link{
			AccessPointID: networkUnit.Links.Data.AccessPointID,
		}
	}

	return &NetworkUnit{
		TenantID:        networkUnit.TenantID,
		NetworkUnitID:   networkUnit.ID,
		NetworkUnitName: networkUnit.Name,
		NetworkAreaID:   networkUnit.NetworkAreaID,
		AccessPoints:    networkUnit.AccessPoints,
		Links:           links,
	}
}

func convertNetworkUnitToTypes(networkArea *NetworkUnit) *types.NetworkUnit {
	links := types.Links{}

	if networkArea.Links != nil {
		if networkArea.Links.Cluster != nil {
			links.Cluster = &types.Link{
				AccessPointID: networkArea.Links.Cluster.AccessPointID,
			}
		}

		if networkArea.Links.File != nil {
			links.File = &types.Link{
				AccessPointID: networkArea.Links.File.AccessPointID,
			}
		}

		if networkArea.Links.Data != nil {
			links.Data = &types.Link{
				AccessPointID: networkArea.Links.Data.AccessPointID,
			}
		}
	}

	return &types.NetworkUnit{
		TenantID:      networkArea.TenantID,
		ID:            networkArea.NetworkUnitID,
		Name:          networkArea.NetworkUnitName,
		NetworkAreaID: networkArea.NetworkAreaID,
		AccessPoints:  networkArea.AccessPoints,
		Links:         links,
	}
}
