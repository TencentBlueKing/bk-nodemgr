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

package config

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (h *handler) getConfigPolicy(nCtx contextx.IContext, configpolicyID []int64) ([]*types.ConfigPolicy, error) {
	// get the config policy info.
	configpolicies, _, err := h.storageConfigPolicy.ListConfigPolicy(nCtx, types.UnlimitedPage(),
		&types.ConfigPolicyCondition{
			ExactInclude: &types.ConfigPolicyExactFields{
				ConfigPolicyID: configpolicyID,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("failed to get config policy: %w", err)
	}

	if len(configpolicies) == 0 {
		return nil, fmt.Errorf("config policy not found")
	}

	return configpolicies, nil
}

// narrowAuthorizedBizIDs narrows the requested biz IDs to those the user is authorized to view config policies for.
func (h *handler) narrowAuthorizedBizIDs(
	rCtx restserver.IContext, requestedIDs []int64,
) ([]int64, bool, error) {

	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, auth.ActionConfigPolicyView, types.AuthResourceTypeBiz)
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
		resources := authRouter.BuildBizResources(requestedIDs...)
		if checkErr := h.authorizer.Check(rCtx, auth.ActionConfigPolicyView, resources); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return narrowedIDs, false, nil
}

// narrowAuthorizedBizIDsForHistory narrows the requested biz IDs to those the user is authorized to view config policy history for.
func (h *handler) narrowAuthorizedBizIDsForHistory(
	rCtx restserver.IContext, requestedIDs []int64,
) ([]int64, bool, error) {

	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, auth.ActionConfigPolicyHistoryView, types.AuthResourceTypeBiz)

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
		resources := authRouter.BuildBizResources(requestedIDs...)
		if checkErr := h.authorizer.Check(rCtx, auth.ActionConfigPolicyHistoryView, resources); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return narrowedIDs, false, nil
}

// narrowConfigPolicyCondition narrows the config policy condition by authorized biz IDs.
func narrowConfigPolicyCondition(condition *types.ConfigPolicyCondition, narrowedIDs []int64, scopeIsAny bool) *types.ConfigPolicyCondition {
	if scopeIsAny {
		return condition
	}

	if condition == nil {
		condition = &types.ConfigPolicyCondition{}
	}
	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.ConfigPolicyExactFields{}
	}
	condition.ExactInclude.BizID = conv.SliceUnique(narrowedIDs)

	return condition
}

// narrowConfigPolicyEventCondition narrows the config policy event condition by authorized biz IDs.
func narrowConfigPolicyEventCondition(condition *types.ConfigPolicyEventCondition, narrowedIDs []int64, scopeIsAny bool,
) *types.ConfigPolicyEventCondition {

	if scopeIsAny {
		return condition
	}

	if condition == nil {
		condition = &types.ConfigPolicyEventCondition{}
	}
	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.ConfigPolicyEventExactFields{}
	}
	condition.ExactInclude.BizID = conv.SliceUnique(narrowedIDs)

	return condition
}
