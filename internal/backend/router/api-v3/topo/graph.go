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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	getGraphNodeGPLimit = 10
)

// GetGraphNode get graph node.
func (h *handler) GetGraphNode(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoGraphNodeGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list graph node, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	requestedIDs := req.GetBkNetworkunitId()
	narrowedIDs, scopeIsAny, authErr := h.narrowAuthorizedNetworkUnitIDs(rCtx, requestedIDs)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to get graph node, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	condition := narrowNetworkUnitCondition(nil, narrowedIDs, scopeIsAny)
	if !scopeIsAny && len(narrowedIDs) == 0 {
		resp := new(protoBackend.TopoGraphNodeGetResp)
		resp.ConvertGrapthNodeInfoFromTypes(nil)

		return resp.GetData(), nil
	}

	networkUnits, _, err := h.storage.ListNetworkUnit(rCtx, types.UnlimitedPage(), condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get graph node, failed to list networkunit")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	networkUnitIDs := make([]int64, len(networkUnits))
	for idx, networkUnit := range networkUnits {
		networkUnitIDs[idx] = networkUnit.ID
	}

	result := make(map[int64]*types.GraphNodeInfo)
	for _, networkUnitID := range networkUnitIDs {
		result[networkUnitID] = &types.GraphNodeInfo{
			NetworkUnitID: networkUnitID,
			IsHealthy:     false,
		}
	}

	// count agent, proxy and check health.
	gp := gopool.NewPool()
	gp.SetLimit(getGraphNodeGPLimit)
	h.countAgents(rCtx, gp, result, networkUnitIDs)
	h.processProxies(rCtx, gp, result, networkUnitIDs)

	// wait until all servers stopped.
	if err := gp.Wait(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get graph node")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoGraphNodeGetResp)
	resp.ConvertGrapthNodeInfoFromTypes(result)

	return resp.GetData(), nil
}

func (h *handler) countAgents(rCtx restserver.IContext, gp gopool.Pool, result map[int64]*types.GraphNodeInfo, ids []int64) {
	if len(ids) == 0 {
		return
	}

	// count total agent
	gp.Go(func() error {
		totalAgents, err := h.storage.CountHostGroupByNetworkUnitID(rCtx, &types.HostCondition{
			DynamicExactInclude: &types.HostDynamicExactFields{
				NetworkUnitID: ids,
				NodeRole:      []types.NodeRole{types.NodeRoleAgent},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to count total agent: %w", err)
		}

		for id, totalAgentCount := range totalAgents {
			result[id].TotalAgent = totalAgentCount
		}

		return nil
	})

	// count running agent
	gp.Go(func() error {
		runningAgents, err := h.storage.CountHostGroupByNetworkUnitID(rCtx, &types.HostCondition{
			DynamicExactInclude: &types.HostDynamicExactFields{
				NetworkUnitID: ids,
				NodeRole:      []types.NodeRole{types.NodeRoleAgent},
				NodeStatus:    []types.NodeStatus{types.NodeStatusRunning},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to count running agent: %w", err)
		}
		for id, runningAgentCount := range runningAgents {
			result[id].RunningAgent = runningAgentCount
		}

		return nil
	})
}

func (h *handler) processProxies(rCtx restserver.IContext, gp gopool.Pool, result map[int64]*types.GraphNodeInfo, ids []int64) {
	gp.Go(func() error {
		proxyDatas, err := h.getProxyDatas(rCtx, ids)
		if err != nil {
			return fmt.Errorf("failed to get proxy data: %w", err)
		}

		for id, proxyData := range proxyDatas {
			result[id].TotalProxy = proxyData.totalProxy
			result[id].RunningProxy = proxyData.runningProxy
			result[id].CycleTimes = proxyData.connCycleTime

			// healthy networkunit should have running proxy and satisfied all required tags.
			if proxyData.runningProxy > 0 && len(proxyData.requiredTagSet) == 0 {
				result[id].IsHealthy = true
			}
		}

		return nil
	})
}

type proxyData struct {
	totalProxy     int64
	runningProxy   int64
	requiredTagSet map[types.ProxyTag]struct{}
	connCycleTime  []types.CycleTime
}

func (h *handler) getProxyDatas(rCtx restserver.IContext, networkUnitIDs []int64) (map[int64]*proxyData, error) {
	if len(networkUnitIDs) == 0 {
		return make(map[int64]*proxyData), nil
	}

	proxies, _, err := h.storage.ListHost(rCtx, types.UnlimitedPage(), &types.HostCondition{
		DynamicExactInclude: &types.HostDynamicExactFields{
			NetworkUnitID: networkUnitIDs,
			NodeRole:      []types.NodeRole{types.NodeRoleProxy},
		},
	})
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*proxyData)
	for _, proxy := range proxies {
		if _, ok := result[proxy.Dynamic.NetworkUnitID]; !ok {
			// fill required tag set.
			// if no tag required, it will be empty, which means all tags are required.
			requiredTagSet := make(map[types.ProxyTag]struct{})
			for _, tag := range types.AllProxyTag() {
				requiredTagSet[tag] = struct{}{}
			}

			result[proxy.Dynamic.NetworkUnitID] = &proxyData{
				totalProxy:     0,
				runningProxy:   0,
				requiredTagSet: requiredTagSet,
				connCycleTime:  make([]types.CycleTime, 0),
			}
		}

		data := result[proxy.Dynamic.NetworkUnitID]
		data.totalProxy++

		if proxy.Dynamic.NodeStatus != types.NodeStatusRunning {
			continue
		}

		data.runningProxy++
		if proxy.Dynamic.ConnCycleTime != "" {
			data.connCycleTime = append(data.connCycleTime, types.CycleTime{
				HostID:    proxy.HostID,
				InnerIP:   proxy.Static.InnerIPList,
				InnerIPV6: proxy.Static.InnerIPV6List,
				AgentID:   proxy.Dynamic.AgentID,
				Time:      proxy.Dynamic.ConnCycleTime,
			})
		}

		// if no tag required, skip
		if len(data.requiredTagSet) == 0 {
			continue
		}

		for _, tag := range proxy.Dynamic.ProxyTags {
			delete(data.requiredTagSet, tag)
		}
	}

	return result, nil
}
