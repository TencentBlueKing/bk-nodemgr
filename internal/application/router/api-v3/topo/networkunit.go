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

	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	maxNetworkUnitLimit = 1000
)

// GetNetworkUnit gets an existing networkunit.
func (h *handler) GetNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(topoNetworkUnitGetReq)
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
	networkUnit, accessPointsMap, err := h.backendHandler.GetNetworkUnit(sCtx, req.GetBkNetworkunitId())
	if err != nil {
		h.logger.Errorf("failed to get networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	data := newEmptyNetworkUnit()
	*data.TenantId = networkUnit.TenantID
	*data.BkNetworkunitId = networkUnit.ID
	*data.BkNetworkunitName = networkUnit.Name
	*data.BkNetworkareaId = networkUnit.NetworkAreaID
	data.AccessPoints = make([]*proto.AccessPoint, len(accessPointsMap))

	idx := 0
	for _, accessPoint := range accessPointsMap {
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
		idx++
	}
	data.Links = convertLinksToProto(networkUnit.Links)

	return data, nil
}

// ListNetworkUnit lists network units.
func (h *handler) ListNetworkUnit(ctx *rest.Context) (interface{}, error) {
	req := new(topoNetworkUnitListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to list networkunit, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list networkunit, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkAreas, num, err := h.backendHandler.ListNetworkUnit(
		sCtx,
		generatePage(req.GetPage(), maxNetworkUnitLimit),
		generateNetworkUnitConditions(req))
	if err != nil {
		h.logger.Errorf("failed to list networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	items := make([]*proto.NetworkUnitBrief, len(networkAreas))
	for idx, networkUnit := range networkAreas {
		item := newEmptyNetworkUnitBrief()
		*item.TenantId = networkUnit.TenantID
		*item.BkNetworkunitId = networkUnit.ID
		*item.BkNetworkunitName = networkUnit.Name
		*item.BkNetworkareaId = networkUnit.NetworkAreaID
		item.AccessPoints = networkUnit.AccessPoints
		item.Links = convertLinksToProto(networkUnit.Links)

		items[idx] = item
	}

	resp := &proto.TopoNetworkUnitListResp_Data{
		Total: num,
		Items: items,
	}

	return resp, nil
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

func generateNetworkUnitConditions(req *topoNetworkUnitListReq) *types.NetworkUnitCondition {
	// exact conditions.
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		conditions := &types.NetworkUnitCondition{
			Type: types.ConditionTypeExactInclude,
		}
		conditions.Exact = &types.NetworkUnitExactFields{
			NetworkUnitID: exactCond.GetBkNetworkunitId(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
		}

		return conditions
	}

	// default empty conditions.
	return nil
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

func newEmptyNetworkUnitBrief() *proto.NetworkUnitBrief {
	return &proto.NetworkUnitBrief{
		TenantId:          new(string),
		BkNetworkunitId:   new(int64),
		BkNetworkunitName: new(string),
		BkNetworkareaId:   new(int64),
		AccessPoints:      make([]int64, 0),
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

type topoNetworkUnitGetReq struct {
	proto.TopoNetworkUnitGetReq
}

// Validate check body.
func (req *topoNetworkUnitGetReq) Validate() error {
	if req.GetBkNetworkunitId() < 0 {
		return errors.New("bk_networkunit_id is invalid")
	}

	return nil
}

type topoNetworkUnitListReq struct {
	proto.TopoNetworkUnitListReq
}

// Validate check body.
func (req *topoNetworkUnitListReq) Validate() error {
	return validateTopoPage(req.GetPage())
}
