/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package bkrepo provides the BKRepo mock API router.
package bkrepo

import (
	"github.com/gin-gonic/gin"
)

// Load registers all BKRepo API routes with the given gin router group.
func Load(rg *gin.RouterGroup, conf *Config) {
	// initialize storage from config.
	bkrepoStorage := newStorage(conf)

	// create a sub router for BKRepo API.
	bkrepoGroup := rg.Group("/bk-repo")

	h := newHandler(bkrepoGroup, bkrepoStorage)

	// upload file.
	h.rg.PUT("/generic/:project/:repo/*path", h.UploadFile)

	// download file.
	h.rg.GET("/generic/:project/:repo/*path", h.DownloadFile)

	// query node info.
	h.rg.GET("/repository/api/node/detail/:project/:repo/*path", h.QueryNodeInfo)

	// list nodes with pagination.
	h.rg.GET("/repository/api/node/page/:project/:repo/*path", h.ListNode)

	// create directory.
	h.rg.POST("/repository/api/node/mkdir/:project/:repo/*path", h.MkDir)
}
