/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topoStg provides the topoStg API handler.
package topo

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg          *gin.RouterGroup
	manager     manager.Manager
	storage     topoStg.IStorage
	cmdbHandler cmdb.IHandler
	logger      logger.Logger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:          rg.Group("/topo"),
		manager:     capability.Manager,
		storage:     capability.TopoStorage,
		cmdbHandler: capability.CmdbHandler,
		logger:      capability.Logger,
	}
}

// Load loads topoStg handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	// business apis.
	h.rg.POST("/business/list", rest.RestHandlerFunc(h.ListBusiness))

	// host apis.
	h.rg.POST("/host/list", rest.RestHandlerFunc(h.ListHost))
	h.rg.POST("/host/distinct", rest.RestHandlerFunc(h.DistinctHost))

	// networkarea apis.
	h.rg.POST("/networkarea/list", rest.RestHandlerFunc(h.ListNetworkArea))
	h.rg.POST("/networkarea/get", rest.RestHandlerFunc(h.GetNetworkArea))
	h.rg.POST("/networkarea/create", rest.RestHandlerFunc(h.CreateNetworkArea))
	h.rg.POST("/networkarea/update", rest.RestHandlerFunc(h.UpdateNetworkArea))
	h.rg.POST("/networkarea/delete", rest.RestHandlerFunc(h.DeleteNetworkArea))

	// networkunit apis.
	h.rg.POST("/networkunit/list", rest.RestHandlerFunc(h.ListNetworkUnit))
	h.rg.POST("/networkunit/get", rest.RestHandlerFunc(h.GetNetworkUnit))
	h.rg.POST("/networkunit/create", rest.RestHandlerFunc(h.CreateNetworkUnit))
	h.rg.POST("/networkunit/update", rest.RestHandlerFunc(h.UpdateNetworkUnit))
	h.rg.POST("/networkunit/delete", rest.RestHandlerFunc(h.DeleteNetworkUnit))

	// accesspoint apis.
	h.rg.POST("/accesspoint/list", rest.RestHandlerFunc(h.ListAccessPoint))

	// topoStg event apis.
	h.rg.POST("/event/list", rest.RestHandlerFunc(h.ListEvent))
	h.rg.POST("/event/distinct", rest.RestHandlerFunc(h.DistinctEvent))

	// constant apis.
	h.rg.POST("/constant/get", rest.RestHandlerFunc(h.GetConstant))
}
