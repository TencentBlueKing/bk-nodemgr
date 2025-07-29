/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package globalsettings provides the global settings API handler.
package globalsettings

import (
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

const (
	maxGlobalSettingsLimit = 500
)

// ListGlobalSettings lists global settings.
func (h *handler) ListGlobalSettings(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.ListGlobalSettingsReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list global settings, failed to decode request body, err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	resp := new(protoBackend.ListGlobalSettingsResp)

	if req.GetOnlyCount() {
		count, err := h.storage.CountGlobalSettings(ctx, req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to count global settings, err: %v", err)
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp.Data = &protoBackend.ListGlobalSettingsResp_Data{
			Total: count,
		}

		return resp.GetData(), nil
	}

	settings, num, err := h.storage.ListGlobalSettings(
		ctx,
		req.ConvertPageToTypes(maxGlobalSettingsLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list global settings, err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp.ConvertGlobalSettingsFromTypes(num, settings)

	return resp.GetData(), nil
}

// GetGlobalSettings gets a global setting by name.
func (h *handler) GetGlobalSetting(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.GetGlobalSettingReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get global setting, failed to decode request body, err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	value, err := h.storage.GetGlobalSetting(ctx, req.GetSettingName())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get global setting, err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.GetGlobalSettingResp)
	resp.Data = &protoBackend.GetGlobalSettingResp_Data{Value: value}

	return resp.GetData(), nil
}

// UpsertManyGlobalSettings upserts many global setting.
func (h *handler) UpsertManyGlobalSettings(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.UpsertGlobalSettingsReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upsert global settings, failed to decode request body, err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.storage.UpsertManyGlobalSettings(ctx, req.ConvertGlobalSettingsToTypes()...); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upsert global settings, err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := &protoBackend.UpsertGlobalSettingsResp{}

	return resp.GetData(), nil
}

// DeleteManyGlobalSettings deletes many global settings by names.
func (h *handler) DeleteManyGlobalSettings(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.DeleteGlobalSettingsReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete global settings, failed to decode request body, err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.storage.DeleteManyGlobalSettings(ctx, req.GetSettingName()...); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete global settings, err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := &protoBackend.DeleteGlobalSettingsResp{}

	return resp.GetData(), nil
}
