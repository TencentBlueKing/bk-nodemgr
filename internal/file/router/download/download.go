/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package download is the file download router.
package download

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	agentFileGroup iface.FileGroup
	proxyFileGroup iface.FileGroup
	logger         logger.Logger
}

func newHandler(rg *gin.RouterGroup, opt *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/download"),
		agentFileGroup: opt.AgentFileGroup,
		proxyFileGroup: opt.ProxyFileGroup,
		logger:         opt.Logger,
	}
}

// Load enables web router into gin.Engine.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/agent", rest.FileHandler(h.Agent))
	h.rg.POST("/proxy", rest.FileHandler(h.Proxy))
}
