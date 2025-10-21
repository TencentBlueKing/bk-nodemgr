/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package globalsettings provides a singleton for global settings.
package globalsettings

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IGlobalSettings defines the interface for global settings.
type IGlobalSettings interface {
	// ListAll list all global settings.
	ListAll(ctx contextx.IContext) ([]*types.GlobalSettings, int64, error)

	// Get gets a global settings by setting name return default value if not exist.
	Get(ctx contextx.IContext, name string, defaultValue string) string

	// Upsert update or insert a global settings.
	Upsert(ctx contextx.IContext, name, value string) error

	// Delete delete global settings.
	Delete(ctx contextx.IContext, names ...string) error
}

// GlobalSettings defines the singleton for global settings.
type GlobalSettings struct {
	stgGlobalSettings IStorage
}

// NewGlobalSettings creates a new instance of GlobalSettings.
func NewGlobalSettings(ctx contextx.IContext, stgGlobalSettings IStorage) (*GlobalSettings, error) {
	gs := &GlobalSettings{stgGlobalSettings: stgGlobalSettings}
	if err := gs.init(ctx); err != nil {
		return nil, err
	}

	return gs, nil
}

// Init initializes the global settings with predefined values.
func (gs *GlobalSettings) init(nCtx contextx.IContext) error {
	for _, setting := range PreDefinition() {
		exist, err := gs.stgGlobalSettings.ExistGlobalSettings(nCtx, setting.SettingName)
		if err != nil {
			return err
		}

		if exist {
			continue
		}

		err = gs.stgGlobalSettings.UpsertGlobalSettings(nCtx, setting)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListAll retrieves global settings with pagination and conditions.
func (gs *GlobalSettings) ListAll(ctx contextx.IContext) ([]*types.GlobalSettings, int64, error) {
	if gs.stgGlobalSettings == nil {
		return nil, 0, ErrUninitialized()
	}

	return gs.stgGlobalSettings.ListGlobalSettings(ctx, types.UnlimitedPage(), nil)
}

// Get retrieves a global setting by its name, returning a default value if not found.
func (gs *GlobalSettings) Get(ctx contextx.IContext, name, defaultValue string) string {
	if gs.stgGlobalSettings == nil {
		return defaultValue
	}

	exist, err := gs.stgGlobalSettings.ExistGlobalSettings(ctx, name)
	if err != nil {
		return defaultValue
	}

	if !exist {
		return defaultValue
	}

	value, err := gs.stgGlobalSettings.GetGlobalSetting(ctx, name)
	if err != nil {
		return defaultValue
	}

	return value
}

// Upsert updates or inserts a global settings.
func (gs *GlobalSettings) Upsert(ctx contextx.IContext, name, value string) error {
	if gs.stgGlobalSettings == nil {
		return ErrUninitialized()
	}

	return gs.stgGlobalSettings.UpsertGlobalSettings(ctx, &types.GlobalSettings{
		SettingName: name,
		Value:       value,
	})
}

// Delete deletes global settings by names.
func (gs *GlobalSettings) Delete(ctx contextx.IContext, names ...string) error {
	if gs.stgGlobalSettings == nil {
		return ErrUninitialized()
	}

	return gs.stgGlobalSettings.DeleteGlobalSettings(ctx, names...)
}
