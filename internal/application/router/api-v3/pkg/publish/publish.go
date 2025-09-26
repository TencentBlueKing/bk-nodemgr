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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
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
		rg:             rg.Group("/publish"),
		backendHandler: capability.BackendHandler,
		fileHandler:    capability.FileHandler,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/release/agent", restserver.Handler(h.PublishReleaseAgent))
	h.rg.POST("/release/proxy", restserver.Handler(h.PublishReleaseProxy))
	h.rg.POST("/release/cert", restserver.Handler(h.PublishReleaseCert))
	h.rg.POST("/release/bintool", restserver.Handler(h.PublishReleaseBinTool))
}

// PublishReleaseAgent publish release agent.
func (h *handler) PublishReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackagePublishReleaseAgentReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.fileHandler.PublishReleaseAgent(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release agent. upload-id(%s): %v", uploadID, err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release agent")

	resp := new(protoApplication.PackagePublishReleaseAgentResp)

	return resp.GetData(), nil
}

// PublishReleaseProxy publish release agent.
func (h *handler) PublishReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackagePublishReleaseProxyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.fileHandler.PublishReleaseProxy(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release proxy. upload-id(%s): %v", uploadID, err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release proxy")

	resp := new(protoApplication.PackagePublishReleaseProxyResp)

	return resp.GetData(), nil
}

// PublishReleaseCert publish release cert.
func (h *handler) PublishReleaseCert(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackagePublishReleaseCertReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release cert, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.fileHandler.PublishReleaseCert(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release cert. upload-id(%s): %v", uploadID, err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release cert")

	resp := new(protoApplication.PackagePublishReleaseCertResp)

	return resp.GetData(), nil
}

// PublishReleaseBinTool publish release bintool.
func (h *handler) PublishReleaseBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackagePublishReleaseBinToolReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release bintool, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.fileHandler.PublishReleaseBinTool(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release bintool. upload-id(%s): %v", uploadID, err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release bintool")

	resp := new(protoApplication.PackagePublishReleaseBinToolResp)

	return resp.GetData(), nil
}
