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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// UploadOriginAgent upload origin agent.
func (h *handler) UploadOriginAgent(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to upload agent, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoFile.UploadOriginAgentReq)
	fileHeader, err := ctx.ParseFileForm(req)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload agent, failed to parse file form: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload agent, failed to open file. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}
	defer func() {
		_ = file.Close()
	}()

	info, err := h.manager.UploadOriginAgent(sCtx, types.Generation(req.GetGeneration()), file)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to upload agent: %v", err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	h.logger.InfoCtxf(sCtx, "uploaded origin agent, generation: %d", req.GetGeneration)

	resp := &protoFile.UploadOriginAgentResp{
		Name: info.Name,
		Size: info.Size,
		Md5:  info.MD5,
	}

	return resp, nil
}

// Agent upload agent package.
func (h *handler) Agent(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to upload agent, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoFile.UploadAgentReq)
	fileHeader, err := ctx.ParseFileForm(req)
	if err != nil {
		h.logger.Errorf("parse file form failed", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.Error("open file failed", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	tenantCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Error("get context failed", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	osType, err := platform.NormalizeOS(req.GetOsType())
	if err != nil {
		h.logger.Error("normalize os type failed", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	cpuArch, err := platform.NormalizeArch(req.GetCpuArch())
	if err != nil {
		h.logger.Error("normalize cpu arch failed", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	pkgName, err := nodepkg.FormatPkgName(
		types.Generation(req.GetGeneration()),
		types.ReleaseTypeAgent,
		platform.Platform{
			OS:   osType,
			Arch: cpuArch,
		},
		req.GetVersion(),
	)
	if err != nil {
		h.logger.Error("format pkg name failed", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.agentFileGroup.Store(
		tenantCtx,
		iface.FileInfo{
			Name: pkgName,
		},
		file,
		req.GetOverwrite()); err != nil {
		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	localFile, err := h.agentFileGroup.GetFile(sCtx, pkgName)
	if err != nil {
		h.logger.Error("get local uploadFile failed", err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	info, _ := localFile.Info(sCtx)
	resp := &protoFile.UploadAgentResp{
		Name: info.Name,
		Size: info.Size,
		Md5:  info.MD5,
	}

	return resp, nil
}
