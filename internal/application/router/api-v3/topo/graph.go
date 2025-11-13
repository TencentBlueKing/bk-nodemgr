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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
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
		types.Page{Limit: maxNetworkUnitLimit},
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

// CountGraphNode counts graph nodes.
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (h *handler) CountGraphNode(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoGraphNodeCountReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count graph node, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkUnitIDs := req.GetBkNetworkunitId()
	if len(networkUnitIDs) == 0 {
		networkUnits, _, err := h.backendHandler.ListNetworkUnit(rCtx, types.Page{}, nil)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count graph node, failed to list networkunit")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		networkUnitIDs = make([]int64, len(networkUnits))
		for idx, networkUnit := range networkUnits {
			networkUnitIDs[idx] = networkUnit.ID
		}
	}

	// init result map.
	result := make(map[int64]*protoApplication.NetworkUnitInfo)
	for _, networkUnitID := range networkUnitIDs {
		result[networkUnitID] = &protoApplication.NetworkUnitInfo{Proxy: 0, Agent: 0}
	}

	gp := gopool.NewPool()
	for _, networkUnitID := range req.GetBkNetworkunitId() {
		id := networkUnitID

		// count agent.
		gp.Go(func() error {
			num, err := h.backendHandler.CountHost(rCtx, &types.HostCondition{
				ExactInclude: &types.HostExactFields{
					NetworkUnitID: []int64{id},
					NodeRole:      []types.NodeRole{types.NodeRoleAgent},
				},
			})
			if err != nil {
				return errors.Join(err, fmt.Errorf("failed to count agent, networkunit-id: %d", id))
			}

			result[id].Agent = num

			return nil
		})

		// count proxy.
		gp.Go(func() error {
			num, err := h.backendHandler.CountHost(rCtx, &types.HostCondition{
				ExactInclude: &types.HostExactFields{
					NetworkUnitID: []int64{id},
					NodeRole:      []types.NodeRole{types.NodeRoleProxy},
				},
			})
			if err != nil {
				return errors.Join(err, fmt.Errorf("failed to count proxy, networkunit-id: %d", id))
			}

			result[id].Proxy = num

			return nil
		})
	}

	// wait until all servers stopped or application error.
	if err := gp.Wait(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count graph node, failed to count host: %v", err)

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoGraphNodeCountResp)
	resp.ConvertNetworkUnitInfoResult(result)

	return resp.GetData(), nil
}
