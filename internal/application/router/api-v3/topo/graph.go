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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"

	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
)

// GetGraph gets a graph descriptions.
func (h *handler) GetGraph(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoGraphGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to get graph, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to get graph, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// list networkunits.
	networkUnits, _, err := h.backendHandler.ListNetworkUnit(
		sCtx,
		types.Page{Limit: maxNetworkUnitLimit},
		&types.NetworkUnitCondition{
			Type: types.ConditionTypeExactInclude,
			Exact: &types.NetworkUnitExactFields{
				NetworkAreaID: req.GetBkNetworkareaId(),
			},
		})
	if err != nil {
		h.logger.Errorf("failed to get graph, failed to list networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// generates links.
	resp := new(proto.TopoGraphGetResp)
	resp.ConvertNetworkUnitsToTypes(networkUnits)

	return resp.Data, nil
}
