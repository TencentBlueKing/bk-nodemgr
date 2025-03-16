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
	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

const (
	maxHostLimit = 1000
)

// ListHost lists hosts with page and conditions.
func (h *handler) ListHost(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoHostListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to list host, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list host, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountHost(
			sCtx,
			req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.Errorf("failed to list host. failed to count host. err: %v", err)
			return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
		}

		resp := new(proto.TopoHostListResp)
		resp.ConvertHostsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	hosts, num, err := h.storage.ListHost(
		sCtx,
		req.ConvertPageToTypes(maxHostLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.Errorf("failed to list host. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(proto.TopoHostListResp)
	resp.ConvertHostsFromTypes(num, hosts)

	return resp.GetData(), nil
}
