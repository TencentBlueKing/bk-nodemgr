/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package download is the file download router.
package download

import (
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/file"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg          *gin.RouterGroup
	fileManager file.IFileManager
}

// Agent download agent package.
func (h *handler) Agent(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	req := new(protoFile.DownloadAgentReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("bind json failed", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	platform, err := platfmt.Normalize(req.GetOsType(), req.GetCpuArch())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if !platform.Validate() {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("invalid platform(%s)", req.GetOsType()))
	}

	pkgName, err := nodepkg.FormatPkgFileName(
		types.Generation2,
		types.ReleaseTypeAgent,
		platform,
		req.GetVersion(),
	)
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	logger.G.Biz(rCtx).With("filename", pkgName).Info("get agent package")

	file, err := h.fileManager.GetFile(rCtx, pkgName)
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("get file failed: %w", err))
	}

	reader, err := file.Content(rCtx)
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("get file content failed: %w", err))
	}

	info := file.Info()
	resp := &restserver.FileResponse{
		Data:        reader,
		Size:        info.Size,
		FilePath:    filepath.Join(".", info.Name),
		FileName:    info.Name,
		ContentType: "application/octet-stream",
	}

	return resp, nil
}

// Proxy download proxy package.
func (h *handler) Proxy(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	req := new(protoFile.DownloadProxyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("bind json failed", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	platform, err := platfmt.Normalize(req.GetOsType(), req.GetCpuArch())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if !platform.Validate() {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("invalid platform(%s)", req.GetOsType()))
	}

	pkgName, err := nodepkg.FormatPkgFileName(
		types.Generation2,
		types.ReleaseTypeProxy,
		platform,
		req.GetVersion(),
	)
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	logger.G.Biz(rCtx).With("filename", pkgName).Info("get proxy package")

	file, err := h.fileManager.GetFile(rCtx, pkgName)
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("get file failed: %w", err))
	}

	reader, err := file.Content(rCtx)
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("get file content failed: %w", err))
	}

	info := file.Info()
	resp := &restserver.FileResponse{
		Data:        reader,
		Size:        info.Size,
		FilePath:    filepath.Join(".", info.Name),
		FileName:    info.Name,
		ContentType: "application/octet-stream",
	}

	return resp, nil
}

// Installer download installer package.
func (h *handler) Installer(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	req := new(protoFile.DownloadInstallerReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download installer, failed to bind json")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	osType, err := platfmt.NormalizeOS(req.GetOsType())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download installer, failed to normalize os")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("normalize os failed: %w", err))
	}

	cpuArch, err := platfmt.NormalizeArch(req.GetCpuArch())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download installer, failed to normalize arch")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("normalize arch failed: %w", err))
	}

	toolName, err := tool.FormatInstallerName(osType, cpuArch)
	if err != nil {
		return nil, err
	}

	file, err := h.fileManager.GetFile(rCtx, toolName)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download installer, failed to get installer")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("get installer failed: %w", err))
	}

	reader, err := file.Content(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download installer, failed to get file content")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("get file content failed: %w", err))
	}

	info := file.Info()
	resp := &restserver.FileResponse{
		Data:        reader,
		Size:        info.Size,
		FilePath:    filepath.Join(".", info.Name),
		FileName:    info.Name,
		ContentType: "application/octet-stream",
	}

	return resp, nil
}

func newHandler(rg *gin.RouterGroup, opt *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:          rg.Group("/download"),
		fileManager: opt.FileManager,
	}
}

// Load enables web router into gin.Engine.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/agent", restserver.FileHandler(h.Agent))
	h.rg.POST("/proxy", restserver.FileHandler(h.Proxy))
	h.rg.POST("/installer", restserver.FileHandler(h.Installer))
}
