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

package proxy

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ProxyUpgradeCheck checks if proxies can be upgraded on hosts.
func (h *handler) ProxyUpgradeCheck(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeProxyUpgradeCheckReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check upgrade proxy, failed to decode request body")
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
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check upgrade proxy, failed to list hosts")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	bizIDMap := make(map[int64]struct{})
	for _, host := range hosts {
		bizIDMap[host.Static.BizID] = struct{}{}
	}
	resources := authRouter.BuildBizResources(conv.MapKeyToSlice(bizIDMap)...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionProxyOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to check upgrade proxy, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	results, err := h.checkProxyUpgrade(rCtx, req.GetHost(), req.GetTargetVersion())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check upgrade proxy")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeProxyUpgradeCheckResp)
	resp.ConvertResultFromTypes(results)

	return resp.GetData(), nil
}

// nolint: cyclop,gocognit,funlen
func (h *handler) checkProxyUpgrade(
	nCtx contextx.IContext,
	reqHosts []*protoBackend.NodeProxyUpgradeCheckReq_Host,
	targetVersions []*protoBackend.TargetVersion,
) ([]*types.NodeProxyUpgradeCheckResult, error) {

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

	checker := newProxyUpgradeChecker(h.storageHost, h.storageNetworkUnit)
	if err := checker.init(nCtx, conv.MapKeyToSlice(hostIDMap), conv.MapKeyToSlice(networkUnitIDMap)); err != nil {
		return nil, fmt.Errorf("failed to init proxy upgrade checker: %w", err)
	}

	versionMap := buildProxyVersionMap(targetVersions)

	results := make([]*types.NodeProxyUpgradeCheckResult, len(reqHosts))
	for idx, reqHost := range reqHosts {
		hostID := reqHost.GetBkHostId()
		networkUnitID := reqHost.GetBkNetworkunitId()
		cpuArch := reqHost.GetCpuArch()

		matchedHost, ok := checker.getHost(hostID)
		if !ok {
			results[idx] = &types.NodeProxyUpgradeCheckResult{
				Status: types.NodeProxyUpgradeCheckStatusHostNotFound,
			}

			continue
		}

		matched := &types.NodeProxyUpgradeCheckMatchedItem{
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
		if !isProxyNodeStatusAllowUpgrade(matchedHost.Dynamic.NodeStatus) {
			results[idx] = &types.NodeProxyUpgradeCheckResult{
				Status:  types.NodeProxyUpgradeCheckStatusNodeStatusNotAllowed,
				Matched: matched,
			}

			continue
		}

		// Check if cpu_arch is present.
		if cpuArch == "" {
			results[idx] = &types.NodeProxyUpgradeCheckResult{
				Status:  types.NodeProxyUpgradeCheckStatusCPUArchMissing,
				Matched: matched,
			}

			continue
		}

		targetNetworkUnitID := resolveUpgradeNetworkUnitID(matchedHost.Dynamic.NetworkUnitID, networkUnitID)
		if targetNetworkUnitID < 0 {
			results[idx] = &types.NodeProxyUpgradeCheckResult{
				Status:  types.NodeProxyUpgradeCheckStatusNetworkUnitNotFound,
				Matched: matched,
			}

			continue
		}

		// Verify target networkunit.
		if networkUnitID >= 0 {
			targetUnit, unitFound := checker.getNetworkUnit(targetNetworkUnitID)
			if !unitFound {
				results[idx] = &types.NodeProxyUpgradeCheckResult{
					Status:  types.NodeProxyUpgradeCheckStatusNetworkUnitNotFound,
					Matched: matched,
				}

				continue
			}

			if targetUnit.NetworkAreaID != matchedHost.Static.NetworkAreaID {
				results[idx] = &types.NodeProxyUpgradeCheckResult{
					Status:  types.NodeProxyUpgradeCheckStatusNetworkUnitMismatch,
					Matched: matched,
				}

				continue
			}
		}

		// Verify version matching by os_type + cpu_arch.
		key := fmt.Sprintf("%s:%s", matchedHost.Dynamic.NodeOsType, cpuArch)
		if _, versionFound := versionMap[key]; !versionFound && len(versionMap) > 0 {
			results[idx] = &types.NodeProxyUpgradeCheckResult{
				Status:  types.NodeProxyUpgradeCheckStatusVersionNotFound,
				Matched: matched,
			}

			continue
		}

		// networkunit change requires confirmation.
		if targetNetworkUnitID != matchedHost.Dynamic.NetworkUnitID {
			results[idx] = &types.NodeProxyUpgradeCheckResult{
				Status:  types.NodeProxyUpgradeCheckStatusNetworkUnitChanged,
				Matched: matched,
			}

			continue
		}

		results[idx] = &types.NodeProxyUpgradeCheckResult{
			Status:  types.NodeProxyUpgradeCheckStatusNormalUpgrade,
			Matched: matched,
		}
	}

	return results, nil
}

func isProxyNodeStatusAllowUpgrade(status types.NodeStatus) bool {
	switch status {
	case types.NodeStatusBusy, types.NodeStatusUpgrade, types.NodeStatusStopping:
		return false
	default:
		return true
	}
}

func buildProxyVersionMap(targetVersions []*protoBackend.TargetVersion) map[string]string {
	m := make(map[string]string, len(targetVersions))
	for _, v := range targetVersions {
		key := fmt.Sprintf("%s:%s", v.GetOsType(), v.GetCpuArch())
		m[key] = v.GetVersion()
	}

	return m
}

func newProxyUpgradeChecker(storageHost topoStg.IStorageHost, storageNetworkUnit topoStg.IStorageNetworkUnit) *proxyUpgradeChecker {
	return &proxyUpgradeChecker{
		storageHost:        storageHost,
		storageNetworkUnit: storageNetworkUnit,
	}
}

type proxyUpgradeChecker struct {
	hostMap        map[int64]*types.Host
	networkUnitMap map[int64]*types.NetworkUnit

	storageHost        topoStg.IStorageHost
	storageNetworkUnit topoStg.IStorageNetworkUnit
}

func (uc *proxyUpgradeChecker) getHost(hostID int64) (*types.Host, bool) {
	host, ok := uc.hostMap[hostID]
	return host, ok
}

func (uc *proxyUpgradeChecker) getNetworkUnit(networkUnitID int64) (*types.NetworkUnit, bool) {
	unit, ok := uc.networkUnitMap[networkUnitID]
	return unit, ok
}

func (uc *proxyUpgradeChecker) init(nCtx contextx.IContext, hostIDs []int64, networkUnitIDs []int64) error {
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
