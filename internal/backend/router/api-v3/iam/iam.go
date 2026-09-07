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

// Package iam provides the IAM resource callback router for bk-nodemgr.
package iam

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	iamv3 "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/iam/v3"
	iamv4 "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/iam/v4"
	"github.com/gin-gonic/gin"
)

// handler holds the router group for IAM routes.
type handler struct {
	rg *gin.RouterGroup
}

// newHandler creates a new handler for IAM routes.
func newHandler(rg *gin.RouterGroup) *handler {
	return &handler{
		// Create sub router for IAM with path /iam
		rg: rg.Group("/iam"),
	}
}

// Load registers the IAM resource callback routes.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg)

	// Load IAM v3 routes
	iamv3.Load(h.rg, capability)

	// Load IAM v4 routes
	iamv4.Load(h.rg, capability)
}
