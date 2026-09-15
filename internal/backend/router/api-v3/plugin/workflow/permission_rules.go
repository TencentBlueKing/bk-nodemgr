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

// Package workflow describes the workflow router.
package workflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (h *handler) narrowAuthorizedBizIDsForPluginHistoryView(rCtx restserver.IContext, requestedIDs []int64) ([]int64, bool, error) {
	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, auth.ActionPluginHistoryView, types.AuthResourceTypeBiz)
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
		if checkErr := h.authorizer.Check(rCtx, auth.ActionPluginHistoryView, authRouter.BuildBizResources(requestedIDs...)); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return narrowedIDs, false, nil
}

func narrowPluginConditionByBiz(condition *types.PluginWorkflowCondition, narrowedBizIDs []int64, scopeIsAny bool) *types.PluginWorkflowCondition {
	if scopeIsAny {
		return condition
	}

	if condition == nil {
		condition = &types.PluginWorkflowCondition{}
	}

	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.PluginWorkflowExactFields{}
	}

	condition.ExactInclude.BizID = conv.SliceUnique(narrowedBizIDs)

	return condition
}
