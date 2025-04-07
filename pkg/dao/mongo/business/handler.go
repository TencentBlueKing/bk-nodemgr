/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package business provides business dao operations.
package business

import (
	"context"
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler business Handler interface.
type IHandler interface {
	// Count counts business by opts.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// List lists business by page and opts.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.Business, int64, error)

	// UpsertMany updates or inserts business.
	UpsertMany(ctx context.Context, bizs ...*types.Business) error
}

// Handler business Handler.
type Handler struct {
	client *mongo.Database
	logger logger.Logger
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao)
	}

	newDaoClient := newDao(tenantID, h.client, h.logger)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure business indexes, err: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New create a new business Handler.
func New(client *mongo.Database, logger logger.Logger) *Handler {
	return &Handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// Count counts business by opts.
func (h *Handler) Count(ctx context.Context, opts ...OptFn) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).Count(ctx, filter)
}

// List list business by page and conditions.
func (h *Handler) List(ctx context.Context, page types.Page, opts ...OptFn) (
	[]*types.Business, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	bizs, err := h.tenantDao(tenantID).List(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.Business, len(bizs))
	for idx, biz := range bizs {
		data[idx] = &types.Business{
			TenantID: biz.TenantID,
			BizID:    biz.BizID,
			BizName:  biz.BizName,
		}
	}

	return data, num, nil
}

// UpsertMany updates or inserts business.
func (h *Handler) UpsertMany(ctx context.Context, bizs ...*types.Business) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(bizs) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*Business, len(bizs))
	for idx, biz := range bizs {
		if biz == nil {
			return base.ErrInvalidItemInParamList()
		}

		data[idx] = &Business{
			TenantID: biz.TenantID,
			BizID:    biz.BizID,
			BizName:  biz.BizName,
		}

		if err = base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.tenantDao(tenantID).upsertMany(ctx, data); err != nil {
		return err
	}

	return nil
}
