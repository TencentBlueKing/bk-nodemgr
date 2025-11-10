/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package publish

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// PublishReleaseAgent publish release agent.
func (h *handler) PublishReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleaseAgentReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleaseAgent(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release agent")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release agent")

	resp := new(protoFile.PublishReleaseAgentResp)

	return resp.GetData(), nil
}

// PublishReleaseProxy publish release agent.
func (h *handler) PublishReleaseProxy(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleaseProxyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleaseProxy(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release proxy")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release proxy")

	resp := new(protoFile.PublishReleaseProxyResp)

	return resp.GetData(), nil
}

// PublishReleaseCert publish release cert.
func (h *handler) PublishReleaseCert(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleaseCertReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release cert, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleaseCert(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release cert")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release cert")

	resp := new(protoFile.PublishReleaseCertResp)

	return resp.GetData(), nil
}

// PublishReleaseBinTool publish release bintool.
func (h *handler) PublishReleaseBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleaseBinToolReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release bintool, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleaseBinTool(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release bintool")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release bintool")

	resp := new(protoFile.PublishReleaseBinToolResp)

	return resp.GetData(), nil
}

// PublishReleasePluginBinToolV2 publish release plugin bintool v2.
func (h *handler) PublishReleasePluginBinToolV2(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleasePluginBinToolV2Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release plugin bintool v2, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleasePluginBinToolV2(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin bintool v2")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release plugin bintool v2")

	resp := new(protoFile.PublishReleasePluginBinToolV2Resp)

	return resp.GetData(), nil
}

// PublishReleasePluginBinToolV3 publish release plugin bintool v3.
func (h *handler) PublishReleasePluginBinToolV3(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleasePluginBinToolV3Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release plugin bintool v3, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleasePluginBinToolV3(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin bintool v3")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release plugin bintool v3")

	resp := new(protoFile.PublishReleasePluginBinToolV3Resp)

	return resp.GetData(), nil
}

// PublishReleasePluginV2 publish release plugin.
func (h *handler) PublishReleasePluginV2(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleasePluginV2Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleasePluginV2(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release plugin")

	resp := new(protoFile.PublishReleasePluginV2Resp)

	return resp.GetData(), nil
}

// PublishReleaseExternalPluginV2 publish release external plugin.
func (h *handler) PublishReleaseExternalPluginV2(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleaseExternalPluginV2Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release external plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleaseExternalPlugin(rCtx, uploadID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release external plugin")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release external plugin")

	resp := new(protoFile.PublishReleaseExternalPluginV2Resp)

	return resp.GetData(), nil
}
