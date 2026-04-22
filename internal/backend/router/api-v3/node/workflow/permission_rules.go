/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// workflowListBizActions determines which view actions are required based on node roles in the query condition.
// Returns ActionAgentView, ActionProxyView, or both depending on the roles present.
func workflowListBizActions(condition *types.NodeWorkflowCondition) []auth.Action {
	// If no NodeRole filter is specified, need to check both Agent and Proxy permissions
	if condition == nil || condition.ExactInclude == nil || len(condition.ExactInclude.NodeRole) == 0 {
		return []auth.Action{auth.ActionAgentHistoryView, auth.ActionProxyHistoryView}
	}

	// Pre-allocate for at most 2 actions: AgentView and ProxyView
	const maxActions = 2
	actions := make([]auth.Action, 0, maxActions)
	needAgentView := false
	needProxyView := false

	for _, nodeRoleStr := range condition.ExactInclude.NodeRole {
		nodeRole := types.NodeRole(nodeRoleStr)
		switch nodeRole {
		case types.NodeRoleBlank, types.NodeRoleAgent:
			needAgentView = true
		case types.NodeRoleProxy:
			needProxyView = true
		default:
			// Unknown role: require both permissions to be safe
			needAgentView = true
			needProxyView = true
		}
	}

	if needAgentView {
		actions = append(actions, auth.ActionAgentView)
	}
	if needProxyView {
		actions = append(actions, auth.ActionProxyView)
	}

	if len(actions) == 0 {
		return []auth.Action{auth.ActionAgentView, auth.ActionProxyView}
	}

	return actions
}

// narrowAuthorizedBizIDsByAction narrows requested business IDs by authorized scope for a specific action.
func (h *handler) narrowAuthorizedBizIDsByAction(
	rCtx restserver.IContext, action auth.Action, requestedIDs []int64,
) ([]int64, bool, error) {

	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, action, types.AuthResourceTypeBiz)

	if err != nil {
		return nil, false, err
	}

	narrowedIDs, scopeIsAny, err := auth.ResolveAuthorizedResourceIDsInt64(
		scope, requestedIDs, types.AuthResourceTypeBiz,
	)

	if err != nil {
		return nil, false, err
	}

	if scopeIsAny {
		return requestedIDs, true, nil
	}

	if len(narrowedIDs) == 0 {
		if checkErr := h.authorizer.Check(rCtx, action, authRouter.BuildBizResources(requestedIDs...)); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return narrowedIDs, false, nil
}

// mergeNarrowedBizIDs merges two narrowed business ID sets according to their scope flags.
// When both scopes are partial, it returns the intersection (union semantically - user needs either permission).
func mergeNarrowedBizIDs(
	leftIDs []int64, leftScopeIsAny bool,
	rightIDs []int64, rightScopeIsAny bool,
	requestedIDs []int64,
) ([]int64, bool) {

	if leftScopeIsAny && rightScopeIsAny {
		return requestedIDs, true
	}

	if leftScopeIsAny {
		return conv.SliceUnique(rightIDs), false
	}

	if rightScopeIsAny {
		return conv.SliceUnique(leftIDs), false
	}

	return conv.SliceUnique(conv.SliceIntersect(leftIDs, rightIDs)), false
}

// narrowAuthorizedBizIDsForWorkflowList narrows requested business IDs by authorized scope,
// considering the NodeRole filter in the query condition.
func (h *handler) narrowAuthorizedBizIDsForWorkflowList(
	rCtx restserver.IContext, requestedIDs []int64, condition *types.NodeWorkflowCondition,
) ([]int64, bool, error) {

	actions := workflowListBizActions(condition)
	if len(actions) == 1 {
		return h.narrowAuthorizedBizIDsByAction(rCtx, actions[0], requestedIDs)
	}

	var (
		mergedIDs        []int64
		mergedScopeIsAny bool
	)

	for idx, action := range actions {
		narrowedIDs, scopeIsAny, err := h.narrowAuthorizedBizIDsByAction(rCtx, action, requestedIDs)
		if err != nil {
			return nil, false, err
		}

		if idx == 0 {
			mergedIDs = narrowedIDs
			mergedScopeIsAny = scopeIsAny

			continue
		}

		mergedIDs, mergedScopeIsAny = mergeNarrowedBizIDs(
			mergedIDs, mergedScopeIsAny,
			narrowedIDs, scopeIsAny,
			requestedIDs,
		)
	}

	if len(requestedIDs) > 0 && !mergedScopeIsAny && len(mergedIDs) == 0 {
		actionResources := make(map[auth.Action][]types.AuthResource, len(actions))
		resources := authRouter.BuildBizResources(requestedIDs...)
		for _, action := range actions {
			actionResources[action] = resources
		}

		if checkErr := h.authorizer.CheckMany(rCtx, actionResources); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return mergedIDs, mergedScopeIsAny, nil
}

// narrowWorkflowConditionByBiz updates the workflow condition with narrowed business IDs.
func narrowWorkflowConditionByBiz(condition *types.NodeWorkflowCondition, narrowedBizIDs []int64, scopeIsAny bool) *types.NodeWorkflowCondition {
	if scopeIsAny {
		return condition
	}

	if condition == nil {
		condition = &types.NodeWorkflowCondition{}
	}

	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.NodeWorkflowExactFields{}
	}

	condition.ExactInclude.BizID = conv.SliceUnique(narrowedBizIDs)

	return condition
}
