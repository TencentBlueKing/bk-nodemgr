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

// Package ops provides out-of-band settings and operation history APIs.
package ops

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/gin-gonic/gin"
)

const (
	asyncPoolNum  = 10
	asyncPoolSize = 100
)

type handler struct {
	rg                *gin.RouterGroup
	daoNodeOpsHistory nodeStg.IDaoNodeOpsHistory
	authorizer        auth.IAuthorizer
	storageHost       topoStg.IStorageHost
	cmdbHandler       cmdb.IHandler
	goAsyncPool       goasync.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	goAsyncPool, _ := goasync.NewHandler(goasync.HandlerOption{
		PoolNum:               asyncPoolNum,
		PerPoolSize:           asyncPoolSize,
		LoadBalancingStrategy: goasync.LoadBalancingStrategyLeastFirst,
	})

	return &handler{
		rg:                rg.Group("/ops"),
		daoNodeOpsHistory: capability.StorageNode,
		authorizer:        capability.Authorizer,
		storageHost:       capability.StorageTopo,
		cmdbHandler:       capability.CmdbHandler,
		goAsyncPool:       goAsyncPool,
	}
}

// Load registers out-of-band settings and operation history routes.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/update_fields/agent", restserver.Handler(h.UpdateFieldsAgent))
	h.rg.POST("/update_fields/proxy", restserver.Handler(h.UpdateFieldsProxy))
	h.rg.POST("/history/list", restserver.Handler(h.HistoryList))
	h.rg.POST("/history/detail", restserver.Handler(h.HistoryDetail))
	h.rg.POST("/list_fields/agent", restserver.Handler(h.ListFieldsAgent))
	h.rg.POST("/list_fields/proxy", restserver.Handler(h.ListFieldsProxy))
}
