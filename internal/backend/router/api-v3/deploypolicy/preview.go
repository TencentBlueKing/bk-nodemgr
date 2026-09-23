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

package deploypolicy

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// Preview synchronously previews deploy policy states without executing changes.
func (h *handler) Preview(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.DeployPolicyPreviewReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to preview deploy policy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to preview deploy policy, failed to validate request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	deployPolicyID := req.GetDeployPolicyId()

	// 1. get deploy policy by id
	deployPolicy, err := h.daoDeployPolicy.GetDeployPolicyByID(rCtx, deployPolicyID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to preview deploy policy, failed to get deploy policy by id")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// 2. check deploy policy is enabled
	if !deployPolicy.Enabled {
		logger.G.Biz(rCtx).Error("failed to preview deploy policy, deploy policy is not enabled")
		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("deploy policy is not enabled"))
	}

	items, err := h.deployPolicyMgr.PreviewDeployPolicy(rCtx, deployPolicy)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to preview deploy policy")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.DeployPolicyPreviewResp)
	if err := resp.ConvertPreviewFromTypes(items); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to convert deploy policy preview")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	return resp.GetData(), nil
}
