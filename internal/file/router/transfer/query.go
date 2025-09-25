/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package transfer

import (
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// TransferQuery query transfer package.
func (h *handler) TransferQuery(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.TransferQueryReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to query transfer, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	upload, download, err := h.manager.QueryTransfer(rCtx, req.GetTaskId())
	if err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to query transfer. err: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	resp := new(protoFile.TransferQueryResp)
	resp.ConvertResult(upload, download)

	return resp.GetData(), nil
}
