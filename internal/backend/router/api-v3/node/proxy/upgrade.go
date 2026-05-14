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

// Upgrade upgrade proxy.
func (h *handler) Upgrade(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeProxyUpgradeReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, err := h.getUpgradeNodeHosts(rCtx, req.GetHost())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade proxy, failed to get host list")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := validateUpgradeHostNetworkUnit(req.GetHost(), hosts); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade proxy, invalid network unit")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	nodeDeployments, bizIDs, err := h.generatesUpgradeNodeDeployments(rCtx, req, hosts)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade proxy, failed to generate node deployments")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	resources := authRouter.BuildBizResources(bizIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionProxyOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to upgrade proxy, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	workflowID, err := h.nodeMgrIface.LaunchUpgradeNode(rCtx, types.UpgradeNodeParam{
		Type:            types.NodeWorkflowTypeUpgradeProxy,
		BizIDs:          bizIDs,
		Operator:        rCtx.BKUsername(),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade proxy")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeProxyUpgradeResp)
	resp.ConvertWorkflowID(workflowID)

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched upgrade proxy workflow")

	return resp.GetData(), nil
}

func (h *handler) getUpgradeNodeHosts(
	nCtx contextx.IContext, reqHosts []*protoBackend.NodeProxyUpgradeReq_Host) (map[int64]*types.Host, error) {

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

// generatesUpgradeNodeDeployments generates upgrade node deployments and get biz id list.
func (h *handler) generatesUpgradeNodeDeployments(
	nCtx contextx.IContext, req *protoBackend.NodeProxyUpgradeReq, typeHosts map[int64]*types.Host,
) ([]*types.NodeDeployment, []int64, error) {

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

		// Use target networkunit if specified; otherwise keep the host's current networkunit.
		targetNetworkUnitID := resolveUpgradeNetworkUnitID(host.Dynamic.NetworkUnitID, reqHost.GetBkNetworkunitId())

		deployHostDynamic := host.Dynamic
		deployHostDynamic.NetworkUnitID = targetNetworkUnitID

		nodeDeployment := types.NewNodeDeployment(&types.DeploymentInfo{
			Host: types.Host{
				TenantID: nCtx.TenantID(),
				HostID:   host.HostID,
				Static:   host.Static,
				Dynamic:  deployHostDynamic,
			},
			RestartOptions: types.DeploymentRestartOptions{
				ForceRestart:           reqHost.GetForce(),
				GracefulRestartTimeout: time.Second * time.Duration(reqHost.GetGracefulRestartTimeoutSec()),
			},
		})

		// set target version.
		nodeDeployment.Info.Host.Dynamic.NodeVersion = reqHost.GetTargetVersion()

		nodeDeployments[idx] = nodeDeployment
	}

	return nodeDeployments, bizIDs, nil
}
