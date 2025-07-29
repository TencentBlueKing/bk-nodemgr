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
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AgentRestart restart agent.
func (h *handler) AgentRestart(ctx *rest.Context) (interface{}, error) {
	req := new(protoBackend.NodeAgentRestartReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to restart agent, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	hosts, err := h.getRestartNodeHosts(ctx, req.GetHost())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to restart agent, failed to get host list. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	reqHosts := req.GetHost()
	nodeDeploys := make([]*types.NodeDeployment, len(hosts))
	for idx := range reqHosts {
		reqHost := reqHosts[idx]

		nodeDeploy, err := h.generatesRestartDeploys(ctx.TenantID, reqHost, hosts)
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to restart agent, failed to generate node deployment. err: %v", err)

			return nil, errf.ErrWrap(errf.InvalidParameter, err)
		}

		nodeDeploys[idx] = nodeDeploy
	}

	workflowID, err := h.manager.LaunchRestartNode(ctx, manager.RestartNodeParam{
		Type:            types.NodeWorkflowTypeRestartAgent,
		BizIDs:          h.getRestartNodeBizIDs(hosts),
		Operator:        ctx.LoginName,
		NodeDeployments: nodeDeploys,
	})
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to restart agent: %v", err)
		return nil, errf.ErrWrap(errf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentRestartResp)
	resp.ConvertWorkflowID(workflowID)

	h.logger.InfoCtxf(ctx, "launched restart agent workflow: %s", workflowID)

	return resp.GetData(), nil
}

func (h *handler) getRestartNodeHosts(
	ctx context.Context, reqHosts []*protoBackend.NodeAgentRestartReq_Host) (map[int64]*types.Host, error) {

	if len(reqHosts) == 0 {
		return nil, errors.New("empty host list")
	}

	hostIDs := make(map[int64]struct{})
	for _, host := range reqHosts {
		hostIDs[host.GetBkHostId()] = struct{}{}
	}

	hosts, _, err := h.storageHost.ListHost(ctx,
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

func (h *handler) getRestartNodeBizIDs(hostMap map[int64]*types.Host) []int64 {
	bizIDs := make(map[int64]struct{})
	for _, host := range hostMap {
		bizIDs[host.Static.BizID] = struct{}{}
	}

	return conv.MapKeyToSlice(bizIDs)
}

func (h *handler) generatesRestartDeploys(
	tenantID string,
	reqHost *protoBackend.NodeAgentRestartReq_Host,
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
