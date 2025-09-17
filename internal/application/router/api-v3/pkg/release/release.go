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
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	backendHandler backend.IHandler
	fileHandler    file.IHandler
	logger         logger.ILogger
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

	h.rg.POST("/list", restserver.Handler(h.ListRelease))
	h.rg.POST("/set_labels", restserver.Handler(h.SetReleaseLabels))
	h.rg.POST("/enable", restserver.Handler(h.EnableRelease))
	h.rg.POST("/disable", restserver.Handler(h.DisableRelease))
	h.rg.POST("/set_as_default", restserver.Handler(h.SetAsDefaultRelease))
	h.rg.POST("/cancel_as_default", restserver.Handler(h.CancelAsDefaultRelease))
	h.rg.POST("/delete", restserver.Handler(h.DeleteRelease))
	h.rg.POST("/deployed_host/count", restserver.Handler(h.CountDeployedHost))
}

const (
	maxReleaseLimit = 1000
)

// ListRelease lists releases with page and conditions.
func (h *handler) ListRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageReleaseListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	releaseType := types.ReleaseType(req.GetReleaseType())
	gen := types.Generation(req.GetGeneration())

	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountRelease(ctx, releaseType, gen, req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to list release. failed to count host. err: %v", err)
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.PackageReleaseListResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	hosts, num, err := h.backendHandler.ListRelease(ctx, releaseType, gen, req.ConvertPageToTypes(maxReleaseLimit), req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list release. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleaseListResp)
	resp.ConvertReleasesFromTypes(num, hosts)

	return resp.GetData(), nil
}

// SetReleaseLabels set release labels.
func (h *handler) SetReleaseLabels(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageReleaseSetLabelsReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to set release labels, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.SetReleaseLabels(ctx, gen, rt, plat, version, req.GetLabels()); err != nil {
		h.logger.ErrorCtxf(ctx,
			"failed to set release labels. gen(%d), release-type(%s), platform(%s), version(%s): %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(ctx, "set release labels. gen(%d), release-type(%s), platform(%s), version(%s), labels(%v)",
		gen, rt, plat, version, req.GetLabels())

	resp := new(protoApplication.PackageReleaseSetLabelsResp)

	return resp.GetData(), nil
}

// EnableRelease enable release.
func (h *handler) EnableRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageReleaseEnableReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to enable release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.EnableRelease(ctx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to enable release. gen(%d), release-type(%s), platform(%s), version(%s): %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(ctx, "enabled release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoApplication.PackageReleaseEnableResp)

	return resp.GetData(), nil
}

// DisableRelease disable release.
func (h *handler) DisableRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageReleaseDisableReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to disable release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.DisableRelease(ctx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to disable release. gen(%d), release-type(%s), platform(%s), version(%s): %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(ctx, "disabled release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoApplication.PackageReleaseDisableResp)

	return resp.GetData(), nil
}

// SetAsDefaultRelease set default release.
func (h *handler) SetAsDefaultRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageReleaseSetAsDefaultReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to set default release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.SetAsDefaultRelease(ctx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(ctx,
			"failed to set default release. gen(%d), release-type(%s), platform(%s), version(%s): %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(ctx, "set default release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoApplication.PackageReleaseSetAsDefaultResp)

	return resp.GetData(), nil
}

// CancelAsDefaultRelease cancel default release.
func (h *handler) CancelAsDefaultRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageReleaseCancelAsDefaultReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to cancel default release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.CancelAsDefaultRelease(ctx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(ctx,
			"failed to cancel default release. gen(%d), release-type(%s), platform(%s), version(%s): %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(ctx, "canceled default release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoApplication.PackageReleaseCancelAsDefaultResp)

	return resp.GetData(), nil
}

func (h *handler) DeleteRelease(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageReleaseDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete release, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.DeleteRelease(ctx, gen, rt, plat, version); err != nil {
		h.logger.ErrorCtxf(ctx,
			"failed to delete release. gen(%d), release-type(%s), platform(%s), version(%s): %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(ctx, "deleted release. gen(%d), release-type(%s), platform(%s), version(%s)",
		gen, rt, plat, version)

	resp := new(protoApplication.PackageReleaseDeleteResp)

	return resp.GetData(), nil
}

// CountDeployedHost count deployed host.
func (h *handler) CountDeployedHost(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageReleaseDeployedHostCountReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to count deployed host, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition, err := req.ConvertConditionsToHostTypes()
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to count deployed host, failed to convert conditions to host types. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, _, err := h.backendHandler.ListHost(ctx, types.UnlimitedPage(), condition)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to count deployed host, failed to list host. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	result, pair, err := req.CountHostsByOsTypeAndArch(hosts)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to count hosts. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	h.logger.InfoCtxf(ctx, "count deployed hosts for release. pair(%d)", pair)

	resp := new(protoApplication.PackageReleaseDeployedHostCountResp)
	resp.ConvertResultFromTypes(result, pair)

	return resp.GetData(), nil
}
