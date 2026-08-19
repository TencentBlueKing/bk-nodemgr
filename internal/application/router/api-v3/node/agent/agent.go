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

// Package agent provides the agent API handler.
package agent

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	backendHandler backend.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/agent"),
		backendHandler: capability.BackendHandler,
	}
}

// Load loads workflow handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/install", restserver.Handler(h.Install))
	h.rg.POST("/upgrade", restserver.Handler(h.Upgrade))
	h.rg.POST("/upgrade_check", restserver.Handler(h.AgentUpgradeCheck))
	h.rg.POST("/reconfig", restserver.Handler(h.Reconfig))
	h.rg.POST("/restart", restserver.Handler(h.Restart))
	h.rg.POST("/uninstall", restserver.Handler(h.Uninstall))
	h.rg.POST("/install_check", restserver.Handler(h.AgentInstallCheck))
	h.rg.POST("/assign_unit", restserver.Handler(h.AssignUnit))
	h.rg.POST("/assign_unit_multi", restserver.Handler(h.AssignUnitMulti))
	h.rg.POST("/update_ops_fields", restserver.Handler(h.AgentUpdateOpsFields))

	h.rg.GET("/install_template/download", restserver.FileHandler(h.DownloadInstallTemplate))
	h.rg.POST("/install_template/upload", restserver.Handler(h.UploadInstallTemplate))
}
