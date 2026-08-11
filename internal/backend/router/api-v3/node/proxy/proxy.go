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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	cipherStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/cipher"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	globalsettingsStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/globalsettings"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/gin-gonic/gin"
)

// handler ...
type handler struct {
	rg                              *gin.RouterGroup
	nodeMgrIface                    managerIface.INodeManager
	storageNetworkUnit              topoStg.IStorageNetworkUnit
	storageNodeDeploymentDomainInit nodeStg.IDaoNodeDeployment
	storageHostCredit               credit.IStorageHostCredit
	storageGlobalSettings           globalsettingsStorage.IStorage
	storageHost                     topoStg.IStorageHost
	storageCipher                   cipherStg.IStorage
	authorizer                      auth.IAuthorizer
	cmdbHandler                     cmdb.IHandler
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                              rg.Group("/proxy"),
		nodeMgrIface:                    capability.Manager,
		storageNetworkUnit:              capability.StorageTopo,
		storageNodeDeploymentDomainInit: capability.StorageNode,
		storageHostCredit:               capability.StorageCredit,
		storageGlobalSettings:           capability.StorageGlobalSettings,
		storageHost:                     capability.StorageTopo,
		storageCipher:                   capability.StorageCipher,
		authorizer:                      capability.Authorizer,
		cmdbHandler:                     capability.CmdbHandler,
	}
}

// Load load agent handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/install", restserver.Handler(h.Install))
	h.rg.POST("/upgrade", restserver.Handler(h.Upgrade))
	h.rg.POST("/upgrade_check", restserver.Handler(h.ProxyUpgradeCheck))
	h.rg.POST("/restart", restserver.Handler(h.Restart))
	h.rg.POST("/reconfig", restserver.Handler(h.Reconfig))
	h.rg.POST("/update", restserver.Handler(h.Update))
	h.rg.POST("/uninstall", restserver.Handler(h.Uninstall))
	h.rg.POST("/install_check", restserver.Handler(h.ProxyInstallCheck))
	h.rg.POST("/assign_unit", restserver.Handler(h.AssignProxyUnit))
	h.rg.POST("/assign_unit_multi", restserver.Handler(h.AssignProxyUnitMulti))
	h.rg.POST("/update_ops_fields", restserver.Handler(h.ProxyUpdateOpsFields))
}
