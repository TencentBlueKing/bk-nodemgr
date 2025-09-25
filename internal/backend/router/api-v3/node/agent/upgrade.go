/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package agent ...
package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AgentUpgrade upgrade agent.
func (h *handler) AgentUpgrade(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeAgentUpgradeReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to upgrade agent, failed to decode request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, err := h.getUpgradeNodeHosts(rCtx, req.GetHost())
	if err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to upgrade agent, failed to get host list: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	unitsMap, err := h.generatesUnitDirectLink(rCtx, hosts)
	if err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to upgrade agent, failed to get network unit info: %v", err)
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	reqHosts := req.GetHost()
	nodeDeploys := make([]*types.NodeDeployment, len(hosts))
	for idx := range reqHosts {
		reqHost := reqHosts[idx]

		nodeDeploy, err := h.generatesUpgradeDeploys(rCtx.TenantID(), reqHost, hosts, unitsMap)
		if err != nil {
			h.logger.ErrorCtxf(rCtx, "failed to upgrade agent, failed to generate node deployment: %v", err)

			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		nodeDeploys[idx] = nodeDeploy
	}

	workflowID, err := h.manager.LaunchUpgradeNode(rCtx, manager.UpgradeNodeParam{
		Type:            types.NodeWorkflowTypeUpgradeAgent,
		BizIDs:          h.getUpgradeNodeBizIDs(hosts),
		Operator:        rCtx.BKUsername(),
		NodeDeployments: nodeDeploys,
	})
	if err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to upgrade agent: %v", err)
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentUpgradeResp)
	resp.ConvertWorkflowID(workflowID)

	h.logger.InfoCtxf(rCtx, "launched upgrade agent workflow: %s", workflowID)

	return resp.GetData(), nil
}

func (h *handler) generatesUnitDirectLink(nCtx contextx.IContext, hostMap map[int64]*types.Host) (map[int64]bool, error) {
	unitInfoMap := make(map[int64]bool)

	unitIDs := make([]int64, 0, len(hostMap))
	for _, host := range hostMap {
		unitIDs = append(unitIDs, host.Dynamic.NetworkUnitID)
	}

	units, _, err := h.storageNetworkUnit.ListNetworkUnit(nCtx,
		types.UnlimitedPage(),
		&types.NetworkUnitCondition{
			ExactInclude: &types.NetworkUnitExactFields{
				NetworkUnitID: unitIDs,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("list network unit failed: %w", err)
	}

	for _, unit := range units {
		unitInfoMap[unit.ID] = unit.IsDirect
	}

	return unitInfoMap, nil
}

func (h *handler) getUpgradeNodeHosts(
	ctx context.Context, reqHosts []*protoBackend.NodeAgentUpgradeReq_Host) (map[int64]*types.Host, error) {

	if len(reqHosts) == 0 {
		return nil, errors.New("empty host list")
	}

	hostIDs := make(map[int64]struct{})
	for _, host := range reqHosts {
		hostIDs[host.GetBkHostId()] = struct{}{}
	}

	hosts, _, err := h.storageHost.ListHost(ctx,
		types.UnlimitedPage(),
		&types.HostCondition{ExactInclude: &types.HostExactFields{
			HostID: conv.MapKeyToSlice(hostIDs),
		}})

	result := make(map[int64]*types.Host)
	for _, host := range hosts {
		result[host.HostID] = host
	}

	return result, err
}

func (h *handler) getUpgradeNodeBizIDs(hostMap map[int64]*types.Host) []int64 {
	bizIDs := make(map[int64]struct{})
	for _, host := range hostMap {
		bizIDs[host.Static.BizID] = struct{}{}
	}

	return conv.MapKeyToSlice(bizIDs)
}

func (h *handler) generatesUpgradeDeploys(
	tenantID string,
	reqHost *protoBackend.NodeAgentUpgradeReq_Host,
	hostMap map[int64]*types.Host,
	unitsMap map[int64]bool) (*types.NodeDeployment, error) {

	hostID := reqHost.GetBkHostId()
	host, ok := hostMap[hostID]
	if !ok {
		return nil, fmt.Errorf("host not found. host-id(%d)", hostID)
	}

	nodeDeployment := types.NewNodeDeployment(&types.DeploymentInfo{
		Host: types.Host{
			TenantID: tenantID,
			HostID:   host.HostID,
			Static:   host.Static,
			Dynamic:  host.Dynamic,
		},
		RestartOptions: types.DeploymentRestartOptions{
			ForceRestart:           reqHost.GetForce(),
			GracefulRestartTimeout: time.Second * time.Duration(reqHost.GetGracefulRestartTimeoutSec()),
		},
		UpgradeOptions: types.DeploymentUpgradeOptions{
			DirectLink: unitsMap[host.Dynamic.NetworkUnitID],
		},
	})

	// set target version.
	nodeDeployment.Info.Host.Dynamic.NodeVersion = reqHost.GetTargetVersion()

	return nodeDeployment, nil
}
