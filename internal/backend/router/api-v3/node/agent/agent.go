/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package agent defines the agent apis.
package agent

import (
	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	cipherStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/cipher"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

// handler ...
type handler struct {
	rg                 *gin.RouterGroup
	nodeMgrIface       managerIface.INodeManager
	storageNetworkUnit topoStg.IStorageNetworkUnit
	storageHost        topoStg.IStorageHost
	storageHostCredit  credit.IStorageHostCredit
	storageCipher      cipherStg.IStorage

	domainNodeInstall topoStg.IDomainNodeInstall
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                 rg.Group("/agent"),
		nodeMgrIface:       capability.Manager,
		storageNetworkUnit: capability.StorageTopo,
		storageHost:        capability.StorageTopo,
		storageHostCredit:  capability.StorageCredit,
		storageCipher:      capability.StorageCipher,
		domainNodeInstall:  capability.StorageTopo,
	}
}

// Load load agent handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/install", restserver.Handler(h.AgentInstall))
	h.rg.POST("/install_check", restserver.Handler(h.AgentInstallCheck))
	h.rg.POST("/upgrade", restserver.Handler(h.AgentUpgrade))
	h.rg.POST("/reconfig", restserver.Handler(h.AgentReconfig))
	h.rg.POST("/restart", restserver.Handler(h.AgentRestart))
	h.rg.POST("/uninstall", restserver.Handler(h.AgentUninstall))
	h.rg.POST("/assign_unit", restserver.Handler(h.AgentAssignUnit))
}
