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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// DeleteConfigPolicy deletes config policy.
func (h *handler) DeleteConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// get the config policy info.
	configpolicy, err := h.getConfigPolicy(rCtx, req.GetConfigpolicyId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// Check permission with biz resources.
	bizIDs := h.getDeleteConfigPolicyBizIDs(configpolicy)
	resources := authRouter.BuildBizResources(bizIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionConfigPolicyManage, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to delete config policy, permission denied")
		return nil, errf.ErrWrap(errf.PermissionDenied, authErr)
	}

	// delete the config policy.
	if err := h.storageConfigPolicy.DeleteManyConfigPolicy(rCtx, req.GetConfigpolicyId()...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordConfigPolicyEventsByPolicies(rCtx, types.ConfigPolicyEventTypeDelete, configpolicy...)

	resp := new(protoBackend.ConfigPolicyDeleteResp)

	return resp.GetData(), nil
}

// getDeleteConfigPolicyBizIDs extracts unique biz IDs from config policies for delete operation.
func (h *handler) getDeleteConfigPolicyBizIDs(policies []*types.ConfigPolicy) []int64 {
	bizIDs := make(map[int64]struct{})
	for _, p := range policies {
		bizIDs[p.BizID] = struct{}{}
	}

	return conv.MapKeyToSlice(bizIDs)
}
