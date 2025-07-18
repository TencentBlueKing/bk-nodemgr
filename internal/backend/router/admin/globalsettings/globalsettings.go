/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package globalsettings provides the global settings API handler.
package globalsettings

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg          *gin.RouterGroup
	manager     manager.Manager
	storage     globalsettings.IStorage
	cmdbHandler cmdb.IHandler
	logger      logger.Logger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:          rg.Group("/globalsettings"),
		manager:     capability.Manager,
		storage:     capability.StorageGlobalSettings,
		cmdbHandler: capability.CmdbHandler,
		logger:      capability.Logger,
	}
}

// Load loads global settings handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", rest.RestHandlerFunc(h.ListGlobalSettings))
	h.rg.POST("/get", rest.RestHandlerFunc(h.GetGlobalSetting))
	h.rg.POST("/upsertmany", rest.RestHandlerFunc(h.UpsertManyGlobalSettings))
	h.rg.POST("/deletemany", rest.RestHandlerFunc(h.DeleteManyGlobalSettings))
}
