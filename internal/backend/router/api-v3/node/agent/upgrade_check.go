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

// AgentUpgradeCheck checks if agents can be upgraded on hosts.
func (h *handler) AgentUpgradeCheck(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeAgentUpgradeCheckReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check upgrade agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hostIDs := make([]int64, 0, len(req.GetHost()))
	for _, host := range req.GetHost() {
		hostIDs = append(hostIDs, host.GetBkHostId())
	}
	hosts, _, err := h.storageHost.ListHost(rCtx, types.UnlimitedPage(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{HostID: hostIDs},
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check upgrade agent, failed to list hosts")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	bizIDMap := make(map[int64]struct{})
	for _, host := range hosts {
		bizIDMap[host.Static.BizID] = struct{}{}
	}
	resources := buildBizResources(conv.MapKeyToSlice(bizIDMap))
	if authErr := h.authorizer.Check(rCtx, auth.ActionAgentOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to check upgrade agent, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	results, err := h.checkUpgrade(rCtx, req.GetHost(), req.GetTargetVersion())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check upgrade agent")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentUpgradeCheckResp)
	resp.ConvertResultFromTypes(results)

	return resp.GetData(), nil
}

// nolint: cyclop,gocognit,funlen
func (h *handler) checkUpgrade(
	nCtx contextx.IContext,
	reqHosts []*protoBackend.NodeAgentUpgradeCheckReq_Host,
	targetVersions []*protoBackend.TargetVersion,
) ([]*types.NodeAgentUpgradeCheckResult, error) {

	hostIDMap := make(map[int64]struct{})
	networkUnitIDMap := make(map[int64]struct{})
	for _, host := range reqHosts {
		if hostID := host.GetBkHostId(); hostID >= 0 {
			hostIDMap[hostID] = struct{}{}
		}
		if networkUnitID := host.GetBkNetworkunitId(); networkUnitID >= 0 {
			networkUnitIDMap[networkUnitID] = struct{}{}
		}
	}

	checker := newUpgradeChecker(h.storageHost, h.storageNetworkUnit)
	if err := checker.init(nCtx, conv.MapKeyToSlice(hostIDMap), conv.MapKeyToSlice(networkUnitIDMap)); err != nil {
		return nil, fmt.Errorf("failed to init upgrade checker: %w", err)
	}

	versionMap := buildVersionMap(targetVersions)

	results := make([]*types.NodeAgentUpgradeCheckResult, len(reqHosts))
	for idx, reqHost := range reqHosts {
		hostID := reqHost.GetBkHostId()
		networkUnitID := reqHost.GetBkNetworkunitId()
		cpuArch := reqHost.GetCpuArch()

		matchedHost, ok := checker.getHost(hostID)
		if !ok {
			results[idx] = &types.NodeAgentUpgradeCheckResult{
				Status: types.NodeAgentUpgradeCheckStatusHostNotFound,
			}

			continue
		}

		matched := &types.NodeAgentUpgradeCheckMatchedItem{
			HostID:        matchedHost.HostID,
			BizID:         matchedHost.Static.BizID,
			NetworkAreaID: matchedHost.Static.NetworkAreaID,
			NetworkUnitID: matchedHost.Dynamic.NetworkUnitID,
			OsType:        matchedHost.Dynamic.NodeOsType,
			NodeRole:      matchedHost.Dynamic.NodeRole,
			InnerIPList:   matchedHost.Static.InnerIPList,
			InnerIPV6List: matchedHost.Static.InnerIPV6List,
		}

		// Check if node status allows upgrade.
		if !isNodeStatusAllowUpgrade(matchedHost.Dynamic.NodeStatus) {
			results[idx] = &types.NodeAgentUpgradeCheckResult{
				Status:  types.NodeAgentUpgradeCheckStatusNodeStatusNotAllowed,
				Matched: matched,
			}

			continue
		}

		// Check if cpu_arch is present.
		if cpuArch == "" {
			results[idx] = &types.NodeAgentUpgradeCheckResult{
				Status:  types.NodeAgentUpgradeCheckStatusCPUArchMissing,
				Matched: matched,
			}

			continue
		}

		targetNetworkUnitID := resolveUpgradeNetworkUnitID(matchedHost.Dynamic.NetworkUnitID, networkUnitID)
		if targetNetworkUnitID < 0 {
			results[idx] = &types.NodeAgentUpgradeCheckResult{
				Status:  types.NodeAgentUpgradeCheckStatusNetworkUnitNotFound,
				Matched: matched,
			}

			continue
		}

		// Verify target networkunit.
		if networkUnitID >= 0 {
			targetUnit, unitFound := checker.getNetworkUnit(targetNetworkUnitID)
			if !unitFound {
				results[idx] = &types.NodeAgentUpgradeCheckResult{
					Status:  types.NodeAgentUpgradeCheckStatusNetworkUnitNotFound,
					Matched: matched,
				}

				continue
			}

			// networkunit must belong to the same networkarea as the host.
			if targetUnit.NetworkAreaID != matchedHost.Static.NetworkAreaID {
				results[idx] = &types.NodeAgentUpgradeCheckResult{
					Status:  types.NodeAgentUpgradeCheckStatusNetworkUnitMismatch,
					Matched: matched,
				}

				continue
			}
		}

		// Verify version matching by os_type + cpu_arch.
		key := fmt.Sprintf("%s:%s", matchedHost.Dynamic.NodeOsType, cpuArch)
		if _, versionFound := versionMap[key]; !versionFound && len(versionMap) > 0 {
			results[idx] = &types.NodeAgentUpgradeCheckResult{
				Status:  types.NodeAgentUpgradeCheckStatusVersionNotFound,
				Matched: matched,
			}

			continue
		}

		// networkunit change requires confirmation.
		if targetNetworkUnitID != matchedHost.Dynamic.NetworkUnitID {
			results[idx] = &types.NodeAgentUpgradeCheckResult{
				Status:  types.NodeAgentUpgradeCheckStatusNetworkUnitChanged,
				Matched: matched,
			}

			continue
		}

		results[idx] = &types.NodeAgentUpgradeCheckResult{
			Status:  types.NodeAgentUpgradeCheckStatusNormalUpgrade,
			Matched: matched,
		}
	}

	return results, nil
}

// isNodeStatusAllowUpgrade returns true if the node status allows upgrade.
func isNodeStatusAllowUpgrade(status types.NodeStatus) bool {
	switch status {
	case types.NodeStatusBusy, types.NodeStatusUpgrade, types.NodeStatusStopping:
		return false
	default:
		return true
	}
}

// buildVersionMap builds a lookup map from os_type:cpu_arch → version.
func buildVersionMap(targetVersions []*protoBackend.TargetVersion) map[string]string {
	m := make(map[string]string, len(targetVersions))
	for _, v := range targetVersions {
		key := fmt.Sprintf("%s:%s", v.GetOsType(), v.GetCpuArch())
		m[key] = v.GetVersion()
	}

	return m
}

func newUpgradeChecker(storageHost topoStg.IStorageHost, storageNetworkUnit topoStg.IStorageNetworkUnit) *upgradeChecker {
	return &upgradeChecker{
		storageHost:        storageHost,
		storageNetworkUnit: storageNetworkUnit,
	}
}

type upgradeChecker struct {
	hostMap        map[int64]*types.Host
	networkUnitMap map[int64]*types.NetworkUnit

	storageHost        topoStg.IStorageHost
	storageNetworkUnit topoStg.IStorageNetworkUnit
}

func (uc *upgradeChecker) getHost(hostID int64) (*types.Host, bool) {
	host, ok := uc.hostMap[hostID]
	return host, ok
}

func (uc *upgradeChecker) getNetworkUnit(networkUnitID int64) (*types.NetworkUnit, bool) {
	unit, ok := uc.networkUnitMap[networkUnitID]
	return unit, ok
}

func (uc *upgradeChecker) init(nCtx contextx.IContext, hostIDs []int64, networkUnitIDs []int64) error {
	if len(hostIDs) > 0 {
		hosts, _, err := uc.storageHost.ListHost(nCtx, types.UnlimitedPage(), &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{HostID: hostIDs},
		})
		if err != nil {
			return fmt.Errorf("failed to list hosts: %w", err)
		}

		uc.hostMap = make(map[int64]*types.Host, len(hosts))
		for _, host := range hosts {
			uc.hostMap[host.HostID] = host
		}
	} else {
		uc.hostMap = make(map[int64]*types.Host)
	}

	if len(networkUnitIDs) > 0 {
		units, _, err := uc.storageNetworkUnit.ListNetworkUnit(nCtx, types.UnlimitedPage(), &types.NetworkUnitCondition{
			ExactInclude: &types.NetworkUnitExactFields{NetworkUnitID: networkUnitIDs},
		})
		if err != nil {
			return fmt.Errorf("failed to list network units: %w", err)
		}

		uc.networkUnitMap = make(map[int64]*types.NetworkUnit, len(units))
		for _, unit := range units {
			uc.networkUnitMap[unit.ID] = unit
		}
	} else {
		uc.networkUnitMap = make(map[int64]*types.NetworkUnit)
	}

	return nil
}
