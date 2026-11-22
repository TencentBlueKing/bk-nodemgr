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

// Package globalsettings provides read-only global settings queries.
package globalsettings

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg         *gin.RouterGroup
	storage    globalsettings.IStorage
	authorizer auth.IAuthorizer
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		rg:         rg.Group("/globalsettings"),
		storage:    capability.StorageGlobalSettings,
		authorizer: capability.Authorizer,
	}
}

// Load registers global settings reads.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)
	h.rg.POST("/list", restserver.Handler(h.ListGlobalSettings))
}

// ListGlobalSettings lists stored global values, including custom settings.
func (h *handler) ListGlobalSettings(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ListGlobalSettingsReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list global settings, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	resources := auth.BuildPlatformResources(types.PlatformResourceGlobalSettings)
	if err := h.authorizer.Check(rCtx, auth.ActionPlatformManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list global settings, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	if req.GetOnlyCount() {
		count, err := h.storage.CountGlobalSettings(rCtx, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count global settings")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		return &protoBackend.ListGlobalSettingsResp_Data{Total: count, Items: []*protoBackend.GlobalSetting{}}, nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list global settings, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	settings, total, err := h.storage.ListGlobalSettings(rCtx, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list global settings")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ListGlobalSettingsResp)
	resp.ConvertGlobalSettingsFromTypes(total, settings)

	return resp.GetData(), nil
}
