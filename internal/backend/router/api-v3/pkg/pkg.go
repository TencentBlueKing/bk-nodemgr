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

// Package pkg describes the package router.
package pkg

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	pkgStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/pkg"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

const (
	asyncPoolNum  = 10
	asyncPoolSize = 100
)

type handler struct {
	rg                      *gin.RouterGroup
	goAsyncPool             goasync.IHandler
	daoPackageEvent         release.IPackageEvent
	daoReleasePlugin        release.IPlugin
	daoReleaseAgent         release.IAgent
	daoReleaseProxy         release.IProxy
	daoReleaseCert          release.ICert
	daoReleaseBinTool       release.IBinTool
	daoReleasePluginBinTool release.IPluginBinTool
	daoPlugin               plugin.IDaoPlugin
	daoPackageWorkflow      pkgStg.IStoragePackageWorkflow
	daoWorkflow             workflow.IStorage
	pkgMgrIface             managerIface.IPackageManager

	daoTenant  tenant.IStorage
	authorizer auth.IAuthorizer
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	goAsyncPool, _ := goasync.NewHandler(goasync.HandlerOption{
		PoolNum:               asyncPoolNum,
		PerPoolSize:           asyncPoolSize,
		LoadBalancingStrategy: goasync.LoadBalancingStrategyLeastFirst,
	})

	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                      rg.Group("/package"),
		goAsyncPool:             goAsyncPool,
		daoPackageEvent:         capability.StorageRelease,
		daoReleasePlugin:        capability.StorageRelease,
		daoReleaseAgent:         capability.StorageRelease,
		daoReleaseProxy:         capability.StorageRelease,
		daoReleaseCert:          capability.StorageRelease,
		daoReleaseBinTool:       capability.StorageRelease,
		daoReleasePluginBinTool: capability.StorageRelease,
		daoPlugin:               capability.StoragePlugin,
		daoPackageWorkflow:      capability.StoragePackage,
		daoWorkflow:             capability.StorageWorkflow,
		pkgMgrIface:             capability.Manager,
		daoTenant:               capability.StorageTenant,
		authorizer:              capability.Authorizer,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	// release agent.
	h.rg.POST("/release/agent/list", restserver.Handler(h.ListReleaseAgent))
	h.rg.POST("/release/agent/list/brief", restserver.Handler(h.ListReleaseAgentBrief))
	h.rg.POST("/release/agent/distinct", restserver.Handler(h.DistinctReleaseAgent))
	h.rg.POST("/release/agent/set_labels_many", restserver.Handler(h.SetReleaseAgentLabelsMany))
	h.rg.POST("/release/agent/enable", restserver.Handler(h.EnableReleaseAgent))
	h.rg.POST("/release/agent/disable", restserver.Handler(h.DisableReleaseAgent))
	h.rg.POST("/release/agent/set_as_default", restserver.Handler(h.SetAsDefaultReleaseAgent))
	h.rg.POST("/release/agent/cancel_as_default", restserver.Handler(h.CancelAsDefaultReleaseAgent))
	h.rg.POST("/release/agent/delete", restserver.Handler(h.DeleteReleaseAgent))

	// release proxy.
	h.rg.POST("/release/proxy/list", restserver.Handler(h.ListReleaseProxy))
	h.rg.POST("/release/proxy/list/brief", restserver.Handler(h.ListReleaseProxyBrief))
	h.rg.POST("/release/proxy/distinct", restserver.Handler(h.DistinctReleaseProxy))
	h.rg.POST("/release/proxy/set_labels_many", restserver.Handler(h.SetReleaseProxyLabelsMany))
	h.rg.POST("/release/proxy/enable", restserver.Handler(h.EnableReleaseProxy))
	h.rg.POST("/release/proxy/disable", restserver.Handler(h.DisableReleaseProxy))
	h.rg.POST("/release/proxy/set_as_default", restserver.Handler(h.SetAsDefaultReleaseProxy))
	h.rg.POST("/release/proxy/cancel_as_default", restserver.Handler(h.CancelAsDefaultReleaseProxy))
	h.rg.POST("/release/proxy/delete", restserver.Handler(h.DeleteReleaseProxy))

	// release plugin.
	h.rg.POST("/release/plugin/list", restserver.Handler(h.ListReleasePlugin))
	h.rg.POST("/release/plugin/list/brief", restserver.Handler(h.ListReleasePluginBrief))
	h.rg.POST("/release/plugin/distinct", restserver.Handler(h.DistinctReleasePlugin))
	h.rg.POST("/release/plugin/enable", restserver.Handler(h.EnableReleasePlugin))
	h.rg.POST("/release/plugin/disable", restserver.Handler(h.DisableReleasePlugin))
	h.rg.POST("/release/plugin/set_hidden", restserver.Handler(h.SetHiddenReleasePlugin))
	h.rg.POST("/release/plugin/set_as_default", restserver.Handler(h.SetAsDefaultReleasePlugin))
	h.rg.POST("/release/plugin/cancel_as_default", restserver.Handler(h.CancelAsDefaultReleasePlugin))
	h.rg.POST("/release/plugin/delete", restserver.Handler(h.DeleteReleasePlugin))
	h.rg.POST("/release/plugin/get_config_variables", restserver.Handler(h.GetConfigVariablesReleasePlugin))

	// release cert.
	h.rg.POST("/release/cert/list", restserver.Handler(h.ListReleaseCert))
	h.rg.POST("/release/cert/delete", restserver.Handler(h.DeleteReleaseCert))

	// release bintool.
	h.rg.POST("/release/bintool/list", restserver.Handler(h.ListReleaseBinTool))
	h.rg.POST("/release/bintool/delete", restserver.Handler(h.DeleteReleaseBinTool))

	// release plugin bintool.
	h.rg.POST("/release/plugin_bintool/list", restserver.Handler(h.ListReleasePluginBinTool))
	h.rg.POST("/release/plugin_bintool/delete", restserver.Handler(h.DeleteReleasePluginBinTool))

	// event
	h.rg.POST("/event/list", restserver.Handler(h.ListPackageEvent))
	h.rg.POST("/event/distinct", restserver.Handler(h.DistinctPackageEvent))

	// workflow
	h.rg.POST("/workflow/import/v3/plugin", restserver.Handler(h.PackagePluginV3Import))
	h.rg.POST("/workflow/import/v2/plugin", restserver.Handler(h.PackagePluginV2Import))
	h.rg.POST("/workflow/import/v2/external_plugin", restserver.Handler(h.PackageExternalPluginV2Import))
	h.rg.POST("/workflow/import_result", restserver.Handler(h.PackageImportResult))
}
