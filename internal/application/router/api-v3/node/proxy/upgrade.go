/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package proxy

import (
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// Upgrade proxy.
func (h *handler) Upgrade(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.NodeProxyUpgradeReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upgrade proxy, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.backendHandler.UpgradeProxy(ctx, req.ConvertParamToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upgrade proxy: %v", err)
		return nil, err
	}
	h.logger.InfoCtxf(ctx, "launched proxy upgrade. workflow-id(%s)", workflowID)

	resp := new(protoApplication.NodeProxyUpgradeResp)
	resp.ConvertWorkflowID(workflowID)

	return resp.GetData(), nil
}
