/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	assignAgentNetworkUnitBatchSize    = 2000
	assignAgentNetworkUnitBatchTimeout = 10 * time.Minute
)

// AssignAgentNetworkUnit assigns a network unit to agent hosts.
func (mgr *Manager) AssignAgentNetworkUnit(nCtx contextx.IContext, param types.NodeAgentAssignUnitParam) (*types.NodeAgentAssignUnitResult, error) {
	networkUnit, err := mgr.conf.StorageTopo.GetNetworkUnit(nCtx, param.NetworkUnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch networkunit: %w", err)
	}

	hosts, missingHostIDs, err := mgr.fetchAgentNetworkUnitAssignmentHosts(nCtx, param.HostIDs)
	if err != nil {
		return nil, err
	}

	if err := validateAgentAssignNetworkUnitHosts(hosts, networkUnit); err != nil {
		return nil, err
	}

	return mgr.assignAgentNetworkUnitToHosts(nCtx, networkUnit.ID, hosts, missingHostIDs...)
}

func (mgr *Manager) fetchAgentNetworkUnitAssignmentHosts(
	nCtx contextx.IContext, hostIDs []int64,
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

	foundIDs := make(map[int64]struct{}, len(uniqueIDs))
	result, err := batchexecutor.Collect(nCtx, uniqueIDs, func(nCtx contextx.IContext, batch []int64) ([]*types.Host, error) {
		hosts, err := mgr.conf.StorageTopo.ListHostWithoutCount(nCtx, types.UnlimitedPage(), &types.HostCondition{
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
	}, batchexecutor.WithBatchSize(assignAgentNetworkUnitBatchSize), batchexecutor.WithTimeout(assignAgentNetworkUnitBatchTimeout))
	if err != nil {
		return nil, nil, err
	}

	missingIDs := make([]int64, 0)
	for _, id := range uniqueIDs {
		if _, ok := foundIDs[id]; !ok {
			missingIDs = append(missingIDs, id)
		}
	}

	return result.Items, missingIDs, nil
}

func validateAgentAssignNetworkUnitHosts(hosts []*types.Host, targetUnit *types.NetworkUnit) error {
	if len(hosts) == 0 {
		return nil
	}

	var firstAreaID int64
	firstAreaIDSet := false
	for _, host := range hosts {
		if host == nil || host.Static == nil {
			return errors.New("empty host static data")
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

func (mgr *Manager) assignAgentNetworkUnitToHosts(
	nCtx contextx.IContext, networkUnitID int64, hosts []*types.Host, missingHostIDs ...int64,
) (*types.NodeAgentAssignUnitResult, error) {

	result := &types.NodeAgentAssignUnitResult{
		FailedReasons: make([]string, 0, len(missingHostIDs)),
	}

	for _, id := range missingHostIDs {
		result.FailedCount++
		result.FailedReasons = append(result.FailedReasons, fmt.Sprintf("host-id(%d) not found", id))
	}

	toUpdate := make([]*types.Host, 0, len(hosts))
	for _, host := range hosts {
		if host == nil || host.Dynamic == nil {
			result.FailedCount++
			result.FailedReasons = append(result.FailedReasons, "empty host dynamic data")

			continue
		}

		if host.Dynamic.NetworkUnitID >= 0 {
			result.FailedCount++
			result.FailedReasons = append(result.FailedReasons,
				fmt.Sprintf("host-id(%d) already assigned to networkunit-id(%d)",
					host.HostID, host.Dynamic.NetworkUnitID))

			continue
		}

		host.Dynamic.NetworkUnitID = networkUnitID
		toUpdate = append(toUpdate, host)
	}

	successTouchIDs := make([]int64, 0, len(toUpdate))
	if err := batchexecutor.Execute(nCtx, toUpdate, func(_ contextx.IContext, batch []*types.Host) error {
		if err := mgr.conf.StorageTopo.UpdateHostDynamicFields(
			nCtx, types.HostDynamicFields{NetworkUnitID: true}, batch...,
		); err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to assign unit, batch update failed")
			result.FailedCount += int64(len(batch))
			result.FailedReasons = append(result.FailedReasons,
				fmt.Sprintf("batch update failed for %d hosts: %v", len(batch), err))

			return nil
		}

		result.SuccessCount += int64(len(batch))
		for _, host := range batch {
			successTouchIDs = append(successTouchIDs, host.HostID)
		}

		return nil
	}, batchexecutor.WithBatchSize(assignAgentNetworkUnitBatchSize), batchexecutor.WithTimeout(assignAgentNetworkUnitBatchTimeout)); err != nil {
		return nil, err
	}

	logger.G.Biz(nCtx).
		With("networkunit-id", networkUnitID).
		With("success-count", result.SuccessCount).
		With("failed-count", result.FailedCount).
		Info("batch assign unit completed")

	if len(successTouchIDs) > 0 {
		if err := mgr.conf.StorageTopo.TouchHostOperationTime(nCtx, successTouchIDs...); err != nil {
			logger.G.Biz(nCtx).WithErr(err).Warn("failed to touch host operation time after assign unit")
		}
	}

	return result, nil
}
