/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package publish provide the api for publish.
package publish

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
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
		rg:             rg.Group("/publish"),
		backendHandler: capability.BackendHandler,
		fileHandler:    capability.FileHandler,
		logger:         capability.Logger,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/release/agent", rest.RestHandlerFunc(h.PublishReleaseAgent))
	h.rg.POST("/release/proxy", rest.RestHandlerFunc(h.PublishReleaseProxy))
	h.rg.POST("/release/cert", rest.RestHandlerFunc(h.PublishReleaseCert))
	h.rg.POST("/release/bintool", rest.RestHandlerFunc(h.PublishReleaseBinTool))
}

// PublishReleaseAgent publish release agent.
func (h *handler) PublishReleaseAgent(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.PackagePublishReleaseAgentReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.fileHandler.PublishReleaseAgent(ctx, uploadID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release agent. upload-id(%s), err: %v", uploadID, err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded and generated release agent. upload-id(%s)", uploadID)

	resp := new(protoApplication.PackagePublishReleaseAgentResp)

	return resp.GetData(), nil
}

// PublishReleaseProxy publish release agent.
func (h *handler) PublishReleaseProxy(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.PackagePublishReleaseProxyReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release proxy, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.fileHandler.PublishReleaseProxy(ctx, uploadID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release proxy. upload-id(%s), err: %v", uploadID, err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded and generated release proxy. upload-id(%s)", uploadID)

	resp := new(protoApplication.PackagePublishReleaseProxyResp)

	return resp.GetData(), nil
}

// PublishReleaseCert publish release cert.
func (h *handler) PublishReleaseCert(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.PackagePublishReleaseCertReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release cert, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.fileHandler.PublishReleaseCert(ctx, uploadID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release cert. upload-id(%s), err: %v", uploadID, err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded and generated release cert. upload-id(%s)", uploadID)

	resp := new(protoApplication.PackagePublishReleaseCertResp)

	return resp.GetData(), nil
}

// PublishReleaseBinTool publish release bintool.
func (h *handler) PublishReleaseBinTool(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.PackagePublishReleaseBinToolReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release bintool, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.fileHandler.PublishReleaseBinTool(ctx, uploadID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release bintool. upload-id(%s), err: %v", uploadID, err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded and generated release bintool. upload-id(%s)", uploadID)

	resp := new(protoApplication.PackagePublishReleaseBinToolResp)

	return resp.GetData(), nil
}
