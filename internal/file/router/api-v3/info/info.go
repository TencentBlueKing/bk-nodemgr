/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package info is the file info query router.
package info

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/options"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg      *gin.RouterGroup
	manager manager.IManager
}

func newHandler(rg *gin.RouterGroup, opt *options.Capability) *handler {
	return &handler{
		rg:      rg.Group("/info"),
		manager: opt.Manager,
	}
}

// Load enables web router into gin.Engine.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/agent", restserver.Handler(h.Agent))
	h.rg.POST("/proxy", restserver.Handler(h.Proxy))
	h.rg.POST("/plugin", restserver.Handler(h.Plugin))
	h.rg.POST("/cert", restserver.Handler(h.Cert))
	h.rg.POST("/bintool", restserver.Handler(h.BinTool))
	h.rg.POST("/installer", restserver.Handler(h.Installer))
	h.rg.POST("/plugin_bintool", restserver.Handler(h.PluginBinTool))
}

// buildInfoResp builds a FileInfoResp_Data from a file.
func buildInfoResp(name string, size int64, md5 string) *protoFile.FileInfoResp_Data {
	return &protoFile.FileInfoResp_Data{
		Name: name,
		Size: size,
		Md5:  md5,
	}
}
