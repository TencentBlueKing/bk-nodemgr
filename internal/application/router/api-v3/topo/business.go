/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

const (
	// not max limit in business.
	// return all data in one request.
	maxBusinessLimit = 0
)

// ListBusiness list business with specified conditions.
func (h *handler) ListBusiness(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.TopoBusinessListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list business, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	bizs, num, err := h.backendHandler.ListBusiness(
		ctx,
		req.ConvertPageToTypes(maxBusinessLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list business: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoBusinessListResp)
	resp.ConvertBusinessFromTypes(num, bizs)

	return resp.GetData(), nil
}
