/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package web is the web router.
package web

import (
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/frontsetting"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg           *gin.RouterGroup
	frontSetting frontsetting.IFrontSetting
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:           rg.Group(""),
		frontSetting: capability.FrontSetting,
	}
}

// Load enables web router into gin.Engine.
func Load(rg *gin.RouterGroup, capability *options.Capability, middlewares ...gin.HandlerFunc) {
	h := newHandler(rg, capability)

	h.rg.Use(middlewares...)

	h.rg.GET("", h.Index)
}

// Index return the index page.
func (h *handler) Index(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"BK_LOGIN_URL":             h.frontSetting.BKLoginURL(),
		"BK_REQUEST_ID_HEADER_KEY": h.frontSetting.BKRequestIDHeaderKey(),
		"BK_PASS_ANALYTICS_SCRIPT": h.frontSetting.BKPassAnalyticsScript(),
		"PASSWORD_VAULT_SWITCH":    h.frontSetting.PasswordVaultSwitch(),
		"PASSWORD_VAULT_NAME":      h.frontSetting.PasswordVaultName(),
		"BK_USER_WEB_URL":          h.frontSetting.BKUserWebURL(),
	})
}
