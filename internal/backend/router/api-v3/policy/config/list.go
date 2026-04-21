/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package config

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// ListConfigPolicy lists config policy with page and conditions.
func (h *handler) ListConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy, failed to convert conditions")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	// Narrow authorized biz IDs.
	var bizIDs []int64
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		bizIDs = exactCond.GetBkBizId()
	}
	narrowedIDs, scopeIsAny, authErr := h.narrowAuthorizedBizIDs(rCtx, bizIDs)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list config policy, permission denied")
		return nil, errf.ErrWrap(errf.PermissionDenied, authErr)
	}
	conditions = narrowConfigPolicyCondition(conditions, narrowedIDs, scopeIsAny)

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storageConfigPolicy.CountConfigPolicy(
			rCtx,
			conditions)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy. failed to count host")

			return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.ConfigPolicyListResp)
		resp.ConvertConfigPoliciesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy, invalid page info")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	hosts, num, err := h.storageConfigPolicy.ListConfigPolicy(rCtx, page, conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyListResp)
	resp.ConvertConfigPoliciesFromTypes(num, hosts)

	return resp.GetData(), nil
}
