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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// Execute executes deploy policy.
func (h *handler) Execute(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.DeployPolicyExecuteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to execute deploy policy, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	triggerID, err := h.backendHandler.ExecuteDeployPolicy(rCtx, req.GetDeployPolicyId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to execute deploy policy")
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.DeployPolicyExecuteResp)
	resp.ConvertTriggerID(triggerID)

	return resp.GetData(), nil
}
