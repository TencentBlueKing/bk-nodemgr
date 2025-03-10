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
	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	maxNetworkUnitLimit = 1000
)

// CreateNetworkUnit creates a new network-unit.
func (h *handler) CreateNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkUnitCreateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to create networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to create networkunit, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// check if networkarea exists.
	networkAreaID := req.GetBkNetworkareaId()
	if _, err = h.storage.GetNetworkArea(sCtx, networkAreaID); err != nil {
		h.logger.Errorf("failed to create networkunit, failed to get networkarea. networkarea-id(%d), err: %v",
			networkAreaID, err)

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// creates networkunit.
	networkUnitID, err := h.storage.CreateNetworkUnit(
		sCtx,
		&types.NetworkUnit{
			TenantID:      ctx.TenantID,
			NetworkAreaID: networkAreaID,
			Name:          req.GetBkNetworkunitName(),
			Links:         req.ConvertLinksToTypes(),
		},
		req.ConvertAccssPointsToTypes(ctx.TenantID, networkAreaID)...)
	if err != nil {
		h.logger.Errorf("failed to create networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := new(proto.TopoNetworkUnitCreateResp)
	resp.ConvertNetworkUnitFromTypes(networkUnitID)

	return resp.Data, nil
}

// UpdateNetworkUnit updates networkunit.
func (h *handler) UpdateNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkUnitUpdateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to update networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to update networkunit, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// check if networkarea exists.
	networkAreaID := req.GetBkNetworkareaId()
	if _, err = h.storage.GetNetworkArea(sCtx, networkAreaID); err != nil {
		h.logger.Errorf("failed to update networkunit, failed to get networkarea. networkarea-id(%d), err: %v",
			networkAreaID, err)

		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	err = h.storage.UpdateNetworkUnit(
		sCtx,
		&types.NetworkUnit{
			TenantID:      ctx.TenantID,
			NetworkAreaID: networkAreaID,
			ID:            req.GetBkNetworkunitId(),
			Name:          req.GetBkNetworkunitName(),
			Links:         req.ConvertLinksToTypes(),
		},
		req.ConvertAccssPointsToTypes(ctx.TenantID, networkAreaID)...)
	if err != nil {
		h.logger.Errorf("failed to update networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := new(proto.TopoNetworkUnitUpdateResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkunitId())

	return resp.Data, nil
}

// GetNetworkUnit gets an existing networkunit.
func (h *handler) GetNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkUnitGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to get networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to get networkunit, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// get networkunit.
	networkUnit, err := h.storage.GetNetworkUnit(sCtx, req.GetBkNetworkunitId())
	if err != nil {
		h.logger.Errorf("failed to get networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// get accesspoints.
	accessPoints := make([]*types.AccessPoint, 0)
	if len(networkUnit.AccessPoints) > 0 {
		accessPoints, _, err = h.storage.ListAccessPoint(
			sCtx,
			types.Page{Offset: 0, Limit: len(networkUnit.AccessPoints)},
			types.AccessPointCondition{
				Type: types.ConditionTypeExactInclude,
				Exact: &types.AccessPointExactFields{
					AccessPointID: networkUnit.AccessPoints,
					NetworkAreaID: []int64{networkUnit.NetworkAreaID},
				},
			})
		if err != nil {
			h.logger.Errorf("failed to get networkunit. failed to list accesspoints. err: %v", err)
			return nil, errf.ErrWrap(errf.InvalidParameter, err)
		}
	}

	resp := new(proto.TopoNetworkUnitGetResp)
	resp.ConvertNetworkUnitFromTypes(networkUnit, accessPoints)

	return resp.Data, nil
}

// ListNetworkUnit lists network units.
func (h *handler) ListNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkUnitListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to list networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list networkunit, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkUnits, num, err := h.storage.ListNetworkUnit(
		sCtx,
		req.ConvertPageToTypes(maxNetworkUnitLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.Errorf("failed to list networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(proto.TopoNetworkUnitListResp)
	resp.ConvertNetworkUnitsFromTypes(num, networkUnits)

	return resp.Data, nil
}

// DeleteNetworkUnit deletes an existing network-unit.
func (h *handler) DeleteNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkUnitDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to delete networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to delete networkunit, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.storage.DeleteManyNetworkUnit(sCtx, req.GetBkNetworkunitId()); err != nil {
		h.logger.Errorf("failed to delete networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(proto.TopoNetworkUnitDeleteResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkunitId())

	return resp, nil
}
