/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package host ...
package host

import (
	"context"
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler host handler interface.
type Handler interface {
	UpsertMany(ctx context.Context, hosts []*types.Host) error
	ListAll(ctx context.Context) ([]*types.Host, error)
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

	d, _ := h.daoMap.LoadOrStore(tenantID, newDao(tenantID, h.client, h.logger))

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New create a new host handler.
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// UpsertMany updates or inserts hosts.
func (h *handler) UpsertMany(ctx context.Context, hosts []*types.Host) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	data := make([]*Host, len(hosts))
	for idx, host := range hosts {
		data[idx] = &Host{
			TenantID:      host.TenantID,
			NetworkAreaID: host.NetworkAreaID,
			BizID:         host.BizID,
			HostID:        host.HostID,
			InnerIP:       host.InnerIP,
			Mac:           host.Mac,
			OSType:        host.OSType,
		}

		if data[idx].TenantID != tenantID {
			return fmt.Errorf("tenantID not match, ctx-tenantID(%s), host-tenantID(%s)", tenantID, host.TenantID)
		}
	}

	if err := h.tenantDao(tenantID).upsertMany(ctx, data); err != nil {
		return err
	}

	return nil
}

// ListAll list all host.
func (h *handler) ListAll(ctx context.Context) ([]*types.Host, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	hosts, err := h.tenantDao(tenantID).listAll(ctx)
	if err != nil {
		return nil, err
	}

	data := make([]*types.Host, len(hosts))
	for idx, host := range hosts {
		data[idx] = &types.Host{
			TenantID:      host.TenantID,
			NetworkAreaID: host.NetworkAreaID,
			BizID:         host.BizID,
			HostID:        host.HostID,
			InnerIP:       host.InnerIP,
			Mac:           host.Mac,
			OSType:        host.OSType,
		}
	}

	return data, nil
}
