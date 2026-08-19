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

package operation

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (h *handler) checkWorkflowOperatePermission(rCtx restserver.IContext, workflowID string) error {
	nodeWorkflow, err := h.daoNodeWorkflow.GetNodeWorkflow(rCtx, workflowID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check workflow operate permission, failed to get the target node workflow")
		return resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	bizIDs := nodeWorkflow.BizIDs
	resources := authRouter.BuildBizResources(bizIDs...)
	actions := workflowOperateActions(nodeWorkflow.NodeRoles)

	actionResources := make(map[auth.Action][]types.AuthResource, len(actions))
	for _, action := range actions {
		actionResources[action] = resources
	}

	if authErr := h.authorizer.CheckMany(rCtx, actionResources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to check workflow operate permission, permission denied")
		return resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	return nil
}

// workflowOperateActions determines which operate actions are required based on node roles.
// Returns ActionAgentOperate, ActionProxyOperate, or both depending on the roles present.
func workflowOperateActions(nodeRoles []types.NodeRole) []auth.Action {
	if len(nodeRoles) == 0 {
		return []auth.Action{auth.ActionAgentOperate, auth.ActionProxyOperate}
	}

	// Pre-allocate for at most 2 actions: AgentOperate and ProxyOperate
	const maxActions = 2
	actions := make([]auth.Action, 0, maxActions)
	needAgentOperate := false
	needProxyOperate := false

	for _, nodeRole := range nodeRoles {
		switch nodeRole {
		case types.NodeRoleBlank, types.NodeRoleAgent:
			needAgentOperate = true
		case types.NodeRoleProxy:
			needProxyOperate = true
		default:
			// Unknown role: require both permissions to be safe
			needAgentOperate = true
			needProxyOperate = true
		}
	}

	if needAgentOperate {
		actions = append(actions, auth.ActionAgentOperate)
	}
	if needProxyOperate {
		actions = append(actions, auth.ActionProxyOperate)
	}

	if len(actions) == 0 {
		return []auth.Action{auth.ActionAgentOperate, auth.ActionProxyOperate}
	}

	return actions
}
