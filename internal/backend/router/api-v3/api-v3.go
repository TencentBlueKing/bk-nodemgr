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

// Package apiv3 defines the api v3 router.
package apiv3

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/callback"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/cipher"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/iam"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/pkg"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/policy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/process"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/proxy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/schedule"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/sync"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/topo"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg *gin.RouterGroup
}

func newHandler(rg *gin.RouterGroup, _ *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg: rg.Group("/api/v3"),
	}
}

// LoadBasicAPIs register the basic apis.
func LoadBasicAPIs(rg *gin.RouterGroup, capability *options.Capability, middlewares ...gin.HandlerFunc) {
	h := newHandler(rg, capability)

	// enable middlewares.
	h.rg.Use(middlewares...)

	sync.Load(h.rg, capability)
	node.Load(h.rg, capability)
	plugin.Load(h.rg, capability)
	topo.Load(h.rg, capability)
	pkg.Load(h.rg, capability)
	schedule.Load(h.rg, capability)
	policy.Load(h.rg, capability)
	process.Load(h.rg, capability)
	deploypolicy.Load(h.rg, capability)
	cipher.Load(h.rg, capability)
	auth.Load(h.rg, capability)
	iam.Load(h.rg, capability)
}

// LoadCallbackAPIs register the callback apis.
func LoadCallbackAPIs(rg *gin.RouterGroup, capability *options.Capability, middlewares ...gin.HandlerFunc) {
	h := newHandler(rg, capability)

	// enable middlewares.
	h.rg.Use(middlewares...)

	callback.Load(h.rg, capability)
}

// LoadProxyAPIs register the proxy apis.
func LoadProxyAPIs(rg *gin.RouterGroup, capability *options.Capability, middlewares ...gin.HandlerFunc) {
	h := newHandler(rg, capability)

	// enable middlewares.
	h.rg.Use(middlewares...)

	proxy.Load(h.rg, capability)
}
