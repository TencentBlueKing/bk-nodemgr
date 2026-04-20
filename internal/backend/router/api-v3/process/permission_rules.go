/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package process defines the process handler.
package process

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (h *handler) narrowAuthorizedBizIDsForPluginView(rCtx restserver.IContext) ([]int64, error) {
	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, auth.ActionPluginView, types.AuthResourceTypeBiz)
	if err != nil {
		return nil, err
	}

	authorizedIDs := make([]int64, 0, len(scope.Resources))
	for _, resource := range scope.Resources {
		if resource.Type != types.AuthResourceTypeBiz {
			continue
		}

		id, convErr := conv.ToInt64(resource.ID)
		if convErr != nil {
			return nil, fmt.Errorf("auth: convert authorized resource id %q: %w", resource.ID, convErr)
		}
		authorizedIDs = append(authorizedIDs, id)
	}
	authorizedIDs = conv.SliceUnique(authorizedIDs)

	return authorizedIDs, nil
}

func narrowProcessConditionByPluginNames(condition *types.ProcessCondition, pluginNames []string) *types.ProcessCondition {
	if condition == nil {
		condition = &types.ProcessCondition{}
	}

	pluginNames = conv.SliceUnique(pluginNames)
	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.ProcessExactFields{
			PluginName: pluginNames,
		}
	}

	condition.ExactInclude.PluginName = conv.SliceIntersect(condition.ExactInclude.PluginName, pluginNames)

	return condition
}
