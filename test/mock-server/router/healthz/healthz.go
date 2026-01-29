/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package healthz defines the healthz router.
package healthz

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handler defines the healthz handler.
type handler struct {
	rg *gin.RouterGroup
}

// newHandler creates a new healthz handler.
func newHandler(rg *gin.RouterGroup) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg: rg.Group("/healthz"),
	}
}

// Load registers healthz routes.
func Load(rg *gin.RouterGroup) {
	h := newHandler(rg)

	h.rg.GET("", h.Healthz)
}

// Healthz check service health.
func (h *handler) Healthz(gCtx *gin.Context) {
	resp := new(Response)
	resp.OK = true
	resp.Manager = "ok"
	gCtx.JSON(http.StatusOK, resp)
}
