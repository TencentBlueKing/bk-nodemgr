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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Distinct queries distinct fields of visible plugins.
func (h *handler) Distinct(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition := req.ConvertConditionsToTypes()
	narrowedBizIDs, scopeIsAny, err := h.narrowAuthorizedBizIDsForPluginView(
		rCtx, req.GetExactIncludeConditions().GetVisibleBizIds())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}
	condition = narrowPluginConditionByBiz(condition, narrowedBizIDs, scopeIsAny)

	resp := new(protoBackend.PluginDistinctResp)
	if !scopeIsAny && len(narrowedBizIDs) == 0 {
		resp.ConvertResultFromTypes(&types.PluginDistinctResult{})
		return resp.GetData(), nil
	}

	result, err := h.daoPlugin.DistinctPlugin(rCtx, req.ConvertDistinctFieldToTypes(), condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp.ConvertResultFromTypes(result)
	return resp.GetData(), nil
}
