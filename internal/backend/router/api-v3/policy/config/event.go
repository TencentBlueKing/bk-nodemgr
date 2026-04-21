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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListConfigPolicyEvent lists events with page and conditions.
func (h *handler) ListConfigPolicyEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyEventListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list policy event, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list policy event, failed to convert conditions")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// Narrow authorized biz IDs.
	var bizIDs []int64
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		bizIDs = exactCond.GetBkBizId()
	}
	narrowedIDs, scopeIsAny, authErr := h.narrowAuthorizedBizIDs(rCtx, bizIDs)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list policy event, permission denied")
		return nil, errf.ErrWrap(errf.PermissionDenied, authErr)
	}
	conditions = narrowConfigPolicyEventCondition(conditions, narrowedIDs, scopeIsAny)

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storageConfigPolicy.CountConfigPolicyEvent(
			rCtx,
			conditions,
		)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list policy event, failed to count event")

			return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.ConfigPolicyEventListResp)
		resp.ConvertConfigPolicyEventsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list policy event, invalid page info")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	events, num, err := h.storageConfigPolicy.ListConfigPolicyEvent(rCtx, page, conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list policy event")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyEventListResp)
	resp.ConvertConfigPolicyEventsFromTypes(num, events)

	return resp.GetData(), nil
}

// DistinctConfigPolicyEvent distincts events with conditions.
// nolint: dupl
func (h *handler) DistinctConfigPolicyEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyEventDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct policy event, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct policy event, failed to convert conditions")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	result, err := h.storageConfigPolicy.DistinctConfigPolicyEvent(
		rCtx,
		types.NewConfigPolicyEventDistinctRequestAllSet(),
		conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct policy event, failed to distinct host fields")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyEventDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
