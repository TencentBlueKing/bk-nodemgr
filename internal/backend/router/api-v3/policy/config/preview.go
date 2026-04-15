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

// PreviewConfigPolicy previews the merged config for each host.
func (h *handler) PreviewConfigPolicy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ConfigPolicyPreviewReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to preview config policy, failed to decode request body")

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	policyType := types.ConfigPolicyType(req.GetPolicyType())
	previewHosts := req.ConvertPreviewHostsToTypes()

	// Check permission with biz resource.
	resources := authRouter.BuildBizResources(req.GetBkBizId())
	if authErr := h.authorizer.Check(rCtx, auth.ActionConfigPolicyView, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to preview config policy, permission denied")
		return nil, errf.ErrWrap(errf.PermissionDenied, authErr)
	}

	results, err := h.storageConfigPolicy.PreviewConfigPolicy(rCtx, req.GetBkBizId(), policyType, previewHosts)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to preview config policy")

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ConfigPolicyPreviewResp)
	resp.ConvertMatchResultsFromTypes(results)

	return resp.GetData(), nil
}
