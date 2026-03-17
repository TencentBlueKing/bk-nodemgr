/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package info

import (
	"fmt"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// Installer returns file info for the installer.
// Unlike release-plugin info handlers that resolve files through EnsureNodeToLocal,
// this handler reads directly from the local installerFileGroup directory via GetInstaller.
// The installer binary is pre-placed on disk; no remote fetch or filecache lookup is needed.
func (h *handler) Installer(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.DownloadInstallerReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get installer info, failed to bind json")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	osType, err := platfmt.NormalizeOS(req.GetOsType())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("normalize os failed: %w", err))
	}

	cpuArch, err := platfmt.NormalizeArch(req.GetCpuArch())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("normalize arch failed: %w", err))
	}

	file, err := h.manager.GetInstaller(rCtx, osType, cpuArch)
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("get installer failed: %w", err))
	}

	info := file.Info()

	return buildInfoResp(info.Name, info.Size, info.MD5), nil
}
