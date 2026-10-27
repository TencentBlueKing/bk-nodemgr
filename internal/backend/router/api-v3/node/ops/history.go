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

package ops

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// HistoryList lists metadata after narrowing the archived business scope.
func (h *handler) HistoryList(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeOpsHistoryListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops history, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	cond, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops history, invalid conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops history, invalid page")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	bizIDs, scopeIsAny, err := h.narrowAuthorizedBizIDs(rCtx, cond.ExactInclude.BizID, cond.ExactInclude.NodeRole)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops history, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	if !scopeIsAny && len(bizIDs) == 0 {
		resp := new(protoBackend.NodeOpsHistoryListResp)
		resp.ConvertNodeOpsHistoriesFromTypes(0, nil)

		return resp.GetData(), nil
	}

	cond = narrowHistoryConditionByBiz(cond, bizIDs, scopeIsAny)

	if req.GetOnlyCount() {
		total, err := h.daoNodeOpsHistory.CountNodeOpsHistory(rCtx, cond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops history, failed to count history")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.NodeOpsHistoryListResp)
		resp.ConvertNodeOpsHistoriesFromTypes(total, nil)

		return resp.GetData(), nil
	}
	selection := &types.NodeOpsHistoryFieldSelection{
		OperateID:   true,
		Operator:    true,
		OperateTime: true,
		NodeRole:    true,
		BizIDs:      true,
	}
	histories, total, err := h.daoNodeOpsHistory.ListNodeOpsHistory(rCtx, page, selection, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops history")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeOpsHistoryListResp)
	resp.ConvertNodeOpsHistoriesFromTypes(total, histories)

	return resp.GetData(), nil
}

// HistoryDetail returns all snapshots after authorizing every archived business.
func (h *handler) HistoryDetail(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeOpsHistoryDetailReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get node ops history, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	history, err := h.daoNodeOpsHistory.GetNodeOpsHistory(rCtx, req.GetOperateId(),
		&types.NodeOpsHistoryFieldSelection{NodeRole: true, BizIDs: true, Snapshots: true})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get node ops history")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeOpsHistoryDetailResp)
	resp.ConvertSnapshotsFromTypes(nil)
	if history == nil {
		return resp.GetData(), nil
	}

	if err := h.checkHistoryViewPermission(rCtx, history); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get node ops history, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	resp.ConvertSnapshotsFromTypes(history.Snapshots)

	return resp.GetData(), nil
}
