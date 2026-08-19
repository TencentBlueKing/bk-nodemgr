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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/compatibility"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AssignProxyUnitMulti batch-assigns multiple network units to proxy hosts.
func (h *handler) AssignProxyUnitMulti(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeProxyAssignUnitMultiReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign multiple proxy units, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	targetUnitByHostID := make(map[int64]int64)
	networkUnitIDs := make([]int64, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		if item == nil {
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("item can not be nil"))
		}
		networkUnitIDs = append(networkUnitIDs, item.GetBkNetworkunitId())
		for _, hostID := range item.GetBkHostId() {
			targetUnitByHostID[hostID] = item.GetBkNetworkunitId()
		}
	}
	networkUnitIDs = conv.SliceUnique(networkUnitIDs)

	networkUnits, err := h.listNetworkUnitsByIDs(rCtx, networkUnitIDs)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign multiple proxy units, failed to fetch network units")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, missingIDs, err := h.fetchHostsInBatches(rCtx, conv.MapKeyToSlice(targetUnitByHostID))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign multiple proxy units, failed to fetch hosts")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	for _, host := range hosts {
		targetUnit := networkUnits[targetUnitByHostID[host.HostID]]
		if err := validateProxyAssignUnitHost(host, targetUnit); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to assign multiple proxy units, precondition validation failed")
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}
	}
	if len(hosts) == 0 {
		resp := &protoBackend.NodeProxyAssignUnitResp{
			Data: &protoBackend.NodeProxyAssignUnitResp_Data{
				FailedCount:   int64(len(missingIDs)),
				FailedReasons: buildMissingHostReasons(missingIDs),
			},
		}

		return resp.GetData(), nil
	}

	if authErr := h.authorizer.Check(
		rCtx, auth.ActionNetworkUnitUseForProxy, authRouter.BuildNetworkUnitResources(networkUnitIDs...),
	); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to assign multiple proxy units, networkunit permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	bizIDMap := make(map[int64]struct{})
	for _, host := range hosts {
		bizIDMap[host.Static.BizID] = struct{}{}
	}
	bizIDs := make([]int64, 0, len(bizIDMap))
	for bizID := range bizIDMap {
		bizIDs = append(bizIDs, bizID)
	}
	if authErr := h.authorizer.Check(rCtx, auth.ActionProxyOperate, authRouter.BuildBizResources(bizIDs...)); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to assign multiple proxy units, proxy operate permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	compatibilityPolicy := h.readCompatibilityModePolicy(rCtx)
	nodeDeployments := make([]*types.NodeDeployment, 0, len(hosts))
	for _, host := range hosts {
		targetUnit := networkUnits[targetUnitByHostID[host.HostID]]
		nodeDeployments = append(nodeDeployments, buildAssignProxyDeployment(
			rCtx, host, targetUnit.ID, compatibilityPolicy,
		))
	}

	workflowID, err := h.nodeMgrIface.LaunchAssignProxyUnit(rCtx, types.AssignProxyUnitParam{
		Type:            types.NodeWorkflowTypeAssignProxyUnit,
		BizIDs:          bizIDs,
		Operator:        rCtx.BKUsername(),
		NodeDeployments: nodeDeployments,
	})
	failedReasons := buildMissingHostReasons(missingIDs)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign multiple proxy units, failed to launch workflow")
		failedReasons = append(failedReasons, fmt.Sprintf("failed to launch workflow: %v", err))
		resp := &protoBackend.NodeProxyAssignUnitResp{
			Data: &protoBackend.NodeProxyAssignUnitResp_Data{
				FailedCount:   int64(len(missingIDs) + len(hosts)),
				FailedReasons: failedReasons,
			},
		}

		return resp.GetData(), nil
	}

	resp := &protoBackend.NodeProxyAssignUnitResp{
		Data: &protoBackend.NodeProxyAssignUnitResp_Data{
			SuccessCount:  int64(len(hosts)),
			FailedCount:   int64(len(missingIDs)),
			FailedReasons: failedReasons,
			WorkflowId:    workflowID,
		},
	}

	return resp.GetData(), nil
}

func (h *handler) listNetworkUnitsByIDs(
	rCtx restserver.IContext, networkUnitIDs []int64,
) (map[int64]*types.NetworkUnit, error) {

	networkUnitList, _, err := h.storageNetworkUnit.ListNetworkUnit(rCtx, types.UnlimitedPage(), &types.NetworkUnitCondition{
		ExactInclude: &types.NetworkUnitExactFields{NetworkUnitID: networkUnitIDs},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list networkunits: %w", err)
	}

	networkUnits := make(map[int64]*types.NetworkUnit, len(networkUnitList))
	for _, networkUnit := range networkUnitList {
		networkUnits[networkUnit.ID] = networkUnit
	}
	for _, networkUnitID := range networkUnitIDs {
		if _, ok := networkUnits[networkUnitID]; !ok {
			return nil, fmt.Errorf("networkunit with id %d not found", networkUnitID)
		}
	}

	return networkUnits, nil
}

func validateProxyAssignUnitHost(host *types.Host, targetUnit *types.NetworkUnit) error {
	if host == nil || host.Static == nil || host.Dynamic == nil {
		return fmt.Errorf("host data is incomplete")
	}
	if targetUnit == nil {
		return fmt.Errorf("target networkunit is not found")
	}
	if host.Dynamic.NodeRole != types.NodeRoleProxy {
		return fmt.Errorf("node role is not proxy. host-id(%d), node-role(%s)", host.HostID, host.Dynamic.NodeRole)
	}
	if host.Dynamic.NodeStatus != types.NodeStatusRunning {
		return fmt.Errorf("host-id(%d) is not online, status(%s)", host.HostID, host.Dynamic.NodeStatus)
	}
	if host.Dynamic.NetworkUnitID >= 0 {
		return fmt.Errorf("host-id(%d) already assigned to networkunit-id(%d)", host.HostID, host.Dynamic.NetworkUnitID)
	}
	if host.Static.NetworkAreaID != targetUnit.NetworkAreaID {
		return fmt.Errorf("host network area (%d) does not match target networkunit area (%d)",
			host.Static.NetworkAreaID, targetUnit.NetworkAreaID)
	}

	return nil
}

func buildAssignProxyDeployment(
	rCtx restserver.IContext, host *types.Host, networkUnitID int64, compatibilityPolicy compatibility.Policy,
) *types.NodeDeployment {

	deployHostDynamic := host.Dynamic
	deployHostDynamic.NetworkUnitID = networkUnitID
	deployHostDynamic.ProxyTags = types.AllProxyTag()

	nodeDeployment := types.NewNodeDeployment(&types.DeploymentInfo{
		Host: types.Host{
			TenantID: rCtx.TenantID(),
			HostID:   host.HostID,
			Static:   host.Static,
			Dynamic:  deployHostDynamic,
		},
	})
	nodeDeployment.Info.InstallOptions.InstallPreOrderedPlugins = true
	nodeDeployment.Info.InstallOptions.EnableCompatibilityMode = compatibility.DecideCompatibilityMode(
		compatibilityPolicy, rCtx.TenantID(), host.Static.BizID, "bkmonitorbeat",
	)

	return nodeDeployment
}

func buildMissingHostReasons(ids []int64) []string {
	reasons := make([]string, 0, len(ids))
	for _, id := range ids {
		reasons = append(reasons, fmt.Sprintf("host-id(%d) not found", id))
	}

	return reasons
}
