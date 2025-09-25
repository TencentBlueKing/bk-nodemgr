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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AgentReconfig reconfig agent.
func (h *handler) AgentReconfig(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeAgentReconfigReq)
	if err := rCtx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to reconfig agent, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, err := h.getReconfigNodeHosts(rCtx, req.GetHost())
	if err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to reconfig agent, failed to get host list. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	reqHosts := req.GetHost()
	nodeDeploys := make([]*types.NodeDeployment, len(hosts))
	for idx := range reqHosts {
		reqHost := reqHosts[idx]

		nodeDeploy, err := h.generatesReconfigDeploys(rCtx.TenantID(), reqHost, hosts)
		if err != nil {
			h.logger.ErrorCtxf(rCtx, "failed to reconfig agent, failed to generate node deployment. err: %v", err)

			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		nodeDeploys[idx] = nodeDeploy
	}

	workflowID, err := h.manager.LaunchReconfigNode(rCtx, manager.ReconfigNodeParam{
		Type:            types.NodeWorkflowTypeReconfigAgent,
		BizIDs:          h.getReconfigNodeBizIDs(hosts),
		Operator:        rCtx.BKUsername(),
		NodeDeployments: nodeDeploys,
	})
	if err != nil {
		h.logger.ErrorCtxf(rCtx, "failed to reconfig agent: %v", err)
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentReconfigResp)
	resp.ConvertWorkflowID(workflowID)

	h.logger.InfoCtxf(rCtx, "launched reconfig agent workflow: %s", workflowID)

	return resp.GetData(), nil
}

func (h *handler) getReconfigNodeHosts(
	nCtx contextx.IContext, reqHosts []*protoBackend.NodeAgentReconfigReq_Host) (map[int64]*types.Host, error) {

	if len(reqHosts) == 0 {
		return nil, errors.New("empty host list")
	}

	hostIDs := make(map[int64]struct{})
	for _, host := range reqHosts {
		hostIDs[host.GetBkHostId()] = struct{}{}
	}

	hosts, _, err := h.storageHost.ListHost(nCtx,
		types.UnlimitedPage(),
		&types.HostCondition{ExactInclude: &types.HostExactFields{
			HostID: conv.MapKeyToSlice(hostIDs),
		}})

	result := make(map[int64]*types.Host)
	for _, host := range hosts {
		result[host.HostID] = host
	}

	return result, err
}

func (h *handler) getReconfigNodeBizIDs(hostMap map[int64]*types.Host) []int64 {
	bizIDs := make(map[int64]struct{})
	for _, host := range hostMap {
		bizIDs[host.Static.BizID] = struct{}{}
	}

	return conv.MapKeyToSlice(bizIDs)
}

func (h *handler) generatesReconfigDeploys(
	tenantID string,
	reqHost *protoBackend.NodeAgentReconfigReq_Host,
	hostMap map[int64]*types.Host) (*types.NodeDeployment, error) {

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
		RestartOptions: types.DeploymentRestartOptions{
			ForceRestart:           reqHost.GetForce(),
			GracefulRestartTimeout: time.Second * time.Duration(reqHost.GetGracefulRestartTimeoutSec()),
		},
		TransferOptions: types.DeploymentTransferOptions{
			SelectDownloads: true,
			EnableInstaller: true,
		},
	})

	return nodeDeployment, nil
}
