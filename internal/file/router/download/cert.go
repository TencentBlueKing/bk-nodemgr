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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Cert download cert package.
func (h *handler) Cert(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	req := new(protoFile.DownloadCertReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("bind json failed")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, _, err := h.manager.EnsureCertToLocal(rCtx, types.Generation(req.GetGeneration()))
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
