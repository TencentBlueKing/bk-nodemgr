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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var (
	errDefaultNetworkAreaNotSupported = errors.New("default networkarea is not supported")
	errDefaultNetworkAreaNotFound     = errors.New("networkarea is not found")
	errDefaultNetworkAreaNotEmpty     = errors.New("networkarea is not empty")
)

// CreateNetworkUnit creates a new network-unit.
func (h *handler) CreateNetworkUnit(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkUnitCreateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create networkunit, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if authErr := h.authorizer.Check(rCtx, auth.ActionNetworkUnitCreate, nil); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to create networkunit, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
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

	resource := authRouter.BuildNetworkUnitResources(networkUnitID)[0]
	if err := h.managerRoleGranter.GrantManagerRole(rCtx, rCtx.BKUsername(), resource); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("resource-type", resource.Type, "resource-id", resource.ID,
			"creator", rCtx.BKUsername(), "tenant-id", rCtx.TenantID()).Error("failed to grant resource creator permissions")
	}

	resp := new(protoBackend.TopoNetworkUnitCreateResp)
	resp.ConvertNetworkUnitFromTypes(networkUnitID)

	return resp.GetData(), nil
}

// CreateDefaultNetworkUnits creates empty non-direct network units in
// multiple empty network areas.
func (h *handler) CreateDefaultNetworkUnits(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkUnitCreateDefaultMultiReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error(
			"failed to create default networkunits, failed to decode request body",
		)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	param := req.ConvertToTypes()
	if authErr := h.authorizer.Check(
		rCtx,
		auth.ActionNetworkUnitCreate,
		nil,
	); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error(
			"failed to create default networkunits, permission denied",
		)

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	if authErr := h.authorizer.Check(
		rCtx,
		auth.ActionNetworkUnitView,
		authRouter.BuildNetworkUnitResources(param.Upstream.NetworkUnitID),
	); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error(
			"failed to create default networkunits, upstream networkunit permission denied",
		)

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	if err := h.validateDefaultNetworkUnitUpstream(rCtx, param.Upstream); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error(
			"failed to create default networkunits, invalid upstream",
		)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result := &types.NetworkUnitCreateDefaultMultiResult{
		Items: make([]*types.NetworkUnitCreateDefaultMultiResultItem, 0, len(param.NetworkAreaIDs)),
	}
	for _, networkAreaID := range param.NetworkAreaIDs {
		item := h.createDefaultNetworkUnit(rCtx, param, networkAreaID)
		result.Items = append(result.Items, item)
		if item.Success {
			result.SuccessCount++
			continue
		}

		result.FailedCount++
	}

	resp := new(protoBackend.TopoNetworkUnitCreateDefaultMultiResp)
	resp.ConvertFromTypes(result)

	logger.G.Biz(rCtx).
		With("success-count", result.SuccessCount).
		With("failed-count", result.FailedCount).
		Info("created default networkunits")

	return resp.GetData(), nil
}

func (h *handler) validateDefaultNetworkUnitUpstream(
	rCtx restserver.IContext, upstream types.Link,
) error {

	upstreamNetworkUnit, err := h.storage.GetNetworkUnit(rCtx, upstream.NetworkUnitID)
	if err != nil {
		if errors.Is(err, base.ErrRecordNoFound()) {
			return fmt.Errorf("upstream networkunit-id(%d) not found", upstream.NetworkUnitID)
		}

		return fmt.Errorf("failed to get upstream networkunit(%d): %w", upstream.NetworkUnitID, err)
	}

	if upstreamNetworkUnit == nil {
		return fmt.Errorf("upstream networkunit-id(%d) not found", upstream.NetworkUnitID)
	}

	if upstreamNetworkUnit.NetworkAreaID != upstream.NetworkAreaID {
		return fmt.Errorf(
			"upstream networkarea not matched, networkarea-id(%d), networkunit-id(%d)",
			upstream.NetworkAreaID,
			upstream.NetworkUnitID,
		)
	}

	for _, accessPointID := range upstreamNetworkUnit.AccessPoints {
		if accessPointID == upstream.AccessPointID {
			return nil
		}
	}

	return fmt.Errorf(
		"upstream accesspoint not found, networkunit-id(%d), accesspoint-id(%d)",
		upstream.NetworkUnitID,
		upstream.AccessPointID,
	)
}

func (h *handler) createDefaultNetworkUnit(
	rCtx restserver.IContext,
	param types.NetworkUnitCreateDefaultMultiParam,
	networkAreaID int64,
) *types.NetworkUnitCreateDefaultMultiResultItem {

	result := &types.NetworkUnitCreateDefaultMultiResultItem{
		NetworkAreaID: networkAreaID,
	}

	networkUnitID, err := h.createDefaultNetworkUnitInArea(rCtx, param, networkAreaID)
	if err != nil {
		result.ErrorCode, result.Message = defaultNetworkUnitCreateError(networkAreaID, err)
		logger.G.Biz(rCtx).WithErr(err).
			With("networkarea-id", networkAreaID).
			Error("failed to create default networkunit")

		return result
	}

	result.Success = true
	result.NetworkUnitID = networkUnitID
	result.Message = "network unit created"

	return result
}

// createDefaultNetworkUnitInArea checks the target network area is empty and
// then creates one default non-direct network unit inside it.
func (h *handler) createDefaultNetworkUnitInArea(
	rCtx restserver.IContext,
	param types.NetworkUnitCreateDefaultMultiParam,
	networkAreaID int64,
) (int64, error) {

	networkArea, err := h.storage.GetNetworkArea(rCtx, networkAreaID)
	if err != nil {
		if errors.Is(err, base.ErrRecordNoFound()) {
			return 0, errDefaultNetworkAreaNotFound
		}

		return 0, fmt.Errorf("failed to get networkarea: %w", err)
	}
	if networkArea == nil {
		return 0, errDefaultNetworkAreaNotFound
	}
	// defensive guard: the request Validate already rejects bk_networkarea_id <= 0,
	// while DefaultNetworkAreaID is 0, so this branch is unreachable in practice.
	if networkAreaID == types.DefaultNetworkAreaID {
		return 0, errDefaultNetworkAreaNotSupported
	}

	_, count, err := h.storage.ListNetworkUnit(
		rCtx,
		types.SingleItemPage(),
		&types.NetworkUnitCondition{
			ExactInclude: &types.NetworkUnitExactFields{
				NetworkAreaID: []int64{networkAreaID},
			},
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to count networkunits: %w", err)
	}
	if count > 0 {
		return 0, errDefaultNetworkAreaNotEmpty
	}

	networkUnit := &types.NetworkUnit{
		TenantID:      rCtx.TenantID(),
		NetworkAreaID: networkAreaID,
		Name:          param.Name,
		IsDirect:      false,
		Links:         defaultNetworkUnitLinks(param.Upstream),
		Generation:    types.Generation2,
	}
	networkUnitID, accessPointResult, err := h.storage.CreateNetworkUnit(rCtx, networkUnit)
	if err != nil {
		return 0, fmt.Errorf("failed to create networkunit: %w", err)
	}

	h.recordNetworkUnitCreateEvents(
		rCtx,
		networkArea,
		networkUnitID,
		param.Name,
		accessPointResult.Created,
	)

	return networkUnitID, nil
}

func defaultNetworkUnitLinks(upstream types.Link) types.Links {
	cluster := upstream
	file := upstream
	data := upstream

	return types.Links{
		Cluster: &cluster,
		File:    &file,
		Data:    &data,
	}
}

func defaultNetworkUnitCreateError(networkAreaID int64, err error) (string, string) {
	switch {
	case errors.Is(err, errDefaultNetworkAreaNotSupported):
		return "default_networkarea_not_supported", "default networkarea is not supported"
	case errors.Is(err, errDefaultNetworkAreaNotFound):
		return "networkarea_not_found", fmt.Sprintf("networkarea-id(%d) not found", networkAreaID)
	case errors.Is(err, errDefaultNetworkAreaNotEmpty):
		return "networkarea_not_empty", fmt.Sprintf("networkarea-id(%d) is not empty", networkAreaID)
	default:
		return "networkunit_create_failed", fmt.Sprintf(
			"failed to create network unit under networkarea-id(%d)",
			networkAreaID,
		)
	}
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

	networkunit := req.GetNetworkunit()
	if authErr := h.authorizer.Check(rCtx, auth.ActionNetworkUnitEdit,
		authRouter.BuildNetworkUnitResources([]int64{networkunit.GetBkNetworkunitId()}...)); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to update networkunit, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	// check if networkarea exists.
	networkAreaID := networkunit.GetBkNetworkareaId()
	networkArea, err := h.storage.GetNetworkArea(rCtx, networkAreaID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("networkarea-id", networkAreaID).Error("failed to update networkunit, failed to get networkarea")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	networkUnitID := networkunit.GetBkNetworkunitId()
	networkUnitName := networkunit.GetBkNetworkunitName()

	// updates networkunit.
	networkUnit := &types.NetworkUnit{
		TenantID:           rCtx.TenantID(),
		NetworkAreaID:      networkAreaID,
		ID:                 networkUnitID,
		Name:               networkunit.GetBkNetworkunitName(),
		IsDirect:           networkunit.GetIsDirect(),
		DirectEndpoints:    req.ConvertDirectEndpointsToTypes(),
		CustomDeployConfig: req.ConvertCustomDeployConfigToTypes(),
	}
	if !networkUnit.IsDirect {
		networkUnit.Links = req.ConvertLinksToTypes()
	}

	accessPointResult, err := h.storage.UpdateNetworkUnit(
		rCtx,
		req.ConvertFieldsToTypes(),
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

	if authErr := h.authorizer.Check(rCtx, auth.ActionNetworkUnitView,
		authRouter.BuildNetworkUnitResources([]int64{req.GetBkNetworkunitId()}...)); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to get networkunit, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
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

	condition := req.ConvertConditionsToTypes()
	var unitIDs []int64
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		unitIDs = exactCond.GetBkNetworkunitId()
	}
	narrowedIDs, scopeIsAny, authErr := h.narrowAuthorizedNetworkUnitIDs(rCtx, unitIDs)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list networkunit, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}
	condition = narrowNetworkUnitCondition(condition, narrowedIDs, scopeIsAny)
	if !scopeIsAny && len(narrowedIDs) == 0 {
		resp := new(protoBackend.TopoNetworkUnitListResp)
		resp.ConvertNetworkUnitsFromTypes(0, nil)

		return resp.GetData(), nil
	}

	networkUnits, num, err := h.storage.ListNetworkUnit(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkunit")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoNetworkUnitListResp)
	resp.ConvertNetworkUnitsFromTypes(num, networkUnits)

	return resp.GetData(), nil
}

// ListNetworkUnitBrief lists network unit briefs.
func (h *handler) ListNetworkUnitBrief(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkUnitListBriefReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkunit brief, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkunit brief, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition := req.ConvertConditionsToTypes()

	networkUnits, num, err := h.storage.ListNetworkUnit(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkunit brief")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoNetworkUnitListBriefResp)
	resp.ConvertNetworkUnitBriefsFromTypes(num, networkUnits)

	return resp.GetData(), nil
}

// GetNetworkUnitDistributionByNetworkAreaID get networkunit distribution by network area id.
func (h *handler) GetNetworkUnitDistributionByNetworkAreaID(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoGetNetworkUnitDistributionByNetworkAreaIDReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get networkunit distribution by network area id, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition := req.ConvertConditionsToTypes()

	result, err := h.storage.GetNetworkUnitDistributionByNetworkAreaID(rCtx, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			Error("failed to get networkunit distribution by network area id. failed to get networkunit distribution by network area id fields: %v", err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoGetNetworkUnitDistributionByNetworkAreaIDResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// RecommendNetworkUnitByNetworkSegment recommends network units by network segment.
// nolint: gocognit
func (h *handler) RecommendNetworkUnitByNetworkSegment(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoRecommendNetworkUnitByNetworkSegmentReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to recommend networkunit by network segment, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	items := req.ConvertItemsToTypes()
	requestedAreaIDs := make([]int64, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		requestedAreaIDs = append(requestedAreaIDs, item.NetworkAreaID)
	}

	narrowedIDs, scopeIsAny, authErr := h.narrowAuthorizedNetworkAreaIDs(rCtx, requestedAreaIDs)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to recommend networkunit by network segment, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	if !scopeIsAny {
		authorizedAreaMap := make(map[int64]struct{}, len(narrowedIDs))
		for _, id := range narrowedIDs {
			authorizedAreaMap[id] = struct{}{}
		}
		for _, item := range items {
			if item == nil {
				continue
			}
			if _, ok := authorizedAreaMap[item.NetworkAreaID]; !ok {
				authErr = h.authorizer.Check(rCtx, auth.ActionNetworkAreaView, buildNetworkAreaResources(item.NetworkAreaID))
				if authErr != nil {
					logger.G.Biz(rCtx).WithErr(authErr).Error("failed to recommend networkunit by network segment, permission denied")
					return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
				}
			}
		}
	}

	results, err := h.storage.RecommendNetworkUnitByNetworkSegment(rCtx, items...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to recommend networkunit by network segment")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoRecommendNetworkUnitByNetworkSegmentResp)
	resp.ConvertResultsFromTypes(results)

	return resp.GetData(), nil
}

// DeleteNetworkUnit deletes an existing network-unit.
func (h *handler) DeleteNetworkUnit(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkUnitDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete networkunit, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if authErr := h.authorizer.Check(rCtx, auth.ActionNetworkUnitDelete,
		authRouter.BuildNetworkUnitResources([]int64{req.GetBkNetworkunitId()}...)); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to delete networkunit, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	networkUnitID := req.GetBkNetworkunitId()

	networkUnit, err := h.storage.GetNetworkUnit(rCtx, networkUnitID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete networkunit, failed to get networkunit")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := h.storage.DeleteManyNetworkUnit(rCtx, networkUnitID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete networkunit")
		if errors.Is(err, topoStg.ErrNetworkUnitDeleteProtected) {
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordNetworkUnitDeleteEvent(rCtx, networkUnit.NetworkAreaID, networkUnitID, networkUnit.Name)

	logger.G.Biz(rCtx).With("networkunit-id", networkUnitID).Info("deleted networkunit")

	resp := new(protoBackend.TopoNetworkUnitDeleteResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkunitId())

	return resp.GetData(), nil
}
