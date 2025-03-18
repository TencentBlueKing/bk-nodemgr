/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodedeployment ...
package nodedeployment

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/base"
)

// IStorage defines the storage interface.
type IStorage interface {
	base.Interface

	// GetGseAgentSetting get gse agent setting.
	GetGseAgentSetting(ctx context.Context, token string) (map[string]any, map[string]any, error)

	// GetGseFileProxySetting get gse file proxy setting.
	GetGseFileProxySetting(ctx context.Context, token string) (map[string]any, map[string]any, error)

	// GetGseDataProxySetting get gse data proxy setting.
	GetGseDataProxySetting(ctx context.Context, token string) (map[string]any, map[string]any, error)

	// GetCheckListSetting get check list setting.
	GetCheckListSetting(ctx context.Context, token string) (map[string]any, map[string]any, error)
}
