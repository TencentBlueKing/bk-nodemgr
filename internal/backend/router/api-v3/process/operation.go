/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package process defines the process handler.
package process

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

const (
	// not max limit in process.
	// return all data in one request.
	maxProcessLimit = 0
)

// List defines the process list handler.
func (h *handler) List(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ProcessListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list processes, failed to decode request query.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if req.GetOnlyCount() {
		cnt, err := h.daoProcess.CountProcesses(
			rCtx,
			req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list processes, failed to count processes.")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.ProcessListResp)
		resp.ConvertProcessFromTypes(cnt, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes(maxProcessLimit)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list processes, invalid page info.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	processes, cnt, err := h.daoProcess.ListProcesses(rCtx, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list processes.")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ProcessListResp)
	resp.ConvertProcessFromTypes(cnt, processes)

	return resp.GetData(), nil
}

// GetDistributionByHostID defines the process distribution by host id handler.
func (h *handler) GetDistributionByHostID(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.GetProcessDistributionByHostIDReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get process distribution by host id, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.daoProcess.GetProcessDistributionByHostID(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			Error("failed to get process distribution by host id. failed to get process distribution by host id fields: %v", err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.GetProcessDistributionByHostIDResp)

	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// GetDistributionByPluginName defines the process distribution by plugin name handler.
func (h *handler) GetDistributionByPluginName(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.GetProcessDistributionByPluginNameReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get process distribution by plugin name, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.daoProcess.GetProcessDistributionByPluginName(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			Error("failed to get process distribution by plugin name. failed to get process distribution by plugin name fields: %v", err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.GetProcessDistributionByPluginNameResp)

	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// Distinct defines the process distinct handler.
func (h *handler) Distinct(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ProcessDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to process distinct, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.daoProcess.DistinctProcess(rCtx,
		req.ConvertSelectorToTypes(),
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to process distinct, failed to distinct process fields")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.ProcessDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
