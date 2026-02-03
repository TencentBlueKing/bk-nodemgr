/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package v3 provides the IAM v3 resource callback router for bk-nodemgr.
package v3

import (
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/iam/v3/provider"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

// handler holds the router group and capabilities for IAM v3 routes.
type handler struct {
	rg         *gin.RouterGroup
	capability *options.Capability
	dispatcher provider.Dispatcher
}

// newHandler creates a new handler for IAM v3 routes.
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// Create sub router for IAM v3 with path /v3
		rg:         rg.Group("/v3"),
		capability: capability,
	}
}

// Load registers the IAM v3 resource callback routes.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	// Create Dispatcher for IAM resource callbacks
	h.dispatcher = provider.NewDispatcher()

	// Register NetworkArea provider
	networkAreaProvider := provider.NewNetworkAreaProvider(capability.StorageTopo)
	h.dispatcher.RegisterProvider(provider.ResourceTypeNetworkArea, networkAreaProvider)

	// Register NetworkUnit provider
	networkUnitProvider := provider.NewNetworkUnitProvider(capability.StorageTopo)
	h.dispatcher.RegisterProvider(provider.ResourceTypeNetworkUnit, networkUnitProvider)

	// Apply Basic Auth middleware to IAM routes
	h.rg.Use(h.basicAuthMiddleware())

	// Register IAM resource callback route
	h.rg.POST("/resource", restserver.Handler(h.handleResourceCallback))
}

// basicAuthMiddleware creates a middleware that validates Basic Auth credentials
// using the IAM system token.
func (h *handler) basicAuthMiddleware() gin.HandlerFunc {
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

		// Create context for IAM handler
		ctx := contextx.FromContext(c.Request.Context())

		// Validate credentials using IAMV3Handler
		if err := h.capability.IAMV3Handler.IsBasicAuthAllowed(ctx, username, password); err != nil {
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
	return h.dispatcher.Dispatch(rCtx)
}
