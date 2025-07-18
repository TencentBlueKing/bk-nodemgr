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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the interface of global settings storage.
type IStorage interface {
	basestorage.Interface

	// ListGlobalSettings lists global settings by page and condition.
	ListGlobalSettings(ctx context.Context, page types.Page, condition *types.GlobalSettingsCondition) (
		[]*types.GlobalSettings, int64, error)

	// CountGlobalSettings counts global settings by condition.
	CountGlobalSettings(ctx context.Context, condition *types.GlobalSettingsCondition) (int64, error)

	// GetGlobalSetting gets a global settings by setting name.
	GetGlobalSetting(ctx context.Context, settingName string) (string, error)

	// UpsertGlobalSettings upserts global settings.
	UpsertManyGlobalSettings(ctx context.Context, settings ...*types.GlobalSettings) error

	// DeleteGlobalSettings delete a new global settings.
	DeleteManyGlobalSettings(ctx context.Context, settingName ...string) error
}
