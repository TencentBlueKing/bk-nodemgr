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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// Remote downloads a verified remote file.
func (h *handler) Remote(rCtx restserver.IContext) (*restserver.FileResponse, error) {
	req := new(protoFile.DownloadRemoteFileReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download remote file, failed to bind json")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, err := h.manager.DownloadRemoteFile(rCtx, req.GetFilename(), req.GetDownloadUrl(), req.GetMd5())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to download remote file")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	content, err := file.Content(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to open remote file content")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	// The FileHandler closes content after the response is written; it must not be closed here.
	return &restserver.FileResponse{
		Data:        content,
		Size:        file.Info().Size,
		FileName:    file.Info().Name,
		ContentType: restserver.MIMETypeBin,
	}, nil
}
