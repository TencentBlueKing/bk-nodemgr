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
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var errNetworkAreaViewDeniedByEmptyScope = errors.New("no authorized network areas")

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
