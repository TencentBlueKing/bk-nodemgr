/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package upload provide the api for upload.
package upload

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
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
		rg:             rg.Group("/upload"),
		backendHandler: capability.BackendHandler,
		fileHandler:    capability.FileHandler,
		logger:         capability.Logger,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/origin/agent", restserver.Handler(h.UploadOriginAgent))
	h.rg.POST("/origin/server", restserver.Handler(h.UploadOriginServer))
	h.rg.POST("/origin/cert", restserver.Handler(h.UploadOriginCert))
	h.rg.POST("/origin/bintool", restserver.Handler(h.UploadOriginBinTool))
}

// UploadOriginAgent upload origin agent.
func (h *handler) UploadOriginAgent(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageUploadOriginAgentReq)
	fileHeader, err := ctx.ParseFileForm(req)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload origin agent, failed to parse file form: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload origin agent, failed to open file. err: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.fileHandler.UploadOriginAgent(ctx, fileHeader.Filename, file)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload origin agent: %v", err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded origin agent, generation: %d, detail: %v", req.GetGeneration(), detail)

	resp := new(protoApplication.PackageUploadOriginAgentResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginServer upload origin server.
func (h *handler) UploadOriginServer(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageUploadOriginServerReq)
	fileHeader, err := ctx.ParseFileForm(req)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload origin server, failed to parse file form: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload origin server, failed to open file. err: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.fileHandler.UploadOriginServer(ctx, fileHeader.Filename, file)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload origin server: %v", err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded origin server, generation: %d, detail: %v", req.GetGeneration(), detail)

	resp := new(protoApplication.PackageUploadOriginServerResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginCert upload origin cert.
func (h *handler) UploadOriginCert(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageUploadOriginCertReq)
	fileHeader, err := ctx.ParseFileForm(req)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload cert, failed to parse file form: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload cert, failed to open file. err: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.fileHandler.UploadOriginCert(ctx, fileHeader.Filename, file)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload cert: %v", err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded origin cert, detail: %v", detail)

	resp := new(protoApplication.PackageUploadOriginCertResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginBinTool upload origin bin tool.
func (h *handler) UploadOriginBinTool(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.PackageUploadOriginBinToolReq)
	fileHeader, err := ctx.ParseFileForm(req)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload origin bintool, failed to parse file form: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload origin bintool, failed to open file. err: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.fileHandler.UploadOriginBinTool(ctx, fileHeader.Filename, file)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to upload origin bintool: %v", err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded origin bintool, detail: %v", detail)

	resp := new(protoApplication.PackageUploadOriginBinToolResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}
