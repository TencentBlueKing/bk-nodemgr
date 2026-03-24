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
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// List defines the handler to list plugins.
func (h *handler) List(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PluginListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugins, failed to decode request query.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if req.GetOnlyCount() {
		cnt, err := h.backendHandler.CountPlugins(rCtx, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugins, failed to count plugins.")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.PluginListResp)
		resp.ConvertPluginFromTypes(cnt, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugins, invalid page info.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	plugins, cnt, err := h.backendHandler.ListPlugins(rCtx, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugins.")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PluginListResp)
	resp.ConvertPluginFromTypes(cnt, plugins)

	return resp.GetData(), nil
}

// ListPerimittedOperations defines the handler to list plugin permitted operations.
func (h *handler) ListPerimittedOperations(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PluginListPermittedOperationReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin permitted operations, failed to decode request query.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugins permitted operations, invalid page info.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	plugins, _, err := h.backendHandler.ListPlugins(rCtx, page, nil)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin permitted operations.")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PluginListPermittedOperationResp)
	resp.ConvertPluginPermittedOperationsFromTypes(plugins...)

	return resp.GetData(), nil
}
