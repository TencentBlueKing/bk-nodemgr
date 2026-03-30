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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AgentInstallCheck checks if an agent can be installed on hosts.
func (h *handler) AgentInstallCheck(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeAgentInstallCheckReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check install agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	bizIDMap := make(map[int64]struct{})
	for _, host := range req.GetHost() {
		bizIDMap[host.GetBkBizId()] = struct{}{}
	}
	bizIDs := make([]int64, 0, len(bizIDMap))
	for bizID := range bizIDMap {
		bizIDs = append(bizIDs, bizID)
	}
	resources := buildBizResources(bizIDs)
	if authErr := h.authorizer.Check(rCtx, auth.ActionAgentOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to check install agent, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	results, err := h.checkInstall(rCtx, req.GetHost())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check install agent")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentInstallCheckResp)
	resp.ConvertResultFromTypes(results)

	return resp.GetData(), nil
}

// nolint: funlen,gocognit,gocyclo,cyclop
func (h *handler) checkInstall(nCtx contextx.IContext, reqHosts []*protoBackend.NodeAgentInstallCheckReq_Host) (
	[]*types.NodeAgentInstallCheckResult, error) {

	// generates the basic host filters.
	ipMap := make(map[string]struct{})
	ipv6Map := make(map[string]struct{})
	networkUnitIDMap := make(map[int64]struct{})
	hostIDMap := make(map[int64]struct{})
	for _, host := range reqHosts {
		for _, ip := range host.GetBkHostInneripList() {
			ipMap[ip] = struct{}{}
		}

		for _, ipv6 := range host.GetBkHostInneripV6List() {
			ipv6Map[ipv6] = struct{}{}
		}

		if networkUnitID := host.GetBkNetworkunitId(); networkUnitID >= 0 {
			networkUnitIDMap[networkUnitID] = struct{}{}
		}

		if hostID := host.GetBkHostId(); hostID >= 0 {
			hostIDMap[hostID] = struct{}{}
		}
	}

	checker := newInstallChecker(h.storageHost, h.domainNodeInstall)
	if err := checker.init(nCtx,
		conv.MapKeyToSlice(networkUnitIDMap),
		conv.MapKeyToSlice(hostIDMap),
		conv.MapKeyToSlice(ipMap),
		conv.MapKeyToSlice(ipv6Map)); err != nil {
		return nil, fmt.Errorf("failed to init install checker: %w", err)
	}

	// check all request hosts.
	results := make([]*types.NodeAgentInstallCheckResult, len(reqHosts))
	for idx, host := range reqHosts {
		reqHostID := host.GetBkHostId()

		// check networkunit.
		matchedNetworkUnit, ok := checker.getNetworkUnit(host.GetBkNetworkunitId())
		if !ok {
			results[idx] = &types.NodeAgentInstallCheckResult{
				Status: types.NodeAgentInstallCheckStatusNetworkUnitNotFound,
			}

			continue
		}

		if !checker.validateNetworkUnit(host.GetBkNetworkunitId()) {
			results[idx] = &types.NodeAgentInstallCheckResult{
				Status: types.NodeAgentInstallCheckStatusNetworkUnitNotSupportInstall,
			}

			continue
		}

		// try to install a brand new host not in CMDB.
		if reqHostID < 0 {
			// check if inner ip duplicated.
			if matchedHost, exist := checker.hasInnerIP(matchedNetworkUnit.NetworkAreaID, host.GetBkHostInneripList()); exist {
				results[idx] = &types.NodeAgentInstallCheckResult{
					Status:  types.NodeAgentInstallCheckStatusDuplicatedInnerIP,
					Matched: types.ConvertHostToNodeAgentInstallCheckMatchedItem(matchedHost),
				}

				continue
			}

			// check if inner ipv6 duplicated.
			if matchedHost, exist := checker.hasInnerIPV6(matchedNetworkUnit.NetworkAreaID, host.GetBkHostInneripV6List()); exist {
				results[idx] = &types.NodeAgentInstallCheckResult{
					Status:  types.NodeAgentInstallCheckStatusDuplicatedInnerIPV6,
					Matched: types.ConvertHostToNodeAgentInstallCheckMatchedItem(matchedHost),
				}

				continue
			}

			// register to CMDB and install.
			results[idx] = &types.NodeAgentInstallCheckResult{
				Status: types.NodeAgentInstallCheckStatusRegisterToCMDBAndInstall,
			}

			continue
		}

		// install an existed host in CMDB.
		matchedHost, ok := checker.getHost(reqHostID)
		if !ok {
			results[idx] = &types.NodeAgentInstallCheckResult{
				Status: types.NodeAgentInstallCheckStatusHostNotFound,
			}

			continue
		}

		// check node-role.
		if matchedHost.Dynamic.NodeRole == types.NodeRoleProxy {
			results[idx] = &types.NodeAgentInstallCheckResult{
				Status:  types.NodeAgentInstallCheckStatusInvalidNodeRole,
				Matched: types.ConvertHostToNodeAgentInstallCheckMatchedItem(matchedHost),
			}

			continue
		}

		// check biz-id.
		if host.GetBkBizId() != matchedHost.Static.BizID {
			results[idx] = &types.NodeAgentInstallCheckResult{
				Status:  types.NodeAgentInstallCheckStatusMismatchedBizID,
				Matched: types.ConvertHostToNodeAgentInstallCheckMatchedItem(matchedHost),
			}

			continue
		}

		// check networkarea-id.
		if matchedNetworkUnit.NetworkAreaID != matchedHost.Static.NetworkAreaID {
			results[idx] = &types.NodeAgentInstallCheckResult{
				Status:  types.NodeAgentInstallCheckStatusMismatchedNetworkAreaID,
				Matched: types.ConvertHostToNodeAgentInstallCheckMatchedItem(matchedHost),
			}

			continue
		}

		// check inner-ip.
		allFound := true
		for _, innerIP := range host.GetBkHostInneripList() {
			found := false
			for _, existIP := range matchedHost.Static.InnerIPList {
				if innerIP == existIP {
					found = true
					break
				}
			}
			allFound = allFound && found
		}
		if !allFound {
			results[idx] = &types.NodeAgentInstallCheckResult{
				Status:  types.NodeAgentInstallCheckStatusMismatchedInnerIP,
				Matched: types.ConvertHostToNodeAgentInstallCheckMatchedItem(matchedHost),
			}

			continue
		}

		// check inner-ipv6
		allFound = true
		for _, innerIPV6 := range host.GetBkHostInneripV6List() {
			found := false
			for _, existIP := range matchedHost.Static.InnerIPV6List {
				if innerIPV6 == existIP {
					found = true
					break
				}
			}
			allFound = allFound && found
		}
		if !allFound {
			results[idx] = &types.NodeAgentInstallCheckResult{
				Status:  types.NodeAgentInstallCheckStatusMismatchedInnerIPV6,
				Matched: types.ConvertHostToNodeAgentInstallCheckMatchedItem(matchedHost),
			}

			continue
		}

		// normal install.
		results[idx] = &types.NodeAgentInstallCheckResult{
			Status:  types.NodeAgentInstallCheckStatusNormalInstall,
			Matched: types.ConvertHostToNodeAgentInstallCheckMatchedItem(matchedHost),
		}
	}

	return results, nil
}

func newInstallChecker(storageHost topoStg.IStorageHost, domainNodeInstall topoStg.IDomainNodeInstall) *installChecker {
	return &installChecker{
		storageHost:       storageHost,
		domainNodeInstall: domainNodeInstall,
	}
}

type installChecker struct {
	networkUnitMap         map[int64]*types.NetworkUnit
	hostMap                map[int64]*types.Host
	networkUnitValidateMap map[int64]bool

	storageHost       topoStg.IStorageHost
	domainNodeInstall topoStg.IDomainNodeInstall
}

func (ic *installChecker) hasInnerIP(networkAreaID int64, innerIPList []string) (*types.Host, bool) {
	for _, host := range ic.hostMap {
		if host.Static.NetworkAreaID != networkAreaID {
			continue
		}

		for _, existIP := range host.Static.InnerIPList {
			for _, givenIP := range innerIPList {
				if existIP == givenIP {
					return host, true
				}
			}
		}
	}

	return nil, false
}

func (ic *installChecker) hasInnerIPV6(networkAreaID int64, innerIPV6 []string) (*types.Host, bool) {
	for _, host := range ic.hostMap {
		if host.Static.NetworkAreaID != networkAreaID {
			continue
		}

		for _, existIP := range host.Static.InnerIPV6List {
			for _, givenIP := range innerIPV6 {
				if existIP == givenIP {
					return host, true
				}
			}
		}
	}

	return nil, false
}

func (ic *installChecker) getHost(hostID int64) (*types.Host, bool) {
	if _, ok := ic.hostMap[hostID]; ok {
		return ic.hostMap[hostID], true
	}

	return nil, false
}

func (ic *installChecker) getNetworkUnit(networkUnitID int64) (*types.NetworkUnit, bool) {
	if _, ok := ic.networkUnitMap[networkUnitID]; ok {
		return ic.networkUnitMap[networkUnitID], true
	}

	return nil, false
}

func (ic *installChecker) validateNetworkUnit(networkUnitID int64) bool {
	if _, ok := ic.networkUnitValidateMap[networkUnitID]; ok {
		return ic.networkUnitValidateMap[networkUnitID]
	}

	return false
}

func (ic *installChecker) init(
	nCtx contextx.IContext, networkUnitIDList []int64, hostIDList []int64, innerIPList []string, innerIPV6List []string) error {

	// fetch networkunits.
	networkUnitMap, err := ic.fetchNetworkUnits(nCtx, networkUnitIDList)
	if err != nil {
		return err
	}
	ic.networkUnitMap = networkUnitMap

	// fetch hosts.
	hostMap, err := ic.fetchHosts(nCtx, hostIDList, innerIPList, innerIPV6List)
	if err != nil {
		return err
	}
	ic.hostMap = hostMap

	// query if networkunit have valid proxy for installation.
	networkUnitValidateMap, err := ic.domainNodeInstall.ExistDedicatedInstallerProxyHost(nCtx, networkUnitIDList)
	if err != nil {
		return fmt.Errorf("failed to check dedicated installer proxy for networkunits: %w", err)
	}
	for _, networkUnit := range networkUnitMap {
		// direct networkunit do not need installer proxy.
		if networkUnit.IsDirect {
			networkUnitValidateMap[networkUnit.ID] = true
		}
	}
	ic.networkUnitValidateMap = networkUnitValidateMap

	return nil
}

func (ic *installChecker) fetchNetworkUnits(nCtx contextx.IContext, networkUnitIDList []int64) (map[int64]*types.NetworkUnit, error) {
	networkUnits := make(map[int64]*types.NetworkUnit)

	if len(networkUnitIDList) == 0 {
		return make(map[int64]*types.NetworkUnit), nil
	}

	// fetch networkunits.
	networkUnitsByID, err := ic.domainNodeInstall.GetNetworkUnitByIDs(nCtx, networkUnitIDList)
	if err != nil {
		return nil, fmt.Errorf("failed to list networkunits by networkunit-id: %w", err)
	}
	for _, networkUnit := range networkUnitsByID {
		networkUnits[networkUnit.ID] = networkUnit
	}

	return networkUnits, nil
}

func (ic *installChecker) fetchHosts(nCtx contextx.IContext, hostIDList []int64, innerIPList []string, innerIPV6List []string) (
	map[int64]*types.Host, error) {

	// TODO: reduce the storage query, combine all hosts list into one db operation.
	hosts := make(map[int64]*types.Host)

	// fetch hosts by host-id.
	if len(hostIDList) > 0 {
		hostsByID, _, err := ic.storageHost.ListHost(nCtx, types.UnlimitedPage(), &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				HostID: hostIDList,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list hosts by host-id: %w", err)
		}
		for _, host := range hostsByID {
			hosts[host.HostID] = host
		}
	}

	// fetch hosts by innerip.
	if len(innerIPList) > 0 {
		hostsByInnerIP, _, err := ic.storageHost.ListHost(nCtx, types.UnlimitedPage(), &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				InnerIP: innerIPList,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list hosts by innerip: %w", err)
		}
		for _, host := range hostsByInnerIP {
			hosts[host.HostID] = host
		}
	}

	// fetch hosts by innerip_v6.
	if len(innerIPV6List) > 0 {
		hostsByInnerIPV6, _, err := ic.storageHost.ListHost(nCtx, types.UnlimitedPage(), &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				InnerIPV6: innerIPV6List,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list hosts by innerip v6: %w", err)
		}
		for _, host := range hostsByInnerIPV6 {
			hosts[host.HostID] = host
		}
	}

	return hosts, nil
}
