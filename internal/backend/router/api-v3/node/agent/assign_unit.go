/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package agent

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const assignUnitBatchSize = 2000

// AgentAssignUnit batch-assigns a network unit to unassigned hosts (metadata-only, no remote operations).
func (h *handler) AgentAssignUnit(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeAgentAssignUnitReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkUnit, err := h.fetchNetworkUnit(rCtx, req.GetBkNetworkunitId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit, failed to fetch network unit")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, missingIDs, err := h.fetchHostsInBatches(rCtx, req.GetBkHostId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit, failed to fetch hosts")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := validateHostNetworkAreaConsistency(hosts, networkUnit); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit, network area validation failed")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	networkUnitResource := buildNetworkUnitResources(networkUnit.ID)
	if authErr := h.authorizer.BatchCheck(rCtx, auth.ActionNetworkUnitUseForAgent, networkUnitResource); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to assign unit, networkunit permission denied")
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
	resources := buildBizResources(bizIDs)
	if authErr := h.authorizer.BatchCheck(rCtx, auth.ActionAgentOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to assign unit, agent operate permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	var successCount int64
	var failedCount int64
	failedReasons := make([]string, 0, len(missingIDs))

	for _, id := range missingIDs {
		failedCount++
		failedReasons = append(failedReasons, fmt.Sprintf("host-id(%d) not found", id))
	}

	toUpdate := make([]*types.Host, 0, len(hosts))
	var successTouchIDs []int64
	for _, host := range hosts {
		if host.Dynamic.NetworkUnitID >= 0 {
			failedCount++
			failedReasons = append(failedReasons,
				fmt.Sprintf("host-id(%d) already assigned to networkunit-id(%d)",
					host.HostID, host.Dynamic.NetworkUnitID))

			continue
		}

		host.Dynamic.NetworkUnitID = networkUnit.ID
		toUpdate = append(toUpdate, host)
	}

	for i := 0; i < len(toUpdate); i += assignUnitBatchSize {
		end := i + assignUnitBatchSize
		if end > len(toUpdate) {
			end = len(toUpdate)
		}

		batch := toUpdate[i:end]
		if err := h.storageHost.UpdateHostDynamicFields(
			rCtx, types.HostDynamicFields{NetworkUnitID: true}, batch...,
		); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit, batch update failed")
			failedCount += int64(len(batch))
			failedReasons = append(failedReasons,
				fmt.Sprintf("batch update failed for %d hosts: %v", len(batch), err))

			continue
		}

		successCount += int64(len(batch))
		for _, host := range batch {
			successTouchIDs = append(successTouchIDs, host.HostID)
		}
	}

	logger.G.Biz(rCtx).
		With("networkunit-id", networkUnit.ID).
		With("success-count", successCount).
		With("failed-count", failedCount).
		Info("batch assign unit completed")

	if len(successTouchIDs) > 0 {
		if err := h.storageHost.TouchHostOperationTime(rCtx, successTouchIDs...); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Warn("failed to touch host operation time after assign unit")
		}
	}

	resp := &protoBackend.NodeAgentAssignUnitResp{
		Data: &protoBackend.NodeAgentAssignUnitResp_Data{
			SuccessCount:  successCount,
			FailedCount:   failedCount,
			FailedReasons: failedReasons,
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

	var allHosts []*types.Host
	foundIDs := make(map[int64]struct{})

	for i := 0; i < len(hostIDs); i += assignUnitBatchSize {
		end := i + assignUnitBatchSize
		if end > len(hostIDs) {
			end = len(hostIDs)
		}
		batch := hostIDs[i:end]

		hosts, _, err := h.storageHost.ListHost(rCtx, types.UnlimitedPage(), &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				HostID: batch,
			},
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch hosts: %w", err)
		}

		for _, host := range hosts {
			foundIDs[host.HostID] = struct{}{}
		}
		allHosts = append(allHosts, hosts...)
	}

	var missingIDs []int64
	for _, id := range hostIDs {
		if _, ok := foundIDs[id]; !ok {
			missingIDs = append(missingIDs, id)
		}
	}

	return allHosts, missingIDs, nil
}

func validateHostNetworkAreaConsistency(hosts []*types.Host, targetUnit *types.NetworkUnit) error {
	if len(hosts) == 0 {
		return nil
	}

	var firstAreaID int64
	firstAreaIDSet := false

	for _, host := range hosts {
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
