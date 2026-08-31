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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ExportPrepareOriginPluginPackage prepares an origin plugin package export artifact.
func (h *handler) ExportPrepareOriginPluginPackage(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.ExportPrepareOriginPluginPackageReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to prepare origin plugin package export, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	param := types.ExportPrepareParam{
		PluginPkgName:       req.GetPluginPkgName(),
		PluginPkgVersion:    req.GetPluginPkgVersion(),
		UploadOriginPkgType: types.UploadCategory(req.GetUploadOriginPkgType()),
		UploadIDs:           req.GetUploadIds(),
		FileNameSuffix:      req.GetFileNameSuffix(),
	}
	info, address, err := h.manager.ExportPrepareOriginPluginPackage(rCtx, param)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("file-name-suffix", req.GetFileNameSuffix(), "plugin-name", req.GetPluginPkgName(),
				"plugin-version", req.GetPluginPkgVersion(), "upload-origin-pkg-type", req.GetUploadOriginPkgType()).
			Error("failed to prepare origin plugin package export")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := new(protoFile.ExportPrepareOriginPluginPackageResp)
	downloadAddress := fmt.Sprintf("%s%s%s?%s=", address, h.rg.BasePath(), originPluginPackageDownloadRoute, exportDownloadTokenParam)
	resp.ConvertResult(info, downloadAddress)

	logger.G.Biz(rCtx).With("file-name-suffix", req.GetFileNameSuffix(), "filename", info.Name).Info("prepared origin plugin package export")

	return resp.GetData(), nil
}

// ExportDownloadOriginPluginPackage downloads an exported origin plugin package.
func (h *handler) ExportDownloadOriginPluginPackage(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	token := rCtx.GContext().Query(exportDownloadTokenParam)
	if token == "" {
		err := errors.New("token is required")
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download origin plugin package export, token is required")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := h.manager.ExportDownloadOriginPluginPackage(rCtx, token)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download origin plugin package export")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	content, err := file.Content(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to open origin plugin package export content")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	info := file.Info()

	return &restserver.FileResponse{
		Data:        content,
		Size:        info.Size,
		FileName:    info.Name,
		ContentType: restserver.MIMETypeBin,
	}, nil
}
