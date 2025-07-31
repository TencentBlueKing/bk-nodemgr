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
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// PublishReleaseAgent publish release agent.
func (h *handler) PublishReleaseAgent(ctx *restserver.Context) (interface{}, error) {
	req := new(protoFile.PublishReleaseAgentReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleaseAgent(ctx, uploadID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release agent. upload-id(%s), err: %v", uploadID, err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded and generated release agent. upload-id(%s)", uploadID)

	resp := new(protoFile.PublishReleaseAgentResp)

	return resp.GetData(), nil
}

// PublishReleaseProxy publish release agent.
func (h *handler) PublishReleaseProxy(ctx *restserver.Context) (interface{}, error) {
	req := new(protoFile.PublishReleaseProxyReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release proxy, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleaseProxy(ctx, uploadID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release proxy. upload-id(%s), err: %v", uploadID, err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded and generated release proxy. upload-id(%s)", uploadID)

	resp := new(protoFile.PublishReleaseProxyResp)

	return resp.GetData(), nil
}

// PublishReleaseCert publish release cert.
func (h *handler) PublishReleaseCert(ctx *restserver.Context) (interface{}, error) {
	req := new(protoFile.PublishReleaseCertReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release cert, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleaseCert(ctx, uploadID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release cert. upload-id(%s), err: %v", uploadID, err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded and generated release cert. upload-id(%s)", uploadID)

	resp := new(protoFile.PublishReleaseCertResp)

	return resp.GetData(), nil
}

// PublishReleaseBinTool publish release bintool.
func (h *handler) PublishReleaseBinTool(ctx *restserver.Context) (interface{}, error) {
	req := new(protoFile.PublishReleaseBinToolReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release bintool, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	uploadID := req.GetUploadId()
	if err := h.manager.PublishReleaseBinTool(ctx, uploadID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to publish release bintool. upload-id(%s), err: %v", uploadID, err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	h.logger.InfoCtxf(ctx, "uploaded and generated release bintool. upload-id(%s)", uploadID)

	resp := new(protoFile.PublishReleaseBinToolResp)

	return resp.GetData(), nil
}
