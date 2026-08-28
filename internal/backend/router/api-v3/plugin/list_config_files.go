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
	backendAuth "github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListConfigFiles defines the handler to list plugin config files.
func (h *handler) ListConfigFiles(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginListConfigFilesReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin config files, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	host, err := h.daoHost.GetHostByID(rCtx, req.GetBkHostId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin config files, failed to get host topo mapping")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := h.authorizer.Check(rCtx, backendAuth.ActionPluginView, authRouter.BuildBizResources(host.Static.BizID)); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin config files, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	configFiles, _, err := h.daoProcessConfig.ListProcessConfigs(rCtx, types.UnlimitedPage(), &types.ProcessConfigCondition{
		ExactInclude: &types.ProcessConfigExactFields{
			ProcessName: []string{req.GetPluginName()},
			HostID:      []int64{req.GetBkHostId()},
		},
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin config files")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PluginListConfigFilesResp)
	if err := resp.ConvertConfigFilesFromTypes(configFiles); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin config files, failed to convert response")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	return resp.GetData(), nil
}
