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

package ops

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// narrowAuthorizedBizIDs intersects the history permissions required by the queried node roles.
// An unrestricted scope still preserves explicitly requested business IDs.
func (h *handler) narrowAuthorizedBizIDs(rCtx restserver.IContext, requestedIDs []int64, roles []string) ([]int64, bool, error) {
	var (
		mergedIDs        []int64
		mergedScopeIsAny bool
	)

	actions := nodeHistoryBizActions(roles)
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
			mergedIDs, mergedScopeIsAny, narrowedIDs, scopeIsAny, requestedIDs,
		)
	}

	return mergedIDs, mergedScopeIsAny, nil
}

// narrowHistoryConditionByBiz updates the history condition with narrowed business IDs.
func narrowHistoryConditionByBiz(condition *types.NodeOpsHistoryCondition, narrowedBizIDs []int64, scopeIsAny bool) *types.NodeOpsHistoryCondition {
	if scopeIsAny {
		return condition
	}
	if condition == nil {
		condition = &types.NodeOpsHistoryCondition{}
	}
	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.NodeOpsHistoryExactFields{}
	}
	condition.ExactInclude.BizID = conv.SliceUnique(narrowedBizIDs)

	return condition
}

func (h *handler) checkHistoryViewPermission(rCtx restserver.IContext, history *types.NodeOpsHistory) error {
	resources := authRouter.BuildBizResources(history.Info.BizIDs...)
	actions := nodeHistoryBizActions([]string{string(history.Info.NodeRole)})
	actionResources := make(map[auth.Action][]types.AuthResource, len(actions))
	for _, action := range actions {
		// Empty archived businesses still require an authorization check.
		actionResources[action] = resources
	}

	return h.authorizer.CheckMany(rCtx, actionResources)
}

func nodeHistoryBizActions(roles []string) []auth.Action {
	if len(roles) == 0 {
		return []auth.Action{auth.ActionAgentHistoryView, auth.ActionProxyHistoryView}
	}

	needAgentView, needProxyView := false, false
	for _, role := range roles {
		switch types.NodeRole(role) {
		case types.NodeRoleBlank, types.NodeRoleAgent:
			needAgentView = true
		case types.NodeRoleProxy:
			needProxyView = true
		default:
			// Unknown roles require both permissions, matching workflow history.
			needAgentView, needProxyView = true, true
		}
	}

	actions := make([]auth.Action, 0)
	if needAgentView {
		actions = append(actions, auth.ActionAgentHistoryView)
	}
	if needProxyView {
		actions = append(actions, auth.ActionProxyHistoryView)
	}

	return actions
}

func (h *handler) narrowAuthorizedBizIDsByAction(rCtx restserver.IContext, action auth.Action, requestedIDs []int64) ([]int64, bool, error) {
	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, action, types.AuthResourceTypeBiz)
	if err != nil {
		return nil, false, err
	}

	narrowedIDs, scopeIsAny, scopeIsEmpty, err := auth.ResolveAuthorizedResourceIDsInt64(
		scope, requestedIDs, types.AuthResourceTypeBiz,
	)
	if err != nil {
		return nil, false, err
	}
	if scopeIsAny {
		return requestedIDs, true, nil
	}
	if scopeIsEmpty {
		if err := h.authorizer.Check(rCtx, action, authRouter.BuildBizResources(requestedIDs...)); err != nil {
			return nil, false, err
		}
	}

	return narrowedIDs, false, nil
}

// mergeNarrowedBizIDs preserves the workflow list's intersection semantics.
func mergeNarrowedBizIDs(leftIDs []int64, leftScopeIsAny bool, rightIDs []int64, rightScopeIsAny bool, requestedIDs []int64) ([]int64, bool) {
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
