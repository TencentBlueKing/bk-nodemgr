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

package topo

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"

	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
)

// GetGraph gets a graph descriptions.
func (h *handler) GetGraph(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoGraphGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get graph, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// list networkunits.
	networkUnits, _, err := h.backendHandler.ListNetworkUnit(
		rCtx,
		types.UnlimitedPage(),
		&types.NetworkUnitCondition{
			ExactInclude: &types.NetworkUnitExactFields{
				NetworkAreaID: req.GetBkNetworkareaId(),
			},
		})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get graph, failed to list networkunit")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// generates links.
	resp := new(protoApplication.TopoGraphGetResp)
	resp.ConvertNetworkUnitsToTypes(networkUnits)

	return resp.GetData(), nil
}

// GetGraphNode gets a graph node.
func (h *handler) GetGraphNode(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoGraphNodeGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get graph node, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	graphNodeInfos, err := h.backendHandler.GetGraphNode(
		rCtx,
		req.GetBkNetworkunitId(),
	)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get graph node, failed to get graph node")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoGraphNodeGetResp)
	resp.ConvertGrapthNodeInfoFromTypes(graphNodeInfos)

	return resp.GetData(), nil
}
