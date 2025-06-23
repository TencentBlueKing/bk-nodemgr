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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

const (
	maxReleaseLimit = 1000
)

// ListRelease lists releases with page and conditions.
func (h *handler) ListRelease(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list release, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.PackageReleaseListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountRelease(
			sCtx,
			req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(sCtx, "failed to list release. failed to count host. err: %v", err)
			return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PackageReleaseListResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	hosts, num, err := h.storage.ListRelease(
		sCtx,
		req.ConvertPageToTypes(maxReleaseLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list release. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleaseListResp)
	resp.ConvertReleasesFromTypes(num, hosts)

	return resp.GetData(), nil
}

// SetReleaseLabels set release labels.
func (h *handler) SetReleaseLabels(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to set release labels, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.PackageReleaseSetLabelsReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to set release labels, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.SetReleaseLabels(sCtx, gen, rt, plat, version, req.GetLabels()); err != nil {
		h.logger.ErrorCtxf(sCtx,
			"failed to set release labels. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "set release labels. gen(%d), release-type(%s), platform(%s), version(%s), labels(%v)",
		gen, rt, plat, version, req.GetLabels())

	resp := new(protoBackend.PackageReleaseSetLabelsResp)

	return resp.GetData(), nil
}

// EnableRelease enable release.
func (h *handler) EnableRelease(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to enable release, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.PackageReleaseEnableReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to enable release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.EnableRelease(sCtx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to enable release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "enabled release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoBackend.PackageReleaseEnableResp)

	return resp.GetData(), nil
}

// DisableRelease disable release.
func (h *handler) DisableRelease(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to disable release, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.PackageReleaseDisableReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to disable release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.DisableRelease(sCtx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to disable release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "disabled release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoBackend.PackageReleaseDisableResp)

	return resp.GetData(), nil
}

// SetAsDefaultRelease set default release.
func (h *handler) SetAsDefaultRelease(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to set as default release, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.PackageReleaseSetAsDefaultReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to set default release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.SetAsDefaultRelease(sCtx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(sCtx,
			"failed to set default release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "set as default release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoBackend.PackageReleaseSetAsDefaultResp)

	return resp.GetData(), nil
}

// CancelAsDefaultRelease cancel default release.
func (h *handler) CancelAsDefaultRelease(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to cancel default release, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.PackageReleaseCancelAsDefaultReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to cancel default release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.CancelAsDefaultRelease(sCtx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(sCtx,
			"failed to cancel default release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "cancel as default release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoBackend.PackageReleaseCancelAsDefaultResp)

	return resp.GetData(), nil
}

func (h *handler) DeleteRelease(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to delete release, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.PackageReleaseDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to delete release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.storage.DeleteRelease(sCtx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(sCtx,
			"failed to delete release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "deleted release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoBackend.PackageReleaseDeleteResp)

	return resp.GetData(), nil
}
