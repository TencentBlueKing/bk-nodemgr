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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	assignUnitBatchSize    = 2000
	assignUnitBatchTimeout = 10 * time.Minute
)

// AssignProxyUnit assigns network unit to proxy hosts and triggers plugin installation workflow.
// nolint: gocognit, funlen, gocyclo, cyclop
func (h *handler) AssignProxyUnit(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeProxyAssignUnitReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign proxy unit, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkUnit, err := h.fetchNetworkUnit(rCtx, req.GetBkNetworkunitId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign proxy unit, failed to fetch network unit")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, missingIDs, err := h.fetchHostsInBatches(rCtx, req.GetBkHostId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign proxy unit, failed to fetch hosts")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := validateProxyAssignUnitPreconditions(hosts, networkUnit); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign proxy unit, precondition validation failed")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkUnitResource := authRouter.BuildNetworkUnitResources([]int64{networkUnit.ID}...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionNetworkUnitUseForProxy, networkUnitResource); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to assign proxy unit, networkunit permission denied")
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
	resources := authRouter.BuildBizResources(bizIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionProxyOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to assign proxy unit, proxy operate permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	// Count missing hosts as per-host failures before workflow launch.
	// After workflow launch, success/failure is at the workflow level:
	// if the workflow is launched successfully, all submitted hosts are counted as success;
	// if the workflow launch fails, all submitted hosts are counted as failed.
	// This differs from agent assign unit, which counts per-host success/failure
	// because it only updates the DB without launching a workflow.
	var successCount int64
	var failedCount int64
	failedReasons := make([]string, 0, len(missingIDs))

	for _, id := range missingIDs {
		failedCount++
		failedReasons = append(failedReasons, fmt.Sprintf("host-id(%d) not found", id))
	}

	nodeDeployments := make([]*types.NodeDeployment, 0, len(hosts))
	for _, host := range hosts {
		deployHostDynamic := host.Dynamic
		deployHostDynamic.NetworkUnitID = networkUnit.ID

		nodeDeployment := types.NewNodeDeployment(&types.DeploymentInfo{
			Host: types.Host{
				TenantID: rCtx.TenantID(),
				HostID:   host.HostID,
				Static:   host.Static,
				Dynamic:  deployHostDynamic,
			},
		})
		nodeDeployment.Info.InstallOptions.InstallPreOrderedPlugins = true

		nodeDeployments = append(nodeDeployments, nodeDeployment)
	}

	workflowID, err := h.nodeMgrIface.LaunchAssignProxyUnit(rCtx, types.AssignProxyUnitParam{
		Type:            types.NodeWorkflowTypeAssignProxyUnit,
		BizIDs:          bizIDs,
		Operator:        rCtx.BKUsername(),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign proxy unit, failed to launch workflow")
		failedCount += int64(len(hosts))
		failedReasons = append(failedReasons, fmt.Sprintf("failed to launch workflow: %v", err))
	} else {
		successCount = int64(len(hosts))
		logger.G.Biz(rCtx).
			With("workflow-id", workflowID).
			With("networkunit-id", networkUnit.ID).
			With("success-count", successCount).
			Info("launched assign proxy unit workflow")
	}

	resp := &protoBackend.NodeProxyAssignUnitResp{
		Data: &protoBackend.NodeProxyAssignUnitResp_Data{
			SuccessCount:  successCount,
			FailedCount:   failedCount,
			FailedReasons: failedReasons,
			WorkflowId:    workflowID,
		},
	}

	return resp.GetData(), nil
}

func (h *handler) fetchNetworkUnit(rCtx restserver.IContext, networkUnitID int64) (*types.NetworkUnit, error) {
	networkUnitList, _, err := h.storageNetworkUnit.ListNetworkUnit(rCtx, types.UnlimitedPage(), &types.NetworkUnitCondition{
		ExactInclude: &types.NetworkUnitExactFields{
			NetworkUnitID: []int64{networkUnitID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch networkunit: %w", err)
	}

	if len(networkUnitList) == 0 {
		return nil, fmt.Errorf("networkunit with id %d not found", networkUnitID)
	}

	return networkUnitList[0], nil
}

func (h *handler) fetchHostsInBatches(
	rCtx restserver.IContext, hostIDs []int64,
) ([]*types.Host, []int64, error) {

	seen := make(map[int64]struct{}, len(hostIDs))
	uniqueIDs := make([]int64, 0, len(hostIDs))

	for _, id := range hostIDs {
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	hostIDs = uniqueIDs

	foundIDs := make(map[int64]struct{})

	result, err := batchexecutor.Collect(rCtx, hostIDs, func(nCtx contextx.IContext, batch []int64) ([]*types.Host, error) {
		hosts, _, err := h.storageHost.ListHost(nCtx, types.UnlimitedPage(), &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				HostID: batch,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch hosts: %w", err)
		}

		for _, host := range hosts {
			foundIDs[host.HostID] = struct{}{}
		}

		return hosts, nil
	}, batchexecutor.WithBatchSize(assignUnitBatchSize), batchexecutor.WithTimeout(assignUnitBatchTimeout))
	if err != nil {
		return nil, nil, err
	}

	var missingIDs []int64
	for _, id := range hostIDs {
		if _, ok := foundIDs[id]; !ok {
			missingIDs = append(missingIDs, id)
		}
	}

	return result.Items, missingIDs, nil
}

func validateProxyAssignUnitPreconditions(hosts []*types.Host, targetUnit *types.NetworkUnit) error {
	if len(hosts) == 0 {
		return nil
	}

	var firstAreaID int64
	firstAreaIDSet := false

	for _, host := range hosts {
		if host.Dynamic.NodeRole != types.NodeRoleProxy {
			return fmt.Errorf("node role is not proxy. host-id(%d), node-role(%s)", host.HostID, host.Dynamic.NodeRole)
		}

		if host.Dynamic.NodeStatus != types.NodeStatusRunning {
			return fmt.Errorf("host-id(%d) is not online, status(%s)", host.HostID, host.Dynamic.NodeStatus)
		}

		if host.Dynamic.NetworkUnitID >= 0 {
			return fmt.Errorf("host-id(%d) already assigned to networkunit-id(%d)", host.HostID, host.Dynamic.NetworkUnitID)
		}

		if !firstAreaIDSet {
			firstAreaID = host.Static.NetworkAreaID
			firstAreaIDSet = true

			continue
		}

		if host.Static.NetworkAreaID != firstAreaID {
			return fmt.Errorf(
				"hosts belong to different network areas (found area %d and %d), cannot assign to a single unit",
				firstAreaID, host.Static.NetworkAreaID)
		}
	}

	if firstAreaIDSet && firstAreaID != targetUnit.NetworkAreaID {
		return fmt.Errorf(
			"host network area (%d) does not match target networkunit area (%d)",
			firstAreaID, targetUnit.NetworkAreaID)
	}

	return nil
}
