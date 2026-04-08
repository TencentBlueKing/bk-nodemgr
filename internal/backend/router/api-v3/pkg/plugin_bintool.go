/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pkg

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authProvider "github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/provider"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListReleasePluginBinTool lists plugin-bintool releases.
func (h *handler) ListReleasePluginBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginBinToolListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin-bintool release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Generation: []types.Generation{gen},
		},
	}

	items, num, err := h.daoReleasePluginBinTool.ListReleasePluginBinTool(rCtx, types.UnlimitedPage(), cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin-bintool release")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleasePluginBinToolListResp)
	resp.ConvertReleasesFromTypes(num, items)

	return resp.GetData(), nil
}

// DeleteReleasePluginBinTool deletes plugin-bintool release.
func (h *handler) DeleteReleasePluginBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginBinToolDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete plugin-bintool release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	name := req.GetName()

	// Check permission.
	resources := authProvider.BuildPackageResources(name)
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete plugin-bintool release, permission denied")
		return nil, err
	}

	key := types.ReleasePluginBinToolKey{
		Generation: gen,
		Name:       name,
	}

	if err := h.daoReleasePluginBinTool.DeleteReleasePluginBinTool(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "name", name).
			Error("failed to delete plugin-bintool release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	h.recordPackageEvents(rCtx, &types.PackageEvent{
		Name:        name,
		ReleaseType: types.ReleaseTypePluginBinTool,
		Generation:  gen,
		EventType:   types.PackageEventTypeDelete,
		OperateTime: time.Now(),
		Operator:    rCtx.Data().GetLoginName(),
	})

	logger.G.Biz(rCtx).
		With("gen", gen, "name", name).
		Info("deleted plugin-bintool release")

	resp := new(protoBackend.PackageReleasePluginBinToolDeleteResp)

	return resp.GetData(), nil
}
