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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// UploadOriginAgent upload origin agent.
func (h *handler) UploadOriginAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.UploadOriginAgentReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload agent, failed to parse file form")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload agent, failed to open file")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.manager.UploadOriginAgent(rCtx, file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload agent")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("gen", req.GetGeneration(), "detail", detail).Info("uploaded origin agent")

	resp := new(protoFile.UploadOriginAgentResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginServer upload origin server.
func (h *handler) UploadOriginServer(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.UploadOriginServerReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload server, failed to parse file form")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload server, failed to open file")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.manager.UploadOriginServer(rCtx, file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload server")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("gen", req.GetGeneration(), "detail", detail).Info("uploaded origin server")

	resp := new(protoFile.UploadOriginServerResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginCert upload origin cert.
func (h *handler) UploadOriginCert(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.UploadOriginCertReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload cert, failed to parse file form")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload cert, failed to open file")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.manager.UploadOriginCert(rCtx, file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload cert")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("detail", detail).Info("uploaded origin cert")

	resp := new(protoFile.UploadOriginCertResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginBinTool upload origin bin tool.
func (h *handler) UploadOriginBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.UploadOriginBinToolReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload origin bintool, failed to parse file form")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload origin bintool, failed to open file")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.manager.UploadOriginBinTool(rCtx, file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload origin bintool")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("detail", detail).Info("uploaded origin bintool")

	resp := new(protoFile.UploadOriginBinToolResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginPluginBinTool upload origin plugin bin tool.
func (h *handler) UploadOriginPluginBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.UploadOriginPluginBinToolReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload origin plugin bintool, failed to parse file form")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload origin plugin bintool, failed to open file")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.manager.UploadOriginPluginBinTool(rCtx, file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload origin plugin bintool")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("detail", detail).Info("uploaded origin plugin bintool")

	resp := new(protoFile.UploadOriginPluginBinToolResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginOfficialPlugin upload origin official plugin.
func (h *handler) UploadOriginOfficialPlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.UploadOriginOfficialPluginReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload official plugin, failed to parse file form")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload official plugin, failed to open file")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.manager.UploadOriginOfficialPlugin(rCtx, file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload official plugin")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("detail", detail).Info("uploaded origin official plugin")

	resp := new(protoFile.UploadOriginOfficialPluginResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}

// UploadOriginExternalPlugin upload origin external plugin.
func (h *handler) UploadOriginExternalPlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.UploadOriginExternalPluginReq)
	fileHeader, err := rCtx.ParseFileForm(req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload external plugin, failed to parse file form")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload external plugin, failed to open file")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detail, err := h.manager.UploadOriginExternalPlugin(rCtx, file)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upload external plugin")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("detail", detail).Info("uploaded origin external plugin")

	resp := new(protoFile.UploadOriginExternalPluginResp)
	resp.ConvertResultFromTypes(false, detail)

	return resp.GetData(), nil
}
