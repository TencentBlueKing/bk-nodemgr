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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Reconfig reconfig proxy.
func (h *handler) Reconfig(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeProxyReconfigReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to reconfig proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	nodeDeployments, bizIDs, err := h.generatesReconfigNodeDeployments(rCtx, req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to reconfig proxy, failed to generate node deployments")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.nodeMgrIface.LaunchReconfigNode(rCtx, types.ReconfigNodeParam{
		Type:            types.NodeWorkflowTypeReconfigProxy,
		BizIDs:          bizIDs,
		Operator:        rCtx.BKUsername(),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to reconfig proxy")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeProxyReconfigResp)
	resp.ConvertWorkflowID(workflowID)

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched reconfig proxy workflow")

	return resp.GetData(), nil
}

func (h *handler) getReconfigNodeHosts(
	nCtx contextx.IContext, reqHosts []*protoBackend.NodeProxyReconfigReq_Host) (map[int64]*types.Host, error) {

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
		if host.Dynamic.NodeRole != types.NodeRoleProxy {
			return nil, fmt.Errorf("node role is not proxy. host-id(%d), node-role(%s)", host.HostID, host.Dynamic.NodeRole)
		}

		result[host.HostID] = host
	}

	return result, err
}

func (h *handler) generatesReconfigNodeDeployments(nCtx contextx.IContext, req *protoBackend.NodeProxyReconfigReq) (
	[]*types.NodeDeployment, []int64, error) {

	typeHosts, err := h.getReconfigNodeHosts(nCtx, req.GetHost())
	if err != nil {
		return nil, nil, err
	}

	// build biz id list.
	bizIDMap := make(map[int64]struct{})
	for _, host := range typeHosts {
		bizIDMap[host.Static.BizID] = struct{}{}
	}
	bizIDs := conv.MapKeyToSlice(bizIDMap)

	nodeDeployments := make([]*types.NodeDeployment, len(req.GetHost()))
	for idx, reqHost := range req.GetHost() {
		host, ok := typeHosts[reqHost.GetBkHostId()]
		if !ok {
			return nil, nil, fmt.Errorf("host not found. host-id(%d)", reqHost.GetBkHostId())
		}

		nodeDeployment := types.NewNodeDeployment(&types.DeploymentInfo{
			Host: types.Host{
				TenantID: nCtx.TenantID(),
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
