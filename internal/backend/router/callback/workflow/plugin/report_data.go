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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// ReportData report plugin installer data.
func (h *handler) ReportData(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoCallback.PluginReportDataReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to report data, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	token := req.GetToken()
	info, err := h.daoPluginDeployment.GetPluginDeploymentInfo(rCtx, token)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to report data, get deployment info failed")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := h.daoPluginDeployment.UpdatePluginDeploymentInfo(rCtx, token, info); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to report data, update deployment info failed")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoCallback.PluginReportDataResp)

	return resp.GetData(), nil
}
