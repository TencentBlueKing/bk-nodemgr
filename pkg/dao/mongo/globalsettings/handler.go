/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package globalsettings provides storage for global settings.
package globalsettings

import (
	"context"
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler global settings handler interface.
type IHandler interface {
	// Get gets global settings by key.
	Get(ctx context.Context, key string) (*types.GlobalSettings, error)

	// Count counts global settings by opts.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// Exist checks if a global settings exists by key.
	Exist(ctx context.Context, key string) (bool, error)

	// List lists global settings by page and opts.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.GlobalSettings, int64, error)

	// Upsert upserts global settings.
	Upsert(ctx context.Context, settings ...*types.GlobalSettings) error

	// Delete deletes global settings.
	Delete(ctx context.Context, name ...string) error
}

// Handler this is a Handler to operate global settings table.
type Handler struct {
	client *mongo.Database
	logger logger.ILogger

	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(tenantID, h.client, h.logger)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure global settings indexes: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new global settings handler.
func New(client *mongo.Database, logger logger.ILogger) *Handler {
	return &Handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// Get gets global settings by key.
func (h *Handler) Get(ctx context.Context, name string) (*types.GlobalSettings, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	filter = WithSettingName(name)(filter)

	data, err := h.tenantDao(tenantID).Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return convertGlobalSettingsToTypes(data), nil
}

// Count counts global settings by opts.
func (h *Handler) Count(ctx context.Context, opts ...OptFn) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	count, err := h.tenantDao(tenantID).Count(ctx, filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Exist checks if a global settings exists by key.
func (h *Handler) Exist(ctx context.Context, name string) (bool, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return false, err
	}

	filter := base.AliveFilter()
	filter = WithSettingName(name)(filter)

	exist, err := h.tenantDao(tenantID).Exist(ctx, filter)
	if err != nil {
		return false, err
	}

	return exist, nil
}

// List lists global settings by page and opts.
func (h *Handler) List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.GlobalSettings, int64, error) {
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

	datas, err := h.tenantDao(tenantID).List(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	globalsettings := make([]*types.GlobalSettings, len(datas))
	for idx, data := range datas {
		globalsettings[idx] = convertGlobalSettingsToTypes(data)
	}

	return globalsettings, num, nil
}

// Upsert upserts global settings.
func (h *Handler) Upsert(ctx context.Context, settings ...*types.GlobalSettings) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(settings) == 0 {
		return errors.New("no global settings to upsert")
	}

	dataList := make([]*GlobalSettings, 0, len(settings))
	for _, s := range settings {
		dataList = append(dataList, convertTypesToGlobalSettings(s))
	}

	if err := h.tenantDao(tenantID).upsertMany(ctx, dataList); err != nil {
		return err
	}

	return nil
}

// Delete deletes global settings.
func (h *Handler) Delete(ctx context.Context, name ...string) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(name) == 0 {
		return nil
	}

	filter := base.AliveFilter()
	filter = WithSettingName(name...)(filter)

	if err := h.tenantDao(tenantID).DeleteMany(ctx, filter); err != nil {
		return err
	}

	return nil
}

// convertGlobalSettingsToTypes converts GlobalSettings to types.GlobalSettings.
func convertGlobalSettingsToTypes(data *GlobalSettings) *types.GlobalSettings {
	if data == nil {
		return nil
	}

	return &types.GlobalSettings{
		SettingName: data.SettingName,
		Value:       data.Value,
	}
}

// convertTypesToGlobalSettings converts types.GlobalSettings to GlobalSettings.
func convertTypesToGlobalSettings(data *types.GlobalSettings) *GlobalSettings {
	if data == nil {
		return nil
	}

	return &GlobalSettings{
		SettingName: data.SettingName,
		Value:       data.Value,
	}
}
