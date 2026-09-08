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

// Package apiv3 defines the api v3 router.
package apiv3

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/api-v3/download"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/api-v3/export"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/api-v3/info"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/api-v3/publish"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/api-v3/transfer"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/api-v3/upload"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

// handler ...
type handler struct {
	rg *gin.RouterGroup
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, _ *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg: rg.Group("/api/v3"),
	}
}

func newDownloadHandler(rg *gin.RouterGroup, _ *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg: rg.Group("/:tenant_id/api/v3"),
	}
}

// LoadBasicAPIs register the basic apis.
func LoadBasicAPIs(rg *gin.RouterGroup, capability *options.Capability, middlewares ...gin.HandlerFunc) {
	h := newHandler(rg, capability)

	// enable middlewares.
	h.rg.Use(middlewares...)

	download.Load(h.rg, capability)
	download.LoadRemoteFile(h.rg, capability)
	export.Load(h.rg, capability)
	info.Load(h.rg, capability)
	publish.Load(h.rg, capability)
	transfer.Load(h.rg, capability)
	upload.Load(h.rg, capability)
}

// LoadDownloadAPIs register the download apis.
func LoadDownloadAPIs(rg *gin.RouterGroup, capability *options.Capability, middlewares ...gin.HandlerFunc) {
	h := newDownloadHandler(rg, capability)

	// enable middlewares.
	h.rg.Use(middlewares...)
	h.rg.Use(restserver.MiddlewarePathTenantID("tenant_id"))

	download.Load(h.rg, capability)
	info.Load(h.rg, capability)
}

// LoadExportAPIs register the export download apis.
func LoadExportAPIs(rg *gin.RouterGroup, capability *options.Capability, middlewares ...gin.HandlerFunc) {
	h := newHandler(rg, capability)

	// enable middlewares.
	h.rg.Use(middlewares...)

	export.LoadDownload(h.rg, capability)
}
