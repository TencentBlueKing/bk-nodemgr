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

// Handler networkarea handler interface.
type Handler interface {
	// Count counts networkarea by conditions.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// List lists networkarea by page and conditions.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.NetworkArea, int64, error)

	// Get gets networkarea by id.
	Get(ctx context.Context, networkAreaID int64) (*types.NetworkArea, error)

	// UpsertMany updates or inserts networkarea.
	UpsertMany(ctx context.Context, networkAreas ...*types.NetworkArea) error

	// UpdateMany updates networkarea.
	UpdateMany(ctx context.Context, networkAreas ...*types.NetworkArea) error

	// DeleteMany deletes networkarea by ids.
	DeleteMany(ctx context.Context, networkAreaIDs ...int64) error
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
		h.logger.Warnf("failed to ensure networkarea indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New create a new networkarea handler.
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// Count counts networkarea by conditions.
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

// List lists networkarea by page and conditions.
func (h *handler) List(ctx context.Context, page types.Page, opts ...OptFn) (
	[]*types.NetworkArea, int64, error) {

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

	networkAreas, err := h.tenantDao(ctx, tenantID).list(ctx, filter, findOpt)
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
func (h *handler) Get(ctx context.Context, networkAreaID int64) (*types.NetworkArea, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	opt := base.WithInt64Values("data.networkarea_id", networkAreaID)
	filter = opt(filter)

	data, err := h.tenantDao(ctx, tenantID).get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return convertNetworkAreaToTypes(data), nil
}

// UpsertMany updates or inserts networkarea.
func (h *handler) UpsertMany(ctx context.Context, networkAreas ...*types.NetworkArea) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	data := make([]*NetworkArea, len(networkAreas))
	for idx, networkArea := range networkAreas {
		data[idx] = convertNetworkAreaFromTypes(networkArea)

		if data[idx].TenantID != tenantID {
			return fmt.Errorf("tenantID not match, ctx-tenantID(%s), networkarea-tenantID(%s)", tenantID, networkArea.TenantID)
		}
	}

	if err := h.tenantDao(ctx, tenantID).upsertMany(ctx, data); err != nil {
		return err
	}

	return nil
}

// UpdateMany updates networkarea.
func (h *handler) UpdateMany(ctx context.Context, networkAreas ...*types.NetworkArea) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	data := make([]*NetworkArea, len(networkAreas))
	for idx, networkArea := range networkAreas {
		data[idx] = convertNetworkAreaFromTypes(networkArea)

		if data[idx].TenantID != tenantID {
			return fmt.Errorf("tenantID not match, ctx-tenantID(%s), networkarea-tenantID(%s)", tenantID, networkArea.TenantID)
		}
	}

	if err := h.tenantDao(ctx, tenantID).updateMany(ctx, data); err != nil {
		return err
	}

	return nil
}

// DeleteMany deletes networkarea by ids.
func (h *handler) DeleteMany(ctx context.Context, networkAreaIDs ...int64) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(networkAreaIDs) == 0 {
		return errors.New("not networkarea-id specified")
	}

	if err := h.tenantDao(ctx, tenantID).deleteMany(ctx, networkAreaIDs...); err != nil {
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
