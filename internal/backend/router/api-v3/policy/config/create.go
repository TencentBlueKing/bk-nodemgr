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

// CreateConfigPolicy creates config policy.
func (h *handler) CreateConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyCreateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// Check permission with biz resource.
	resources := authRouter.BuildBizResources(req.GetBkBizId())
	if authErr := h.authorizer.Check(rCtx, auth.ActionConfigPolicyManage, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to create config policy, permission denied")
		return nil, errf.ErrWrap(errf.PermissionDenied, authErr)
	}

	configPolicy := req.ConvertConfigPolicyToTypes()
	configPolicy.TenantID = rCtx.TenantID()
	configPolicy.Version = initVersion
	configPolicyID, err := h.storageConfigPolicy.CreateConfigPolicy(rCtx, configPolicy)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordConfigPolicyEventsByPolicyIDs(rCtx, types.ConfigPolicyEventTypeCreate, configPolicyID)

	resp := new(protoBackend.ConfigPolicyCreateResp)
	resp.ConvertConfigPolicyID(configPolicyID)

	return resp.GetData(), nil
}
