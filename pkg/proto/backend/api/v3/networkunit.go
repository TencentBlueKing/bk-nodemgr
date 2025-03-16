/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *TopoNetworkUnitCreateReq) Validate() error {
	if x.GetBkNetworkunitName() == "" {
		return errors.New("bk_networkunit_name is required")
	}

	if x.GetBkNetworkareaId() < 0 {
		return errors.New("bk_networkarea_id is required")
	}

	return nil
}

// ConvertAccssPointsToTypes convert accesspoint from proto to types.
func (x *TopoNetworkUnitCreateReq) ConvertAccssPointsToTypes(
	tenantID string,
	networkAreaID int64) []*types.AccessPoint {

	return convertAccesspointsToTypes(tenantID, networkAreaID, x.GetAccesspoints())
}

// ConvertAccesspointsFromTypes convert accesspoint from types to proto.
func (x *TopoNetworkUnitCreateReq) ConvertAccesspointsFromTypes(accessPoints []*types.AccessPoint) {
	x.Accesspoints = convertAccesspointsFromTypes(accessPoints)
}

// ConvertLinksToTypes convert links from proto to types.
func (x *TopoNetworkUnitCreateReq) ConvertLinksToTypes() types.Links {
	return convertLinksToTypes(x.GetLinks())
}

// ConvertLinksFromTypes convert links from types to proto.
func (x *TopoNetworkUnitCreateReq) ConvertLinksFromTypes(links types.Links) {
	x.Links = convertLinksFromTypes(links)
}

// ConvertNetworkUnitFromTypes convert networkunit from types to proto.
func (x *TopoNetworkUnitCreateResp) ConvertNetworkUnitFromTypes(networkUnitID int64) {
	data := &TopoNetworkUnitCreateResp_Data{BkNetworkunitId: new(int64)}
	*data.BkNetworkunitId = networkUnitID

	x.Data = data
}

// Validate check body.
func (x *TopoNetworkUnitUpdateReq) Validate() error {
	if x.GetBkNetworkunitName() == "" {
		return errors.New("bk_networkunit_name is required")
	}

	return nil
}

// ConvertAccssPointsToTypes convert access points from proto to types.
func (x *TopoNetworkUnitUpdateReq) ConvertAccssPointsToTypes(
	tenantID string,
	networkAreaID int64) []*types.AccessPoint {

	return convertAccesspointsToTypes(tenantID, networkAreaID, x.GetAccesspoints())
}

// ConvertAccesspointsFromTypes convert access points from types to proto.
func (x *TopoNetworkUnitUpdateReq) ConvertAccesspointsFromTypes(accessPoints []*types.AccessPoint) {
	x.Accesspoints = convertAccesspointsFromTypes(accessPoints)
}

// ConvertLinksToTypes convert links from proto to types.
func (x *TopoNetworkUnitUpdateReq) ConvertLinksToTypes() types.Links {
	return convertLinksToTypes(x.GetLinks())
}

// ConvertLinksFromTypes convert links from types to proto.
func (x *TopoNetworkUnitUpdateReq) ConvertLinksFromTypes(links types.Links) {
	x.Links = convertLinksFromTypes(links)
}

// ConvertNetworkUnitFromTypes convert networkunit from types to proto.
func (x *TopoNetworkUnitUpdateResp) ConvertNetworkUnitFromTypes(networkUnitID int64) {
	data := &TopoNetworkUnitUpdateResp_Data{BkNetworkunitId: new(int64)}
	*data.BkNetworkunitId = networkUnitID

	x.Data = data
}

// Validate check body.
func (x *TopoNetworkUnitGetReq) Validate() error {
	if x.GetBkNetworkunitId() < 0 {
		return errors.New("bk_networkunit_id is invalid")
	}

	return nil
}

// ConvertNetworkUnitFromTypes convert networkunit from types to proto.
func (x *TopoNetworkUnitGetResp) ConvertNetworkUnitFromTypes(
	networkUnit *types.NetworkUnit, accessPoints []*types.AccessPoint) {

	data := newEmptyNetworkUnit()
	*data.TenantId = networkUnit.TenantID
	*data.BkNetworkunitId = networkUnit.ID
	*data.BkNetworkunitName = networkUnit.Name
	*data.BkNetworkareaId = networkUnit.NetworkAreaID

	data.Accesspoints = convertAccesspointsFromTypes(accessPoints)
	data.Links = convertLinksFromTypes(networkUnit.Links)

	x.Data = data
}

// ConvertNetworkUnitToTypes convert networkunit from proto to types.
func (x *TopoNetworkUnitGetResp) ConvertNetworkUnitToTypes() (*types.NetworkUnit, map[int64]*types.AccessPoint) {
	data := x.GetData()
	if data == nil {
		return nil, nil
	}

	accessPoints := convertAccesspointsToTypes(data.GetTenantId(), data.GetBkNetworkareaId(), data.GetAccesspoints())
	accessPointsMap := make(map[int64]*types.AccessPoint)
	accessPointIDs := make([]int64, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		accessPointsMap[accessPoint.ID] = accessPoint
		accessPointIDs[idx] = accessPoint.ID
	}

	return &types.NetworkUnit{
		TenantID:      data.GetTenantId(),
		NetworkAreaID: data.GetBkNetworkareaId(),
		ID:            data.GetBkNetworkunitId(),
		Name:          data.GetBkNetworkunitName(),
		AccessPoints:  accessPointIDs,
		Links:         convertLinksToTypes(data.GetLinks()),
	}, accessPointsMap
}

// Validate check body.
func (x *TopoNetworkUnitListReq) Validate() error {
	return validateTopoPage(x.GetPage())
}

// ConvertPageToTypes convert page to types.
func (x *TopoNetworkUnitListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoNetworkUnitListReq) ConvertConditionsToTypes() types.NetworkUnitCondition {
	// exact conditions.
	if exactCond := x.GetExactIncludeConditions(); exactCond != nil {
		conditions := types.NetworkUnitCondition{
			Type: types.ConditionTypeExactInclude,
		}
		conditions.Exact = &types.NetworkUnitExactFields{
			NetworkUnitID: exactCond.GetBkNetworkunitId(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
		}

		return conditions
	}

	// default empty conditions.
	return types.NetworkUnitCondition{
		Type: types.ConditionTypeExactInclude,
	}
}

// ConvertConditionsFromTypes convert conditions from types to proto.
func (x *TopoNetworkUnitListReq) ConvertConditionsFromTypes(condition *types.NetworkUnitCondition) error {
	if condition == nil {
		return nil
	}

	switch condition.Type {
	case types.ConditionTypeExactInclude:
		if condition.Exact != nil {
			x.ExactIncludeConditions = &TopoNetworkUnitListReq_ExactConditions{
				BkNetworkunitId: condition.Exact.NetworkUnitID,
				BkNetworkareaId: condition.Exact.NetworkAreaID,
			}
		}

	default:
		return fmt.Errorf("unknown condition type: %s", condition.Type)
	}

	return nil
}

// ConvertNetworkUnitsFromTypes convert networkunits from types to proto.
func (x *TopoNetworkUnitListResp) ConvertNetworkUnitsFromTypes(total int64, networkUnits []*types.NetworkUnit) {
	items := make([]*NetworkUnitBrief, len(networkUnits))
	for idx, networkUnit := range networkUnits {
		item := newEmptyNetworkUnitBrief()
		*item.TenantId = networkUnit.TenantID
		*item.BkNetworkunitId = networkUnit.ID
		*item.BkNetworkunitName = networkUnit.Name
		*item.BkNetworkareaId = networkUnit.NetworkAreaID
		item.Accesspoints = networkUnit.AccessPoints
		item.Links = convertLinksFromTypes(networkUnit.Links)

		items[idx] = item
	}

	x.Data = &TopoNetworkUnitListResp_Data{
		Total: total,
		Items: items,
	}
}

// ConvertNetworkUnitsToTypes convert networkunits from proto to types.
func (x *TopoNetworkUnitListResp) ConvertNetworkUnitsToTypes() (int64, []*types.NetworkUnit) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.NetworkUnit, len(items))
	for idx, item := range items {
		result[idx] = &types.NetworkUnit{
			TenantID:      item.GetTenantId(),
			NetworkAreaID: item.GetBkNetworkareaId(),
			ID:            item.GetBkNetworkunitId(),
			Name:          item.GetBkNetworkunitName(),
			AccessPoints:  item.GetAccesspoints(),
			Links:         convertLinksToTypes(item.GetLinks()),
		}
	}

	return data.GetTotal(), result
}

// Validate check body.
func (x *TopoNetworkUnitDeleteReq) Validate() error {
	if x.GetBkNetworkunitId() < 0 {
		return errors.New("bk_networkunit_id is invalid")
	}

	return nil
}

// ConvertNetworkUnitFromTypes convert networkunit from types to proto.
func (x *TopoNetworkUnitDeleteResp) ConvertNetworkUnitFromTypes(networkUnitID int64) {
	data := &TopoNetworkUnitDeleteResp_Data{BkNetworkunitId: new(int64)}
	*data.BkNetworkunitId = networkUnitID

	x.Data = data
}

func newEmptyNetworkUnit() *NetworkUnit {
	return &NetworkUnit{
		TenantId:          new(string),
		BkNetworkunitId:   new(int64),
		BkNetworkunitName: new(string),
		BkNetworkareaId:   new(int64),
		Accesspoints:      make([]*AccessPoint, 0),
		Links: &Links{
			Cluster: newEmptyLink(),
			File:    newEmptyLink(),
			Data:    newEmptyLink(),
		},
	}
}

func newEmptyNetworkUnitBrief() *NetworkUnitBrief {
	return &NetworkUnitBrief{
		TenantId:          new(string),
		BkNetworkunitId:   new(int64),
		BkNetworkunitName: new(string),
		BkNetworkareaId:   new(int64),
		Accesspoints:      make([]int64, 0),
		Links: &Links{
			Cluster: newEmptyLink(),
			File:    newEmptyLink(),
			Data:    newEmptyLink(),
		},
	}
}

func newEmptyLink() *Link {
	return &Link{
		BkNetworkareaId: -1,
		BkNetworkunitId: -1,
		AccesspointId:   -1,
	}
}

func newEmptyAccessPoint() *AccessPoint {
	return &AccessPoint{
		TenantId:        new(string),
		AccesspointId:   new(int64),
		AccesspointName: new(string),
		BkNetworkareaId: new(int64),
		Endpoints:       &AccessPoint_Endpoints{},
	}
}

func convertAccesspointsToTypes(
	tenantID string, networkAreaID int64, accessPoints []*AccessPoint) []*types.AccessPoint {

	data := make([]*types.AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		data[idx] = &types.AccessPoint{
			TenantID:      tenantID,
			NetworkAreaID: networkAreaID,
			ID:            accessPoint.GetAccesspointId(),
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

func convertAccesspointsFromTypes(accessPoints []*types.AccessPoint) []*AccessPoint {
	data := make([]*AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		protoAccessPoint := newEmptyAccessPoint()

		*protoAccessPoint.TenantId = accessPoint.TenantID
		*protoAccessPoint.AccesspointId = accessPoint.ID
		*protoAccessPoint.AccesspointName = accessPoint.Name
		*protoAccessPoint.BkNetworkareaId = accessPoint.NetworkAreaID
		protoAccessPoint.Endpoints = &AccessPoint_Endpoints{
			Cluster: accessPoint.Endpoints.Cluster,
			File:    accessPoint.Endpoints.File,
			Data:    accessPoint.Endpoints.Data,
		}

		data[idx] = protoAccessPoint
	}

	return data
}

func convertLinksToTypes(links *Links) types.Links {
	data := types.Links{}
	if links != nil {
		data.Cluster = convertLinkToTypes(links.GetCluster())
		data.File = convertLinkToTypes(links.GetFile())
		data.Data = convertLinkToTypes(links.GetData())
	}

	return data
}

func convertLinksFromTypes(links types.Links) *Links {
	return &Links{
		Cluster: convertLinkFromTypes(links.Cluster),
		File:    convertLinkFromTypes(links.File),
		Data:    convertLinkFromTypes(links.Data),
	}
}

func convertLinkToTypes(link *Link) *types.Link {
	if link == nil {
		return nil
	}

	return &types.Link{
		NetworkAreaID: link.GetBkNetworkareaId(),
		NetworkUnitID: link.GetBkNetworkunitId(),
		AccessPointID: link.GetAccesspointId(),
	}
}

func convertLinkFromTypes(link *types.Link) *Link {
	if link == nil {
		return nil
	}

	return &Link{
		BkNetworkareaId: link.NetworkAreaID,
		BkNetworkunitId: link.NetworkUnitID,
		AccesspointId:   link.AccessPointID,
	}
}
