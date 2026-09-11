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

// Package v4 provides the IAM V4 resource callback router.
package v4

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/v4/provider"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv4"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-Id"

type handler struct {
	rg           *gin.RouterGroup
	dispatcher   provider.IDispatcher
	iamV4Handler iamv4.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{rg: rg.Group("/v4"), dispatcher: capability.AuthProviderV4Handler, iamV4Handler: capability.IAMV4Handler}
}

// Load registers the dedicated IAM V4 callback transport.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	if capability.IAMV4Handler == nil {
		return
	}
	h := newHandler(rg, capability)
	h.rg.Use(h.basicAuthMiddleware())
	h.rg.POST("/resource", h.handleResourceCallback)
}

func (h *handler) basicAuthMiddleware() gin.HandlerFunc {
	tenantIDPattern := regexp.MustCompile(restserver.TenantIDRegexp)
	return func(gCtx *gin.Context) {
		gCtx.Header(requestIDHeader, gCtx.GetHeader(requestIDHeader))
		rCtx, err := restserver.GenRestContext(gCtx)
		if err != nil {
			callbackError(gCtx, contextx.FromContext(gCtx.Request.Context()), http.StatusInternalServerError, err)
			return
		}
		var tenantID string
		switch tenant.GetMode() {
		case tenant.ModeSingle:
			tenantID = tenant.SingleModeTenantID
		case tenant.ModeMultiple:
			tenantID = gCtx.GetHeader(apigwheader.BKGWTenantIDKey)
			if !tenantIDPattern.MatchString(tenantID) {
				callbackError(gCtx, rCtx, http.StatusBadRequest, fmt.Errorf("invalid tenant id"))
				return
			}
		default:
			callbackError(gCtx, rCtx, http.StatusInternalServerError, fmt.Errorf("invalid tenant mode"))
			return
		}
		// Token validation and downstream GenRestContext must use the same tenant.
		rCtx.Data().SetTenantID(tenantID)
		ctx := contextx.From(rCtx, contextx.WithTenantID(tenantID))
		username, password, ok := gCtx.Request.BasicAuth()
		if !ok {
			callbackError(gCtx, ctx, http.StatusUnauthorized, iamv4.ErrInvalidCredentials)
			return
		}
		if err := h.iamV4Handler.IsBasicAuthAllowed(ctx, username, password); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, iamv4.ErrInvalidCredentials) {
				status = http.StatusUnauthorized
			}
			callbackError(gCtx, ctx, status, err)

			return
		}
		gCtx.Next()
	}
}

func callbackError(gCtx *gin.Context, ctx contextx.IContext, status int, err error) {
	code, message := "INTERNAL", "resource query failed"
	switch status {
	case http.StatusBadRequest:
		code, message = "INVALID_ARGUMENT", "invalid callback arguments"
	case http.StatusUnauthorized:
		code, message = "UNAUTHENTICATED", "invalid callback credentials"
		gCtx.Header("WWW-Authenticate", `Basic realm="IAM"`)
	case http.StatusNotFound:
		code, message = "NOT_FOUND", "resource type or method not found"
	}
	logger.G.Biz(ctx).WithErr(err).With("request-id", gCtx.GetHeader(requestIDHeader)).Error("IAM V4 callback failed")
	gCtx.Header(requestIDHeader, gCtx.GetHeader(requestIDHeader))
	gCtx.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func (h *handler) handleResourceCallback(gCtx *gin.Context) {
	rCtx, err := restserver.GenRestContext(gCtx)
	if err != nil {
		callbackError(gCtx, contextx.FromContext(gCtx.Request.Context()), http.StatusInternalServerError, err)
		return
	}
	req := new(protoBackend.IAMV4ResourceCallbackReq)
	if err := gCtx.ShouldBindJSON(req); err != nil {
		callbackError(gCtx, rCtx, http.StatusBadRequest, err)
		return
	}
	if err := req.Validate(); err != nil {
		callbackError(gCtx, rCtx, http.StatusBadRequest, err)
		return
	}
	var page types.Page
	method := provider.RequestMethod(req.GetMethod())
	if method == provider.RequestMethodListInstance {
		page, err = req.GetPage().ConvertToTypes()
		if err != nil {
			callbackError(gCtx, rCtx, http.StatusBadRequest, err)
			return
		}
	}
	if h.dispatcher == nil {
		callbackError(gCtx, rCtx, http.StatusInternalServerError, fmt.Errorf("resource dispatcher is nil"))
		return
	}
	result, err := h.dispatcher.DispatchMethod(rCtx, req.GetType(), method, req.GetFilter().AsMap(), page, req.GetRequires())
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, provider.ErrInvalidArgument):
			status = http.StatusBadRequest
		case errors.Is(err, provider.ErrNotFound):
			status = http.StatusNotFound
		}
		callbackError(gCtx, rCtx, status, err)

		return
	}
	// Encode before writing headers so serialization failures retain the V4 error envelope.
	body, err := json.Marshal(gin.H{"data": result})
	if err != nil {
		callbackError(gCtx, rCtx, http.StatusInternalServerError, err)
		return
	}
	logger.G.Biz(rCtx).With("request-id", gCtx.GetHeader(requestIDHeader)).Debug("IAM V4 callback completed")
	gCtx.Data(http.StatusOK, "application/json; charset=utf-8", body)
}
