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
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
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

	if err := validateTopoNetworkUnitCreateReq(req); err != nil {
		h.logger.Errorf("failed to create networkunit, failed to validate request body. err: %v", err)
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

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// creates networkunit.
	networkUnitID, err := h.storage.CreateNetworkUnit(
		sCtx,
		&types.NetworkUnit{
			TenantID:      ctx.TenantID,
			NetworkAreaID: networkAreaID,
			Name:          req.GetBkNetworkunitName(),
			Links:         convertLinks(req.GetLinks()),
		},
		convertAccssPoints(ctx, networkAreaID, req.GetAccessPoints())...)
	if err != nil {
		h.logger.Errorf("failed to create networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := &proto.TopoNetworkUnitCreateResp_Data{BkNetworkunitId: new(int64)}
	*resp.BkNetworkunitId = networkUnitID

	return resp, nil
}

// UpdateNetworkUnit updates networkunit.
func (h *handler) UpdateNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkUnitUpdateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to update networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := validateTopoNetworkUnitUpdateReq(req); err != nil {
		h.logger.Errorf("failed to update networkunit, failed to validate request body. err: %v", err)
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

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	err = h.storage.UpdateNetworkUnit(
		sCtx,
		&types.NetworkUnit{
			TenantID:      ctx.TenantID,
			NetworkAreaID: networkAreaID,
			ID:            req.GetBkNetworkunitId(),
			Name:          req.GetBkNetworkunitName(),
			Links:         convertLinks(req.GetLinks()),
		},
		convertAccssPoints(ctx, networkAreaID, req.GetAccessPoints())...)
	if err != nil {
		h.logger.Errorf("failed to update networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := &proto.TopoNetworkUnitUpdateResp_Data{BkNetworkunitId: new(int64)}
	*resp.BkNetworkunitId = req.GetBkNetworkunitId()

	return resp, nil
}

// GetNetworkUnit gets an existing networkunit.
func (h *handler) GetNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkUnitGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to get networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := validateTopoNetworkUnitGetReq(req); err != nil {
		h.logger.Errorf("failed to get networkunit, failed to validate request body. err: %v", err)
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
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// get accesspoints.
	accessPoints := make([]*types.AccessPoint, 0)
	if len(networkUnit.AccessPoints) > 0 {
		accessPoints, _, err = h.storage.ListAccessPoint(
			sCtx,
			types.Page{Offset: 0, Limit: len(networkUnit.AccessPoints)},
			topo.AccessPointCondition{
				Type: types.ConditionTypeExactInclude,
				Exact: &topo.AccessPointExactFields{
					AccessPointID: networkUnit.AccessPoints,
					NetworkAreaID: []int64{networkUnit.NetworkAreaID},
				},
			})
		if err != nil {
			h.logger.Errorf("failed to get networkunit. failed to list accesspoints. err: %v", err)
			return nil, errf.ErrWrap(errf.InvalidParameter, err)
		}
	}

	data := newEmptyNetworkUnit()
	*data.TenantId = networkUnit.TenantID
	*data.BkNetworkunitId = networkUnit.ID
	*data.BkNetworkunitName = networkUnit.Name
	*data.BkNetworkareaId = networkUnit.NetworkAreaID
	data.AccessPoints = make([]*proto.AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		protoAccessPoint := newEmptyAccessPoint()

		*protoAccessPoint.TenantId = accessPoint.TenantID
		*protoAccessPoint.AccesspointId = accessPoint.ID
		*protoAccessPoint.AccesspointName = accessPoint.Name
		protoAccessPoint.Endpoints = &proto.AccessPoint_Endpoints{
			Cluster: accessPoint.Endpoints.Cluster,
			File:    accessPoint.Endpoints.File,
			Data:    accessPoint.Endpoints.Data,
		}

		data.AccessPoints[idx] = protoAccessPoint
	}
	data.Links = convertLinksToProto(networkUnit.Links)

	return data, nil
}

// ListNetworkUnit lists network units.
func (h *handler) ListNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkUnitListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to list networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := validateTopoNetworkUnitListReq(req); err != nil {
		h.logger.Errorf("failed to list networkunit, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list networkunit, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkAreas, num, err := h.storage.ListNetworkUnit(
		sCtx,
		generatePage(req.GetPage(), maxNetworkUnitLimit),
		generateNetworkUnitConditions(req))
	if err != nil {
		h.logger.Errorf("failed to list networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	items := make([]*proto.NetworkUnit, len(networkAreas))
	for idx, networkUnit := range networkAreas {
		item := newEmptyNetworkUnit()
		*item.TenantId = networkUnit.TenantID
		*item.BkNetworkunitId = networkUnit.ID
		*item.BkNetworkunitName = networkUnit.Name
		*item.BkNetworkareaId = networkUnit.NetworkAreaID
		item.Links = convertLinksToProto(networkUnit.Links)

		items[idx] = item
	}

	resp := &proto.TopoNetworkUnitListResp_Data{
		Total: num,
		Items: items,
	}

	return resp, nil
}

// DeleteNetworkUnit deletes an existing network-unit.
func (h *handler) DeleteNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkUnitDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to delete networkunit, failed to decode request body. err: %v", err)
		return nil, err
	}

	if err := validateTopoNetworkUnitDeleteReq(req); err != nil {
		h.logger.Errorf("failed to delete networkunit, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to delete networkunit, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.storage.DeleteManyNetworkUnit(sCtx, req.GetBkNetworkunitId()); err != nil {
		h.logger.Errorf("failed to delete networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	resp := &proto.TopoNetworkAreaDeleteResp_Data{BkNetworkareaId: new(int64)}
	*resp.BkNetworkareaId = req.GetBkNetworkunitId()

	return resp, nil
}

func convertAccssPoints(
	ctx *rest.Context,
	networkAreaID int64,
	accessPoints []*proto.AccessPoint) []*types.AccessPoint {

	data := make([]*types.AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		data[idx] = &types.AccessPoint{
			TenantID:      ctx.TenantID,
			NetworkAreaID: networkAreaID,
			Name:          accessPoint.GetAccesspointName(),
		}

		if endpoints := accessPoint.GetEndpoints(); endpoints != nil {
			data[idx].Endpoints.Cluster = endpoints.GetCluster()
			data[idx].Endpoints.File = endpoints.GetFile()
			data[idx].Endpoints.Data = endpoints.GetData()
		}
	}

	return data
}

func convertLinks(links *proto.Links) types.Links {
	data := types.Links{}
	if links != nil {
		if reqLink := links.GetCluster(); reqLink != nil {
			data.Cluster = &types.Link{AccessPointID: reqLink.GetAccesspointId()}
		}
		if reqLink := links.GetFile(); reqLink != nil {
			data.File = &types.Link{AccessPointID: reqLink.GetAccesspointId()}
		}
		if reqLink := links.GetData(); reqLink != nil {
			data.Data = &types.Link{AccessPointID: reqLink.GetAccesspointId()}
		}
	}

	return data
}

func convertLinksToProto(links types.Links) *proto.Links {
	data := &proto.Links{
		Cluster: &proto.Link{},
		File:    &proto.Link{},
		Data:    &proto.Link{},
	}

	if links.Cluster != nil {
		data.Cluster = &proto.Link{
			AccesspointId: links.Cluster.AccessPointID,
		}
	}
	if links.File != nil {
		data.File = &proto.Link{
			AccesspointId: links.File.AccessPointID,
		}
	}
	if links.Data != nil {
		data.Data = &proto.Link{
			AccesspointId: links.Data.AccessPointID,
		}
	}

	return data
}

func validateTopoNetworkUnitCreateReq(req *proto.TopoNetworkUnitCreateReq) error {
	if req.GetBkNetworkunitName() == "" {
		return errors.New("bk_networkunit_name is required")
	}

	if req.GetBkNetworkareaId() < 0 {
		return errors.New("bk_networkarea_id is required")
	}

	return nil
}

func validateTopoNetworkUnitUpdateReq(req *proto.TopoNetworkUnitUpdateReq) error {
	if req.GetBkNetworkunitName() == "" {
		return errors.New("bk_networkunit_name is required")
	}

	return nil
}

func validateTopoNetworkUnitGetReq(req *proto.TopoNetworkUnitGetReq) error {
	if req.GetBkNetworkunitId() < 0 {
		return errors.New("bk_networkunit_id is invalid")
	}

	return nil
}

func validateTopoNetworkUnitListReq(req *proto.TopoNetworkUnitListReq) error {
	if err := validateTopoPage(req.GetPage()); err != nil {
		return err
	}

	return nil
}

func validateTopoNetworkUnitDeleteReq(req *proto.TopoNetworkUnitDeleteReq) error {
	if req.GetBkNetworkunitId() < 0 {
		return errors.New("bk_networkunit_id is invalid")
	}

	return nil
}

func generateNetworkUnitConditions(req *proto.TopoNetworkUnitListReq) topo.NetworkUnitCondition {
	// exact conditions.
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		conditions := topo.NetworkUnitCondition{
			Type: types.ConditionTypeExactInclude,
		}
		conditions.Exact = &topo.NetworkUnitExactFields{
			NetworkUnitID: exactCond.GetBkNetworkunitId(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
		}

		return conditions
	}

	// default empty conditions.
	return topo.NetworkUnitCondition{
		Type: types.ConditionTypeExactInclude,
	}
}

func newEmptyNetworkUnit() *proto.NetworkUnit {
	return &proto.NetworkUnit{
		TenantId:          new(string),
		BkNetworkunitId:   new(int64),
		BkNetworkunitName: new(string),
		BkNetworkareaId:   new(int64),
		AccessPoints:      make([]*proto.AccessPoint, 0),
		Links: &proto.Links{
			Cluster: &proto.Link{},
			File:    &proto.Link{},
			Data:    &proto.Link{},
		},
	}
}

func newEmptyAccessPoint() *proto.AccessPoint {
	return &proto.AccessPoint{
		TenantId:        new(string),
		AccesspointId:   new(int64),
		AccesspointName: new(string),
		Endpoints:       &proto.AccessPoint_Endpoints{},
	}
}
