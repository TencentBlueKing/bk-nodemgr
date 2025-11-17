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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// checkAgentInstallMaxPageSize defines the max page size for page executor.
	// In the scenario of 40,000 hosts, a single request for 1,000 hosts requires 400 table lookups.
	checkAgentInstallMaxPageSize = 1000

	// defaultCheckConcurrency defines the default check concurrency.
	defaultCheckConcurrency = 100
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

	ips := make([]string, len(hosts))
	unitIDs := make([]int64, len(hosts))
	for i, host := range hosts {
		unitIDs[i] = host.GetBkNetworkunitId()
		ips[i] = host.GetBkHostInnerip()
	}

	// fetch network-units.
	unitsIDMap, err := h.getNetworkUnitByIDs(nCtx, unitIDs) // unitID -> networkUnit
	if err != nil {
		return nil, fmt.Errorf("failed to get network-unit by ids: %w", err)
	}

	// fetch host.
	ipHostMap, err := h.getHostByIPs(nCtx, ips) // ip -> host
	if err != nil {
		return nil, fmt.Errorf("failed to get host by ips: %w", err)
	}

	// fetch pagent install eligibility.
	unitPagentEligMap, err := h.getPagentInstallEligs(nCtx, unitsIDMap) // unitID -> canInstallPagent
	if err != nil {
		return nil, fmt.Errorf("failed to get pagent install eligibility: %w", err)
	}

	gp := gopool.NewPool()
	gp.SetLimit(defaultCheckConcurrency)
	results := make([]*types.NodeAgentInstallCheckResult, len(hosts))

	for i := range hosts {
		idx := i
		gp.Go(func() error {
			result, err := h.processHost(hosts[idx], ipHostMap, unitsIDMap, unitPagentEligMap)
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
func (h *handler) processHost(host *protoBackend.NodeAgentInstallCheckReq_Host,
	ipHostMap map[string][]*types.Host,
	unitsIDMap map[int64]*types.NetworkUnit,
	unitPagentEligMap map[int64]bool) (*types.NodeAgentInstallCheckResult, error) {

	networkUnitID := host.GetBkNetworkunitId()
	innerIP := host.GetBkHostInnerip()

	networkUnit, exists := unitsIDMap[networkUnitID]
	if !exists {
		return nil, fmt.Errorf("failed to get networkunit by id, networkunit-id(%d)", networkUnitID)
	}

	canInstallPagent := unitPagentEligMap[networkUnitID]
	if !canInstallPagent {
		return &types.NodeAgentInstallCheckResult{
			InnerIP:     innerIP,
			InstallElig: types.NodeAgentInstallEligNotExistRelay,
		}, nil
	}

	allHostsWithSameIP := ipHostMap[innerIP]
	sameAreaHosts := make([]*types.Host, 0)
	for _, h := range allHostsWithSameIP {
		if h.Static.NetworkAreaID != networkUnit.NetworkAreaID {
			continue
		}
		sameAreaHosts = append(sameAreaHosts, h)
	}

	return h.checkAgentInstallElig(host.GetBkHostId(), host.GetBkBizId(), innerIP, sameAreaHosts)
}

func (h *handler) checkAgentInstallElig(hostID,
	bizID int64, innerIP string, needCheckHosts []*types.Host) (*types.NodeAgentInstallCheckResult, error) {

	result := &types.NodeAgentInstallCheckResult{
		InnerIP:     innerIP,
		InstallElig: types.NodeAgentInstallEligNormalInstall,
	}

	switch {
	// if no install record found, we mark it as "import cmdb and normal install".
	case len(needCheckHosts) == 0:
		result.InstallElig = types.NodeAgentInstallEligImportCmdbAndNormalInstall

	// if only one install record found, and it's a proxy, we can't install agent on proxy.
	case len(needCheckHosts) == 1 && needCheckHosts[0].Dynamic.NodeRole == types.NodeRoleProxy:
		result.InstallElig = types.NodeAgentInstallEligExistProxy

	// if only one install record found, and it's the same host id, it means it's reinstall.
	case len(needCheckHosts) == 1 && needCheckHosts[0].HostID == hostID:
		result.InstallElig = types.NodeAgentInstallEligNormalInstall

	// if one or more install records found, we need to check host.
	default:
		result = handleMultipleHosts(needCheckHosts, innerIP, bizID)
	}

	return result, nil
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

func (h *handler) getPagentInstallEligs(nCtx contextx.IContext,
	unitsIDMap map[int64]*types.NetworkUnit) (map[int64]bool, error) {

	result := make(map[int64]bool)

	for unitID, unit := range unitsIDMap {
		if unit.IsDirect {
			result[unitID] = true
			continue
		}

		inDirectUnits, err := h.domainNodeInstall.ExistDedicatedInstallerProxyHost(nCtx, unitID)
		if err != nil {
			return nil, fmt.Errorf("failed to check dedicated installer proxy host: %w", err)
		}
		result[unitID] = inDirectUnits
	}

	return result, nil
}

func (h *handler) getHostByIPs(nCtx contextx.IContext,
	innerIPs []string) (map[string][]*types.Host, error) {

	executor := pageexecutor.NewPageExecutor[*types.Host](checkAgentInstallMaxPageSize, 1*time.Hour)

	fn := func(_ context.Context, p types.Page) ([]*types.Host, error) {
		cond := &types.HostCondition{
			ExactInclude: &types.HostExactFields{
				InnerIP: innerIPs,
			},
		}
		host, _, err := h.storageHost.ListHost(nCtx, p, cond)

		return host, err
	}

	pageResult, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return nil, fmt.Errorf("failed to get host by inner: %w", err)
	}

	// ip -> host
	ipHostMap := make(map[string][]*types.Host)
	for _, host := range pageResult.Items {
		if len(host.Static.InnerIPList) == 0 {
			return nil, fmt.Errorf("failed to get host byinner ip: %w", err)
		}

		ip := host.Static.InnerIPList[0]
		ipHostMap[ip] = append(ipHostMap[ip], host)
	}

	return ipHostMap, nil
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
