/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package process

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// GetDistributionByHostID defines the process distribution by host id handler.
func (h *handler) GetDistributionByHostID(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.GetProcessDistributionByPluginNameReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get process distribution by plugin name, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.backendHandler.GetProcessDistributionByPluginName(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			Error("failed to get process distribution by plugin name. failed to get process distribution by plugin name fields: %v", err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.GetProcessDistributionByPluginNameResp)

	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// GetDistributionByPluginName defines the process distribution by plugin name handler.
func (h *handler) GetDistributionByPluginName(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.GetProcessDistributionByPluginNameReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get process distribution by plugin name, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.backendHandler.GetProcessDistributionByPluginName(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			Error("failed to get process distribution by plugin name. failed to get process distribution by plugin name fields: %v", err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.GetProcessDistributionByPluginNameResp)

	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
