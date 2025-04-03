/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package download

import (
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Proxy download agent package.
func (h *handler) Proxy(ctx *rest.Context) (*rest.FileResponse, error) {
	req := new(protoFile.DownloadProxyReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Error("bind json failed", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	pkgName, err := nodepkg.FormatPkgName(
		types.NodeRoleProxy,
		types.NodeGeneration(req.GetGeneration()),
		req.GetVersion(),
		req.GetCpuArch(),
		req.GetOsType(),
	)
	if err != nil {
		h.logger.Error("format pkg name failed", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	file, err := h.proxyFileGroup.GetFile(pkgName)
	if err != nil {
		return nil, errf.ErrWrap(errf.InvalidParameter, fmt.Errorf("get file failed, err: %w", err))
	}

	reader, err := file.Content()
	if err != nil {
		return nil, errf.ErrWrap(errf.InvalidParameter, fmt.Errorf("get file content failed, err: %w", err))
	}

	info := file.Info()
	resp := &rest.FileResponse{
		Data:        reader,
		Size:        info.Size,
		FilePath:    filepath.Join(".", info.Name),
		FileName:    info.Name,
		ContentType: "application/octet-stream",
	}

	return resp, nil
}
