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

package agent

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	syncUnassignedAgentNetworkUnitPageSize = 2000
	unassignedAgentNetworkUnitID           = -1
)

type syncUnassignedNetworkUnitReq struct {
	BKBizID []int64 `json:"bk_biz_id"`
}

func (r *syncUnassignedNetworkUnitReq) Validate() error {
	return nil
}

func (r *syncUnassignedNetworkUnitReq) AutoConvert() {}

type syncUnassignedNetworkUnitResp struct {
	SuccessCount  int64    `json:"success_count"`
	FailedCount   int64    `json:"failed_count"`
	FailedReasons []string `json:"failed_reasons"`
}

// SyncUnassignedNetworkUnit syncs unassigned agent hosts to recommended network units.
func (h *handler) SyncUnassignedNetworkUnit(rCtx server.IContext) (any, error) {
	req := new(syncUnassignedNetworkUnitReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error(
			"failed to sync unassigned network unit, failed to decode request body",
		)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.syncUnassignedAgentNetworkUnit(rCtx, req.BKBizID...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to sync unassigned network unit")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	return &syncUnassignedNetworkUnitResp{
		SuccessCount:  result.SuccessCount,
		FailedCount:   result.FailedCount,
		FailedReasons: result.FailedReasons,
	}, nil
}

func (h *handler) syncUnassignedAgentNetworkUnit(
	rCtx server.IContext, bizIDs ...int64,
) (*types.NodeAgentAssignUnitResult, error) {

	bizs, err := h.listAgentNetworkUnitSyncBusinesses(rCtx, bizIDs...)
	if err != nil {
		return nil, err
	}

	result := &types.NodeAgentAssignUnitResult{FailedReasons: []string{}}
	for _, biz := range bizs {
		if err := h.syncUnassignedAgentNetworkUnitByBiz(rCtx, biz.BizID, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (h *handler) listAgentNetworkUnitSyncBusinesses(rCtx server.IContext, bizIDs ...int64) ([]*types.Business, error) {
	conditions := make([]*types.BusinessCondition, 0, 1)
	if len(bizIDs) > 0 {
		conditions = append(conditions, &types.BusinessCondition{
			ExactInclude: &types.BusinessExactFields{BizID: bizIDs},
		})
	}

	bizs, _, err := h.storageTopo.ListBusinesses(rCtx, types.UnlimitedPage(), conditions...)
	if err != nil {
		return nil, fmt.Errorf("failed to list businesses: %w", err)
	}

	return bizs, nil
}

func (h *handler) syncUnassignedAgentNetworkUnitByBiz(
	rCtx server.IContext,
	bizID int64,
	result *types.NodeAgentAssignUnitResult,
) error {

	hosts, err := h.listUnassignedAgentHostsByBiz(rCtx, bizID)
	if err != nil {
		return err
	}

	for offset := 0; offset < len(hosts); offset += syncUnassignedAgentNetworkUnitPageSize {
		end := min(offset+syncUnassignedAgentNetworkUnitPageSize, len(hosts))

		if err := h.assignRecommendedNetworkUnits(rCtx, hosts[offset:end], result); err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) listUnassignedAgentHostsByBiz(rCtx server.IContext, bizID int64) ([]*types.Host, error) {
	condition := &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			BizID: []int64{bizID},
		},
		DynamicExactInclude: &types.HostDynamicExactFields{
			NodeRole:      []types.NodeRole{types.NodeRoleAgent},
			NetworkUnitID: []int64{unassignedAgentNetworkUnitID},
		},
	}

	hosts := make([]*types.Host, 0)
	for offset := 0; ; offset += syncUnassignedAgentNetworkUnitPageSize {
		page := types.Page{Offset: offset, Limit: syncUnassignedAgentNetworkUnitPageSize}
		pageHosts, total, err := h.storageTopo.ListHost(rCtx, page, condition)
		if err != nil {
			return nil, fmt.Errorf("failed to list unassigned agent hosts for biz-id(%d): %w", bizID, err)
		}

		if len(pageHosts) == 0 {
			break
		}

		hosts = append(hosts, pageHosts...)
		if int64(offset+len(pageHosts)) >= total {
			break
		}
	}

	return hosts, nil
}

func (h *handler) assignRecommendedNetworkUnits(
	rCtx server.IContext,
	hosts []*types.Host,
	result *types.NodeAgentAssignUnitResult,
) error {

	candidates, items := buildNetworkUnitRecommendationItems(hosts, result)
	if len(items) == 0 {
		return nil
	}

	recommendations, err := h.storageTopo.RecommendNetworkUnitByNetworkSegment(rCtx, items...)
	if err != nil {
		return fmt.Errorf("failed to recommend network unit by segment: %w", err)
	}

	if len(recommendations) != len(items) {
		return fmt.Errorf("network unit recommendation result count mismatch: got %d, want %d", len(recommendations), len(items))
	}

	hostIDsByNetworkUnitID := make(map[int64][]int64)
	for idx, recommendation := range recommendations {
		candidate := candidates[idx]
		if recommendation == nil {
			result.FailedCount++
			result.FailedReasons = append(result.FailedReasons,
				fmt.Sprintf("host-id(%d) recommend networkunit failed: empty recommendation result", candidate.HostID))

			continue
		}

		if recommendation.NetworkUnitID < 0 {
			result.FailedCount++
			result.FailedReasons = append(result.FailedReasons,
				fmt.Sprintf("host-id(%d) recommend networkunit failed: %s", candidate.HostID, recommendation.Message))

			continue
		}

		hostIDsByNetworkUnitID[recommendation.NetworkUnitID] = append(
			hostIDsByNetworkUnitID[recommendation.NetworkUnitID], candidate.HostID)
	}

	return h.assignNetworkUnits(rCtx, hostIDsByNetworkUnitID, result)
}

func buildNetworkUnitRecommendationItems(
	hosts []*types.Host,
	result *types.NodeAgentAssignUnitResult,
) ([]*types.Host, []*types.NetworkUnitSegmentRecommendationItem) {

	candidates := make([]*types.Host, 0, len(hosts))
	items := make([]*types.NetworkUnitSegmentRecommendationItem, 0, len(hosts))

	for _, host := range hosts {
		if host == nil {
			result.FailedCount++
			result.FailedReasons = append(result.FailedReasons, "empty host data")

			continue
		}

		if host.Static == nil {
			result.FailedCount++
			result.FailedReasons = append(result.FailedReasons,
				fmt.Sprintf("host-id(%d) has no static data", host.HostID))

			continue
		}

		if len(host.Static.InnerIPList) == 0 {
			result.FailedCount++
			result.FailedReasons = append(result.FailedReasons,
				fmt.Sprintf("host-id(%d) has no inner ip", host.HostID))

			continue
		}

		candidates = append(candidates, host)
		items = append(items, &types.NetworkUnitSegmentRecommendationItem{
			NetworkAreaID: host.Static.NetworkAreaID,
			IP:            host.Static.InnerIPList[0],
		})
	}

	return candidates, items
}

func (h *handler) assignNetworkUnits(
	rCtx server.IContext,
	hostIDsByNetworkUnitID map[int64][]int64,
	result *types.NodeAgentAssignUnitResult,
) error {

	networkUnitIDs := conv.MapKeyToSlice(hostIDsByNetworkUnitID)
	for _, networkUnitID := range networkUnitIDs {
		assignResult, err := h.nodeMgrIface.AssignAgentNetworkUnit(rCtx, types.NodeAgentAssignUnitParam{
			HostIDs:       hostIDsByNetworkUnitID[networkUnitID],
			NetworkUnitID: networkUnitID,
		})
		if err != nil {
			return fmt.Errorf("failed to assign networkunit-id(%d): %w", networkUnitID, err)
		}

		mergeNodeAgentAssignUnitResult(result, assignResult)
	}

	return nil
}

func mergeNodeAgentAssignUnitResult(dst *types.NodeAgentAssignUnitResult, src *types.NodeAgentAssignUnitResult) {
	if src == nil {
		return
	}

	dst.SuccessCount += src.SuccessCount
	dst.FailedCount += src.FailedCount
	dst.FailedReasons = append(dst.FailedReasons, src.FailedReasons...)
}
