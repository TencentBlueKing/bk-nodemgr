/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package apiv3 defines the api v3 router.
package apiv3

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/pkg"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/schedule"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/sync"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/gin-gonic/gin"
)

// handler ...
type handler struct {
	rg           *gin.RouterGroup
	authIdentity rest.AuthIdentity
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:           rg.Group("/api/v3"),
		authIdentity: capability.AuthIdentity,
	}
}

// Load ter register the api v3 router.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	//h.rg.Use(rest.MiddlewareAuth(h.authIdentity))

	sync.Load(h.rg, capability)
	node.Load(h.rg, capability)
	topo.Load(h.rg, capability)
	pkg.Load(h.rg, capability)
	schedule.Load(h.rg, capability)
}
