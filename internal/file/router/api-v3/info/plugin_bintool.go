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

package info

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// PluginBinTool returns file info for the plugin bintool package.
func (h *handler) PluginBinTool(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.DownloadPluginBinToolReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get plugin bintool info, failed to bind json")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	file, _, err := h.manager.EnsurePluginBinToolToLocal(rCtx, types.Generation(req.GetGeneration()), req.GetName())
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("get file failed: %w", err))
	}

	info := file.Info()

	return buildInfoResp(info.Name, info.Size, info.MD5), nil
}
