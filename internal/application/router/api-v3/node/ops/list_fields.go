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

// ListFieldsAgent reads current agent settings from backend.
func (h *handler) ListFieldsAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeOpsFieldsListAgentReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list agent node ops fields, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, err := h.backendHandler.ListAgentOpsFields(rCtx, req.GetBkHostId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list agent node ops fields")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.NodeOpsFieldsListAgentResp)
	resp.ConvertHostsFromTypes(hosts)

	return resp.GetData(), nil
}

// ListFieldsProxy reads current proxy settings from backend.
func (h *handler) ListFieldsProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeOpsFieldsListProxyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list proxy node ops fields, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, err := h.backendHandler.ListProxyOpsFields(rCtx, req.GetBkHostId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list proxy node ops fields")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.NodeOpsFieldsListProxyResp)
	resp.ConvertHostsFromTypes(hosts)

	return resp.GetData(), nil
}
