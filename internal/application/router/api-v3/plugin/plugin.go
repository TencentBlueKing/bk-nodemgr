/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package plugin define plugin api v3 router.
package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/api-v3/plugin/workflow"
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
		rg:             rg.Group("/plugin"),
		backendHandler: capability.BackendHandler,
	}
}

// Load loads topo handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.List))
	h.rg.POST("/install", restserver.Handler(h.Install))
	h.rg.POST("/upgrade", restserver.Handler(h.Upgrade))
	h.rg.POST("/uninstall", restserver.Handler(h.Uninstall))
	h.rg.POST("/apply_subconfig", restserver.Handler(h.ApplySubConfig))
	h.rg.POST("/set_memo", restserver.Handler(h.SetMemo))
	h.rg.POST("/list_permitted_operations", restserver.Handler(h.ListPerimittedOperations))

	workflow.Load(h.rg, capability)
}
