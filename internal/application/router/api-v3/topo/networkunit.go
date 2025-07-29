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
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// not max limit in networkunit.
	// return all data in one request.
	maxNetworkUnitLimit = 0
)

// CreateNetworkUnit creates a new network-unit.
func (h *handler) CreateNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.TopoNetworkUnitCreateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// check if networkarea exists.
	networkAreaID := req.GetBkNetworkareaId()
	if _, err := h.backendHandler.GetNetworkArea(ctx, networkAreaID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create networkunit, failed to get networkarea. networkarea-id(%d), err: %v",
			networkAreaID, err)

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// creates networkunit.
	networkUnitID, err := h.backendHandler.CreateNetworkUnit(
		ctx,
		&types.NetworkUnit{
			TenantID:      ctx.TenantID,
			NetworkAreaID: networkAreaID,
			Name:          req.GetBkNetworkunitName(),
			Links:         req.ConvertLinksToTypes(),
		},
		req.ConvertAccssPointsToTypes(ctx.TenantID, networkAreaID)...)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := new(protoApplication.TopoNetworkUnitCreateResp)
	resp.ConvertNetworkUnitFromTypes(networkUnitID)

	return resp.GetData(), nil
}

// UpdateNetworkUnit updates networkunit.
func (h *handler) UpdateNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.TopoNetworkUnitUpdateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to update networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// check if networkarea exists.
	networkAreaID := req.GetBkNetworkareaId()
	if _, err := h.backendHandler.GetNetworkArea(ctx, networkAreaID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to update networkunit, failed to get networkarea. networkarea-id(%d), err: %v",
			networkAreaID, err)

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	err := h.backendHandler.UpdateNetworkUnit(
		ctx,
		&types.NetworkUnit{
			TenantID:      ctx.TenantID,
			NetworkAreaID: networkAreaID,
			ID:            req.GetBkNetworkunitId(),
			Name:          req.GetBkNetworkunitName(),
			Links:         req.ConvertLinksToTypes(),
		},
		req.ConvertAccssPointsToTypes(ctx.TenantID, networkAreaID)...)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to update networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := new(protoApplication.TopoNetworkUnitUpdateResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkunitId())

	return resp.GetData(), nil
}

// GetNetworkUnit gets an existing networkunit.
func (h *handler) GetNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.TopoNetworkUnitGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// get networkunit.
	networkUnit, accessPointsMap, err := h.backendHandler.GetNetworkUnit(ctx, req.GetBkNetworkunitId())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	accessPoints := make([]*types.AccessPoint, len(accessPointsMap))
	idx := 0
	for _, accessPoint := range accessPointsMap {
		accessPoints[idx] = accessPoint
		idx++
	}

	resp := new(protoApplication.TopoNetworkUnitGetResp)
	resp.ConvertNetworkUnitFromTypes(networkUnit, accessPoints)

	return resp.GetData(), nil
}

// ListNetworkUnit lists network units.
func (h *handler) ListNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.TopoNetworkUnitListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkUnits, num, err := h.backendHandler.ListNetworkUnit(
		ctx,
		req.ConvertPageToTypes(maxNetworkUnitLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// get accesspoints.
	idMap := make(map[int64]bool, 0)
	for _, networkUnit := range networkUnits {
		for _, accessPointID := range networkUnit.AccessPoints {
			idMap[accessPointID] = true
		}
	}
	ids := make([]int64, len(idMap))
	index := 0
	for id := range idMap {
		ids[index] = id
		index++
	}

	accessPoints := make([]*types.AccessPoint, 0)
	if len(ids) > 0 {
		accessPoints, _, err = h.backendHandler.ListAccessPoint(
			ctx, types.Page{Limit: len(ids)}, &types.AccessPointCondition{
				ExactInclude: &types.AccessPointExactFields{
					AccessPointID: ids,
				}})
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to list networkunit, failed to list accesspoint. err: %v", err)
			return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
		}
	}

	resp := new(protoApplication.TopoNetworkUnitListResp)
	resp.ConvertNetworkUnitsFromTypes(num, networkUnits, accessPoints)

	return resp.GetData(), nil
}

// DeleteNetworkUnit deletes an existing network-unit.
func (h *handler) DeleteNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.TopoNetworkUnitDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.backendHandler.DeleteNetworkUnit(ctx, req.GetBkNetworkunitId()); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.TopoNetworkUnitDeleteResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkunitId())

	return resp.GetData(), nil
}
