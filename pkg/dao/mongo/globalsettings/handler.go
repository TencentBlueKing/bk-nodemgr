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
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler global settings handler interface.
type IHandler interface {
	// Get gets global settings by key.
	Get(nCtx contextx.IContext, key string) (*types.GlobalSettings, error)

	// Count counts global settings by opts.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// Exist checks if a global settings exists by key.
	Exist(nCtx contextx.IContext, key string) (bool, error)

	// List lists global settings by page and opts.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.GlobalSettings, int64, error)

	// ListWithoutCount lists global settings by page and opts without count.
	ListWithoutCount(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.GlobalSettings, error)

	// Upsert upserts global settings.
	Upsert(nCtx contextx.IContext, settings ...*types.GlobalSettings) error

	// Delete deletes global settings.
	Delete(nCtx contextx.IContext, name ...string) error
}

// Handler this is a Handler to operate global settings table.
type Handler struct {
	client *mongo.Database
	dao    *dao
}

// New create a new global settings handler.
func New(client *mongo.Database) *Handler {
	d := newDao(client)
	if err := d.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("table-name", TableName()).Warn("failed to ensure tenant indexes")
	}

	return &Handler{
		client: client,
		dao:    d,
	}
}

// Get gets global settings by key.
func (h *Handler) Get(nCtx contextx.IContext, name string) (*types.GlobalSettings, error) {
	filter := base.AliveFilter()
	filter = WithSettingName(name)(filter)

	data, err := h.dao.Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertGlobalSettingsToTypes(data), nil
}

// Count counts global settings by opts.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	count, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Exist checks if a global settings exists by key.
func (h *Handler) Exist(nCtx contextx.IContext, name string) (bool, error) {
	filter := base.AliveFilter()
	filter = WithSettingName(name)(filter)

	exist, err := h.dao.Exist(nCtx, filter)
	if err != nil {
		return false, err
	}

	return exist, nil
}

// List lists global settings by page and opts.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.GlobalSettings, int64, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	datas, err := h.dao.List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	globalsettings := make([]*types.GlobalSettings, len(datas))
	for idx, data := range datas {
		globalsettings[idx] = convertGlobalSettingsToTypes(data)
	}

	return globalsettings, num, nil
}

// ListWithoutCount lists global settings by page and opts without count.
func (h *Handler) ListWithoutCount(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.GlobalSettings, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	findOpt := base.ParsePage(page)

	datas, err := h.dao.List(nCtx, filter, findOpt)
	if err != nil {
		return nil, err
	}

	globalsettings := make([]*types.GlobalSettings, len(datas))
	for idx, data := range datas {
		globalsettings[idx] = convertGlobalSettingsToTypes(data)
	}

	return globalsettings, nil
}

// Upsert upserts global settings.
func (h *Handler) Upsert(nCtx contextx.IContext, settings ...*types.GlobalSettings) error {
	if len(settings) == 0 {
		return errors.New("no global settings to upsert")
	}

	dataList := make([]*GlobalSettings, 0, len(settings))
	for _, s := range settings {
		dataList = append(dataList, convertTypesToGlobalSettings(s))
	}

	if err := h.dao.upsertMany(nCtx, dataList); err != nil {
		return err
	}

	return nil
}

// Delete deletes global settings.
func (h *Handler) Delete(nCtx contextx.IContext, name ...string) error {
	if len(name) == 0 {
		return nil
	}

	filter := base.AliveFilter()
	filter = WithSettingName(name...)(filter)

	if err := h.dao.DeleteMany(nCtx, filter); err != nil {
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
