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

package backend

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerGlobalSettings provides read-only access to shared stored global settings.
type IHandlerGlobalSettings interface {
	ListGlobalSettings(ctx contextx.IContext, page types.Page, condition *types.GlobalSettingsCondition) (
		[]*types.GlobalSettings, int64, error)
	CountGlobalSettings(ctx contextx.IContext, condition *types.GlobalSettingsCondition) (int64, error)
}

// ListGlobalSettings reads one page of stored raw values, including custom settings.
func (h *Handler) ListGlobalSettings(ctx contextx.IContext, page types.Page, condition *types.GlobalSettingsCondition) (
	[]*types.GlobalSettings, int64, error) {

	req := &protoBackend.ListGlobalSettingsReq{Page: convertPage(page)}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, fmt.Errorf("failed to convert global settings conditions: %w", err)
	}
	resp, err := h.cli.listGlobalSettings(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list global settings: %w", err)
	}
	items, total := resp.ConvertGlobalSettingsToTypes()

	return items, total, nil
}

// CountGlobalSettings counts stored settings matching the supplied conditions.
func (h *Handler) CountGlobalSettings(ctx contextx.IContext, condition *types.GlobalSettingsCondition) (int64, error) {
	req := &protoBackend.ListGlobalSettingsReq{OnlyCount: true}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, fmt.Errorf("failed to convert global settings conditions: %w", err)
	}
	resp, err := h.cli.listGlobalSettings(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to count global settings: %w", err)
	}

	return resp.GetData().GetTotal(), nil
}
