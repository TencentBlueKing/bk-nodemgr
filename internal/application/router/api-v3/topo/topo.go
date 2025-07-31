/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo provides the topo API handler.
package topo

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	backendHandler backend.Handler
	logger         logger.Logger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/topo"),
		backendHandler: capability.BackendHandler,
		logger:         capability.Logger,
	}
}

// Load loads topo handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	// business apis.
	h.rg.POST("/business/list", restserver.Handler(h.ListBusiness))

	// host apis.
	h.rg.POST("/host/list", restserver.Handler(h.ListHost))
	h.rg.POST("/host/distinct", restserver.Handler(h.DistinctHost))

	// networkarea apis.
	h.rg.POST("/networkarea/list", restserver.Handler(h.ListNetworkArea))
	h.rg.POST("/networkarea/statistics", restserver.Handler(h.StatisticsNetworkArea))
	h.rg.POST("/networkarea/get", restserver.Handler(h.GetNetworkArea))
	h.rg.POST("/networkarea/create", restserver.Handler(h.CreateNetworkArea))
	h.rg.POST("/networkarea/update", restserver.Handler(h.UpdateNetworkArea))
	h.rg.POST("/networkarea/delete", restserver.Handler(h.DeleteNetworkArea))

	// networkunit apis.
	h.rg.POST("/networkunit/list", restserver.Handler(h.ListNetworkUnit))
	h.rg.POST("/networkunit/get", restserver.Handler(h.GetNetworkUnit))
	h.rg.POST("/networkunit/create", restserver.Handler(h.CreateNetworkUnit))
	h.rg.POST("/networkunit/update", restserver.Handler(h.UpdateNetworkUnit))
	h.rg.POST("/networkunit/delete", restserver.Handler(h.DeleteNetworkUnit))

	// topo graph.
	h.rg.POST("/graph/get", restserver.Handler(h.GetGraph))
	h.rg.POST("/graph/node/count", restserver.Handler(h.CountGraphNode))

	// topo event apis.
	h.rg.POST("/event/list", restserver.Handler(h.ListEvent))
	h.rg.POST("/event/distinct", restserver.Handler(h.DistinctEvent))

	// constant apis.
	h.rg.POST("/constant/get", restserver.Handler(h.GetConstant))
}
