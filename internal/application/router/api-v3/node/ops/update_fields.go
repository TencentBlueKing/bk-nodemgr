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

package ops

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// UpdateFieldsAgent batch-updates out-of-band (ops) fields for hosts.
func (h *handler) UpdateFieldsAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeOpsUpdateFieldsAgentReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update agent ops fields, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.backendHandler.UpdateAgentOpsFields(rCtx, req.ConvertParamToTypes()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update agent ops fields, failed to update host")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("count", len(req.GetHosts())).Info("updated agent ops fields")

	resp := new(protoApplication.NodeOpsUpdateFieldsAgentResp)

	return resp.GetData(), nil
}

// UpdateFieldsProxy batch-updates out-of-band (ops) fields for proxy hosts.
func (h *handler) UpdateFieldsProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeOpsUpdateFieldsProxyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update proxy ops fields, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.backendHandler.UpdateProxyOpsFields(rCtx, req.ConvertParamToTypes()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update proxy ops fields, failed to update host")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("count", len(req.GetHosts())).Info("updated proxy ops fields")

	resp := new(protoApplication.NodeOpsUpdateFieldsProxyResp)

	return resp.GetData(), nil
}
