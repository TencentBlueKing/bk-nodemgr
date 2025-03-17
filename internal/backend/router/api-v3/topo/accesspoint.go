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
	maxAccessPointLimit = 1000
)

// ListAccessPoint lists accesspoints with page and conditions.
func (h *handler) ListAccessPoint(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoAccessPointListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to list accesspoint, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list accesspoint, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountAccessPoint(
			sCtx,
			req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.Errorf("failed to list accesspoint. failed to count accesspoint. err: %v", err)
			return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
		}

		resp := new(proto.TopoAccessPointListResp)
		resp.ConvertAccessPointsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	accesspoints, num, err := h.storage.ListAccessPoint(
		sCtx,
		req.ConvertPageToTypes(maxAccessPointLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.Errorf("failed to list accesspoint. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(proto.TopoAccessPointListResp)
	resp.ConvertAccessPointsFromTypes(num, accesspoints)

	return resp.GetData(), nil
}
