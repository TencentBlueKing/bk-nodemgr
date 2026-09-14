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

// BuildNodeWorkflowHistoryResources maps stored workflow roles to history permissions.
func BuildNodeWorkflowHistoryResources(workflow *types.NodeWorkflow) map[auth.Action][]types.AuthResource {
	resources := BuildBizResources(workflow.BizIDs...)
	actionResources := make(map[auth.Action][]types.AuthResource)
	for _, nodeRole := range workflow.NodeRoles {
		switch nodeRole {
		case types.NodeRoleBlank, types.NodeRoleAgent:
			actionResources[auth.ActionAgentHistoryView] = resources
		case types.NodeRoleProxy:
			actionResources[auth.ActionProxyHistoryView] = resources
		default:
			// Unknown roles require both history view permissions.
			actionResources[auth.ActionAgentHistoryView] = resources
			actionResources[auth.ActionProxyHistoryView] = resources
		}
	}
	if len(actionResources) == 0 {
		// Missing roles require both history view permissions.
		actionResources[auth.ActionAgentHistoryView] = resources
		actionResources[auth.ActionProxyHistoryView] = resources
	}

	return actionResources
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

// BuildNetworkAreaResources constructs types.AuthResource slice for network
// area IDs.
func BuildNetworkAreaResources(networkAreaIDs ...int64) []types.AuthResource {
	resources := make([]types.AuthResource, 0, len(networkAreaIDs))
	for _, id := range networkAreaIDs {
		resources = append(resources, types.AuthResource{
			SystemID: types.SystemIDNodeMgr,
			Type:     types.AuthResourceTypeNetworkArea,
			ID:       fmt.Sprintf("%d", id),
		})
	}

	return resources
}
