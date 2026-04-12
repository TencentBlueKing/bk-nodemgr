/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// ListAccessPoint lists accesspoints with page and conditions.
func (h *handler) ListAccessPoint(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoAccessPointListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list accesspoint, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// Extract requested AccessPoint IDs from conditions.
	condition := req.ConvertConditionsToTypes()
	var accessPointIDs []int64
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		accessPointIDs = exactCond.GetAccesspointId()
	}

	// Get authorized AccessPoint IDs and narrow the scope.
	narrowedIDs, scopeIsAny, authErr := h.narrowAuthorizedAccessPointIDs(rCtx, accessPointIDs)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list accesspoint, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}
	condition = narrowAccessPointCondition(condition, narrowedIDs, scopeIsAny)

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountAccessPoint(rCtx, condition)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list accesspoint. failed to count accesspoint")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.TopoAccessPointListResp)
		resp.ConvertAccessPointsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list accesspoint, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	accesspoints, num, err := h.storage.ListAccessPoint(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list accesspoint")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoAccessPointListResp)
	resp.ConvertAccessPointsFromTypes(num, accesspoints)

	return resp.GetData(), nil
}

// ListAccessPointBrief lists accesspoint briefs with page and conditions.
func (h *handler) ListAccessPointBrief(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoAccessPointListBriefReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list accesspoint brief, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition := req.ConvertConditionsToTypes()
	var accessPointIDs []int64
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		accessPointIDs = exactCond.GetAccesspointId()
	}

	narrowedIDs, scopeIsAny, authErr := h.narrowAuthorizedAccessPointIDs(rCtx, accessPointIDs)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list accesspoint brief, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}
	condition = narrowAccessPointCondition(condition, narrowedIDs, scopeIsAny)

	if req.GetOnlyCount() {
		num, err := h.storage.CountAccessPoint(rCtx, condition)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list accesspoint brief. failed to count accesspoint")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.TopoAccessPointListBriefResp)
		resp.ConvertAccessPointsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list accesspoint brief, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	accesspoints, num, err := h.storage.ListAccessPoint(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list accesspoint brief")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoAccessPointListBriefResp)
	resp.ConvertAccessPointsFromTypes(num, accesspoints)

	return resp.GetData(), nil
}
