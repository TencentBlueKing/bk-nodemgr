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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AgentAssignUnitMulti batch-assigns multiple network units to agent hosts.
func (h *handler) AgentAssignUnitMulti(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.NodeAgentAssignUnitMultiReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign multiple units, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkUnitIDs := make([]int64, 0, len(req.GetItems()))
	networkUnits := make(map[int64]*types.NetworkUnit, len(req.GetItems()))
	allHostIDs := make([]int64, 0)
	for _, item := range req.GetItems() {
		if item == nil {
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, fmt.Errorf("item can not be nil"))
		}
		networkUnitIDs = append(networkUnitIDs, item.GetBkNetworkunitId())
		allHostIDs = append(allHostIDs, item.GetBkHostId()...)
	}
	networkUnitIDs = uniqueInt64s(networkUnitIDs)
	networkUnitList, _, err := h.storageNetworkUnit.ListNetworkUnit(rCtx, types.UnlimitedPage(), &types.NetworkUnitCondition{
		ExactInclude: &types.NetworkUnitExactFields{NetworkUnitID: networkUnitIDs},
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign multiple units, failed to fetch network units")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}
	for _, networkUnit := range networkUnitList {
		networkUnits[networkUnit.ID] = networkUnit
	}
	for _, networkUnitID := range networkUnitIDs {
		if _, ok := networkUnits[networkUnitID]; !ok {
			err := fmt.Errorf("networkunit with id %d not found", networkUnitID)
			logger.G.Biz(rCtx).WithErr(err).Error("failed to assign multiple units, network unit not found")

			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}
	}
	bizIDs, err := h.getAssignAgentNetworkUnitBizIDs(rCtx, allHostIDs)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign multiple units, failed to fetch host bizs")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if authErr := h.authorizer.Check(
		rCtx, auth.ActionNetworkUnitUseForAgent, authRouter.BuildNetworkUnitResources(networkUnitIDs...),
	); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to assign multiple units, networkunit permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	if authErr := h.authorizer.Check(rCtx, auth.ActionAgentOperate, authRouter.BuildBizResources(bizIDs...)); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to assign multiple units, agent operate permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	result := &types.NodeAgentAssignUnitResult{FailedReasons: make([]string, 0)}
	for _, item := range req.GetItems() {
		networkUnit := networkUnits[item.GetBkNetworkunitId()]
		assignResult, err := h.nodeMgrIface.AssignAgentNetworkUnit(rCtx, types.NodeAgentAssignUnitParam{
			HostIDs:       item.GetBkHostId(),
			NetworkUnitID: networkUnit.ID,
		})
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).
				With("networkunit-id", networkUnit.ID).
				Error("failed to assign multiple units, failed to assign network unit")
			result.FailedCount += int64(len(item.GetBkHostId()))
			result.FailedReasons = append(result.FailedReasons, fmt.Sprintf(
				"failed to assign networkunit-id(%d): %v", networkUnit.ID, err,
			))

			continue
		}
		mergeAgentAssignUnitResult(result, assignResult)
	}

	resp := &protoBackend.NodeAgentAssignUnitResp{
		Data: &protoBackend.NodeAgentAssignUnitResp_Data{
			SuccessCount:  result.SuccessCount,
			FailedCount:   result.FailedCount,
			FailedReasons: result.FailedReasons,
		},
	}

	return resp.GetData(), nil
}

func uniqueInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}

func mergeAgentAssignUnitResult(dst, src *types.NodeAgentAssignUnitResult) {
	if src == nil {
		return
	}

	dst.SuccessCount += src.SuccessCount
	dst.FailedCount += src.FailedCount
	dst.FailedReasons = append(dst.FailedReasons, src.FailedReasons...)
}
