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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
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
	h.rg.POST("/business/list", rest.RestHandlerFunc(h.ListBusiness))

	// host apis.
	h.rg.POST("/host/list", rest.RestHandlerFunc(h.ListHost))

	// networkarea apis.
	h.rg.POST("/networkarea/list", rest.RestHandlerFunc(h.ListNetworkArea))
	h.rg.POST("/networkarea/get", rest.RestHandlerFunc(h.GetNetworkArea))
	// h.rg.POST("/networkarea/create", rest.RestHandlerFunc(h.CreateNetworkArea))
	// h.rg.POST("/networkarea/update", rest.RestHandlerFunc(h.UpdateNetworkArea))
	// h.rg.POST("/networkarea/delete", rest.RestHandlerFunc(h.DeleteNetworkArea))

	// networkunit apis.
	h.rg.POST("/networkunit/list", rest.RestHandlerFunc(h.ListNetworkUnit))
	h.rg.POST("/networkunit/get", rest.RestHandlerFunc(h.GetNetworkUnit))
	h.rg.POST("/networkunit/create", rest.RestHandlerFunc(h.CreateNetworkUnit))
	h.rg.POST("/networkunit/update", rest.RestHandlerFunc(h.UpdateNetworkUnit))
	h.rg.POST("/networkunit/delete", rest.RestHandlerFunc(h.DeleteNetworkUnit))

	// topo graph.
	h.rg.POST("/graph/get", rest.RestHandlerFunc(h.GetGraph))
}
