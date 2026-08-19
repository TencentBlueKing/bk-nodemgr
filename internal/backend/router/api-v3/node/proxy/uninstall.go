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

// Package proxy ...
package proxy

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Uninstall uninstalls proxy.
func (h *handler) Uninstall(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeProxyUninstallReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, err := h.getUninstallNodeHosts(rCtx, req.GetHost())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall proxy, failed to get host list")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := validateHostNetworkUnit(hosts); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall proxy, invalid network unit")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	bizIDs := h.getUninstallNodeBizIDs(hosts)
	resources := authRouter.BuildBizResources(bizIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionProxyOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to uninstall proxy, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	originUnitDirectMap, err := h.generatesUninstallOriginUnitDirectLink(rCtx, hosts)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall proxy, failed to get origin network unit direct link")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	reqHosts := req.GetHost()
	nodeDeploys := make([]*types.NodeDeployment, len(hosts))
	for idx := range reqHosts {
		reqHost := reqHosts[idx]

		nodeDeploy, err := h.generatesUninstallDeploys(rCtx.TenantID(), reqHost, hosts, originUnitDirectMap)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall proxy, failed to generate node deployment")

			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		nodeDeploys[idx] = nodeDeploy
	}

	workflowID, err := h.nodeMgrIface.LaunchUninstallNode(rCtx, types.UninstallNodeParam{
		Type:            types.NodeWorkflowTypeUninstallProxy,
		BizIDs:          bizIDs,
		Operator:        rCtx.BKUsername(),
		NodeDeployments: nodeDeploys,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall proxy")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeProxyUninstallResp)
	resp.ConvertWorkflowID(workflowID)

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched uninstall proxy workflow")

	return resp.GetData(), nil
}

func (h *handler) getUninstallNodeBizIDs(hostMap map[int64]*types.Host) []int64 {
	bizIDs := make(map[int64]struct{})
	for _, host := range hostMap {
		bizIDs[host.Static.BizID] = struct{}{}
	}

	return conv.MapKeyToSlice(bizIDs)
}

func (h *handler) generatesUninstallOriginUnitDirectLink(
	nCtx contextx.IContext, hostMap map[int64]*types.Host) (map[int64]bool, error) {

	unitIDSet := make(map[int64]struct{})
	for _, host := range hostMap {
		unitIDSet[host.Dynamic.ProxyInstallOriginUnitID] = struct{}{}
	}

	units, _, err := h.storageNetworkUnit.ListNetworkUnit(nCtx,
		types.UnlimitedPage(),
		&types.NetworkUnitCondition{
			ExactInclude: &types.NetworkUnitExactFields{
				NetworkUnitID: conv.MapKeyToSlice(unitIDSet),
			},
		})
	if err != nil {
		return nil, fmt.Errorf("list origin network unit failed: %w", err)
	}

	unitInfoMap := make(map[int64]bool)
	for _, unit := range units {
		unitInfoMap[unit.ID] = unit.IsDirect
	}

	return unitInfoMap, nil
}

func (h *handler) getUninstallNodeHosts(
	nCtx contextx.IContext, reqHosts []*protoBackend.NodeProxyUninstallReq_Host) (map[int64]*types.Host, error) {

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

func (h *handler) generatesUninstallDeploys(
	tenantID string,
	reqHost *protoBackend.NodeProxyUninstallReq_Host,
	hostMap map[int64]*types.Host,
	originUnitDirectMap map[int64]bool) (*types.NodeDeployment, error) {

	hostID := reqHost.GetBkHostId()
	host, ok := hostMap[hostID]
	if !ok {
		return nil, fmt.Errorf("host not found. host-id(%d)", hostID)
	}

	originUnitID := host.Dynamic.ProxyInstallOriginUnitID
	crossUnit := originUnitID != host.Dynamic.NetworkUnitID

	// uninstall ,we need to clear the proxy dynamic config
	nodeDeployment := types.NewNodeDeployment(&types.DeploymentInfo{
		Host: types.Host{
			TenantID: tenantID,
			HostID:   host.HostID,
			Static:   host.Static,
			Dynamic: &types.HostDynamic{
				NodeRole:                 host.Dynamic.NodeRole,
				NodeStatus:               host.Dynamic.NodeStatus,
				NodeGeneration:           host.Dynamic.NodeGeneration,
				NodeOsType:               host.Dynamic.NodeOsType,
				NodeCPUArch:              host.Dynamic.NodeCPUArch,
				AgentID:                  host.Dynamic.AgentID,
				NetworkUnitID:            host.Dynamic.NetworkUnitID,
				ProxyInstallOriginUnitID: originUnitID,
			},
		},
		UninstallOptions: types.DeploymentUninstallOptions{
			DirectLink: originUnitDirectMap[originUnitID],
			SkipReport: crossUnit,
		},
		TransferOptions: types.DeploymentTransferOptionsOnlyTransferInstaller(),
	})

	return nodeDeployment, nil
}
