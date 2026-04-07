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

// ListReleaseCert lists cert releases.
func (h *handler) ListReleaseCert(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseCertListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list cert release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Generation: []types.Generation{gen},
		},
	}

	items, num, err := h.daoReleaseCert.ListReleaseCert(rCtx, types.UnlimitedPage(), cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list cert release")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleaseCertListResp)
	resp.ConvertReleasesFromTypes(num, items)

	return resp.GetData(), nil
}

// DeleteReleaseCert deletes cert release.
func (h *handler) DeleteReleaseCert(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleaseCertDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete cert release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// check permission.
	resources := authProvider.BuildPackageResources(string(types.ReleaseTypeCert))
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete cert release, permission denied")
		return nil, err
	}

	gen := types.Generation(req.GetGeneration())
	key := types.ReleaseCertKey{
		Generation: gen,
	}

	if err := h.daoReleaseCert.DeleteReleaseCert(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen).
			Error("failed to delete cert release")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	h.recordPackageEvents(rCtx, &types.PackageEvent{
		Name:        types.ReleaseNameCert,
		ReleaseType: types.ReleaseTypeCert,
		Generation:  gen,
		EventType:   types.PackageEventTypeDelete,
		OperateTime: time.Now(),
		Operator:    rCtx.Data().GetLoginName(),
	})

	logger.G.Biz(rCtx).
		With("gen", gen).
		Info("deleted cert release")

	resp := new(protoBackend.PackageReleaseCertDeleteResp)

	return resp.GetData(), nil
}
