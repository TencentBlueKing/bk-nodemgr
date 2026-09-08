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

// Package v4 provides the IAM v4 resource callback router for bk-nodemgr.
package v4

import (
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/v4/provider"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv4"
	"github.com/gin-gonic/gin"
)

// handler holds the router group and capabilities for IAM v4 routes.
type handler struct {
	rg           *gin.RouterGroup
	dispatcher   provider.IDispatcher
	iamV4Handler iamv4.IHandler
}

// newHandler creates a new handler for IAM v4 routes.
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// Create sub router for IAM v4 with path /v4
		rg:           rg.Group("/v4"),
		dispatcher:   capability.AuthProviderV4Handler,
		iamV4Handler: capability.IAMV4Handler,
	}
}

// Load registers the IAM v4 resource callback routes.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	if capability.IAMV4Handler == nil {
		return
	}

	h := newHandler(rg, capability)
	h.rg.Use(h.basicAuthMiddleware())
	h.rg.POST("/resource", restserver.Handler(h.handleResourceCallback))
}

// basicAuthMiddleware creates a middleware that validates Basic Auth credentials
// using the IAM system token.
// nolint: varnamelen
func (h *handler) basicAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="IAM"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "missing basic auth credentials",
			})

			return
		}

		ctx := contextx.FromContext(c.Request.Context())
		if err := h.iamV4Handler.IsBasicAuthAllowed(ctx, username, password); err != nil {
			c.Header("WWW-Authenticate", `Basic realm="IAM"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "invalid credentials",
			})

			return
		}

		rCtx, err := restserver.GenRestContext(c)
		if err != nil {
			c.Next()

			return
		}

		if rCtx.Data().GetTenantID() == "" {
			rCtx.Data().SetTenantID(tenant.SingleModeTenantID)
		}

		c.Next()
	}
}

// handleResourceCallback handles IAM resource callback requests.
func (h *handler) handleResourceCallback(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.IAMResourceCallbackReq)
	if err := rCtx.GContext().ShouldBindJSON(req); err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	var filterMap map[string]interface{}
	if req.GetFilter() != nil {
		filterMap = req.GetFilter().AsMap()
	}

	page, err := protoBackend.ConvIAMCallbackPageToTypes(req.GetPage())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, dispatchErr := h.dispatcher.DispatchMethod(
		rCtx,
		req.GetType(),
		provider.RequestMethod(req.GetMethod()),
		filterMap,
		page,
	)
	if dispatchErr != nil {
		return nil, resterrf.ErrWrap(resterrf.Unknown, dispatchErr)
	}

	return result, nil
}
