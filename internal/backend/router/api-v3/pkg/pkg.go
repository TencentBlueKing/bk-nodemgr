/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package pkg describes the package router.
package pkg

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/tenant"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg               *gin.RouterGroup
	daoPackageEvent  release.IPackageEvent
	daoReleasePlugin release.IPlugin
	daoRelease       release.IRelease
	daoReleaseAgent  release.IAgent
	daoReleaseProxy  release.IProxy
	daoPlugin        plugin.IDaoPlugin
	daoTenant        tenant.IStorage
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:               rg.Group("/package"),
		daoPackageEvent:  capability.StorageRelease,
		daoReleasePlugin: capability.StorageRelease,
		daoReleaseAgent:  capability.StorageRelease,
		daoReleaseProxy:  capability.StorageRelease,
		daoRelease:       capability.StorageRelease,
		daoPlugin:        capability.StoragePlugin,
		daoTenant:        capability.StorageTenant,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/release/list", restserver.Handler(h.ListRelease))
	h.rg.POST("/release/distinct", restserver.Handler(h.DistinctRelease))
	h.rg.POST("/release/set_labels", restserver.Handler(h.SetReleaseLabels))
	h.rg.POST("/release/set_labels_many", restserver.Handler(h.SetReleaseLabelsMany))
	h.rg.POST("/release/enable", restserver.Handler(h.EnableRelease))
	h.rg.POST("/release/disable", restserver.Handler(h.DisableRelease))
	h.rg.POST("/release/set_as_default", restserver.Handler(h.SetAsDefaultRelease))
	h.rg.POST("/release/cancel_as_default", restserver.Handler(h.CancelAsDefaultRelease))
	h.rg.POST("/release/delete", restserver.Handler(h.DeleteRelease))
	h.rg.POST("/release_plugin/list", restserver.Handler(h.ListReleasePlugin))
	h.rg.POST("/release_plugin/enable", restserver.Handler(h.EnableReleasePlugin))
	h.rg.POST("/release_plugin/disable", restserver.Handler(h.DisableReleasePlugin))
	h.rg.POST("/release_plugin/set_as_default", restserver.Handler(h.SetAsDefaultReleasePlugin))
	h.rg.POST("/release_plugin/cancel_as_default", restserver.Handler(h.CancelAsDefaultReleasePlugin))
	h.rg.POST("/release_plugin/delete", restserver.Handler(h.DeleteReleasePlugin))
	h.rg.POST("/release_agent/list", restserver.Handler(h.ListReleaseAgent))
	h.rg.POST("/release_proxy/list", restserver.Handler(h.ListReleaseProxy))
	
	h.rg.POST("/event/list", restserver.Handler(h.ListPackageEvent))
	h.rg.POST("/event/distinct", restserver.Handler(h.DistinctPackageEvent))
}
