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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AgentInstallCheck checks if an agent can be installed on hosts.
func (h *handler) AgentInstallCheck(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeAgentInstallCheckReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to check install agent, failed to decode request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	results, err := h.checkHostStatus(ctx, req.GetHost())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to check install agent: %v", err)
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentInstallCheckResp)
	resp.ConvertResultFromTypes(results, len(results))

	h.logger.InfoCtxf(ctx, "checked install agent")

	return resp.GetData(), nil
}

func (h *handler) checkHostStatus(ctx contextx.ITenantContext,
	hosts []*protoBackend.NodeAgentInstallCheckReq_Host) ([]*types.NodeAgentInstallCheckResult, error) {

	unitIDToAreaMap, err := h.mapUnitToArea(ctx, hosts)
	if err != nil {
		return nil, err
	}

	gp := gopool.NewPool()
	results := make([]*types.NodeAgentInstallCheckResult, len(hosts))

	for i, host := range hosts {
		i, host := i, host
		gp.Go(func() error {
			result, err := h.checkSingleHostStatus(ctx, host, unitIDToAreaMap[host.GetBkNetworkunitId()])
			if err != nil {
				h.logger.ErrorCtxf(ctx, "failed to check status for host. inner-ip(%s), network-unit-id(%d): %v",
					host.GetBkHostInnerip(), host.GetBkNetworkunitId(), err)

				return err
			}
			results[i] = result

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// mapUnitToArea maps unit IDs to their network area IDs.
func (h *handler) mapUnitToArea(ctx contextx.ITenantContext, hosts []*protoBackend.NodeAgentInstallCheckReq_Host) (map[int64]int64, error) {
	unitIDs := make([]int64, 0, len(hosts))
	for _, host := range hosts {
		unitIDs = append(unitIDs, host.GetBkNetworkunitId())
	}

	networkUnits, _, err := h.storageNetworkUnit.ListNetworkUnit(ctx, types.UnlimitedPage(), &types.NetworkUnitCondition{
		ExactInclude: &types.NetworkUnitExactFields{
			NetworkUnitID: unitIDs,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list network units: %w", err)
	}

	unitIDToAreaMap := make(map[int64]int64, len(networkUnits))
	for _, networkUnit := range networkUnits {
		unitIDToAreaMap[networkUnit.ID] = networkUnit.NetworkAreaID
	}

	return unitIDToAreaMap, nil
}

func (h *handler) checkSingleHostStatus(ctx contextx.ITenantContext, host *protoBackend.NodeAgentInstallCheckReq_Host, networkAreaID int64) (
	*types.NodeAgentInstallCheckResult, error) {

	hosts, _, err := h.storageHost.ListHost(ctx, types.UnlimitedPage(), &types.HostCondition{
		ExactInclude: &types.HostExactFields{
			NetworkAreaID: []int64{networkAreaID},
			InnerIP:       []string{host.GetBkHostInnerip()},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(hosts) == 0 {
		return &types.NodeAgentInstallCheckResult{
			InnerIP: host.GetBkHostInnerip(),
			State:   types.NodeAgentInstallCheckStateClean,
		}, nil
	}

	return h.determineHostState(hosts, host.GetBkBizId()), nil
}

func (h *handler) determineHostState(
	hosts []*types.Host, bizID int64) *types.NodeAgentInstallCheckResult {

	if len(hosts) > 1 {
		// if install record with the same ip already exists under dynamic addressing. we need to check ip conflict.
		if duplicateHostIDs := getDynamicDuplicateIPHostIDs(hosts); len(duplicateHostIDs) > 0 {
			return &types.NodeAgentInstallCheckResult{
				InnerIP:          hosts[0].Static.InnerIP,
				State:            types.NodeAgentInstallCheckStateDuplicateIP,
				DuplicateHostIDs: duplicateHostIDs,
			}
		}

		// if install record with the same ip already exists under biz. we need to check ip conflict.
		if conflictHostIDs := getIPConflictHostIDsByBizID(hosts, bizID); len(conflictHostIDs) > 0 {
			return &types.NodeAgentInstallCheckResult{
				InnerIP:          hosts[0].Static.InnerIP,
				State:            types.NodeAgentInstallCheckStateConflictIP,
				DuplicateHostIDs: conflictHostIDs,
			}
		}
	}

	host := hosts[0]
	result := &types.NodeAgentInstallCheckResult{
		InnerIP: host.Static.InnerIP,
	}

	if host.Dynamic.NodeStatus == types.NodeStatusRunning {
		// node already exists proxy. we can't install agent.
		if host.Dynamic.NodeRole == types.NodeRoleProxy {
			result.State = types.NodeAgentInstallCheckStateExistProxy
			return result
		}

		// node already exists agent. we can't reinstall agent.
		if host.Dynamic.NodeRole == types.NodeRoleAgent {
			result.State = types.NodeAgentInstallCheckStateExistAgent
			return result
		}
	}

	// node is not running, we can install agent.
	result.State = types.NodeAgentInstallCheckStateNormal

	return result
}

func getDynamicDuplicateIPHostIDs(hosts []*types.Host) []int64 {
	hostIDs := make([]int64, 0)
	for _, host := range hosts {
		if host.Static.Addressing == types.AddressingDynamic {
			hostIDs = append(hostIDs, host.HostID)
		}
	}

	return hostIDs
}

func getIPConflictHostIDsByBizID(hosts []*types.Host, bizID int64) []int64 {
	hostIDs := make([]int64, 0)
	for _, host := range hosts {
		if host.Static.BizID == bizID {
			hostIDs = append(hostIDs, host.HostID)
		}
	}

	return hostIDs
}
