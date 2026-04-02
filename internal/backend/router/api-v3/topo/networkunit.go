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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// CreateNetworkUnit creates a new network-unit.
func (h *handler) CreateNetworkUnit(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkUnitCreateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create networkunit, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// check if networkarea exists.
	networkAreaID := req.GetBkNetworkareaId()
	networkArea, err := h.storage.GetNetworkArea(rCtx, networkAreaID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("networkarea-id", networkAreaID).Error("failed to create networkunit, failed to get networkarea")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// creates networkunit.
	networkUnit := &types.NetworkUnit{
		TenantID:           rCtx.TenantID(),
		NetworkAreaID:      networkAreaID,
		Name:               req.GetBkNetworkunitName(),
		IsDirect:           req.GetIsDirect(),
		DirectEndpoints:    req.ConvertDirectEndpointsToTypes(),
		Generation:         types.Generation(req.GetGeneration()),
		CustomDeployConfig: req.ConvertCustomDeployConfigToTypes(),
	}
	if !networkUnit.IsDirect {
		networkUnit.Links = req.ConvertLinksToTypes()
	}
	networkUnitID, accessPointResult, err := h.storage.CreateNetworkUnit(
		rCtx,
		networkUnit,
		req.ConvertAccssPointsToTypes(rCtx.TenantID(), networkAreaID)...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create networkunit")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// record event.
	h.recordNetworkUnitCreateEvents(rCtx, networkArea, networkUnitID, req.GetBkNetworkunitName(), accessPointResult.Created)

	logger.G.Biz(rCtx).With("networkunit-id", networkUnitID, "created-accesspoints", len(accessPointResult.Created)).Info("created networkunit")

	resp := new(protoBackend.TopoNetworkUnitCreateResp)
	resp.ConvertNetworkUnitFromTypes(networkUnitID)

	return resp.GetData(), nil
}

// UpdateNetworkUnit updates networkunit.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (h *handler) UpdateNetworkUnit(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkUnitUpdateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update networkunit, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// check if networkarea exists.
	networkAreaID := req.GetBkNetworkareaId()
	networkArea, err := h.storage.GetNetworkArea(rCtx, networkAreaID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("networkarea-id", networkAreaID).Error("failed to update networkunit, failed to get networkarea")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	networkUnitID := req.GetBkNetworkunitId()
	networkUnitName := req.GetBkNetworkunitName()

	// updates networkunit.
	networkUnit := &types.NetworkUnit{
		TenantID:           rCtx.TenantID(),
		NetworkAreaID:      networkAreaID,
		ID:                 networkUnitID,
		Name:               req.GetBkNetworkunitName(),
		IsDirect:           req.GetIsDirect(),
		DirectEndpoints:    req.ConvertDirectEndpointsToTypes(),
		Generation:         types.Generation(req.GetGeneration()),
		CustomDeployConfig: req.ConvertCustomDeployConfigToTypes(),
	}
	if !networkUnit.IsDirect {
		networkUnit.Links = req.ConvertLinksToTypes()
	}
	accessPointResult, err := h.storage.UpdateNetworkUnit(
		rCtx,
		networkUnit,
		req.ConvertAccssPointsToTypes(rCtx.TenantID(), networkAreaID)...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update networkunit")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// record event.
	h.recordNetworkUnitUpdateEvents(rCtx, networkArea, networkUnitID, networkUnitName, accessPointResult)

	logger.G.Biz(rCtx).
		With("networkunit-id", networkUnitID,
			"created-accesspoints", len(accessPointResult.Created),
			"updated-accesspoints", len(accessPointResult.Updated)).
		Info("updated networkunit")

	resp := new(protoBackend.TopoNetworkUnitUpdateResp)
	resp.ConvertNetworkUnitFromTypes(networkUnitID)

	return resp.GetData(), nil
}

// GetNetworkUnit gets an existing networkunit.
func (h *handler) GetNetworkUnit(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkUnitGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get networkunit, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// get networkunit.
	networkUnit, err := h.storage.GetNetworkUnit(rCtx, req.GetBkNetworkunitId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get networkunit")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// get accesspoints.
	accessPoints := make([]*types.AccessPoint, 0)
	if len(networkUnit.AccessPoints) > 0 {
		accessPoints, _, err = h.storage.ListAccessPoint(
			rCtx,
			types.Page{Offset: 0, Limit: len(networkUnit.AccessPoints)},
			&types.AccessPointCondition{
				ExactInclude: &types.AccessPointExactFields{
					AccessPointID: networkUnit.AccessPoints,
					NetworkAreaID: []int64{networkUnit.NetworkAreaID},
				},
			})
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to get networkunit. failed to list accesspoints")
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}
	}

	resp := new(protoBackend.TopoNetworkUnitGetResp)
	resp.ConvertNetworkUnitFromTypes(networkUnit, accessPoints)

	return resp.GetData(), nil
}

// ListNetworkUnit lists network units.
func (h *handler) ListNetworkUnit(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkUnitListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkunit, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkunit, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkUnits, num, err := h.storage.ListNetworkUnit(rCtx, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkunit")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoNetworkUnitListResp)
	resp.ConvertNetworkUnitsFromTypes(num, networkUnits)

	return resp.GetData(), nil
}

// DeleteNetworkUnit deletes an existing network-unit.
func (h *handler) DeleteNetworkUnit(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkUnitDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete networkunit, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkUnitID := req.GetBkNetworkunitId()

	networkUnit, err := h.storage.GetNetworkUnit(rCtx, networkUnitID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete networkunit, failed to get networkunit")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := h.storage.DeleteManyNetworkUnit(rCtx, networkUnitID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete networkunit")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordNetworkUnitDeleteEvent(rCtx, networkUnit.NetworkAreaID, networkUnitID, networkUnit.Name)

	logger.G.Biz(rCtx).With("networkunit-id", networkUnitID).Info("deleted networkunit")

	resp := new(protoBackend.TopoNetworkUnitDeleteResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkunitId())

	return resp.GetData(), nil
}
