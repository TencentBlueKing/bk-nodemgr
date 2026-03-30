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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AgentUninstall uninstall agent.
func (h *handler) AgentUninstall(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeAgentUninstallReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, err := h.getUninstallNodeHosts(rCtx, req.GetHost())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall agent, failed to get host list")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := validateHostNetworkUnit(hosts); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall agent, invalid network unit")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	unitsMap, err := h.generatesUnitDirectLink(rCtx, hosts)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall agent, failed to get network unit info")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	bizIDs := h.getUninstallNodeBizIDs(hosts)
	resources := buildBizResources(bizIDs)
	if authErr := h.authorizer.Check(rCtx, auth.ActionAgentOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to uninstall agent, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	reqHosts := req.GetHost()
	nodeDeploys := make([]*types.NodeDeployment, len(hosts))
	for idx := range reqHosts {
		reqHost := reqHosts[idx]

		nodeDeploy, err := h.generatesUninstallDeploys(rCtx.TenantID(), reqHost, hosts, unitsMap)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall agent, failed to generate node deployment")

			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		nodeDeploys[idx] = nodeDeploy
	}

	workflowID, err := h.nodeMgrIface.LaunchUninstallNode(rCtx, types.UninstallNodeParam{
		Type:            types.NodeWorkflowTypeUninstallAgent,
		BizIDs:          bizIDs,
		Operator:        rCtx.BKUsername(),
		NodeDeployments: nodeDeploys,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to uninstall agent")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentUninstallResp)
	resp.ConvertWorkflowID(workflowID)

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched uninstall agent workflow")

	return resp.GetData(), nil
}

func (h *handler) getUninstallNodeBizIDs(hostMap map[int64]*types.Host) []int64 {
	bizIDs := make(map[int64]struct{})
	for _, host := range hostMap {
		bizIDs[host.Static.BizID] = struct{}{}
	}

	return conv.MapKeyToSlice(bizIDs)
}

func (h *handler) getUninstallNodeHosts(
	nCtx contextx.IContext, reqHosts []*protoBackend.NodeAgentUninstallReq_Host) (map[int64]*types.Host, error) {

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
	reqHost *protoBackend.NodeAgentUninstallReq_Host,
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
		UninstallOptions: types.DeploymentUninstallOptions{
			DirectLink: unitsMap[host.Dynamic.NetworkUnitID],
		},
		TransferOptions: types.DeploymentTransferOptions{
			SelectDownloads: true,
			EnableInstaller: true,
		},
	})

	return nodeDeployment, nil
}
