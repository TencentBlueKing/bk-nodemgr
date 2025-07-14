/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package release provides the release router.
package release

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	backendHandler backend.Handler
	fileHandler    file.IHandler
	logger         logger.Logger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/release"),
		backendHandler: capability.BackendHandler,
		fileHandler:    capability.FileHandler,
		logger:         capability.Logger,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", rest.RestHandlerFunc(h.ListRelease))
	h.rg.POST("/set_labels", rest.RestHandlerFunc(h.SetReleaseLabels))
	h.rg.POST("/enable", rest.RestHandlerFunc(h.EnableRelease))
	h.rg.POST("/disable", rest.RestHandlerFunc(h.DisableRelease))
	h.rg.POST("/set_as_default", rest.RestHandlerFunc(h.SetAsDefaultRelease))
	h.rg.POST("/cancel_as_default", rest.RestHandlerFunc(h.CancelAsDefaultRelease))
	h.rg.POST("/delete", rest.RestHandlerFunc(h.DeleteRelease))
	h.rg.POST("/deployed_host/count", rest.RestHandlerFunc(h.CountDeployedHost))
}

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

	req := new(protoApplication.PackageReleaseListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountRelease(
			sCtx,
			req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(sCtx, "failed to list release. failed to count host. err: %v", err)
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.PackageReleaseListResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	hosts, num, err := h.backendHandler.ListRelease(
		sCtx,
		req.ConvertPageToTypes(maxReleaseLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list release. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleaseListResp)
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

	req := new(protoApplication.PackageReleaseSetLabelsReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to set release labels, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.SetReleaseLabels(sCtx, gen, rt, plat, version, req.GetLabels()); err != nil {
		h.logger.ErrorCtxf(sCtx,
			"failed to set release labels. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "set release labels. gen(%d), release-type(%s), platform(%s), version(%s), labels(%v)",
		gen, rt, plat, version, req.GetLabels())

	resp := new(protoApplication.PackageReleaseSetLabelsResp)

	return resp.GetData(), nil
}

// EnableRelease enable release.
func (h *handler) EnableRelease(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to enable release, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.PackageReleaseEnableReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to enable release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.EnableRelease(sCtx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to enable release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "enabled release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoApplication.PackageReleaseEnableResp)

	return resp.GetData(), nil
}

// DisableRelease disable release.
func (h *handler) DisableRelease(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to disable release, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.PackageReleaseDisableReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to disable release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.DisableRelease(sCtx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to disable release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "disabled release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoApplication.PackageReleaseDisableResp)

	return resp.GetData(), nil
}

// SetAsDefaultRelease set default release.
func (h *handler) SetAsDefaultRelease(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to set as default release, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.PackageReleaseSetAsDefaultReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to set default release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.SetAsDefaultRelease(sCtx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(sCtx,
			"failed to set default release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "set default release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoApplication.PackageReleaseSetAsDefaultResp)

	return resp.GetData(), nil
}

// CancelAsDefaultRelease cancel default release.
func (h *handler) CancelAsDefaultRelease(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to cancel default release, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.PackageReleaseCancelAsDefaultReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to cancel default release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.CancelAsDefaultRelease(sCtx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(sCtx,
			"failed to cancel default release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "canceled default release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoApplication.PackageReleaseCancelAsDefaultResp)

	return resp.GetData(), nil
}

func (h *handler) DeleteRelease(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to delete release, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.PackageReleaseDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to delete release, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.DeleteRelease(sCtx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(sCtx,
			"failed to delete release. gen(%d), release-type(%s), platform(%s), version(%s), err: %v",
			gen, rt, plat, version, err)

		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "deleted release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoApplication.PackageReleaseDeleteResp)

	return resp.GetData(), nil
}

// CountDeployedHost count deployed host.
func (h *handler) CountDeployedHost(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to count deployed host, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.PackageReleaseDeployedHostCountReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to count deployed host, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	hosts, _, err := h.backendHandler.ListHost(sCtx, types.UnlimitedPage(), req.ConvertConditionsToHostTypes())
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list host. err: %w", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	result, pair, err := req.CountHostsByOsType(hosts)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to count hosts. err: %w", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(sCtx, "count deployed hosts for release. pair(%d)", pair)

	resp := new(protoApplication.PackageReleaseDeployedHostCountResp)
	resp.ConvertResultFromTypes(result, pair)

	return resp.GetData(), nil
}
