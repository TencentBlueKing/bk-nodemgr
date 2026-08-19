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

package topo

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

type topoEventIDNarrower func(requestedIDs []int64) ([]int64, bool, error)

var (
	errNetworkUnitHistoryViewDeniedByEmptyScope = errors.New("no authorized network unit history scope")
)

func narrowAuthorizedHistoryResourceIDs(
	rCtx restserver.IContext, authorizer auth.IAuthorizer, action auth.Action,
	resourceType types.AuthResourceType, requestedIDs []int64, buildResources func(...int64,
	) []types.AuthResource) ([]int64, bool, error) {

	scope, err := authorizer.ListAuthorizedInstances(rCtx, action, resourceType)
	if err != nil {
		return nil, false, err
	}

	narrowedIDs, scopeIsAny, err := auth.ResolveAuthorizedResourceIDsInt64(scope, requestedIDs, resourceType)
	if err != nil {
		return nil, false, err
	}

	if scopeIsAny {
		return requestedIDs, true, nil
	}

	if len(narrowedIDs) == 0 {
		if checkErr := authorizer.Check(rCtx, action, buildResources(requestedIDs...)); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return narrowedIDs, false, nil
}

func narrowTopoEventCondition(
	condition *types.TopoEventCondition,
	narrowNetworkAreaIDs topoEventIDNarrower,
	narrowNetworkUnitIDs topoEventIDNarrower,
	narrowAccessPointIDs topoEventIDNarrower,
) (*types.TopoEventCondition, error) {

	if condition == nil {
		condition = &types.TopoEventCondition{}
	}
	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.TopoEventExactFields{}
	}

	networkAreaIDs, networkAreaScopeIsAny, err := narrowNetworkAreaIDs(condition.ExactInclude.NetworkAreaID)
	if err != nil {
		return nil, err
	}
	if !networkAreaScopeIsAny {
		condition.ExactInclude.NetworkAreaID = conv.SliceUnique(networkAreaIDs)
	}

	networkUnitIDs, scopeIsAny, err := narrowNetworkUnitIDs(condition.ExactInclude.NetworkUnitID)
	if err != nil {
		return nil, err
	}
	if !scopeIsAny {
		condition.ExactInclude.NetworkUnitID = conv.SliceUnique(networkUnitIDs)
	}

	if len(condition.ExactInclude.AccessPointID) == 0 {
		return condition, nil
	}

	accessPointIDs, accessPointScopeIsAny, err := narrowAccessPointIDs(condition.ExactInclude.AccessPointID)
	if err != nil {
		return nil, err
	}
	if !accessPointScopeIsAny {
		condition.ExactInclude.AccessPointID = conv.SliceUnique(accessPointIDs)
	}

	return condition, nil
}

func (h *handler) narrowAuthorizedNetworkUnitHistoryIDs(
	rCtx restserver.IContext, requestedIDs []int64,
) ([]int64, bool, error) {

	return narrowAuthorizedHistoryResourceIDs(
		rCtx,
		h.authorizer,
		auth.ActionNetworkUnitHistoryView,
		types.AuthResourceTypeNetworkUnit,
		requestedIDs,
		authRouter.BuildNetworkUnitResources)
}

func (h *handler) narrowAuthorizedNetworkAreaHistoryIDs(
	rCtx restserver.IContext, requestedIDs []int64,
) ([]int64, bool, error) {

	return narrowAuthorizedHistoryResourceIDs(rCtx,
		h.authorizer,
		auth.ActionNetworkAreaHistoryView,
		types.AuthResourceTypeNetworkArea,
		requestedIDs,
		buildNetworkAreaResources)
}

func (h *handler) narrowAuthorizedAccessPointHistoryIDs(
	rCtx restserver.IContext, requestedIDs []int64,
) ([]int64, bool, error) {

	// AccessPoint authorization is anchored to NetworkUnit history scope.
	var targetNetworkUnitIDs []int64
	if len(requestedIDs) > 0 {
		var err error
		targetNetworkUnitIDs, err = h.storage.GetNetworkUnitIDsByAccessPoints(rCtx, requestedIDs)
		if err != nil {
			return nil, false, err
		}
		if len(targetNetworkUnitIDs) == 0 {
			return nil, false, errNetworkUnitHistoryViewDeniedByEmptyScope
		}
	}

	authorizedNetworkUnitIDs, scopeIsAny, err := h.narrowAuthorizedNetworkUnitHistoryIDs(rCtx, targetNetworkUnitIDs)
	if err != nil {
		return nil, false, err
	}
	if scopeIsAny {
		return requestedIDs, true, nil
	}
	if len(authorizedNetworkUnitIDs) == 0 {
		return nil, false, errNetworkUnitHistoryViewDeniedByEmptyScope
	}

	networkUnits, err := h.storage.GetNetworkUnitByIDs(rCtx, authorizedNetworkUnitIDs)
	if err != nil {
		return nil, false, err
	}

	authorizedAccessPointIDs := make([]int64, 0)
	for _, unit := range networkUnits {
		if unit != nil && len(unit.AccessPoints) > 0 {
			authorizedAccessPointIDs = append(authorizedAccessPointIDs, unit.AccessPoints...)
		}
	}

	if len(requestedIDs) == 0 {
		return conv.SliceUnique(authorizedAccessPointIDs), false, nil
	}

	narrowedIDs := conv.SliceIntersect(requestedIDs, authorizedAccessPointIDs)
	if len(narrowedIDs) == 0 && len(targetNetworkUnitIDs) > 0 {
		resources := authRouter.BuildNetworkUnitResources(targetNetworkUnitIDs...)
		if checkErr := h.authorizer.Check(rCtx, auth.ActionNetworkUnitHistoryView, resources); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return conv.SliceUnique(narrowedIDs), false, nil
}

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
	cond, err = narrowTopoEventCondition(cond,
		func(requestedIDs []int64) ([]int64, bool, error) {
			return h.narrowAuthorizedNetworkAreaHistoryIDs(rCtx, requestedIDs)
		},
		func(requestedIDs []int64) ([]int64, bool, error) {
			return h.narrowAuthorizedNetworkUnitHistoryIDs(rCtx, requestedIDs)
		},
		func(requestedIDs []int64) ([]int64, bool, error) {
			return h.narrowAuthorizedAccessPointHistoryIDs(rCtx, requestedIDs)
		},
	)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
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
