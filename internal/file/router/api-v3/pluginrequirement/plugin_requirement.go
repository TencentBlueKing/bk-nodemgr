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

// Package pluginrequirement provides the plugin requirement file router.
package pluginrequirement

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/topo"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg          *gin.RouterGroup
	manager     manager.IManager
	storageTopo topo.IStorage
}

func newHandler(rg *gin.RouterGroup, opt *options.Capability) *handler {
	return &handler{
		rg:          rg.Group("/plugin/requirement"),
		manager:     opt.Manager,
		storageTopo: opt.StorageTopo,
	}
}

// Load enables the plugin requirement file routes.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/create", restserver.Handler(h.Create))
	h.rg.POST("/update", restserver.Handler(h.Update))
	h.rg.POST("/get", restserver.Handler(h.Get))
	h.rg.GET("/download", restserver.FileHandler(h.Download))
	h.rg.POST("/transfer", restserver.Handler(h.Transfer))
}
