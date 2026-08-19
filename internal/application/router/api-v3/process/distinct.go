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

// Package process define process api v3 router.
package process

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Distinct defines the process distinct handler.
//
// When plugin_name is set in the request, the result is served from the
// in-memory distinct cache (keyed by tenant id x plugin name, refreshed by a
// background goroutine). Otherwise the handler falls back to querying the
// backend directly with the provided conditions (e.g. for per-host distinct).
func (h *handler) Distinct(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.DistinctProcessReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to process distinct, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if pluginName := req.GetPluginName(); pluginName != "" {
		result := h.distinctCache.GetProcess(rCtx.TenantID(), pluginName)
		if result == nil {
			// cold-start window before the first sync completes; real sync
			// failures are already logged at Error level in the syncer.
			logger.G.Biz(rCtx).
				With("tenant-id", rCtx.TenantID(), "plugin-name", pluginName).
				Debug("process distinct cache miss, return empty result")
			result = new(types.ProcessDistinctResult)
		}

		resp := new(protoApplication.DistinctProcessResp)
		resp.ConvertResultFromTypes(result)

		return resp.GetData(), nil
	}

	result, err := h.backendHandler.DistinctProcess(rCtx, types.NewProcessDistinctSelectorAllSet(), req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to process distinct, failed to distinct process fields")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.DistinctProcessResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
