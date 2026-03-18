/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package agent

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// AssignUnit batch-assigns a network unit to unassigned hosts.
func (h *handler) AssignUnit(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeAgentAssignUnitReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	successCount, failedCount, failedReasons, err := h.backendHandler.AssignUnitAgent(
		rCtx, req.GetBkHostId(), req.GetBkNetworkunitId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit agent")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("success_count", successCount).With("failed_count", failedCount).
		Info("batch assign unit completed")

	resp := new(protoApplication.NodeAgentAssignUnitResp)
	resp.ConvertResult(successCount, failedCount, failedReasons)

	return resp.GetData(), nil
}
