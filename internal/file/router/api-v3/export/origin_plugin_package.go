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

package export

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// ExportGetOriginPluginPackageDownloadAddress returns a temporary package download address.
func (h *handler) ExportGetOriginPluginPackageDownloadAddress(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ExportGetOriginPluginPackageDownloadAddressReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get origin plugin package download address, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	downloadURL, expiredAt, err := h.manager.ExportGetOriginPluginPackageDownloadAddress(rCtx, req.GetExportId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("export-id", req.GetExportId()).
			Error("failed to get origin plugin package download address")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := new(protoFile.ExportGetOriginPluginPackageDownloadAddressResp)
	resp.ConvertResult(downloadURL, expiredAt)

	return resp.GetData(), nil
}

// ExportPrepareOriginPluginPackage prepares an origin plugin package export artifact.
func (h *handler) ExportPrepareOriginPluginPackage(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ExportPrepareOriginPluginPackageReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to prepare origin plugin package export, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	exportID, err := h.manager.ExportPrepareOriginPluginPackage(rCtx, req.GetPluginPkgName(), req.GetPluginPkgVersion())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("plugin-name", req.GetPluginPkgName(), "plugin-version", req.GetPluginPkgVersion()).
			Error("failed to prepare origin plugin package export")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := new(protoFile.ExportPrepareOriginPluginPackageResp)
	resp.ConvertResult(exportID)

	logger.G.Biz(rCtx).With("export-id", exportID).
		Info("prepared origin plugin package export")

	return resp.GetData(), nil
}

const (
	exportDownloadTokenParam = "token"
)

// ExportDownloadOriginPluginPackage downloads an exported origin plugin package.
func (h *handler) ExportDownloadOriginPluginPackage(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	token := rCtx.GContext().Query(exportDownloadTokenParam)
	if token == "" {
		err := errors.New("token is required")
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download origin plugin package export, token is required")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	record, content, err := h.manager.ExportDownloadOriginPluginPackage(rCtx, token)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download origin plugin package export")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	return &restserver.FileResponse{
		Data:        content,
		Size:        record.Size,
		FileName:    record.DownloadName,
		ContentType: restserver.MIMETypeBin,
	}, nil
}
