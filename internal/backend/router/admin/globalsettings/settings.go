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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

const (
	maxGlobalSettingsLimit = 500
)

// ListGlobalSettings lists global settings.
func (h *handler) ListGlobalSettings(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ListGlobalSettingsReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list global settings, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	resp := new(protoBackend.ListGlobalSettingsResp)

	if req.GetOnlyCount() {
		count, err := h.storage.CountGlobalSettings(rCtx, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count global settings")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp.Data = &protoBackend.ListGlobalSettingsResp_Data{
			Total: count,
		}

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes(maxGlobalSettingsLimit)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list global settings, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	settings, num, err := h.storage.ListGlobalSettings(rCtx, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list global settings")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp.ConvertGlobalSettingsFromTypes(num, settings)

	return resp.GetData(), nil
}

// GetGlobalSettings gets a global setting by name.
func (h *handler) GetGlobalSetting(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.GetGlobalSettingReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get global setting, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	value, err := h.storage.GetGlobalSetting(rCtx, req.GetSettingName())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get global setting")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.GetGlobalSettingResp)
	resp.Data = &protoBackend.GetGlobalSettingResp_Data{Value: value}

	return resp.GetData(), nil
}

// UpsertManyGlobalSettings upserts many global setting.
func (h *handler) UpsertManyGlobalSettings(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.UpsertGlobalSettingsReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upsert global settings, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.storage.UpsertGlobalSettings(rCtx, req.ConvertGlobalSettingsToTypes()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upsert global settings")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := &protoBackend.UpsertGlobalSettingsResp{}

	return resp.GetData(), nil
}

// DeleteManyGlobalSettings deletes many global settings by names.
func (h *handler) DeleteManyGlobalSettings(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.DeleteGlobalSettingsReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete global settings, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.storage.DeleteGlobalSettings(rCtx, req.GetSettingName()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete global settings")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := &protoBackend.DeleteGlobalSettingsResp{}

	return resp.GetData(), nil
}
