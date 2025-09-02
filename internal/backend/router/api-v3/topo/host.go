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
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	maxHostLimit = 1000
)

// ListHost lists hosts with page and conditions.
func (h *handler) ListHost(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.TopoHostListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list host, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountHost(
			ctx,
			req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to list host. failed to count host. err: %v", err)
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.TopoHostListResp)
		resp.ConvertHostsFromTypes(num, nil, nil)

		return resp.GetData(), nil
	}

	hosts, num, err := h.storage.ListHostOrderByUpdateTime(
		ctx,
		req.ConvertPageToTypes(maxHostLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list host. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// get credit status.
	creditIDMap := make(map[string]struct{})
	for _, host := range hosts {
		if host.Dynamic.LoginCreditID != "" {
			creditIDMap[host.Dynamic.LoginCreditID] = struct{}{}
		}
	}
	hostCredits, err := h.storageHostCredit.CheckHostCreditValid(ctx, conv.MapKeyToSlice(creditIDMap)...)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list host. failed to get host credit status. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoHostListResp)
	resp.ConvertHostsFromTypes(num, hosts, hostCredits)

	return resp.GetData(), nil
}

// DistinctHost get distinct host fields.
func (h *handler) DistinctHost(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.TopoHostDistinctReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to distinct host, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.storage.DistinctHost(
		ctx,
		types.NewHostDistinctRequestAllSet(),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to distinct host. failed to distinct host fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoHostDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
