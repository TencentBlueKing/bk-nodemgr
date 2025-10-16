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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	backendHandler backend.IHandler
	fileHandler    file.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/release"),
		backendHandler: capability.BackendHandler,
		fileHandler:    capability.FileHandler,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListRelease))
	h.rg.POST("/set_labels", restserver.Handler(h.SetReleaseLabels))
	h.rg.POST("/set_labels_many", restserver.Handler(h.SetReleaseLabelsMany))
	h.rg.POST("/enable", restserver.Handler(h.EnableRelease))
	h.rg.POST("/disable", restserver.Handler(h.DisableRelease))
	h.rg.POST("/set_as_default", restserver.Handler(h.SetAsDefaultRelease))
	h.rg.POST("/cancel_as_default", restserver.Handler(h.CancelAsDefaultRelease))
	h.rg.POST("/delete", restserver.Handler(h.DeleteRelease))
	h.rg.POST("/deployed_host/count", restserver.Handler(h.CountDeployedHost))

	h.rg.POST("/agent/download", restserver.StreamHandler(h.AgentDownload))
}

const (
	maxReleaseLimit = 1000
)

// ListRelease lists releases with page and conditions.
func (h *handler) ListRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	releaseType := types.ReleaseType(req.GetReleaseType())
	gen := types.Generation(req.GetGeneration())

	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountRelease(rCtx, releaseType, gen, req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list release. failed to count host")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.PackageReleaseListResp)
		resp.ConvertReleasesFromTypes(num, nil)

		return resp.GetData(), nil
	}

	releases, num, err := h.backendHandler.ListRelease(rCtx, releaseType, gen, req.ConvertPageToTypes(maxReleaseLimit), req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list release")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.PackageReleaseListResp)
	resp.ConvertReleasesFromTypes(num, releases)

	return resp.GetData(), nil
}

// SetReleaseLabels set release labels.
func (h *handler) SetReleaseLabels(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseSetLabelsReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set release labels, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.SetReleaseLabels(rCtx, gen, rt, plat, version, req.GetLabels()); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "plat", plat, "version", version).
			Error("failed to set release labels")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "release-type", rt, "plat", plat, "version", version, "labels", req.GetLabels()).Info("set release labels")

	resp := new(protoApplication.PackageReleaseSetLabelsResp)

	return resp.GetData(), nil
}

// SetReleaseLabelsMany set release labels.
func (h *handler) SetReleaseLabelsMany(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseSetLabelsManyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set many release labels, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	rt, gen, plat, version := req.GetIdentifiers()
	if err := h.backendHandler.SetReleaseLabelsMany(rCtx, rt, gen, plat, version, req.GetLabels()); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "plat", plat, "version", version).
			Error("failed to set many release labels")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "release-type", rt, "plat", plat, "version", version, "labels", req.GetLabels()).Info("set many release labels")

	resp := new(protoApplication.PackageReleaseSetLabelsResp)

	return resp.GetData(), nil
}

// EnableRelease enable release.
func (h *handler) EnableRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.EnableRelease(rCtx, gen, rt, plat, version); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable release. gen(%d), release-type(%s), platform(%s), version(%s): %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "release-type", rt, "plat", plat, "version", version).Info("enabled release")

	resp := new(protoApplication.PackageReleaseEnableResp)

	return resp.GetData(), nil
}

// DisableRelease disable release.
func (h *handler) DisableRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.DisableRelease(rCtx, gen, rt, plat, version); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable release. gen(%d), release-type(%s), platform(%s), version(%s): %v",
			gen, rt, plat, version, err)

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "release-type", rt, "plat", plat, "version", version).Info("disabled release")

	resp := new(protoApplication.PackageReleaseDisableResp)

	return resp.GetData(), nil
}

// SetAsDefaultRelease set default release.
func (h *handler) SetAsDefaultRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseSetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set default release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.SetAsDefaultRelease(rCtx, gen, rt, plat, version); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "plat", plat, "version", version).
			Error("failed to set default release")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "release-type", rt, "plat", plat, "version", version).Info("set default release")

	resp := new(protoApplication.PackageReleaseSetAsDefaultResp)

	return resp.GetData(), nil
}

// CancelAsDefaultRelease cancel default release.
func (h *handler) CancelAsDefaultRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseCancelAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel default release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.CancelAsDefaultRelease(rCtx, gen, rt, plat, version); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "plat", plat, "version", version).
			Error("failed to cancel default release")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "release-type", rt, "plat", plat, "version", version).Info("canceled default release")

	resp := new(protoApplication.PackageReleaseCancelAsDefaultResp)

	return resp.GetData(), nil
}

func (h *handler) DeleteRelease(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, rt, plat, version := req.GetIdentifier()
	if err := h.backendHandler.DeleteRelease(rCtx, gen, rt, plat, version); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "release-type", rt, "plat", plat, "version", version).
			Error("failed to delete default release")

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "release-type", rt, "plat", plat, "version", version).Info("deleted default release")

	resp := new(protoApplication.PackageReleaseDeleteResp)

	return resp.GetData(), nil
}

// CountDeployedHost count deployed host.
func (h *handler) CountDeployedHost(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageReleaseDeployedHostCountReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count deployed host, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition, err := req.ConvertConditionsToHostTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count deployed host, failed to convert conditions to host types")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, _, err := h.backendHandler.ListHost(rCtx, types.UnlimitedPage(), condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count deployed host, failed to list host")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	result, pair, err := req.CountHostsByOsTypeAndArch(hosts)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count hosts")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("pair", pair).Info("count deployed hosts for release")

	resp := new(protoApplication.PackageReleaseDeployedHostCountResp)
	resp.ConvertResultFromTypes(result, pair)

	return resp.GetData(), nil
}

func (h *handler) AgentDownload(rCtx restserver.IContext) (*restserver.StreamResponse, error) {
	req := new(protoApplication.PackageReleaseAgentDownloadReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen, plat, version := req.GetIdentifier()
	resp, err := h.fileHandler.DownloadReleaseAgent(rCtx, gen, plat, version)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download release agent: %v", err)

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	logger.G.Biz(rCtx).With("gen", gen, "plat", plat, "version", version).Info("downloaded release agent")

	return resp, nil
}
