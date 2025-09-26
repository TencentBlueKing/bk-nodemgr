/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package tenant ...
package tenant

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler tenant handler interface.
type Handler interface {
	Upsert(nCtx contextx.IContext, tenant *types.Tenant) error
	ListAll(nCtx contextx.IContext) ([]*types.Tenant, error)
}

type handler struct {
	dao *dao
}

// New create a new host handler.
func New(client *mongo.Database) Handler {
	return &handler{
		dao: newDao(client),
	}
}

// Upsert updates or inserts a host.
func (h *handler) Upsert(nCtx contextx.IContext, tenant *types.Tenant) error {
	data := &Tenant{
		ID:     tenant.ID,
		Name:   tenant.Name,
		Status: tenant.Status,
	}
	if err := h.dao.upsert(nCtx, data); err != nil {
		return err
	}

	return nil
}

// ListAll list all host.
func (h *handler) ListAll(nCtx contextx.IContext) ([]*types.Tenant, error) {
	tenants, err := h.dao.listAll(nCtx)
	if err != nil {
		return nil, err
	}

	data := make([]*types.Tenant, len(tenants))
	for idx, tenant := range tenants {
		data[idx] = &types.Tenant{
			ID:     tenant.ID,
			Name:   tenant.Name,
			Status: tenant.Status,
		}
	}

	return data, nil
}
