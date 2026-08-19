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

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	types "github.com/TencentBlueKing/bk-nodemgr/internal/application/types/router/healthz"
	"github.com/gin-gonic/gin"
)

// handler ...
type handler struct {
	rg *gin.RouterGroup
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, _ *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg: rg.Group("/healthz"),
	}
}

// Load ...
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.GET("", h.Healthz)
}

// Healthz check service health.
func (h *handler) Healthz(ctx *gin.Context) {
	resp := new(types.Response)
	resp.OK = true
	resp.Manager = "ok"
	ctx.JSON(http.StatusOK, resp)
}
