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
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IGlobalSettings defines the interface for global settings.
type IGlobalSettings interface {
	// Init initializes the global settings with predefined values.
	Init(ctx context.Context) error

	// ListAll list all global settings.
	ListAll(ctx context.Context) ([]*types.GlobalSettings, int64, error)

	// Get gets a global settings by setting name.
	Get(ctx context.Context, name string) (string, error)

	// Upsert update or insert a global settings.
	Upsert(ctx context.Context, name, value string) error

	// Delete delete global settings.
	Delete(ctx context.Context, names ...string) error
}

// GlobalSettings defines the singleton for global settings.
type GlobalSettings struct {
	stgGlobalSettings IStorage
}

// NewGlobalSettings creates a new instance of GlobalSettings.
func NewGlobalSettings(ctx context.Context, stgGlobalSettings IStorage) (*GlobalSettings, error) {
	gs := &GlobalSettings{stgGlobalSettings: stgGlobalSettings}
	if err := gs.init(ctx); err != nil {
		return nil, err
	}

	return gs, nil
}

// Init initializes the global settings with predefined values.
func (gs *GlobalSettings) init(ctx context.Context) error {
	for _, setting := range PreDefinition() {
		exist, err := gs.stgGlobalSettings.ExistGlobalSettings(ctx, setting.SettingName)
		if err != nil {
			return err
		}

		if exist {
			continue
		}

		err = gs.stgGlobalSettings.UpsertGlobalSettings(ctx, setting)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListAll retrieves global settings with pagination and conditions.
func (gs *GlobalSettings) ListAll(ctx context.Context) ([]*types.GlobalSettings, int64, error) {
	if gs.stgGlobalSettings == nil {
		return nil, 0, ErrUninitialized()
	}

	return gs.stgGlobalSettings.ListGlobalSettings(ctx, types.UnlimitedPage(), nil)
}

// Get retrieves a global setting by its name, returning a default value if not found.
func (gs *GlobalSettings) Get(ctx context.Context, name string) (string, error) {
	if gs.stgGlobalSettings == nil {
		return "", ErrUninitialized()
	}

	exist, err := gs.stgGlobalSettings.ExistGlobalSettings(ctx, name)
	if err != nil {
		return "", err
	}

	if !exist {
		return "", ErrNonexist()
	}

	return gs.stgGlobalSettings.GetGlobalSetting(ctx, name)
}

// Upsert updates or inserts a global settings.
func (gs *GlobalSettings) Upsert(ctx context.Context, name, value string) error {
	if gs.stgGlobalSettings == nil {
		return ErrUninitialized()
	}

	return gs.stgGlobalSettings.UpsertGlobalSettings(ctx, &types.GlobalSettings{
		SettingName: name,
		Value:       value,
	})
}

// Delete deletes global settings by names.
func (gs *GlobalSettings) Delete(ctx context.Context, names ...string) error {
	if gs.stgGlobalSettings == nil {
		return ErrUninitialized()
	}

	return gs.stgGlobalSettings.DeleteGlobalSettings(ctx, names...)
}
