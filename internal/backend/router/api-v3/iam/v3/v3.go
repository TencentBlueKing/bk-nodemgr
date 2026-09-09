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

// Package v3 provides the IAM v3 resource callback router for bk-nodemgr.
package v3

import (
	"net/http"
	"regexp"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/v3/provider"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv3"
	"github.com/gin-gonic/gin"
)

// handler holds the router group and capabilities for IAM v3 routes.
type handler struct {
	rg           *gin.RouterGroup
	dispatcher   provider.IDispatcher
	iamV3Handler iamv3.IHandler
}

// newHandler creates a new handler for IAM v3 routes.
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// Create sub router for IAM v3 with path /v3
		rg:           rg.Group("/v3"),
		dispatcher:   capability.AuthProviderV3Handler,
		iamV3Handler: capability.IAMV3Handler,
	}
}

// Load registers the IAM v3 resource callback routes.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	// Apply Basic Auth middleware to IAM routes
	h.rg.Use(h.basicAuthMiddleware())

	// Register IAM resource callback route
	h.rg.POST("/resource", restserver.Handler(h.handleResourceCallback))
}

// basicAuthMiddleware creates a middleware that validates Basic Auth credentials
// using the IAM system token.
// nolint: varnamelen
func (h *handler) basicAuthMiddleware() gin.HandlerFunc {
	tenantIDPattern := regexp.MustCompile(restserver.TenantIDRegexp)

	return func(c *gin.Context) {
		// Extract Basic Auth credentials from HTTP Header
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="IAM"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "missing basic auth credentials",
			})

			return
		}

		rCtx, err := restserver.GenRestContext(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized, "message": "missing rest context",
			})

			return
		}

		var tenantID string
		switch tenant.GetMode() {
		case tenant.ModeSingle:
			tenantID = tenant.SingleModeTenantID
		case tenant.ModeMultiple:
			tenantID = c.GetHeader(apigwheader.BKGWTenantIDKey)
			if !tenantIDPattern.MatchString(tenantID) {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"code": http.StatusBadRequest, "message": "invalid tenant id",
				})

				return
			}
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError, "message": "invalid tenant mode",
			})

			return
		}

		// Keep token validation and downstream provider queries in the same tenant.
		rCtx.Data().SetTenantID(tenantID)
		ctx := contextx.From(rCtx, contextx.WithTenantID(tenantID))

		// Validate credentials using IAMV3Handler
		if err := h.iamV3Handler.IsBasicAuthAllowed(ctx, username, password); err != nil {
			c.Header("WWW-Authenticate", `Basic realm="IAM"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "invalid credentials",
			})

			return
		}

		// Authentication successful, continue processing
		c.Next()
	}
}

// handleResourceCallback handles IAM resource callback requests.
func (h *handler) handleResourceCallback(rCtx restserver.IContext) (interface{}, error) {
	// Bind request body to proto message
	req := new(protoBackend.IAMResourceCallbackReq)
	if err := rCtx.GContext().ShouldBindJSON(req); err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// Extract filter as map
	var filterMap map[string]interface{}
	if req.GetFilter() != nil {
		filterMap = req.GetFilter().AsMap()
	}

	// Convert and validate page parameters
	page, err := protoBackend.ConvIAMCallbackPageToTypes(req.GetPage())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	resourceType := req.GetType()
	requestMethod := provider.RequestMethod(req.GetMethod())

	// Dispatch to handler
	result, dispatchErr := h.dispatcher.DispatchMethod(
		rCtx,
		resourceType,
		requestMethod,
		filterMap,
		page,
	)
	if dispatchErr != nil {
		return nil, resterrf.ErrWrap(resterrf.Unknown, dispatchErr)
	}

	return result, nil
}
