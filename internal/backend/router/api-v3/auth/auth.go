/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package auth provides proactive permission verification endpoints.
package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

type handler struct {
	rg         *gin.RouterGroup
	authorizer auth.IAuthorizer
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		rg:         rg.Group("/auth"),
		authorizer: capability.Authorizer,
	}
}

// Load registers the auth verification routes.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/verify", restserver.Handler(h.Verify))
	h.rg.POST("/authorized", restserver.Handler(h.Authorized))
}

// BuildBizResources constructs types.AuthResource slice for business IDs.
func BuildBizResources(bizIDs ...int64) []types.AuthResource {
	resources := make([]types.AuthResource, 0, len(bizIDs))
	for _, bizID := range bizIDs {
		resources = append(resources, types.AuthResource{
			SystemID: types.SystemIDCMDB,
			Type:     types.AuthResourceTypeBiz,
			ID:       fmt.Sprintf("%d", bizID),
		})
	}

	return resources
}

// BuildNetworkUnitResources constructs types.AuthResource slice for network unit IDs.
func BuildNetworkUnitResources(networkUnitIDs ...int64) []types.AuthResource {
	resources := make([]types.AuthResource, 0, len(networkUnitIDs))
	for _, id := range networkUnitIDs {
		resources = append(resources, types.AuthResource{
			SystemID: types.SystemIDNodeMgr,
			Type:     types.AuthResourceTypeNetworkUnit,
			ID:       fmt.Sprintf("%d", id),
		})
	}

	return resources
}
