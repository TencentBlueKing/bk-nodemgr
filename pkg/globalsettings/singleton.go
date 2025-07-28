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
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IGlobalSettings defines the interface for global settings.
type IGlobalSettings interface {
	// List lists global settings by page and condition.
	List(ctx context.Context, page types.Page, condition *types.GlobalSettingsCondition) (
		[]*types.GlobalSettings, int64, error)

	// Count counts global settings by condition.
	Count(ctx context.Context, condition *types.GlobalSettingsCondition) (int64, error)

	// Get gets a global settings by setting name.
	Get(ctx context.Context, settingName string) (string, error)

	// UpsertMany upserts global settings.
	UpsertMany(ctx context.Context, settings ...*types.GlobalSettings) error

	// DeleteMany delete a new global settings.
	DeleteMany(ctx context.Context, settingName ...string) error
}

var (
	// nolint: gochecknoglobals
	instance *Globalsettings
	// nolint: gochecknoglobals
	once sync.Once
)

// InitGlobalSettings initializes the global settings storage.
func InitGlobalSettings(stgGlobalSettings IStorageGlobalSettings) {
	once.Do(func() {
		instance = &Globalsettings{
			stgGlobalSettings: stgGlobalSettings,
		}
	})
}

// GetInstance returns the singleton instance of globalsettings.
func GetInstance() *Globalsettings {
	if instance == nil {
		panic("globalsettings not initialized, please call InitGlobalSettings first")
	}

	return instance
}

// Globalsettings defines the singleton for global settings.
type Globalsettings struct {
	stgGlobalSettings IStorageGlobalSettings
}

// List retrieves global settings with pagination and conditions.
func (gs *Globalsettings) List(
	ctx context.Context, page types.Page, cond *types.GlobalSettingsCondition) ([]*types.GlobalSettings, int64, error) {

	if gs.stgGlobalSettings == nil {
		return nil, 0, errors.New("global settings storage not initialized")
	}

	return gs.stgGlobalSettings.ListGlobalSettings(ctx, page, cond)
}

// Count counts the number of global settings based on conditions.
func (gs *Globalsettings) Count(ctx context.Context, cond *types.GlobalSettingsCondition) (int64, error) {
	if gs.stgGlobalSettings == nil {
		return 0, errors.New("global settings storage not initialized")
	}

	return gs.stgGlobalSettings.CountGlobalSettings(ctx, cond)
}

// Get retrieves a specific global setting by name.
func (gs *Globalsettings) Get(ctx context.Context, name string) (string, error) {
	if gs.stgGlobalSettings == nil {
		return "", errors.New("global settings storage not initialized")
	}

	return gs.stgGlobalSettings.GetGlobalSetting(ctx, name)
}

// UpsertMany updates or inserts many global settings.
func (gs *Globalsettings) UpsertMany(ctx context.Context, settings ...*types.GlobalSettings) error {
	if gs.stgGlobalSettings == nil {
		return errors.New("global settings storage not initialized")
	}

	return gs.stgGlobalSettings.UpsertManyGlobalSettings(ctx, settings...)
}

// DeleteMany deletes many global settings by names.
func (gs *Globalsettings) DeleteMany(ctx context.Context, settingNames ...string) error {
	if gs.stgGlobalSettings == nil {
		return errors.New("global settings storage not initialized")
	}

	return gs.stgGlobalSettings.DeleteManyGlobalSettings(ctx, settingNames...)
}
