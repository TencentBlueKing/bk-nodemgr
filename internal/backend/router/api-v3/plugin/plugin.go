/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package plugin defines the router to handle the plugin request.
package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/plugin/workflow"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg              *gin.RouterGroup
	daoNodeWorkflow nodeStg.IDaoNodeWorkflow
	daoPlugin       pluginStg.IDaoPlugin
	daoHost         topoStg.IStorageHost
	domainPlugin    pluginStg.IDomainPlugin
	pluginMgrIface  managerIface.IPluginManager
	authorizer      auth.IAuthorizer
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:              rg.Group("/plugin"),
		daoNodeWorkflow: capability.StorageNode,
		daoPlugin:       capability.StoragePlugin,
		daoHost:         capability.StorageTopo,
		domainPlugin:    capability.StoragePlugin,
		pluginMgrIface:  capability.Manager,
		authorizer:      capability.Authorizer,
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

	h.rg.POST("/restart", restserver.Handler(h.Restart))
	h.rg.POST("/migrate_from_v2", restserver.Handler(h.MigrateFromV2))
	h.rg.POST("/stop", restserver.Handler(h.Stop))

	workflow.Load(h.rg, capability)
}
