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
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// not max limit in networkunit.
	// return all data in one request.
	maxNetworkUnitLimit = 0
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
		TenantID:        rCtx.TenantID(),
		NetworkAreaID:   networkAreaID,
		Name:            req.GetBkNetworkunitName(),
		IsDirect:        req.GetIsDirect(),
		DirectEndpoints: req.ConvertDirectEndpointsToTypes(),
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

	// generates access point event.
	events := make([]*types.TopoEvent, len(accessPointResult.Created))
	for idx, accessPoint := range accessPointResult.Created {
		events[idx] = &types.TopoEvent{
			TenantID:        rCtx.TenantID(),
			Type:            types.TopoEventAccessPointCreate,
			NetworkAreaID:   networkAreaID,
			NetworkAreaName: networkArea.Name,
			NetworkUnitID:   networkUnitID,
			NetworkUnitName: req.GetBkNetworkunitName(),
			AccessPointID:   accessPoint.ID,
			AccessPointName: accessPoint.Name,
			OperateTime:     time.Now(),
			Operator:        rCtx.BKUsername(),
		}
	}

	// record event.
	go func() {
		if err := h.storage.CreateManyTopoEvent(contextx.New(context.Background()),
			append(events, &types.TopoEvent{
				TenantID:        rCtx.TenantID(),
				Type:            types.TopoEventNetworkUnitCreate,
				NetworkAreaID:   networkArea.ID,
				NetworkAreaName: networkArea.Name,
				NetworkUnitID:   networkUnitID,
				NetworkUnitName: req.GetBkNetworkunitName(),
				OperateTime:     time.Now(),
				Operator:        rCtx.BKUsername(),
			})...); err != nil {
			logger.G.Biz(rCtx).WithErr(err).With("networkunit-id", networkUnitID).Warn("failed to record topo event in networkunit create")
		}
	}()

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
		TenantID:        rCtx.TenantID(),
		NetworkAreaID:   networkAreaID,
		ID:              networkUnitID,
		Name:            req.GetBkNetworkunitName(),
		IsDirect:        req.GetIsDirect(),
		DirectEndpoints: req.ConvertDirectEndpointsToTypes(),
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

	// generates access point event.
	cEvents := make([]*types.TopoEvent, len(accessPointResult.Created))
	for idx, accessPoint := range accessPointResult.Created {
		cEvents[idx] = &types.TopoEvent{
			TenantID:        rCtx.TenantID(),
			Type:            types.TopoEventAccessPointCreate,
			NetworkAreaID:   networkAreaID,
			NetworkAreaName: networkArea.Name,
			NetworkUnitID:   networkUnitID,
			NetworkUnitName: networkUnitName,
			AccessPointID:   accessPoint.ID,
			AccessPointName: accessPoint.Name,
			OperateTime:     time.Now(),
			Operator:        rCtx.BKUsername(),
		}
	}
	uEvents := make([]*types.TopoEvent, len(accessPointResult.Updated))
	for idx, accessPoint := range accessPointResult.Updated {
		uEvents[idx] = &types.TopoEvent{
			TenantID:        rCtx.TenantID(),
			Type:            types.TopoEventAccessPointUpdate,
			NetworkAreaID:   networkAreaID,
			NetworkAreaName: networkArea.Name,
			NetworkUnitID:   networkUnitID,
			NetworkUnitName: networkUnitName,
			AccessPointID:   accessPoint.ID,
			AccessPointName: accessPoint.Name,
			OperateTime:     time.Now(),
			Operator:        rCtx.BKUsername(),
		}
	}
	events := append(cEvents, uEvents...) // nolint:gocritic

	// record event.
	go func() {
		if err := h.storage.CreateManyTopoEvent(contextx.New(context.Background()),
			append(events, &types.TopoEvent{
				TenantID:        rCtx.TenantID(),
				Type:            types.TopoEventNetworkUnitUpdate,
				NetworkAreaID:   networkArea.ID,
				NetworkAreaName: networkArea.Name,
				NetworkUnitID:   networkUnitID,
				NetworkUnitName: networkUnitName,
				OperateTime:     time.Now(),
				Operator:        rCtx.BKUsername(),
			})...); err != nil {
			logger.G.Biz(rCtx).WithErr(err).With("networkunit-id", networkUnitID).Warn("failed to record topo event in networkunit update")
		}
	}()

	logger.G.Biz(rCtx).
		With("networkunit-id", networkUnitID, "created-accesspoints", len(accessPointResult.Created), "updated-accesspoints", len(accessPointResult.Updated)).
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

	networkUnits, num, err := h.storage.ListNetworkUnit(
		rCtx,
		req.ConvertPageToTypes(maxNetworkUnitLimit),
		req.ConvertConditionsToTypes())
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
	go func() {
		// try get networkarea.
		networkAreaName := ""
		networkArea, err := h.storage.GetNetworkArea(contextx.New(context.Background()), networkUnit.NetworkAreaID)
		if err == nil {
			networkAreaName = networkArea.Name
		}

		if err := h.storage.CreateManyTopoEvent(contextx.New(context.Background()),
			&types.TopoEvent{
				TenantID:        rCtx.TenantID(),
				Type:            types.TopoEventNetworkUnitDelete,
				NetworkAreaID:   networkUnit.NetworkAreaID,
				NetworkAreaName: networkAreaName,
				NetworkUnitID:   networkUnitID,
				NetworkUnitName: networkUnit.Name,
				OperateTime:     time.Now(),
				Operator:        rCtx.BKUsername(),
			}); err != nil {
			logger.G.Biz(rCtx).WithErr(err).With("networkunit-id", networkUnitID).Warn("failed to record topo event in networkunit delete")
		}
	}()

	logger.G.Biz(rCtx).With("networkunit-id", networkUnitID).Info("deleted networkunit")

	resp := new(protoBackend.TopoNetworkUnitDeleteResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkunitId())

	return resp.GetData(), nil
}
