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
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler business Handler interface.
type IHandler interface {
	// Count counts business by opts.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists business by page and opts.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Business, int64, error)

	// ListWithoutCount lists business by page and opts, without count.
	ListWithoutCount(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Business, error)

	// UpsertMany updates or inserts business.
	UpsertMany(nCtx contextx.IContext, bizs ...*types.Business) error
}

// Handler business Handler.
type Handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		// nolint: forcetypeassert
		return d.(*dao)
	}

	newDaoClient := newDao(tenantID, h.client)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure business indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	// nolint: forcetypeassert
	return d.(*dao)
}

// New create a new business Handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Count counts business by opts.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
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

// List list business by page and conditions.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) (
	[]*types.Business, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	bizs, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt)
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

// ListWithoutCount lists business by page and opts, without count.
func (h *Handler) ListWithoutCount(nCtx contextx.IContext, page types.Page, opts ...OptFn) (
	[]*types.Business, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	findOpt := base.ParsePage(page)

	bizs, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	data := make([]*types.Business, len(bizs))
	for idx, biz := range bizs {
		data[idx] = &types.Business{
			TenantID: biz.TenantID,
			BizID:    biz.BizID,
			BizName:  biz.BizName,
		}
	}

	return data, nil
}

// UpsertMany updates or inserts business.
func (h *Handler) UpsertMany(nCtx contextx.IContext, bizs ...*types.Business) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

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

		if err := base.CheckTenantIDMatched(tenantID, data[idx].TenantID); err != nil {
			return err
		}
	}

	if err := h.tenantDao(tenantID).upsertMany(nCtx, data); err != nil {
		return err
	}

	return nil
}
