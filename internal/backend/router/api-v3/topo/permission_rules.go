/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	routerAuth "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var (
	errNetworkAreaViewDeniedByEmptyScope = errors.New("no authorized network areas")
	errNetworkUnitViewDeniedByEmptyScope = errors.New("no authorized network units")
	errBizViewDeniedByEmptyScope         = errors.New("no authorized businesses")
)

// buildBizResources is deprecated. Use auth.BuildBizResources instead.
func buildBizResources(bizIDs []int64) []auth.Resource {
	return routerAuth.BuildBizResources(bizIDs)
}

// buildNetworkAreaResources constructs IAM resource descriptors for the given network area IDs.
func buildNetworkAreaResources(ids []int64) []auth.Resource {
	resources := make([]auth.Resource, 0, len(ids))
	for _, id := range ids {
		resources = append(resources, auth.Resource{
			SystemID: auth.SystemIDNodeMgr,
			Type:     auth.ResourceTypeNetworkArea,
			ID:       fmt.Sprintf("%d", id),
		})
	}

	return resources
}

// buildNetworkUnitResources is deprecated. Use auth.BuildNetworkUnitResources instead.
func buildNetworkUnitResources(ids []int64) []auth.Resource {
	return routerAuth.BuildNetworkUnitResources(ids)
}

func (h *handler) narrowAuthorizedNetworkAreaIDs(
	rCtx restserver.IContext, requestedIDs []int64,
) ([]int64, bool, error) {

	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, auth.ActionNetworkAreaView, auth.ResourceTypeNetworkArea)

	if err != nil {
		return nil, false, err
	}

	narrowedIDs, scopeIsAny, hasAuthorized, err := auth.ResolveAuthorizedResourceIDsInt64(
		scope, requestedIDs, auth.ResourceTypeNetworkArea,
	)

	if err != nil {
		return nil, false, err
	}

	if !hasAuthorized {
		if checkErr := h.authorizer.Check(rCtx, auth.ActionNetworkAreaView, nil); checkErr != nil {
			return nil, false, checkErr
		}

		return nil, false, errNetworkAreaViewDeniedByEmptyScope
	}

	if scopeIsAny {
		return requestedIDs, true, nil
	}

	if len(requestedIDs) > 0 && len(narrowedIDs) == 0 {
		if checkErr := h.authorizer.Check(rCtx, auth.ActionNetworkAreaView, buildNetworkAreaResources(requestedIDs)); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return narrowedIDs, false, nil
}

func narrowNetworkAreaCondition(condition *types.NetworkAreaCondition, narrowedIDs []int64, scopeIsAny bool) *types.NetworkAreaCondition {
	if scopeIsAny {
		return condition
	}

	if condition == nil {
		condition = &types.NetworkAreaCondition{}
	}
	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.NetworkAreaExactFields{}
	}
	condition.ExactInclude.NetworkAreaID = conv.SliceUnique(narrowedIDs)

	return condition
}

func (h *handler) narrowAuthorizedNetworkUnitIDs(
	rCtx restserver.IContext, requestedIDs []int64,
) ([]int64, bool, error) {

	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, auth.ActionNetworkUnitView, auth.ResourceTypeNetworkUnit)

	if err != nil {
		return nil, false, err
	}

	narrowedIDs, scopeIsAny, hasAuthorized, err := auth.ResolveAuthorizedResourceIDsInt64(
		scope, requestedIDs, auth.ResourceTypeNetworkUnit,
	)

	if err != nil {
		return nil, false, err
	}

	if !hasAuthorized {
		if checkErr := h.authorizer.Check(rCtx, auth.ActionNetworkUnitView, nil); checkErr != nil {
			return nil, false, checkErr
		}

		return nil, false, errNetworkUnitViewDeniedByEmptyScope
	}

	if scopeIsAny {
		return requestedIDs, true, nil
	}

	if len(requestedIDs) > 0 && len(narrowedIDs) == 0 {
		if checkErr := h.authorizer.Check(rCtx, auth.ActionNetworkUnitView, buildNetworkUnitResources(requestedIDs)); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return narrowedIDs, false, nil
}

func narrowNetworkUnitCondition(condition *types.NetworkUnitCondition, narrowedIDs []int64, scopeIsAny bool) *types.NetworkUnitCondition {
	if scopeIsAny {
		return condition
	}

	if condition == nil {
		condition = &types.NetworkUnitCondition{}
	}

	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.NetworkUnitExactFields{}
	}

	condition.ExactInclude.NetworkUnitID = conv.SliceUnique(narrowedIDs)

	return condition
}

func (h *handler) narrowAuthorizedBizIDsByAction(
	rCtx restserver.IContext, action auth.Action, requestedIDs []int64,
) ([]int64, bool, error) {

	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, action, auth.ResourceTypeBiz)

	if err != nil {
		return nil, false, err
	}

	narrowedIDs, scopeIsAny, hasAuthorized, err := auth.ResolveAuthorizedResourceIDsInt64(
		scope, requestedIDs, auth.ResourceTypeBiz,
	)

	if err != nil {
		return nil, false, err
	}

	if !hasAuthorized {
		if checkErr := h.authorizer.Check(rCtx, action, nil); checkErr != nil {
			return nil, false, checkErr
		}

		return nil, false, errBizViewDeniedByEmptyScope
	}

	if scopeIsAny {
		return requestedIDs, true, nil
	}

	if len(requestedIDs) > 0 && len(narrowedIDs) == 0 {
		if checkErr := h.authorizer.Check(rCtx, action, buildBizResources(requestedIDs)); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return narrowedIDs, false, nil
}

func hostListBizActions(condition *types.HostCondition) []auth.Action {
	if condition == nil || condition.DynamicExactInclude == nil || len(condition.DynamicExactInclude.NodeRole) == 0 {
		return []auth.Action{auth.ActionAgentView, auth.ActionProxyView}
	}

	actions := make([]auth.Action, 0, 2)
	needAgentView := false
	needProxyView := false

	for _, nodeRole := range condition.DynamicExactInclude.NodeRole {
		switch nodeRole {
		case types.NodeRoleBlank, types.NodeRoleAgent:
			needAgentView = true
		case types.NodeRoleProxy:
			needProxyView = true
		default:
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

func (h *handler) narrowAuthorizedBizIDsForHostList(
	rCtx restserver.IContext, requestedIDs []int64, condition *types.HostCondition,
) ([]int64, bool, error) {
	actions := hostListBizActions(condition)
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
		actionResources := make(map[auth.Action][]auth.Resource, len(actions))
		resources := buildBizResources(requestedIDs)
		for _, action := range actions {
			actionResources[action] = resources
		}

		if checkErr := h.authorizer.CheckMany(rCtx, actionResources); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return mergedIDs, mergedScopeIsAny, nil
}

func narrowHostConditionByBiz(condition *types.HostCondition, narrowedBizIDs []int64, scopeIsAny bool) *types.HostCondition {
	if scopeIsAny {
		return condition
	}

	if condition == nil {
		condition = &types.HostCondition{}
	}

	if condition.StaticExactInclude == nil {
		condition.StaticExactInclude = &types.HostStaticExactFields{}
	}

	condition.StaticExactInclude.BizID = conv.SliceUnique(narrowedBizIDs)

	return condition
}
