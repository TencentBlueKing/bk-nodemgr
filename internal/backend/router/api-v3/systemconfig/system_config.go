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

// Package systemconfig provides the read-only Backend business configuration API.
package systemconfig

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/systemconfig"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg           *gin.RouterGroup
	authorizer   auth.IAuthorizer
	systemConfig systemconfig.ISystemConfig
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		rg:           rg.Group("/system_config"),
		authorizer:   capability.Authorizer,
		systemConfig: capability.SystemConfig,
	}
}

// Load registers the query under the caller mode's existing authentication middleware.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)
	h.rg.POST("/get", restserver.Handler(h.GetSystemConfig))
}

// GetSystemConfig returns the same parameter-free view for application-state and user-state callers.
func (h *handler) GetSystemConfig(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.GetSystemConfigReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get system config, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	resources := auth.BuildPlatformResources(types.PlatformResourceSystemConfig)
	if err := h.authorizer.Check(rCtx, auth.ActionPlatformManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get system config, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	resp := new(protoBackend.GetSystemConfigResp)
	resp.ConvertSystemConfigFromTypes(h.systemConfig.Get())

	return resp.GetData(), nil
}
