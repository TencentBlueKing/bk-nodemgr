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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	assignAgentNetworkUnitPermissionBatchSize    = 2000
	assignAgentNetworkUnitPermissionBatchTimeout = 10 * time.Minute
)

// AgentAssignUnit batch-assigns a network unit to unassigned hosts (metadata-only, no remote operations).
func (h *handler) AgentAssignUnit(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.NodeAgentAssignUnitReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkUnit, err := h.storageNetworkUnit.GetNetworkUnit(rCtx, req.GetBkNetworkunitId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit, failed to fetch network unit")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	bizIDs, err := h.getAssignAgentNetworkUnitBizIDs(rCtx, req.GetBkHostId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit, failed to fetch host bizs")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := h.checkAssignAgentNetworkUnitPermission(rCtx, networkUnit, bizIDs); err != nil {
		return nil, err
	}

	result, err := h.nodeMgrIface.AssignAgentNetworkUnit(rCtx, types.NodeAgentAssignUnitParam{
		HostIDs:       req.GetBkHostId(),
		NetworkUnitID: networkUnit.ID,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to assign unit")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
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

func (h *handler) getAssignAgentNetworkUnitBizIDs(rCtx restserver.IContext, hostIDs []int64) ([]int64, error) {
	hostIDs = conv.SliceUnique(hostIDs)
	bizIDMap := make(map[int64]struct{}, len(hostIDs))
	selection := &types.HostFieldSelection{BizID: true}

	err := batchexecutor.Execute(
		rCtx,
		hostIDs,
		func(nCtx contextx.IContext, batch []int64) error {
			hosts, _, err := h.storageHost.ListHostWithFields(
				nCtx,
				types.UnlimitedPage(),
				selection,
				&types.HostCondition{
					StaticExactInclude: &types.HostStaticExactFields{HostID: batch},
				},
			)
			if err != nil {
				return err
			}

			for _, host := range hosts {
				if host == nil || host.Static == nil {
					continue
				}

				bizIDMap[host.Static.BizID] = struct{}{}
			}

			return nil
		},
		batchexecutor.WithBatchSize(assignAgentNetworkUnitPermissionBatchSize),
		batchexecutor.WithTimeout(assignAgentNetworkUnitPermissionBatchTimeout),
	)
	if err != nil {
		return nil, err
	}

	return conv.MapKeyToSlice(bizIDMap), nil
}

func (h *handler) checkAssignAgentNetworkUnitPermission(
	rCtx restserver.IContext, networkUnit *types.NetworkUnit, bizIDs []int64,
) error {

	if authErr := h.authorizer.Check(
		rCtx, auth.ActionNetworkUnitUseForAgent, authRouter.BuildNetworkUnitResources(networkUnit.ID),
	); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to assign unit, networkunit permission denied")
		return resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	if authErr := h.authorizer.Check(rCtx, auth.ActionAgentOperate, authRouter.BuildBizResources(bizIDs...)); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to assign unit, agent operate permission denied")
		return resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	return nil
}
