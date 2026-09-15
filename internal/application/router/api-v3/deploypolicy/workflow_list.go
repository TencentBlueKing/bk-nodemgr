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

// WorkflowList lists tenant-scoped deploy policy workflows through the backend.
func (h *handler) WorkflowList(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.DeployPolicyWorkflowListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list deploy policy workflows, failed to decode request body")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	condition, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list deploy policy workflows, failed to convert conditions")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	if req.GetOnlyCount() {
		total, err := h.backendHandler.CountDeployPolicyWorkflows(rCtx, condition)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count deploy policy workflows")
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}
		resp := new(protoApplication.DeployPolicyWorkflowListResp)
		resp.ConvertDeployPolicyWorkflowsFromTypes(total, nil)

		return resp.GetData(), nil
	}
	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list deploy policy workflows, invalid page info")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	workflows, total, err := h.backendHandler.ListDeployPolicyWorkflows(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list deploy policy workflows")
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}
	resp := new(protoApplication.DeployPolicyWorkflowListResp)
	resp.ConvertDeployPolicyWorkflowsFromTypes(total, workflows)

	return resp.GetData(), nil
}
