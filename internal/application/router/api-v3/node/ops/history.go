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
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// HistoryList lists authorized operation metadata from the backend.
func (h *handler) HistoryList(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeOpsHistoryListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops history, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops history, invalid conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops history, invalid page")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if req.GetOnlyCount() {
		total, err := h.backendHandler.CountNodeOpsHistory(rCtx, condition)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count node ops history")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}
		resp := new(protoApplication.NodeOpsHistoryListResp)
		resp.ConvertNodeOpsHistoriesFromTypes(total, nil)

		return resp.GetData(), nil
	}

	histories, total, err := h.backendHandler.ListNodeOpsHistory(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node ops history")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.NodeOpsHistoryListResp)
	resp.ConvertNodeOpsHistoriesFromTypes(total, histories)

	return resp.GetData(), nil
}

// HistoryDetail returns all authorized archived snapshots from the backend.
func (h *handler) HistoryDetail(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.NodeOpsHistoryDetailReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get node ops history, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	snapshots, err := h.backendHandler.GetNodeOpsHistory(rCtx, req.GetOperateId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get node ops history")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.NodeOpsHistoryDetailResp)
	resp.ConvertSnapshotsFromTypes(snapshots)

	return resp.GetData(), nil
}
