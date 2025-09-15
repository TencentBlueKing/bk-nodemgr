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
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	maxReleaseLimit = 1000
)

// ListRelease lists releases with page and conditions.
func (h *handler) ListRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.PackageReleaseListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	releaseType := types.ReleaseType(req.GetReleaseType())
	gen := types.Generation(req.GetGeneration())
	exactIncludeCond := req.ConvertExactIncludeConditionsToTypes()
	exactIncludeCond.Generation = append(exactIncludeCond.Generation, gen)
	cond := &types.ReleaseCondition{
		ExactInclude: exactIncludeCond,
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountRelease(ctx, releaseType, cond)
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to list release. failed to count host. err: %v", err)
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PackageReleaseListResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page := req.ConvertPageToTypes(maxReleaseLimit)
	hosts, num, err := h.storage.ListRelease(ctx, releaseType, page, cond)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list release. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleaseListResp)
	resp.ConvertReleasesFromTypes(num, hosts)

	return resp.GetData(), nil
}

// DistinctRelease distinct releases.
func (h *handler) DistinctRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.PackageReleaseDistinctReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to distinct release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	releaseType := types.ReleaseType(req.GetReleaseType())
	gen := types.Generation(req.GetGeneration())
	exactIncludeCond := req.ConvertExactIncludeConditionsToTypes()
	exactIncludeCond.Generation = append(exactIncludeCond.Generation, gen)
	cond := &types.ReleaseCondition{
		ExactInclude: exactIncludeCond,
	}

	distinctField := types.ReleaseDistinctField{
		OSType:  req.GetDistinctField().GetOsType(),
		CPUArch: req.GetDistinctField().GetCpuArch(),
	}

	result, err := h.storage.DistinctRelease(ctx, releaseType, distinctField, cond)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to distinct host. failed to distinct host fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleaseDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// SetReleaseLabels set release labels.
func (h *handler) SetReleaseLabels(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.PackageReleaseSetLabelsReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to set release labels, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.SetReleaseLabels(ctx, gen, rt, plat, version, req.GetLabels()); err != nil {
		h.logger.ErrorCtxf(ctx,
			"failed to set release labels. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(ctx, "set release labels. gen(%d), release-type(%s), platform(%s), version(%s), labels(%v)",
		gen, rt, plat, version, req.GetLabels())

	resp := new(protoBackend.PackageReleaseSetLabelsResp)

	return resp.GetData(), nil
}

// EnableRelease enable release.
func (h *handler) EnableRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.PackageReleaseEnableReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to enable release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.EnableRelease(ctx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to enable release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(ctx, "enabled release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoBackend.PackageReleaseEnableResp)

	return resp.GetData(), nil
}

// DisableRelease disable release.
func (h *handler) DisableRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.PackageReleaseDisableReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to disable release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.DisableRelease(ctx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to disable release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(ctx, "disabled release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoBackend.PackageReleaseDisableResp)

	return resp.GetData(), nil
}

// SetAsDefaultRelease set default release.
func (h *handler) SetAsDefaultRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.PackageReleaseSetAsDefaultReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to set default release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.SetAsDefaultRelease(ctx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(ctx,
			"failed to set default release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(ctx, "set as default release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoBackend.PackageReleaseSetAsDefaultResp)

	return resp.GetData(), nil
}

// CancelAsDefaultRelease cancel default release.
func (h *handler) CancelAsDefaultRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.PackageReleaseCancelAsDefaultReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to cancel default release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.CancelAsDefaultRelease(ctx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(ctx,
			"failed to cancel default release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(ctx, "cancel as default release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoBackend.PackageReleaseCancelAsDefaultResp)

	return resp.GetData(), nil
}

func (h *handler) DeleteRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.PackageReleaseDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.DeleteRelease(ctx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(ctx,
			"failed to delete release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(ctx, "deleted release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoBackend.PackageReleaseDeleteResp)

	return resp.GetData(), nil
}
