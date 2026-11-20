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

// Package excel provides Excel template download and upload APIs for node
// agent and proxy installation, and for out-of-band settings. Template
// rendering and parsing live in the workbook subpackage.
package excel

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/frontsetting"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	backendHandler backend.IHandler
	frontSetting   frontsetting.IFrontSetting
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/excel"),
		backendHandler: capability.BackendHandler,
		frontSetting:   capability.FrontSetting,
	}
}

// Load loads excel handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/agent_install_template/parse", restserver.Handler(h.ExcelParseAgentInstallTemplate))
	h.rg.POST("/proxy_install_template/parse", restserver.Handler(h.ExcelParseProxyInstallTemplate))
	h.rg.POST("/ops_template/parse", restserver.Handler(h.ExcelParseOpsTemplate))

	h.rg.POST("/agent_install_template/download", restserver.FileHandler(h.ExcelDownloadAgentInstallTemplate))
	h.rg.POST("/proxy_install_template/download", restserver.FileHandler(h.ExcelDownloadProxyInstallTemplate))
	h.rg.POST("/ops_template/download", restserver.FileHandler(h.ExcelDownloadOpsTemplate))
}
