/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package publish

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// PublishReleaseAgent publish release agent.
func (h *handler) PublishReleaseAgent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleaseAgentReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	shared := req.GetIsShared()
	if err := h.manager.PublishReleaseAgent(rCtx, uploadID, shared); err != nil {
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
	shared := req.GetIsShared()
	uploadCategory := types.UploadCategory(req.GetUploadOriginPkgType())
	switch uploadCategory {
	case types.UploadCategoryOriginProxy:
		if err := h.manager.PublishReleaseProxyFromProxyPkg(rCtx, uploadID, shared); err != nil {
			logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release proxy from origin proxy pkg")

			return nil, resterrf.ErrWrap(resterrf.Aborted, err)
		}
	case types.UploadCategoryOriginServer:
		if err := h.manager.PublishReleaseProxyFromServerPkg(rCtx, uploadID, shared); err != nil {
			logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release proxy from origin server pkg")

			return nil, resterrf.ErrWrap(resterrf.Aborted, err)
		}
	default:
		logger.G.Biz(rCtx).
			With("upload-origin-pkg-type", req.GetUploadOriginPkgType()).
			Error("failed to publish release proxy, unsupported upload origin pkg type")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("unsupported upload origin pkg type: %s", req.GetUploadOriginPkgType()))
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
	shared := req.GetIsShared()
	if err := h.manager.PublishReleaseCert(rCtx, uploadID, shared); err != nil {
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
	shared := req.GetIsShared()
	if err := h.manager.PublishReleaseBinTool(rCtx, uploadID, shared); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release bintool")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release bintool")

	resp := new(protoFile.PublishReleaseBinToolResp)

	return resp.GetData(), nil
}

// PublishReleasePluginBinTool publish release plugin bintool .
func (h *handler) PublishReleasePluginBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleasePluginBinToolReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release plugin bintool, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	shared := req.GetIsShared()
	if err := h.manager.PublishReleasePluginBinTool(rCtx, uploadID, shared); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin bintool")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release plugin bintool")

	resp := new(protoFile.PublishReleasePluginBinToolResp)

	return resp.GetData(), nil
}

// PublishReleasePluginV2 publish release plugin.
func (h *handler) PublishReleasePluginV2(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleasePluginV2Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release plugin v2, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	shared := req.GetIsShared()
	if err := h.manager.PublishReleasePluginV2(rCtx, uploadID, shared); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin v2")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release plugin v2")

	resp := new(protoFile.PublishReleasePluginV2Resp)

	return resp.GetData(), nil
}

// PublishReleaseExternalPluginV2 publish release external plugin.
func (h *handler) PublishReleaseExternalPluginV2(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleaseExternalPluginV2Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release external plugin v2, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	shared := req.GetIsShared()
	if err := h.manager.PublishReleaseExternalPlugin(rCtx, uploadID, shared); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release external plugin v2")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release external plugin v2")

	resp := new(protoFile.PublishReleaseExternalPluginV2Resp)

	return resp.GetData(), nil
}

// PublishReleasePluginV3 publish release plugin.
func (h *handler) PublishReleasePluginV3(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PublishReleasePluginV3Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to publish release plugin v3, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	shared := req.GetIsShared()
	if err := h.manager.PublishReleasePluginV3(rCtx, uploadID, shared); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin v3")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("upload-id", uploadID).Info("uploaded and generated release plugin v3")

	resp := new(protoFile.PublishReleasePluginV3Resp)

	return resp.GetData(), nil
}
