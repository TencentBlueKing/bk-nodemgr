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

package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ReportStatus report plugin installer status.
func (h *handler) ReportStatus(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoCallback.PluginReportStatusReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to report status, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	token := req.GetToken()
	operInstID := req.GetOperInstId()
	status := req.GetStatus()

	info, err := h.daoPluginDeployment.GetPluginDeploymentInfo(rCtx, token)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to report status, failed to get info")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	dataMap := map[string]any{
		types.PDKeyInstallerReportStatus: status,
	}

	if err := h.stgWorkflow.UpsertActionInstancePrivateData(rCtx, req.GetOperInstId(), info.BlockingActionName, dataMap); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to report result, failed to update action private data")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	logger.G.Biz(rCtx).
		With("oper-inst-id", operInstID, "action", info.BlockingActionName, "status", status).
		Info("report status")

	resp := new(protoCallback.PluginReportStatusResp)

	return resp.GetData(), nil
}
