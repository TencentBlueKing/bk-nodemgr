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

package download

import (
	"fmt"
	"path/filepath"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Plugin download plugin package.
func (h *handler) Plugin(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	req := new(protoFile.DownloadPluginReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("bind json failed")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginPkgName := req.GetPluginPkgName()

	os, err := platfmt.NormalizeOS(req.GetOsType())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("normalize os failed: %w", err))
	}

	arch, err := platfmt.NormalizeArch(req.GetCpuArch())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("normalize arch failed: %w", err))
	}

	version := req.GetVersion()
	platform := platfmt.Platform{OS: os, Arch: arch}

	file, _, err := h.manager.EnsurePluginToLocal(rCtx, pluginPkgName, types.Generation2, platform, version)
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
		ContentType: restserver.MIMETypeBin,
	}

	return resp, nil
}
