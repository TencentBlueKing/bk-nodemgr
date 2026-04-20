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
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var (
	errBizViewDeniedByEmptyScope = errors.New("no authorized businesses")
)

func (h *handler) narrowAuthorizedPluginNamesForView(rCtx restserver.IContext) ([]string, bool, error) {
	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, auth.ActionPluginView, types.AuthResourceTypeBiz)
	if err != nil {
		return nil, false, err
	}

	narrowedIDs, scopeIsAny, hasAuthorized, err := auth.ResolveAuthorizedResourceIDsInt64(
		scope, nil, types.AuthResourceTypeBiz,
	)
	if err != nil {
		return nil, false, err
	}

	if !hasAuthorized {
		if checkErr := h.authorizer.Check(rCtx, auth.ActionPluginView, nil); checkErr != nil {
			return nil, false, checkErr
		}

		return nil, false, errBizViewDeniedByEmptyScope
	}

	if scopeIsAny {
		return nil, true, nil
	}

	authorizedPlugin, err := h.domainProcess.ListVisiblePluginByBizIDs(rCtx, narrowedIDs)
	if err != nil {
		return nil, false, err
	}

	pluginNames := conv.SliceToSlice(authorizedPlugin, func(item *types.Plugin) string {
		return item.Name
	})

	return pluginNames, false, nil
}

func narrowProcessCondition(condition *types.ProcessCondition, pluginName []string, scopeIsAny bool) *types.ProcessCondition {
	if condition == nil {
		condition = &types.ProcessCondition{}
	}

	if scopeIsAny {
		return condition
	}

	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.ProcessExactFields{
			PluginName: pluginName,
		}
	}

	condition.ExactInclude.PluginName = conv.SliceIntersect(condition.ExactInclude.PluginName, pluginName)

	return condition
}
