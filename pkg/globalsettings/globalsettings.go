/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package globalsettings provides a singleton for global settings.
package globalsettings

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Handler defines the singleton for global settings.
type Handler struct {
	stg IStorage
}

// NewHandler creates a new instance of Handler.
func NewHandler(nCtx contextx.IContext, stg IStorage) (*Handler, error) {
	gs := &Handler{stg: stg}
	if err := gs.registerPreDefinition(nCtx); err != nil {
		return nil, err
	}

	return gs, nil
}

// Init initializes the global settings with predefined values.
func (h *Handler) registerPreDefinition(nCtx contextx.IContext) error {
	for _, setting := range PreDefinition() {
		exist, err := h.stg.ExistGlobalSettings(nCtx, setting.SettingName)
		if err != nil {
			return err
		}

		if exist {
			continue
		}

		err = h.stg.UpsertGlobalSettings(nCtx, setting)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListAll retrieves global settings with pagination and conditions.
func (h *Handler) ListAll(nCtx contextx.IContext) ([]*types.GlobalSettings, int64, error) {
	if h.stg == nil {
		return nil, 0, ErrUninitialized()
	}

	return h.stg.ListGlobalSettings(nCtx, types.UnlimitedPage(), nil)
}

// Get retrieves a global setting by its name, returning a default value if not found.
func (h *Handler) Get(nCtx contextx.IContext, name, defaultValue string) string {
	if h.stg == nil {
		return defaultValue
	}

	exist, err := h.stg.ExistGlobalSettings(nCtx, name)
	if err != nil {
		return defaultValue
	}

	if !exist {
		return defaultValue
	}

	value, err := h.stg.GetGlobalSetting(nCtx, name)
	if err != nil {
		return defaultValue
	}

	return value
}

// Upsert updates or inserts a global settings.
func (h *Handler) Upsert(nCtx contextx.IContext, name, value string) error {
	if h.stg == nil {
		return ErrUninitialized()
	}

	return h.stg.UpsertGlobalSettings(nCtx, &types.GlobalSettings{
		SettingName: name,
		Value:       value,
	})
}

// Delete deletes global settings by names.
func (h *Handler) Delete(nCtx contextx.IContext, names ...string) error {
	if h.stg == nil {
		return ErrUninitialized()
	}

	return h.stg.DeleteGlobalSettings(nCtx, names...)
}
