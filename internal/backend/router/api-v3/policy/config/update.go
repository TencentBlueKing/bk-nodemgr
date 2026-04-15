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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// UpdateConfigPolicy updates config policy.
func (h *handler) UpdateConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyUpdateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// Get config policy to check permission.
	policies, err := h.getConfigPolicy(rCtx, []int64{req.GetConfigpolicyId()})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update config policy, failed to get policy info")
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// Check permission with biz resource.
	resources := authRouter.BuildBizResources(policies[0].BizID)
	if authErr := h.authorizer.Check(rCtx, auth.ActionConfigPolicyManage, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to update config policy, permission denied")
		return nil, errf.ErrWrap(errf.PermissionDenied, authErr)
	}

	configPolicy := req.ConvertConfigPolicyToTypes()
	configPolicy.TenantID = rCtx.TenantID()
	if err := h.storageConfigPolicy.UpdateConfigPolicy(rCtx, configPolicy); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordConfigPolicyEventsByPolicyIDs(rCtx, types.ConfigPolicyEventTypeUpdate, configPolicy.ID)

	resp := new(protoBackend.ConfigPolicyUpdateResp)
	resp.ConvertConfigPolicyID(req.GetConfigpolicyId())

	return resp.GetData(), nil
}
