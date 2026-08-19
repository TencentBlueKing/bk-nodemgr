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

// Package web is the web router.
package web

import (
	"net/http"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/frontsetting"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	bksaasbklogin "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bksaas/bklogin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
	"github.com/gin-gonic/gin"
)

const (
	// webUserInfoTimeout is the timeout for getting web user info from bklogin API.
	webUserInfoTimeout = 3 * time.Second
)

type handler struct {
	rg             *gin.RouterGroup
	frontSetting   frontsetting.IFrontSetting
	bkloginHandler bksaasbklogin.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group(""),
		frontSetting:   capability.FrontSetting,
		bkloginHandler: capability.BKLoginHandler,
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
	// Get tenant ID
	var tenantID string
	if tenant.GetMode() == tenant.ModeSingle {
		tenantID = tenant.SingleModeTenantID
	} else {
		// Multiple tenant mode: get tenant ID from request header
		tenantID = header.BKTenantIDGetter(ctx.Request)
		if tenantID == "" {
			// Use default value if tenant ID is not found in header
			tenantID = tenant.SingleModeTenantID
		}
	}

	// Get login name from bklogin API for display purposes only.
	// This is not used for actual authentication logic.
	// Returns empty string if auth cookie is missing or GetWebUserInfo call fails.
	var (
		bkUsername   = ""
		loginName    = ""
		userTimeZone = ""
		userEmail    = ""
	)
	token, cookieErr := ctx.Cookie(h.bkloginHandler.GetAuthType())
	if cookieErr == nil && token != "" {
		nCtx, cancel := contextx.WithTimeout(contextx.New(ctx.Request.Context()), webUserInfoTimeout)
		defer cancel()

		info, err := h.bkloginHandler.GetWebUserInfo(nCtx, token)
		if err != nil {
			logger.G.Biz(nCtx).Warn("failed to get web user info")
		}

		if info != nil {
			loginName = info.LoginName
			userTimeZone = info.TimeZone
			userEmail = info.Email
			bkUsername = info.BKUsername
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"BK_LOGIN_URL":                h.frontSetting.BKLoginURL(),
		"BK_REQUEST_ID_HEADER_KEY":    h.frontSetting.BKRequestIDHeaderKey(),
		"BK_PASS_ANALYTICS_SCRIPT":    h.frontSetting.BKPassAnalyticsScript(),
		"BK_USERNAME":                 bkUsername,
		"BK_IAM_SAAS_HOST":            h.frontSetting.BKIamSaaSHost(),
		"BK_USER_SAAS_HOST":           h.frontSetting.BKUserSaaSHost(),
		"USER_TIMEZONE":               userTimeZone,
		"USER_EMAIL":                  userEmail,
		"PASSWORD_VAULT_SWITCH":       h.frontSetting.PasswordVaultSwitch(),
		"PASSWORD_VAULT_NAME":         h.frontSetting.PasswordVaultName(),
		"BK_USER_WEB_URL":             h.frontSetting.BKUserWebURL(),
		"BK_TENANT":                   tenantID,
		"BK_DOMAIN":                   h.frontSetting.BKDomain(),
		"BK_DOCS_CENTER_URL":          h.frontSetting.BKDocsCenterURL(),
		"BKAPP_NAV_OPEN_SOURCE_URL":   h.frontSetting.BKAppNavOpenSourceURL(),
		"WINDOWS_WMI_PORT_DEFAULT":    h.frontSetting.WindowsWMIPortDefault(),
		"UNIX_SSH_PORT_DEFAULT":       h.frontSetting.UnixSSHPortDefault(),
		"ENABLE_NOTICE":               h.frontSetting.EnableNotice(),
		"APP_VERSION":                 version.VERSION,
		"LOGIN_NAME":                  loginName,
		"BK_IAM_SYSTEM_ID_BK_NODEMGR": h.frontSetting.BKIamSystemIDBKNodemgr(),
		"BK_IAM_SYSTEM_ID_BK_CMDB":    h.frontSetting.BKIamSystemIDBKCmdb(),
	})
}
