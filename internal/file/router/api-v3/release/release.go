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

// Package release provides release operations for the file service.
package release

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/options"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg      *gin.RouterGroup
	manager manager.IManager
}

func newHandler(rg *gin.RouterGroup, opt *options.Capability) *handler {
	return &handler{
		rg:      rg.Group("/release"),
		manager: opt.Manager,
	}
}

// Load enables release routes.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	// agent releases.
	h.rg.POST("/agent/list", restserver.Handler(h.ListReleaseAgent))
	h.rg.POST("/agent/count", restserver.Handler(h.CountReleaseAgent))
	h.rg.POST("/agent/get", restserver.Handler(h.GetReleaseAgent))
	h.rg.POST("/agent/distinct", restserver.Handler(h.DistinctReleaseAgent))
	h.rg.POST("/agent/enable", restserver.Handler(h.EnableReleaseAgent))
	h.rg.POST("/agent/disable", restserver.Handler(h.DisableReleaseAgent))
	h.rg.POST("/agent/set_as_default", restserver.Handler(h.SetAsDefaultReleaseAgent))
	h.rg.POST("/agent/cancel_as_default", restserver.Handler(h.CancelAsDefaultReleaseAgent))
	h.rg.POST("/agent/delete", restserver.Handler(h.DeleteReleaseAgent))
	h.rg.POST("/agent/set_labels_many", restserver.Handler(h.SetReleaseAgentLabelsMany))

	// proxy releases.
	h.rg.POST("/proxy/list", restserver.Handler(h.ListReleaseProxy))
	h.rg.POST("/proxy/count", restserver.Handler(h.CountReleaseProxy))
	h.rg.POST("/proxy/get", restserver.Handler(h.GetReleaseProxy))
	h.rg.POST("/proxy/distinct", restserver.Handler(h.DistinctReleaseProxy))
	h.rg.POST("/proxy/enable", restserver.Handler(h.EnableReleaseProxy))
	h.rg.POST("/proxy/disable", restserver.Handler(h.DisableReleaseProxy))
	h.rg.POST("/proxy/set_as_default", restserver.Handler(h.SetAsDefaultReleaseProxy))
	h.rg.POST("/proxy/cancel_as_default", restserver.Handler(h.CancelAsDefaultReleaseProxy))
	h.rg.POST("/proxy/delete", restserver.Handler(h.DeleteReleaseProxy))
	h.rg.POST("/proxy/set_labels_many", restserver.Handler(h.SetReleaseProxyLabelsMany))

	// plugin releases.
	h.rg.POST("/plugin/list", restserver.Handler(h.ListReleasePlugin))
	h.rg.POST("/plugin/count", restserver.Handler(h.CountReleasePlugin))
	h.rg.POST("/plugin/get", restserver.Handler(h.GetReleasePlugin))
	h.rg.POST("/plugin/distinct", restserver.Handler(h.DistinctReleasePlugin))
	h.rg.POST("/plugin/exist", restserver.Handler(h.ExistReleasePlugin))
	h.rg.POST("/plugin/default_version", restserver.Handler(h.GetReleasePluginDefaultVersion))
	h.rg.POST("/plugin/distinct_name", restserver.Handler(h.DistinctNameReleasePlugin))
	h.rg.POST("/plugin/enable", restserver.Handler(h.EnableReleasePlugin))
	h.rg.POST("/plugin/disable", restserver.Handler(h.DisableReleasePlugin))
	h.rg.POST("/plugin/set_as_default", restserver.Handler(h.SetAsDefaultReleasePlugin))
	h.rg.POST("/plugin/cancel_as_default", restserver.Handler(h.CancelAsDefaultReleasePlugin))
	h.rg.POST("/plugin/delete", restserver.Handler(h.DeleteReleasePlugin))
	h.rg.POST("/plugin/set_hidden", restserver.Handler(h.SetHiddenReleasePlugin))
	h.rg.POST("/plugin/cancel_hidden", restserver.Handler(h.CancelHiddenReleasePlugin))

	// cert releases.
	h.rg.POST("/cert/list", restserver.Handler(h.ListReleaseCert))
	h.rg.POST("/cert/delete", restserver.Handler(h.DeleteReleaseCert))

	// bintool releases.
	h.rg.POST("/bintool/list", restserver.Handler(h.ListReleaseBinTool))
	h.rg.POST("/bintool/delete", restserver.Handler(h.DeleteReleaseBinTool))

	// plugin_bintool releases.
	h.rg.POST("/plugin_bintool/list", restserver.Handler(h.ListReleasePluginBinTool))
	h.rg.POST("/plugin_bintool/distinct_name", restserver.Handler(h.DistinctNameReleasePluginBinTool))
	h.rg.POST("/plugin_bintool/delete", restserver.Handler(h.DeleteReleasePluginBinTool))
}
