/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"fmt"

	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	pluginInstallerStatusSuccess = "success"
	pluginInstallerStatusFailed  = "failed"
)

// ReportStatus report plugin installer status.
func (h *handler) ReportStatus(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoCallback.PluginReportStatusReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to report status, failed to decode request body: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	token := req.GetToken()
	operInstID := req.GetOperInstId()
	status := req.GetStatus()

	h.logger.Infof("operatin instance(%s), status(%s)", operInstID, status)

	var state action.State
	switch status {
	case pluginInstallerStatusSuccess:
		state = action.StateSuccess
	case pluginInstallerStatusFailed:
		state = action.StateFailed
	default:
		h.logger.Errorf("failed to report status, invalid status: %s", status)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("invalid status: %s", req.GetStatus()))
	}

	info, err := h.daoPluginDeployment.GetPluginDeploymentInfo(rCtx, token)
	if err != nil {
		h.logger.Errorf("failed to report status, failed to get info: %v", err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err = h.stgWorkflow.UpdateOperInstActionStatus(rCtx, operInstID, info.BlockingActionName, state); err != nil {
		h.logger.Errorf("failed to report status, failed to update action status: %v", err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoCallback.PluginReportStatusResp)

	return resp.GetData(), nil
}
