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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
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

	networkUnits, _, err := h.storage.ListNetworkUnit(rCtx, types.UnlimitedPage(), condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get graph node, failed to list networkunit")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	networkUnitIDs := make([]int64, len(networkUnits))
	for idx, networkUnit := range networkUnits {
		networkUnitIDs[idx] = networkUnit.ID
	}

	result := make(map[int64]*protoBackend.GraphNodeInfo)
	for _, networkUnitID := range networkUnitIDs {
		result[networkUnitID] = &protoBackend.GraphNodeInfo{
			BkNetworkunitID: networkUnitID,
			IsHealthy:       true,
		}
	}

	// count agent, proxy and check health.
	gp := gopool.NewPool()

	h.countAgents(rCtx, gp, result, networkUnitIDs)

	for _, networkUnitID := range networkUnitIDs {
		id := networkUnitID

		h.processProxies(rCtx, gp, result, id)
	}

	// wait until all servers stopped.
	if err := gp.Wait(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get graph node")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoGraphNodeGetResp)
	resp.ConvertGrapthNodeInfoFromTypes(result)

	return resp.GetData(), nil
}

func (h *handler) countAgents(rCtx restserver.IContext, gp gopool.Pool, result map[int64]*protoBackend.GraphNodeInfo, ids []int64) {
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

func (h *handler) processProxies(rCtx restserver.IContext, gp gopool.Pool, result map[int64]*protoBackend.GraphNodeInfo, id int64) {
	gp.Go(func() error {
		proxyData, err := h.getProxyData(rCtx, id)
		if err != nil {
			return fmt.Errorf("failed to get proxy data, networkunit-id(%d): %w", id, err)
		}

		result[id].TotalProxy = proxyData.totalProxy
		result[id].RunningProxy = proxyData.runningProxy

		// if proxy is not running or required tag is not empty, set node unhealthy
		if proxyData.runningProxy == 0 || len(proxyData.requiredTagSet) > 0 {
			result[id].IsHealthy = false

			return nil
		}

		// get cycle times
		cycleTimes, err := h.getAgentCycleTimes(rCtx, proxyData.agentIDs)
		if err != nil {
			return fmt.Errorf("failed to get agent cycle times, networkunit-id(%d): %w", id, err)
		}
		result[id].CycleTimes = cycleTimes

		return nil
	})
}

type proxyData struct {
	totalProxy     int64
	runningProxy   int64
	requiredTagSet map[types.ProxyTag]struct{}
	agentIDs       []string
}

func (h *handler) getProxyData(rCtx restserver.IContext, networkUnitID int64) (*proxyData, error) {
	proxies, totalProxy, err := h.storage.ListHost(rCtx, types.UnlimitedPage(), &types.HostCondition{
		DynamicExactInclude: &types.HostDynamicExactFields{
			NetworkUnitID: []int64{networkUnitID},
			NodeRole:      []types.NodeRole{types.NodeRoleProxy},
		},
	})
	if err != nil {
		return nil, err
	}

	// fill required tag set.
	// if no tag required, it will be empty, which means all tags are required
	allRequiredTags := types.AllProxyTag()
	requiredTagSet := make(map[types.ProxyTag]struct{}, len(allRequiredTags))
	for _, tag := range allRequiredTags {
		requiredTagSet[tag] = struct{}{}
	}

	data := &proxyData{
		totalProxy:     totalProxy,
		runningProxy:   0,
		requiredTagSet: requiredTagSet,
		agentIDs:       make([]string, 0),
	}

	for _, proxy := range proxies {
		if proxy.Dynamic.NodeStatus != types.NodeStatusRunning {
			continue
		}

		// count running proxy
		data.runningProxy++
		data.agentIDs = append(data.agentIDs, proxy.Dynamic.AgentID)

		// if no tag required, skip
		if len(data.requiredTagSet) == 0 {
			continue
		}

		for _, tag := range proxy.Dynamic.ProxyTags {
			delete(data.requiredTagSet, tag)
		}
	}

	return data, nil
}

func (h *handler) getAgentCycleTimes(rCtx restserver.IContext, agentIDs []string) ([]string, error) {
	if len(agentIDs) == 0 {
		return []string{}, nil
	}

	agentInfos, err := h.gseHandler.ListAgentInfo(rCtx, agentIDs...)
	if err != nil {
		return nil, err
	}

	cycleTimes := make([]string, len(agentInfos))
	for idx, info := range agentInfos {
		cycleTimes[idx] = info.ConnCycleTime
	}

	return cycleTimes, nil
}
