/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed on the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tenant

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler tenant handler interface.
type IHandler interface {
	// Create create a new tenant.
	Create(nCtx contextx.IContext, tenant *types.Tenant) error

	// Count count tenants by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists tenants by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Tenant, int64, error)

	// Exist check a tenant exist by conditions.
	Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error)

	// CreateMany create many tenants.
	CreateMany(nCtx contextx.IContext, tenants ...*types.Tenant) error

	// UpdateMany update many tenants.
	UpdateMany(nCtx contextx.IContext, tenantMap map[string]*types.Tenant) error

	// DeleteMany delete many tenants.
	DeleteMany(nCtx contextx.IContext, ids ...string) error
}

var _ IHandler = &Handler{}

// Handler this is a Handler to operate tenant table.
type Handler struct {
	client *mongo.Database
	dao    *dao
}

// New create a new tenant handler.
func New(client *mongo.Database) *Handler {
	d := newDao(client, TableName())
	if err := d.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("table-name", TableName()).Warn("failed to ensure tenant indexes")
	}

	return &Handler{
		client: client,
		dao:    d,
	}
}

// Create create a new tenant.
func (h *Handler) Create(nCtx contextx.IContext, tenant *types.Tenant) error {
	data := convTenantFromTypes(tenant)

	if err := h.dao.Create(nCtx, data); err != nil {
		return fmt.Errorf("failed to create tenant, err: %w", err)
	}

	return nil
}

func convTenantFromTypes(tenant *types.Tenant) *Tenant {
	data := &Tenant{
		ID:      tenant.ID,
		Name:    tenant.Name,
		Enabled: tenant.Enabled,
	}

	return data
}

// Count count tenant by conditions.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.dao.Count(nCtx, filter)
}

// List list tenant by page and conditions.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Tenant, int64, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	data, err := h.dao.List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	tenants := make([]*types.Tenant, len(data))
	for idx, tenant := range data {
		tenants[idx] = convertTenantToTypes(tenant)
	}

	return tenants, num, nil
}

// Exist check a tenant exist by conditions.
func (h *Handler) Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	exist, err := h.dao.Exist(nCtx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check tenant exist, err: %w", err)
	}

	return exist, nil
}

func convertTenantToTypes(data *Tenant) *types.Tenant {
	tenant := &types.Tenant{
		ID:      data.ID,
		Name:    data.Name,
		Enabled: data.Enabled,
	}

	return tenant
}

// CreateMany create many tenant.
func (h *Handler) CreateMany(nCtx contextx.IContext, tenants ...*types.Tenant) error {
	datas := make([]*Tenant, len(tenants))
	for idx, tenant := range tenants {
		datas[idx] = convTenantFromTypes(tenant)
	}

	if err := h.dao.CreateMany(nCtx, datas); err != nil {
		return fmt.Errorf("failed to create many tenant: %w", err)
	}

	return nil
}

// DeleteMany delete many tenant.
func (h *Handler) DeleteMany(nCtx contextx.IContext, ids ...string) error {
	if len(ids) == 0 {
		return base.ErrEmptyParamData()
	}

	filter := base.AliveFilter()
	filter = WithID(ids...)(filter)
	if err := h.dao.DeleteMany(nCtx, filter); err != nil {
		return fmt.Errorf("failed to delete many tenant: %w", err)
	}

	return nil
}

// UpdateMany update many tenant.
func (h *Handler) UpdateMany(nCtx contextx.IContext, tenantMap map[string]*types.Tenant) error {
	if len(tenantMap) == 0 {
		return base.ErrEmptyParamData()
	}
	updates := make([]*base.DocumentFieldUpdate, 0, len(tenantMap))

	for tenantID, tenant := range tenantMap {
		filter := base.AliveFilter()
		opts := []base.OptFn{
			WithID(tenantID),
		}

		for _, opt := range opts {
			filter = opt(filter)
		}

		data := convTenantFromTypes(tenant)

		update := &base.DocumentFieldUpdate{
			Filter: filter,
			Fields: map[string]any{
				FieldKeyName:    data.Name,
				FieldKeyEnabled: data.Enabled,
			},
		}

		updates = append(updates, update)
	}

	if err := h.dao.UpdateFieldsBulk(nCtx, updates); err != nil {
		return fmt.Errorf("failed to update many tenant: %w", err)
	}

	return nil
}
