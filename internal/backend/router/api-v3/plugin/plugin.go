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

// Package plugin defines the router to handle the plugin request.
package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/plugin/workflow"
	globalsettingsStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/globalsettings"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg                    *gin.RouterGroup
	daoNodeWorkflow       nodeStg.IDaoNodeWorkflow
	daoPlugin             pluginStg.IDaoPlugin
	daoPluginWorkflow     pluginStg.IDaoPluginWorkflow
	daoHost               topoStg.IStorageHost
	daoProcessConfig      pluginStg.IDaoProcessConfig
	storageGlobalSettings globalsettingsStorage.IStorage
	domainPlugin          pluginStg.IDomainPlugin
	pluginMgrIface        managerIface.IPluginManager
	authorizer            auth.IAuthorizer
	cmdbHandler           cmdb.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                    rg.Group("/plugin"),
		daoNodeWorkflow:       capability.StorageNode,
		daoPlugin:             capability.StoragePlugin,
		daoPluginWorkflow:     capability.StoragePlugin,
		daoHost:               capability.StorageTopo,
		daoProcessConfig:      capability.StoragePlugin,
		storageGlobalSettings: capability.StorageGlobalSettings,
		domainPlugin:          capability.StoragePlugin,
		pluginMgrIface:        capability.Manager,
		authorizer:            capability.Authorizer,
		cmdbHandler:           capability.CmdbHandler,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.List))
	h.rg.POST("/install", restserver.Handler(h.Install))
	h.rg.POST("/upgrade", restserver.Handler(h.Upgrade))
	h.rg.POST("/uninstall", restserver.Handler(h.Uninstall))
	h.rg.POST("/apply_subconfig", restserver.Handler(h.ApplySubConfig))
	h.rg.POST("/set_memo", restserver.Handler(h.SetMemo))
	h.rg.POST("/list_config_files", restserver.Handler(h.ListConfigFiles))

	h.rg.POST("/start", restserver.Handler(h.Start))
	h.rg.POST("/restart", restserver.Handler(h.Restart))
	h.rg.POST("/migrate_from_v2", restserver.Handler(h.MigrateFromV2))
	h.rg.POST("/stop", restserver.Handler(h.Stop))

	h.rg.POST("/start_debug", restserver.Handler(h.StartDebug))
	h.rg.POST("/stop_debug", restserver.Handler(h.StopDebug))

	workflow.Load(h.rg, capability)
}
