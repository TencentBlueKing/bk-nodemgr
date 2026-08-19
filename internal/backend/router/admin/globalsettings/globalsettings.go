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

// Package globalsettings provides the global settings API handler.
package globalsettings

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/globalsettings"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg          *gin.RouterGroup
	manager     manager.IManager
	storage     globalsettings.IStorage
	cmdbHandler cmdb.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:          rg.Group("/globalsettings"),
		manager:     capability.Manager,
		storage:     capability.StorageGlobalSettings,
		cmdbHandler: capability.CmdbHandler,
	}
}

// Load loads global settings handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListGlobalSettings))
	h.rg.POST("/get", restserver.Handler(h.GetGlobalSetting))
	h.rg.POST("/upsertmany", restserver.Handler(h.UpsertManyGlobalSettings))
	h.rg.POST("/deletemany", restserver.Handler(h.DeleteManyGlobalSettings))
	h.rg.POST("/plugin_compatibility_mode_policy/get", restserver.Handler(h.GetPluginCompatibilityModePolicy))
	h.rg.POST("/plugin_compatibility_mode_policy/upsert", restserver.Handler(h.UpsertPluginCompatibilityModePolicy))
}
