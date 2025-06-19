/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package upload

import (
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

// UploadOriginAgent upload origin agent.
func (h *handler) UploadOriginAgent(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to upload origin agent, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.PackageUploadOriginAgentReq)
	fileHeader, err := ctx.ParseFileForm(req)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload origin agent, failed to parse file form: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload origin agent, failed to open file. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.fileHandler.UploadOriginAgent(sCtx, fileHeader.Filename, file)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload origin agent: %v", err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	h.logger.InfoCtxf(sCtx, "uploaded origin agent, generation: %d, detail: %v", req.GetGeneration, detail)

	resp := new(protoApplication.PackageUploadOriginAgentResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginServer upload origin server.
func (h *handler) UploadOriginServer(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to upload origin server, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.PackageUploadOriginServerReq)
	fileHeader, err := ctx.ParseFileForm(req)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload origin server, failed to parse file form: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload origin server, failed to open file. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.fileHandler.UploadOriginServer(sCtx, fileHeader.Filename, file)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload origin server: %v", err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	h.logger.InfoCtxf(sCtx, "uploaded origin server, generation: %d, detail: %v", req.GetGeneration, detail)

	resp := new(protoApplication.PackageUploadOriginServerResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginCert upload origin cert.
func (h *handler) UploadOriginCert(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to upload cert, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.PackageUploadOriginCertReq)
	fileHeader, err := ctx.ParseFileForm(req)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload cert, failed to parse file form: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload cert, failed to open file. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.fileHandler.UploadOriginCert(sCtx, fileHeader.Filename, file)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload cert: %v", err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	h.logger.InfoCtxf(sCtx, "uploaded origin cert, detail: %v", detail)

	resp := new(protoApplication.PackageUploadOriginCertResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginBinTool upload origin bin tool.
func (h *handler) UploadOriginBinTool(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to upload origin bintool, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.PackageUploadOriginBinToolReq)
	fileHeader, err := ctx.ParseFileForm(req)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload origin bintool, failed to parse file form: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload origin bintool, failed to open file. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.fileHandler.UploadOriginBinTool(sCtx, fileHeader.Filename, file)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload origin bintool: %v", err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	h.logger.InfoCtxf(sCtx, "uploaded origin bintool, detail: %v", detail)

	resp := new(protoApplication.PackageUploadOriginBinToolResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}
