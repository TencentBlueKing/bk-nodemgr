/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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

// ProxyInstallCheck checks if a proxy can be installed on hosts.
func (h *handler) ProxyInstallCheck(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeProxyInstallCheckReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check install proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	bizIDs := conv.SliceUnique(
		conv.SliceToSlice[*types.Host, int64](req.ConvertHostsToTypes(), func(host *types.Host) int64 {
			return host.Static.BizID
		}))
	resources := authRouter.BuildBizResources(bizIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionProxyOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to check install proxy, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	results, err := h.checkProxyInstall(rCtx, req.GetHost())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check install proxy")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeProxyInstallCheckResp)
	resp.ConvertResultFromTypes(results)

	return resp.GetData(), nil
}

// nolint: funlen,gocognit,gocyclo,cyclop,maintidx,nestif
// NOCC: golint/fnsize(proxy install precheck is one validation flow).
func (h *handler) checkProxyInstall(nCtx contextx.IContext, reqHosts []*protoBackend.NodeProxyInstallCheckReq_Host) (
	[]*types.NodeProxyInstallCheckResult, error) {

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

	checker := newProxyInstallChecker(h.storageHost, h.storageNetworkUnit)
	if err := checker.init(nCtx,
		conv.MapKeyToSlice(networkUnitIDMap),
		conv.MapKeyToSlice(hostIDMap),
		conv.MapKeyToSlice(ipMap),
		conv.MapKeyToSlice(ipv6Map)); err != nil {
		return nil, fmt.Errorf("failed to init proxy install checker: %w", err)
	}

	results := make([]*types.NodeProxyInstallCheckResult, len(reqHosts))
	for idx, host := range reqHosts {
		reqHostID := host.GetBkHostId()
		matchedNetworkUnit, ok := checker.getNetworkUnit(host.GetBkNetworkunitId())
		if !ok {
			results[idx] = &types.NodeProxyInstallCheckResult{
				Status:    types.NodeProxyInstallCheckStatusNetworkUnitNotFound,
				MessageEn: "Networkunit does not exist",
				MessageZh: "所属管控单元不存在",
				Category:  types.InstallCheckCategoryError,
			}

			continue
		}

		if reqHostID < 0 {
			if matchedHost, exist := checker.hasInnerIP(matchedNetworkUnit.NetworkAreaID, host.GetBkHostInneripList()); exist {
				// verify all request IPs belong to the same matched host.
				if !allIPsBelongToHost(host.GetBkHostInneripList(), matchedHost.Static.InnerIPList) {
					results[idx] = &types.NodeProxyInstallCheckResult{
						Status:    types.NodeProxyInstallCheckStatusMismatchedInnerIP,
						Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
						MessageEn: "Inner IPV4 does not match CMDB configuration",
						MessageZh: "内网IPV4与CMDB配置不符",
						Category:  types.InstallCheckCategoryError,
					}

					continue
				}

				// check biz-id matches.
				if host.GetBkBizId() != matchedHost.Static.BizID {
					results[idx] = &types.NodeProxyInstallCheckResult{
						Status:    types.NodeProxyInstallCheckStatusMismatchedBizID,
						Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
						MessageEn: "Business ID does not match CMDB configuration",
						MessageZh: "所属业务与CMDB配置不符",
						Category:  types.InstallCheckCategoryError,
					}

					continue
				}

				results[idx] = &types.NodeProxyInstallCheckResult{
					Status:    types.NodeProxyInstallCheckStatusDuplicatedInnerIP,
					Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
					MessageEn: "Inner IPV4 already exists in networkarea, will reinstall this host as Proxy",
					MessageZh: "内网IPV4在该管控区域下已经存在, 将重装该主机为Proxy",
					Category:  types.InstallCheckCategoryNeedConfirm,
				}

				continue
			}
			if matchedHost, exist := checker.hasInnerIPV6(matchedNetworkUnit.NetworkAreaID, host.GetBkHostInneripV6List()); exist {
				// verify all request IPv6s belong to the same matched host.
				if !allIPsBelongToHost(host.GetBkHostInneripV6List(), matchedHost.Static.InnerIPV6List) {
					results[idx] = &types.NodeProxyInstallCheckResult{
						Status:    types.NodeProxyInstallCheckStatusMismatchedInnerIPV6,
						Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
						MessageEn: "Inner IPV6 does not match CMDB configuration",
						MessageZh: "内网IPV6与CMDB配置不符",
						Category:  types.InstallCheckCategoryError,
					}

					continue
				}

				// check biz-id matches.
				if host.GetBkBizId() != matchedHost.Static.BizID {
					results[idx] = &types.NodeProxyInstallCheckResult{
						Status:    types.NodeProxyInstallCheckStatusMismatchedBizID,
						Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
						MessageEn: "Business ID does not match CMDB configuration",
						MessageZh: "所属业务与CMDB配置不符",
						Category:  types.InstallCheckCategoryError,
					}

					continue
				}

				results[idx] = &types.NodeProxyInstallCheckResult{
					Status:    types.NodeProxyInstallCheckStatusDuplicatedInnerIPV6,
					Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
					MessageEn: "Inner IPV6 already exists in networkarea, will reinstall this host as Proxy",
					MessageZh: "内网IPV6在该管控区域下已经存在, 将重装该主机为Proxy",
					Category:  types.InstallCheckCategoryNeedConfirm,
				}

				continue
			}
			results[idx] = &types.NodeProxyInstallCheckResult{
				Status:    types.NodeProxyInstallCheckStatusRegisterToCMDBAndInstall,
				MessageEn: "Import node to CMDB and install Proxy",
				MessageZh: "将节点导入CMDB并安装Proxy",
				Category:  types.InstallCheckCategoryRegisterToCMDBAndInstall,
			}

			continue
		}

		matchedHost, ok := checker.getHost(reqHostID)
		if !ok {
			results[idx] = &types.NodeProxyInstallCheckResult{
				Status:    types.NodeProxyInstallCheckStatusHostNotFound,
				MessageEn: "Host does not exist",
				MessageZh: "该主机不存在",
				Category:  types.InstallCheckCategoryError,
			}

			continue
		}

		if matchedHost.Dynamic.NodeRole == types.NodeRoleAgent {
			results[idx] = &types.NodeProxyInstallCheckResult{
				Status:    types.NodeProxyInstallCheckStatusInvalidNodeRole,
				Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
				MessageEn: "Node role does not allow Proxy installation, please uninstall the Agent first",
				MessageZh: "节点角色不允许安装Proxy, 请先卸载Agent",
				Category:  types.InstallCheckCategoryError,
			}

			continue
		}

		if host.GetBkBizId() != matchedHost.Static.BizID {
			results[idx] = &types.NodeProxyInstallCheckResult{
				Status:    types.NodeProxyInstallCheckStatusMismatchedBizID,
				Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
				MessageEn: "Business ID does not match CMDB configuration",
				MessageZh: "所属业务与CMDB配置不符",
				Category:  types.InstallCheckCategoryError,
			}

			continue
		}

		if matchedNetworkUnit.NetworkAreaID != matchedHost.Static.NetworkAreaID {
			results[idx] = &types.NodeProxyInstallCheckResult{
				Status:    types.NodeProxyInstallCheckStatusMismatchedNetworkAreaID,
				Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
				MessageEn: "Networkarea ID does not match CMDB configuration",
				MessageZh: "所属管控区域与CMDB配置不符",
				Category:  types.InstallCheckCategoryError,
			}

			continue
		}

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
			results[idx] = &types.NodeProxyInstallCheckResult{
				Status:    types.NodeProxyInstallCheckStatusMismatchedInnerIP,
				Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
				MessageEn: "Inner IPV4 does not match CMDB configuration",
				MessageZh: "内网IPV4与CMDB配置不符",
				Category:  types.InstallCheckCategoryError,
			}

			continue
		}

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
			results[idx] = &types.NodeProxyInstallCheckResult{
				Status:    types.NodeProxyInstallCheckStatusMismatchedInnerIPV6,
				Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
				MessageEn: "Inner IPV6 does not match CMDB configuration",
				MessageZh: "内网IPV6与CMDB配置不符",
				Category:  types.InstallCheckCategoryError,
			}

			continue
		}

		if host.GetBkNetworkunitId() != matchedHost.Dynamic.NetworkUnitID {
			results[idx] = &types.NodeProxyInstallCheckResult{
				Status:    types.NodeProxyInstallCheckStatusNormalInstall,
				Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
				MessageEn: "Install Proxy into networkunit",
				MessageZh: "安装Proxy到新的管控单元",
				Category:  types.InstallCheckCategoryNeedConfirm,
			}
		} else {
			results[idx] = &types.NodeProxyInstallCheckResult{
				Status:    types.NodeProxyInstallCheckStatusNormalInstall,
				Matched:   types.ConvertHostToNodeProxyInstallCheckMatchedItem(matchedHost),
				MessageEn: "Install Proxy",
				MessageZh: "安装Proxy",
				Category:  types.InstallCheckCategoryNormalInstall,
			}
		}
	}

	return results, nil
}

func newProxyInstallChecker(
	storageHost topoStg.IStorageHost, storageNetworkUnit topoStg.IStorageNetworkUnit) *proxyInstallChecker {

	return &proxyInstallChecker{
		storageHost:        storageHost,
		storageNetworkUnit: storageNetworkUnit,
	}
}

type proxyInstallChecker struct {
	hostMap            map[int64]*types.Host
	networkUnitMap     map[int64]*types.NetworkUnit
	storageHost        topoStg.IStorageHost
	storageNetworkUnit topoStg.IStorageNetworkUnit
}

func (ic *proxyInstallChecker) hasInnerIP(networkAreaID int64, innerIPList []string) (*types.Host, bool) {
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

func (ic *proxyInstallChecker) hasInnerIPV6(networkAreaID int64, innerIPV6 []string) (*types.Host, bool) {
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

func (ic *proxyInstallChecker) getHost(hostID int64) (*types.Host, bool) {
	host, ok := ic.hostMap[hostID]
	return host, ok
}

func (ic *proxyInstallChecker) getNetworkUnit(networkUnitID int64) (*types.NetworkUnit, bool) {
	networkUnit, ok := ic.networkUnitMap[networkUnitID]
	return networkUnit, ok
}

func (ic *proxyInstallChecker) init(
	nCtx contextx.IContext, networkUnitIDList []int64, hostIDList []int64, innerIPList []string, innerIPV6List []string) error {

	networkUnitMap, err := ic.fetchNetworkUnits(nCtx, networkUnitIDList)
	if err != nil {
		return err
	}
	ic.networkUnitMap = networkUnitMap

	hostMap, err := ic.fetchHosts(nCtx, hostIDList, innerIPList, innerIPV6List)
	if err != nil {
		return err
	}
	ic.hostMap = hostMap

	return nil
}

func (ic *proxyInstallChecker) fetchNetworkUnits(
	nCtx contextx.IContext, networkUnitIDList []int64) (map[int64]*types.NetworkUnit, error) {

	networkUnits := make(map[int64]*types.NetworkUnit)
	if len(networkUnitIDList) == 0 {
		return networkUnits, nil
	}

	networkUnitList, _, err := ic.storageNetworkUnit.ListNetworkUnit(nCtx, types.UnlimitedPage(), &types.NetworkUnitCondition{
		ExactInclude: &types.NetworkUnitExactFields{
			NetworkUnitID: networkUnitIDList,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list networkunits by networkunit-id: %w", err)
	}

	for _, networkUnit := range networkUnitList {
		networkUnits[networkUnit.ID] = networkUnit
	}

	return networkUnits, nil
}

func (ic *proxyInstallChecker) fetchHosts(nCtx contextx.IContext, hostIDList []int64, innerIPList []string, innerIPV6List []string) (
	map[int64]*types.Host, error) {

	hosts := make(map[int64]*types.Host)

	if len(hostIDList) > 0 {
		hostsByID, _, err := ic.storageHost.ListHost(nCtx, types.UnlimitedPage(), &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{HostID: hostIDList},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list hosts by host-id: %w", err)
		}
		for _, host := range hostsByID {
			hosts[host.HostID] = host
		}
	}

	if len(innerIPList) > 0 {
		hostsByInnerIP, _, err := ic.storageHost.ListHost(nCtx, types.UnlimitedPage(), &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{InnerIP: innerIPList},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list hosts by innerip: %w", err)
		}
		for _, host := range hostsByInnerIP {
			hosts[host.HostID] = host
		}
	}

	if len(innerIPV6List) > 0 {
		hostsByInnerIPV6, _, err := ic.storageHost.ListHost(nCtx, types.UnlimitedPage(), &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{InnerIPV6: innerIPV6List},
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

// allIPsBelongToHost checks if all given IPs exist in the host's IP list.
func allIPsBelongToHost(givenIPs []string, hostIPs []string) bool {
	hostIPSet := make(map[string]struct{}, len(hostIPs))
	for _, ip := range hostIPs {
		hostIPSet[ip] = struct{}{}
	}
	for _, ip := range givenIPs {
		if _, ok := hostIPSet[ip]; !ok {
			return false
		}
	}

	return true
}
