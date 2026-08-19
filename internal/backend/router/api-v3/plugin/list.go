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

package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// List defines the handler to list plugins.
func (h *handler) List(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugins, failed to decode request query.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition := req.ConvertConditionsToTypes()
	var bizIDs []int64
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		bizIDs = exactCond.GetVisibleBizIds()
	}
	narrowedBizIDs, scopeIsAny, authErr := h.narrowAuthorizedBizIDsForPluginView(rCtx, bizIDs)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list plugin, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}
	condition = narrowPluginConditionByBiz(condition, narrowedBizIDs, scopeIsAny)

	if req.GetOnlyCount() {
		cnt, err := h.daoPlugin.CountPlugins(rCtx, condition)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugins, failed to count plugins.")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PluginListResp)
		resp.ConvertPluginFromTypes(cnt, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugins, invalid page info.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	plugins, cnt, err := h.daoPlugin.ListPlugins(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugins.")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PluginListResp)
	resp.ConvertPluginFromTypes(cnt, plugins)

	return resp.GetData(), nil
}
