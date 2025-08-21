/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package proxy defines the agent apis.
package proxy

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/gin-gonic/gin"
)

// handler ...
type handler struct {
	rg                              *gin.RouterGroup
	manager                         manager.Manager
	storageNetworkUnit              topo.IStorageNetworkUnit
	storageNodeDeploymentDomainInit nodedeployment.IStorageDomainInit
	storageHostCredit               credit.IStorageHostCredit
	storageHost                     topo.IStorageHost

	logger logger.Logger
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                              rg.Group("/proxy"),
		manager:                         capability.Manager,
		storageNetworkUnit:              capability.StorageTopo,
		storageNodeDeploymentDomainInit: capability.StorageNodeDeployment,
		logger:                          capability.Logger,
		storageHostCredit:               capability.StorageCredit,
		storageHost:                     capability.StorageTopo,
	}
}

// Load load agent handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/install", restserver.Handler(h.Install))
	h.rg.POST("/upgrade", restserver.Handler(h.Upgrade))
	h.rg.POST("/restart", restserver.Handler(h.Restart))
	h.rg.POST("/reconfig", restserver.Handler(h.Reconfig))
}
