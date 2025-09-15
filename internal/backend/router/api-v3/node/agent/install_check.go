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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AgentInstallCheck checks if an agent can be installed on hosts.
func (h *handler) AgentInstallCheck(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeAgentInstallCheckReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to check install agent, failed to decode request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	results, err := h.checkInstall(ctx, req.GetHost())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to check install agent: %v", err)
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentInstallCheckResp)

	resp.ConvertResultFromTypes(results, len(results))

	return resp.GetData(), nil
}

func (h *handler) checkInstall(ctx contextx.ITenantContext,
	hosts []*protoBackend.NodeAgentInstallCheckReq_Host) ([]*types.NodeAgentInstallCheckResult, error) {

	unitIDToAreaMap, err := h.getNetworkAreaByNetworkUnit(ctx, hosts)
	if err != nil {
		return nil, err
	}

	gp := gopool.NewPool()
	results := make([]*types.NodeAgentInstallCheckResult, len(hosts))

	for i := range hosts {
		idx := i
		gp.Go(func() error {
			bizID := hosts[idx].GetBkBizId()
			innerIP := hosts[idx].GetBkHostInnerip()
			networkunitID := hosts[idx].GetBkNetworkunitId()
			networkareaID := unitIDToAreaMap[networkunitID]

			result, err := h.checkInstallEligibility(ctx, innerIP, networkareaID, bizID)
			if err != nil {
				h.logger.ErrorCtxf(ctx, "failed to check install eligibility. inner-ip(%s), network-unit-id(%d), biz-id(%d): %v",
					innerIP, networkunitID, bizID, err)

				return err
			}
			results[idx] = result

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// getNetworkAreaByNetworkUnit get network area id by network unit id.
func (h *handler) getNetworkAreaByNetworkUnit(ctx contextx.ITenantContext, hosts []*protoBackend.NodeAgentInstallCheckReq_Host) (
	map[int64]int64, error) {

	unitIDs := make([]int64, 0, len(hosts))
	for _, host := range hosts {
		unitIDs = append(unitIDs, host.GetBkNetworkunitId())
	}

	networkUnits, err := h.domainNodeInstall.GetNetworkUnitByAreaIDs(ctx, unitIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list network units: %w", err)
	}

	unitIDToAreaMap := make(map[int64]int64, len(networkUnits))
	for _, networkUnit := range networkUnits {
		unitIDToAreaMap[networkUnit.ID] = networkUnit.NetworkAreaID
	}

	return unitIDToAreaMap, nil
}

func (h *handler) checkInstallEligibility(ctx contextx.ITenantContext, innerIP string,
	networkAreaID, bizID int64) (*types.NodeAgentInstallCheckResult, error) {

	hosts, err := h.domainNodeInstall.GetHostsByAreaAndIP(ctx, networkAreaID, innerIP)
	if err != nil {
		return nil, err
	}

	return h.determineHost(hosts, innerIP, bizID), nil
}

func (h *handler) determineHost(
	hosts []*types.Host, innerIP string, bizID int64) *types.NodeAgentInstallCheckResult {

	result := &types.NodeAgentInstallCheckResult{
		InnerIP: innerIP,
	}
	switch {
	// if no install record found, we can clean install agent.
	case len(hosts) == 0:
		result.InstallEligibilitiy = types.NodeAgentInstallEligibilityNormal
		return result

	// if more than one install record found, we need to check conflict.
	case len(hosts) > 1:
		return h.handleMultipleHosts(hosts, innerIP, bizID)

	// if only one install record found, we need to check conflict is running or not.
	default:
		return h.handleSingleHost(hosts[0], innerIP)
	}
}

func (h *handler) handleSingleHost(host *types.Host, innerIP string) *types.NodeAgentInstallCheckResult {
	result := &types.NodeAgentInstallCheckResult{
		InnerIP:             innerIP,
		InstallEligibilitiy: types.NodeAgentInstallEligibilityNormal,
	}

	if host.Dynamic.NodeStatus != types.NodeStatusRunning {
		return result
	}

	switch host.Dynamic.NodeRole {
	// node already exists proxy. we can't install agent.
	case types.NodeRoleProxy:
		result.InstallEligibilitiy = types.NodeAgentInstallEligibilityExistProxy

	// node already exists agent. we can't reinstall agent.
	case types.NodeRoleAgent:
		result.InstallEligibilitiy = types.NodeAgentInstallEligibilityExistAgent

	// node is blank, we can install agent normally.
	default:
		result.InstallEligibilitiy = types.NodeAgentInstallEligibilityNormal
	}

	return result
}

func (h *handler) handleMultipleHosts(hosts []*types.Host, innerIP string, bizID int64) *types.NodeAgentInstallCheckResult {
	if duplicateHostIDs := getDynamicDuplicateIPHostIDs(hosts); len(duplicateHostIDs) > 0 {
		return &types.NodeAgentInstallCheckResult{
			InnerIP:             innerIP,
			InstallEligibilitiy: types.NodeAgentInstallEligibilityDuplicateIP,
			DuplicateHostIDs:    duplicateHostIDs,
		}
	}

	if conflictHostIDs := getIPConflictHostIDsByBizID(hosts, bizID); len(conflictHostIDs) > 0 {
		return &types.NodeAgentInstallCheckResult{
			InnerIP:             innerIP,
			InstallEligibilitiy: types.NodeAgentInstallEligibilityConflictIP,
			DuplicateHostIDs:    conflictHostIDs,
		}
	}

	return &types.NodeAgentInstallCheckResult{
		InnerIP:             innerIP,
		InstallEligibilitiy: types.NodeAgentInstallEligibilityNormal,
	}
}

func getDynamicDuplicateIPHostIDs(hosts []*types.Host) []int64 {
	hostIDs := make([]int64, 0)
	for _, host := range hosts {
		if host.Static.Addressing != types.AddressingDynamic {
			continue
		}
		if host.Dynamic.NodeStatus == types.NodeStatusRunning {
			continue
		}

		hostIDs = append(hostIDs, host.HostID)
	}

	return hostIDs
}

func getIPConflictHostIDsByBizID(hosts []*types.Host, bizID int64) []int64 {
	hostIDs := make([]int64, 0)
	for _, host := range hosts {
		if host.Static.BizID != bizID {
			continue
		}
		if host.Dynamic.NodeStatus == types.NodeStatusRunning {
			continue
		}

		hostIDs = append(hostIDs, host.HostID)
	}

	return hostIDs
}
