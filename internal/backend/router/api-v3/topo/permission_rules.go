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
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var (
	errNetworkUnitViewDeniedByEmptyScope = errors.New("no authorized network units")
	errBizViewDeniedByEmptyScope         = errors.New("no authorized businesses")
	errAccessPointViewDeniedByEmptyScope = errors.New("no authorized access points")
)

// buildBizResources is deprecated. Use auth.BuildBizResources instead.
func buildBizResources(bizIDs []int64) []types.AuthResource {
	return authRouter.BuildBizResources(bizIDs)
}

// buildNetworkAreaResources constructs IAM resource descriptors for the given network area IDs.
func buildNetworkAreaResources(ids []int64) []types.AuthResource {
	resources := make([]types.AuthResource, 0, len(ids))
	for _, id := range ids {
		resources = append(resources, types.AuthResource{
			SystemID: types.SystemIDNodeMgr,
			Type:     types.AuthResourceTypeNetworkArea,
			ID:       fmt.Sprintf("%d", id),
		})
	}

	return resources
}

// buildNetworkUnitResources is deprecated. Use auth.BuildNetworkUnitResources instead.
func buildNetworkUnitResources(ids []int64) []types.AuthResource {
	return authRouter.BuildNetworkUnitResources(ids)
}

func (h *handler) narrowAuthorizedNetworkUnitIDs(
	rCtx restserver.IContext, requestedIDs []int64,
) ([]int64, bool, error) {

	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, auth.ActionNetworkUnitView, types.AuthResourceTypeNetworkUnit)

	if err != nil {
		return nil, false, err
	}

	narrowedIDs, scopeIsAny, hasAuthorized, err := auth.ResolveAuthorizedResourceIDsInt64(
		scope, requestedIDs, types.AuthResourceTypeNetworkUnit,
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

	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, action, types.AuthResourceTypeBiz)

	if err != nil {
		return nil, false, err
	}

	narrowedIDs, scopeIsAny, hasAuthorized, err := auth.ResolveAuthorizedResourceIDsInt64(
		scope, requestedIDs, types.AuthResourceTypeBiz,
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

	// Pre-allocate for at most 2 actions: AgentView and ProxyView
	const maxActions = 2
	actions := make([]auth.Action, 0, maxActions)
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
		actionResources := make(map[auth.Action][]types.AuthResource, len(actions))
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

func (h *handler) narrowAuthorizedAccessPointIDs(
	rCtx restserver.IContext, requestedIDs []int64,
) ([]int64, bool, error) {

	// AccessPoint authorization is based on NetworkUnit ownership.
	// Step 1: If specific AccessPoint IDs are requested, find which NetworkUnits own them.
	var targetNetworkUnitIDs []int64
	if len(requestedIDs) > 0 {
		var err error
		targetNetworkUnitIDs, err = h.storage.GetNetworkUnitIDsByAccessPoints(rCtx, requestedIDs)
		if err != nil {
			return nil, false, err
		}
		if len(targetNetworkUnitIDs) == 0 {
			// Requested AccessPoints don't exist or don't belong to any NetworkUnit.
			return nil, false, errAccessPointViewDeniedByEmptyScope
		}
	}

	// Step 2: Get authorized NetworkUnit IDs (either for specific units or all units).
	authorizedNetworkUnitIDs, scopeIsAny, err := h.narrowAuthorizedNetworkUnitIDs(rCtx, targetNetworkUnitIDs)
	if err != nil {
		return nil, false, err
	}

	// Step 3: If user has full access to all NetworkUnits, return requested IDs or all.
	if scopeIsAny {
		return requestedIDs, true, nil
	}

	// Step 4: Query authorized NetworkUnits to get their AccessPoints.
	if len(authorizedNetworkUnitIDs) == 0 {
		return nil, false, errAccessPointViewDeniedByEmptyScope
	}

	networkUnits, err := h.storage.GetNetworkUnitByIDs(rCtx, authorizedNetworkUnitIDs)
	if err != nil {
		return nil, false, err
	}

	// Step 5: Extract all AccessPoint IDs from authorized NetworkUnits.
	authorizedAccessPointIDs := make([]int64, 0)
	for _, unit := range networkUnits {
		if unit != nil && len(unit.AccessPoints) > 0 {
			authorizedAccessPointIDs = append(authorizedAccessPointIDs, unit.AccessPoints...)
		}
	}

	// Step 6: If no specific IDs requested, return all authorized AccessPoint IDs.
	if len(requestedIDs) == 0 {
		return conv.SliceUnique(authorizedAccessPointIDs), false, nil
	}

	// Step 7: Intersect requested IDs with authorized IDs.
	narrowedIDs := conv.SliceIntersect(requestedIDs, authorizedAccessPointIDs)

	// Step 8: If requested specific IDs but none are authorized, build resources and check.
	if len(narrowedIDs) == 0 {
		// Build NetworkUnit resources for permission denied error.
		if len(targetNetworkUnitIDs) > 0 {
			resources := buildNetworkUnitResources(targetNetworkUnitIDs)
			if checkErr := h.authorizer.Check(rCtx, auth.ActionNetworkUnitView, resources); checkErr != nil {
				return nil, false, checkErr
			}
		}
	}

	return conv.SliceUnique(narrowedIDs), false, nil
}

func narrowAccessPointCondition(condition *types.AccessPointCondition, narrowedIDs []int64, scopeIsAny bool) *types.AccessPointCondition {
	if scopeIsAny {
		return condition
	}

	if condition == nil {
		condition = &types.AccessPointCondition{}
	}
	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.AccessPointExactFields{}
	}
	condition.ExactInclude.AccessPointID = conv.SliceUnique(narrowedIDs)

	return condition
}
