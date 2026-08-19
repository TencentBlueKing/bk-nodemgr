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

// Package versionlog provides version changelog routes.
package versionlog

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg *gin.RouterGroup
}

func newHandler(rg *gin.RouterGroup) *handler {
	return &handler{
		rg: rg.Group("/version_log"),
	}
}

// Load registers the version log routes.
func Load(rg *gin.RouterGroup, _ *options.Capability) {
	h := newHandler(rg)

	h.rg.GET("/version_logs_list/", restserver.Handler(h.VersionLogsList))
	h.rg.GET("/changelog/:version", restserver.Handler(h.VersionLogDetail))
}
