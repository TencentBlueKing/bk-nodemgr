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

// Package download is the file download router.
package download

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
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:      rg.Group("/download"),
		manager: opt.Manager,
	}
}

// Load enables web router into gin.Engine.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/installer", restserver.FileHandler(h.Installer))
	h.rg.POST("/agent", restserver.FileHandler(h.Agent))
	h.rg.POST("/proxy", restserver.FileHandler(h.Proxy))
	h.rg.POST("/plugin", restserver.FileHandler(h.Plugin))
	h.rg.POST("/cert", restserver.FileHandler(h.Cert))
	h.rg.POST("/bintool", restserver.FileHandler(h.BinTool))
	h.rg.POST("/plugin_bintool", restserver.FileHandler(h.PluginBinTool))
}

// LoadRemoteFile load remote file handler.
func LoadRemoteFile(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/remote_file", restserver.FileHandler(h.Remote))
}
