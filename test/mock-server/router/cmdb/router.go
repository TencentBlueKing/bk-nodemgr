/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb provides the CMDB mock API router.
package cmdb

import (
	"github.com/gin-gonic/gin"
)

// handler defines the CMDB mock API handler.
type handler struct {
	rg    *gin.RouterGroup
	store *storage
}

// newHandler creates a new CMDB handler.
func newHandler(rg *gin.RouterGroup, store *storage) *handler {
	return &handler{
		// create a sub router for CMDB API v3.
		rg:    rg.Group("/api/v3"),
		store: store,
	}
}

// Load registers all CMDB API routes with the given gin router group.
func Load(rg *gin.RouterGroup, conf *Config) {
	// initialize storage from config.
	cmdbStorage := newStorage(conf)

	// create a sub router for CMDB API v3.
	cmdbGroup := rg.Group("/cmdb")

	h := newHandler(cmdbGroup, cmdbStorage)

	// search businesses.
	h.rg.POST("/biz/search/:supplier_account", h.SearchBusiness)
}
