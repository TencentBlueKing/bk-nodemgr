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
	"time"

	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListEvent lists events with page and conditions.
func (h *handler) ListEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoEventListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	cond, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, failed to convert conditions to types")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountTopoEvent(
			rCtx,
			cond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, failed to count event")

			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.TopoEventListResp)
		resp.ConvertTopoEventsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, invalid page info")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	events, num, err := h.storage.ListTopoEvent(rCtx, page, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoEventListResp)
	resp.ConvertTopoEventsFromTypes(num, events)

	return resp.GetData(), nil
}

// DistinctEvent distincts events with conditions.
// nolint: dupl
func (h *handler) DistinctEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoEventDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct topoevent, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	cond, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct topoevent, failed to convert conditions to types")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.storage.DistinctTopoEvent(
		rCtx,
		types.NewTopoEventDistinctRequestAllSet(),
		cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct topoevent, failed to distinct host fields")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoEventDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

func (h *handler) recordTopoEvents(rCtx restserver.IContext, events ...*types.TopoEvent) {
	err := h.goAsyncPool.Run(
		rCtx,
		func(nCtx contextx.IContext) error {
			return h.storage.CreateManyTopoEvent(nCtx, events...)
		},
		goasync.WithName("record_topo_event"),
	)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to submit topo event recording task")
	}
}

func (h *handler) recordNetworkUnitCreateEvents(
	rCtx restserver.IContext,
	networkArea *types.NetworkArea,
	networkUnitID int64,
	networkUnitName string,
	createdAccessPoints []*types.AccessPoint,
) {

	events := make([]*types.TopoEvent, len(createdAccessPoints)+1)
	for idx, accessPoint := range createdAccessPoints {
		events[idx] = &types.TopoEvent{
			TenantID:        rCtx.TenantID(),
			Type:            types.TopoEventAccessPointCreate,
			NetworkAreaID:   networkArea.ID,
			NetworkAreaName: networkArea.Name,
			NetworkUnitID:   networkUnitID,
			NetworkUnitName: networkUnitName,
			AccessPointID:   accessPoint.ID,
			AccessPointName: accessPoint.Name,
			OperateTime:     time.Now(),
			Operator:        rCtx.BKUsername(),
		}
	}

	events[len(createdAccessPoints)] = &types.TopoEvent{
		TenantID:        rCtx.TenantID(),
		Type:            types.TopoEventNetworkUnitCreate,
		NetworkAreaID:   networkArea.ID,
		NetworkAreaName: networkArea.Name,
		NetworkUnitID:   networkUnitID,
		NetworkUnitName: networkUnitName,
		OperateTime:     time.Now(),
		Operator:        rCtx.BKUsername(),
	}

	h.recordTopoEvents(rCtx, events...)
}

func (h *handler) recordNetworkUnitUpdateEvents(
	rCtx restserver.IContext,
	networkArea *types.NetworkArea,
	networkUnitID int64,
	networkUnitName string,
	accessPointResult *topoStg.AccessPointResult,
) {

	events := make([]*types.TopoEvent, 0, len(accessPointResult.Created)+len(accessPointResult.Updated)+1)
	for _, accessPoint := range accessPointResult.Created {
		events = append(events, &types.TopoEvent{
			TenantID:        rCtx.TenantID(),
			Type:            types.TopoEventAccessPointCreate,
			NetworkAreaID:   networkArea.ID,
			NetworkAreaName: networkArea.Name,
			NetworkUnitID:   networkUnitID,
			NetworkUnitName: networkUnitName,
			AccessPointID:   accessPoint.ID,
			AccessPointName: accessPoint.Name,
			OperateTime:     time.Now(),
			Operator:        rCtx.BKUsername(),
		})
	}
	for _, accessPoint := range accessPointResult.Updated {
		events = append(events, &types.TopoEvent{
			TenantID:        rCtx.TenantID(),
			Type:            types.TopoEventAccessPointUpdate,
			NetworkAreaID:   networkArea.ID,
			NetworkAreaName: networkArea.Name,
			NetworkUnitID:   networkUnitID,
			NetworkUnitName: networkUnitName,
			AccessPointID:   accessPoint.ID,
			AccessPointName: accessPoint.Name,
			OperateTime:     time.Now(),
			Operator:        rCtx.BKUsername(),
		})
	}

	events = append(events, &types.TopoEvent{
		TenantID:        rCtx.TenantID(),
		Type:            types.TopoEventNetworkUnitUpdate,
		NetworkAreaID:   networkArea.ID,
		NetworkAreaName: networkArea.Name,
		NetworkUnitID:   networkUnitID,
		NetworkUnitName: networkUnitName,
		OperateTime:     time.Now(),
		Operator:        rCtx.BKUsername(),
	})

	h.recordTopoEvents(rCtx, events...)
}

func (h *handler) recordNetworkUnitDeleteEvent(
	rCtx restserver.IContext,
	networkAreaID int64,
	networkUnitID int64,
	networkUnitName string,
) {

	tenantID := rCtx.TenantID()
	operator := rCtx.BKUsername()
	operateTime := time.Now()

	err := h.goAsyncPool.Run(
		rCtx,
		func(nCtx contextx.IContext) error {
			networkArea, err := h.storage.GetNetworkArea(nCtx, networkAreaID)
			if err != nil {
				logger.G.Sys().WithErr(err).With("networkarea-id", networkAreaID).Error(
					"failed to record network unit delete topo event: get network area")

				return err
			}

			return h.storage.CreateManyTopoEvent(nCtx, &types.TopoEvent{
				TenantID:        tenantID,
				Type:            types.TopoEventNetworkUnitDelete,
				NetworkAreaID:   networkArea.ID,
				NetworkAreaName: networkArea.Name,
				NetworkUnitID:   networkUnitID,
				NetworkUnitName: networkUnitName,
				OperateTime:     operateTime,
				Operator:        operator,
			})
		},
		goasync.WithName("record_topo_event"),
	)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to submit topo event recording task")
	}
}
