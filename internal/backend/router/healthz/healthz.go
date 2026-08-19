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

// Package healthz defines the healthz router.
package healthz

import (
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/gin-gonic/gin"
)

// handler ...
type handler struct {
	rg      *gin.RouterGroup
	manager manager.IManager
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:      rg.Group("/healthz"),
		manager: capability.Manager,
	}
}

// Load ...
func Load(rg *gin.RouterGroup, capability *options.Capability, middlewares ...gin.HandlerFunc) {
	h := newHandler(rg, capability)

	// enable middlewares.
	h.rg.Use(middlewares...)

	h.rg.GET("", h.Healthz)
}

// Healthz check service health.
func (h *handler) Healthz(ctx *gin.Context) {
	resp := new(Response)

	if h.manager == nil {
		resp.OK = false
		resp.Manager = "not initialized"
		ctx.JSON(http.StatusInternalServerError, resp)

		return
	}

	if err := h.manager.CheckHealth(); err != nil {
		resp.OK = false
		resp.Manager = err.Error()
		ctx.JSON(http.StatusInternalServerError, resp)

		return
	}

	resp.OK = true
	resp.Manager = "ok"
	ctx.JSON(http.StatusOK, resp)
}
