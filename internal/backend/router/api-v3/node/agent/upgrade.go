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
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
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
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, err := h.getUpgradeNodeHosts(rCtx, req.GetHost())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade agent, failed to get host list")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := validateUpgradeHostNetworkUnit(req.GetHost(), hosts); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade agent, invalid network unit")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	bizIDs := h.getUpgradeNodeBizIDs(hosts)
	resources := buildBizResources(bizIDs)
	if authErr := h.authorizer.Check(rCtx, auth.ActionAgentOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to upgrade agent, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	extraUnitIDs := make([]int64, 0)
	for _, reqHost := range req.GetHost() {
		if id := reqHost.GetBkNetworkunitId(); id >= 0 {
			extraUnitIDs = append(extraUnitIDs, id)
		}
	}
	unitsMap, err := h.generatesUnitDirectLinkWithExtra(rCtx, hosts, extraUnitIDs)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade agent, failed to get network unit info")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}
	reqHosts := req.GetHost()
	nodeDeploys := make([]*types.NodeDeployment, len(reqHosts))
	for idx := range reqHosts {
		reqHost := reqHosts[idx]

		nodeDeploy, err := h.generatesUpgradeDeploys(rCtx.TenantID(), reqHost, hosts, unitsMap)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade agent, failed to generate node deployment")

			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		nodeDeploys[idx] = nodeDeploy
	}

	workflowID, err := h.nodeMgrIface.LaunchUpgradeNode(rCtx, types.UpgradeNodeParam{
		Type:            types.NodeWorkflowTypeUpgradeAgent,
		BizIDs:          bizIDs,
		Operator:        rCtx.BKUsername(),
		NodeDeployments: nodeDeploys,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade agent")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentUpgradeResp)
	resp.ConvertWorkflowID(workflowID)

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched upgrade agent workflow")

	return resp.GetData(), nil
}

func (h *handler) generatesUnitDirectLink(nCtx contextx.IContext, hostMap map[int64]*types.Host) (map[int64]bool, error) {
	return h.generatesUnitDirectLinkWithExtra(nCtx, hostMap, nil)
}

func (h *handler) generatesUnitDirectLinkWithExtra(
	nCtx contextx.IContext, hostMap map[int64]*types.Host, extraUnitIDs []int64) (map[int64]bool, error) {

	unitInfoMap := make(map[int64]bool)

	unitIDSet := make(map[int64]struct{})
	for _, host := range hostMap {
		unitIDSet[host.Dynamic.NetworkUnitID] = struct{}{}
	}
	for _, id := range extraUnitIDs {
		unitIDSet[id] = struct{}{}
	}

	units, _, err := h.storageNetworkUnit.ListNetworkUnit(nCtx,
		types.UnlimitedPage(),
		&types.NetworkUnitCondition{
			ExactInclude: &types.NetworkUnitExactFields{
				NetworkUnitID: conv.MapKeyToSlice(unitIDSet),
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
	nCtx contextx.IContext, reqHosts []*protoBackend.NodeAgentUpgradeReq_Host) (map[int64]*types.Host, error) {

	if len(reqHosts) == 0 {
		return nil, errors.New("empty host list")
	}

	hostIDs := make(map[int64]struct{})
	for _, host := range reqHosts {
		hostIDs[host.GetBkHostId()] = struct{}{}
	}

	hosts, _, err := h.storageHost.ListHost(nCtx,
		types.UnlimitedPage(),
		&types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
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

	// Use target networkunit if specified; otherwise keep the host's current networkunit.
	targetNetworkUnitID := resolveUpgradeNetworkUnitID(host.Dynamic.NetworkUnitID, reqHost.GetBkNetworkunitId())

	deployHost := host.Dynamic
	deployHost.NetworkUnitID = targetNetworkUnitID

	nodeDeployment := types.NewNodeDeployment(&types.DeploymentInfo{
		Host: types.Host{
			TenantID: tenantID,
			HostID:   host.HostID,
			Static:   host.Static,
			Dynamic:  deployHost,
		},
		RestartOptions: types.DeploymentRestartOptions{
			ForceRestart:           reqHost.GetForce(),
			GracefulRestartTimeout: time.Second * time.Duration(reqHost.GetGracefulRestartTimeoutSec()),
		},
		UpgradeOptions: types.DeploymentUpgradeOptions{
			DirectLink: unitsMap[targetNetworkUnitID],
		},
	})

	// set target version.
	nodeDeployment.Info.Host.Dynamic.NodeVersion = reqHost.GetTargetVersion()

	return nodeDeployment, nil
}
