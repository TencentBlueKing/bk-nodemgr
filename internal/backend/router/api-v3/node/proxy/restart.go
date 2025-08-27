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

// Restart restart proxy.
func (h *handler) Restart(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeProxyRestartReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to restart proxy, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	nodeDeployments, bizIDs, err := h.generatesRestartNodeDeployments(ctx, req)
	if err != nil {
		h.logger.Errorf("failed to restart proxy, failed to generate node deployments. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.manager.LaunchRestartNode(ctx, manager.RestartNodeParam{
		Type:            types.NodeWorkflowTypeRestartProxy,
		BizIDs:          bizIDs,
		Operator:        ctx.BKUsername(),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to restart proxy: %v", err)
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeProxyRestartResp)
	resp.ConvertWorkflowID(workflowID)

	h.logger.InfoCtxf(ctx, "launched restart proxy workflow: %s", workflowID)

	return resp.GetData(), nil
}

func (h *handler) getRestartNodeHosts(
	ctx contextx.ITenantContext, reqHosts []*protoBackend.NodeProxyRestartReq_Host) (map[int64]*types.Host, error) {

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
		if host.Dynamic.NodeRole != types.NodeRoleProxy {
			return nil, fmt.Errorf("node role is not proxy. host-id(%d), node-role(%s)", host.HostID, host.Dynamic.NodeRole)
		}

		result[host.HostID] = host
	}

	return result, err
}

func (h *handler) generatesRestartNodeDeployments(ctx contextx.ITenantContext, req *protoBackend.NodeProxyRestartReq) (
	[]*types.NodeDeployment, []int64, error) {

	typeHosts, err := h.getRestartNodeHosts(ctx, req.GetHost())
	if err != nil {
		return nil, nil, err
	}

	// build biz id list.
	bizIDMap := make(map[int64]struct{})
	for _, host := range typeHosts {
		bizIDMap[host.Static.BizID] = struct{}{}
	}
	bizIDs := conv.MapKeyToSlice(bizIDMap)

	nodeDeployments := make([]*types.NodeDeployment, len(req.Host))
	for idx, reqHost := range req.Host {
		host, ok := typeHosts[reqHost.GetBkHostId()]
		if !ok {
			return nil, nil, fmt.Errorf("host not found. host-id(%d)", reqHost.GetBkHostId())
		}

		nodeDeployment := types.NewNodeDeployment(&types.DeploymentInfo{
			Host: types.Host{
				TenantID: ctx.TenantID(),
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

		nodeDeployments[idx] = nodeDeployment
	}

	return nodeDeployments, bizIDs, nil
}
