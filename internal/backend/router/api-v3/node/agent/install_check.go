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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AgentInstallCheck checks if an agent can be installed on hosts.
func (h *handler) AgentInstallCheck(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeAgentInstallCheckReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check install agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	results, err := h.checkInstall(rCtx, req.GetHost())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check install agent")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentInstallCheckResp)
	resp.ConvertResultFromTypes(results, len(results))

	return resp.GetData(), nil
}

func (h *handler) checkInstall(nCtx contextx.IContext,
	hosts []*protoBackend.NodeAgentInstallCheckReq_Host) ([]*types.NodeAgentInstallCheckResult, error) {

	if len(hosts) == 0 {
		return []*types.NodeAgentInstallCheckResult{}, nil
	}

	unitIDs := make([]int64, len(hosts))
	for i, host := range hosts {
		unitIDs[i] = host.GetBkNetworkunitId()
	}

	unitsMap, err := h.getNetworkUnitByIDs(nCtx, unitIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get network-unit by ids: %w", err)
	}

	gp := gopool.NewPool()
	results := make([]*types.NodeAgentInstallCheckResult, len(hosts))

	for i := range hosts {
		idx := i
		gp.Go(func() error {
			result, err := h.processHost(nCtx, hosts[idx], unitsMap)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).With("host", hosts[idx]).Error("failed to process host install check")
				return fmt.Errorf("failed to process host install check: %w", err)
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

// processHost process a single host install check.
func (h *handler) processHost(nCtx contextx.IContext,
	host *protoBackend.NodeAgentInstallCheckReq_Host, unitsMap map[int64]*types.NetworkUnit) (*types.NodeAgentInstallCheckResult, error) {

	networkunitID := host.GetBkNetworkunitId()
	networkUnit, exists := unitsMap[networkunitID]
	if !exists {
		return nil, fmt.Errorf("failed to get networkunit by id, networkunit-id(%d)", networkunitID)
	}

	canInstallPagent, err := h.checkPagentInstallElig(nCtx, networkUnit)
	if err != nil {
		return nil, fmt.Errorf("failed to check pagent eligibility: %w", err)
	}

	if !canInstallPagent {
		return &types.NodeAgentInstallCheckResult{
			InnerIP:     host.GetBkHostInnerip(),
			InstallElig: types.NodeAgentInstallEligNotExistRelay,
		}, nil
	}

	return h.checkAgentInstallElig(nCtx, host.GetBkHostInnerip(), networkUnit.NetworkAreaID, host.GetBkBizId())
}

func (h *handler) checkAgentInstallElig(nCtx contextx.IContext, innerIP string,
	networkAreaID, bizID int64) (*types.NodeAgentInstallCheckResult, error) {

	// find same inner ip and area hosts.
	hosts, err := h.domainNodeInstall.GetHostsByAreaAndInnerIP(nCtx, networkAreaID, innerIP)
	if err != nil {
		return nil, err
	}

	result := &types.NodeAgentInstallCheckResult{
		InnerIP:     innerIP,
		InstallElig: types.NodeAgentInstallEligNormalInstall,
	}

	switch {
	// if no install record found, we mark it as "import cmdb and normal install".
	case len(hosts) == 0:
		result.InstallElig = types.NodeAgentInstallEligImportCmdbAndNormalInstall

	// if only one install record found, and it's a proxy, we can't install agent on proxy.
	case len(hosts) == 1 && hosts[0].Dynamic.NodeRole == types.NodeRoleProxy:
		result.InstallElig = types.NodeAgentInstallEligExistProxy

	// if more than one install record found, we need to check conflict.
	default:
		result = handleMultipleHosts(hosts, innerIP, bizID)
	}

	return result, nil
}

func (h *handler) checkPagentInstallElig(nCtx contextx.IContext, networkUnit *types.NetworkUnit) (bool, error) {
	if networkUnit.IsDirect {
		return true, nil
	}

	// check can be installed proxy host
	num, err := h.domainNodeInstall.CountDedicatedInstallerProxyHost(nCtx, networkUnit.ID)
	if err != nil {
		return false, fmt.Errorf("failed to count dedicated installer proxy host: %w", err)
	}

	return num > 0, nil
}

func handleMultipleHosts(hosts []*types.Host, innerIP string, bizID int64) *types.NodeAgentInstallCheckResult {
	if duplicateHostIDs := getDynamicDuplicateIPHostIDs(hosts); len(duplicateHostIDs) > 0 {
		return &types.NodeAgentInstallCheckResult{
			InnerIP:        innerIP,
			InstallElig:    types.NodeAgentInstallEligDuplicateIP,
			PendingHostIDs: duplicateHostIDs,
		}
	}

	if conflictHostIDs := getIPConflictHostIDsByBizID(hosts, bizID); len(conflictHostIDs) > 0 {
		return &types.NodeAgentInstallCheckResult{
			InnerIP:        innerIP,
			InstallElig:    types.NodeAgentInstallEligConflictIP,
			PendingHostIDs: conflictHostIDs,
		}
	}

	return &types.NodeAgentInstallCheckResult{
		InnerIP:     innerIP,
		InstallElig: types.NodeAgentInstallEligNormalInstall,
	}
}

func (h *handler) getNetworkUnitByIDs(nCtx contextx.IContext, unitIDs []int64) (
	map[int64]*types.NetworkUnit, error) {

	units, err := h.domainNodeInstall.GetNetworkUnitByIDs(nCtx, unitIDs)
	if err != nil {
		return nil, err
	}

	networkUnitsMap := make(map[int64]*types.NetworkUnit, len(units))
	for _, unit := range units {
		networkUnitsMap[unit.ID] = unit
	}

	return networkUnitsMap, nil
}

func getDynamicDuplicateIPHostIDs(hosts []*types.Host) []int64 {
	hostIDs := make([]int64, 0)
	for _, host := range hosts {
		if host.Static.Addressing != types.AddressingDynamic {
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
		hostIDs = append(hostIDs, host.HostID)
	}

	return hostIDs
}
