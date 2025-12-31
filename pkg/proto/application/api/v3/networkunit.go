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

// AutoConvert auto convert.
func (x *TopoNetworkUnitCreateReq) AutoConvert() {
	for _, accesspoint := range x.GetAccesspoints() {
		accesspoint.autoConvert()
	}
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

// ConvertDirectEndpointsToTypes convert endpoints from proto to types.
func (x *TopoNetworkUnitCreateReq) ConvertDirectEndpointsToTypes() *types.Endpoints {
	return convertEndpointToTypes(x.GetDirectEndpoints())
}

// ConvertDirectEndpointsFromTypes convert endpoints from types to proto.
func (x *TopoNetworkUnitCreateReq) ConvertDirectEndpointsFromTypes(endpoints *types.Endpoints) {
	x.DirectEndpoints = convertEndpointFromTypes(endpoints)
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

// AutoConvert auto convert.
func (x *TopoNetworkUnitUpdateReq) AutoConvert() {
	for _, accesspoint := range x.GetAccesspoints() {
		accesspoint.autoConvert()
	}
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

// ConvertDirectEndpointsToTypes convert endpoints from proto to types.
func (x *TopoNetworkUnitUpdateReq) ConvertDirectEndpointsToTypes() *types.Endpoints {
	return convertEndpointToTypes(x.GetDirectEndpoints())
}

// ConvertDirectEndpointsFromTypes convert endpoints from types to proto.
func (x *TopoNetworkUnitUpdateReq) ConvertDirectEndpointsFromTypes(endpoints *types.Endpoints) {
	x.DirectEndpoints = convertEndpointFromTypes(endpoints)
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

// AutoConvert auto convert.
func (x *TopoNetworkUnitGetReq) AutoConvert() {
}

// ConvertNetworkUnitFromTypes convert networkunit from types to proto.
func (x *TopoNetworkUnitGetResp) ConvertNetworkUnitFromTypes(
	networkUnit *types.NetworkUnit, accessPoints []*types.AccessPoint) {

	data := newEmptyNetworkUnit()
	*data.TenantId = networkUnit.TenantID
	*data.BkNetworkunitId = networkUnit.ID
	*data.BkNetworkunitName = networkUnit.Name
	*data.BkNetworkareaId = networkUnit.NetworkAreaID
	*data.IsDirect = networkUnit.IsDirect

	data.Accesspoints = convertAccesspointsFromTypes(accessPoints)
	data.Links = convertLinksFromTypes(networkUnit.Links)
	data.DirectEndpoints = convertEndpointFromTypes(networkUnit.DirectEndpoints)

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
		TenantID:        data.GetTenantId(),
		NetworkAreaID:   data.GetBkNetworkareaId(),
		ID:              data.GetBkNetworkunitId(),
		Name:            data.GetBkNetworkunitName(),
		AccessPoints:    accessPointIDs,
		Links:           convertLinksToTypes(data.GetLinks()),
		IsDirect:        data.GetIsDirect(),
		DirectEndpoints: convertEndpointToTypes(data.GetDirectEndpoints()),
	}, accessPointsMap
}

// Validate check body.
func (x *TopoNetworkUnitListReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert auto convert.
func (x *TopoNetworkUnitListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *TopoNetworkUnitListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoNetworkUnitListReq) ConvertConditionsToTypes() *types.NetworkUnitCondition {
	condition := &types.NetworkUnitCondition{}

	// exact conditions.
	if exactCond := x.GetExactIncludeConditions(); exactCond != nil {
		condition.ExactInclude = &types.NetworkUnitExactFields{
			NetworkUnitID: exactCond.GetBkNetworkunitId(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			IsDirect:      exactCond.GetIsDirect(),
		}
	}

	return condition
}

// ConvertConditionsFromTypes convert conditions from types to proto.
func (x *TopoNetworkUnitListReq) ConvertConditionsFromTypes(condition *types.NetworkUnitCondition) error {
	if condition == nil {
		return nil
	}

	if condition.ExactInclude != nil {
		x.ExactIncludeConditions = &TopoNetworkUnitListReq_ExactConditions{
			BkNetworkunitId: condition.ExactInclude.NetworkUnitID,
			BkNetworkareaId: condition.ExactInclude.NetworkAreaID,
		}
	}

	if condition.FuzzyInclude != nil || condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return fmt.Errorf("fuzzy-include, exact-exclude and fuzzy-exclude not supported")
	}

	return nil
}

// ConvertNetworkUnitsFromTypes convert networkunits from types to proto.
func (x *TopoNetworkUnitListResp) ConvertNetworkUnitsFromTypes(total int64, networkUnits []*types.NetworkUnit, accessPoints []*types.AccessPoint) {
	accessPointMap := make(map[int64]*types.AccessPoint)
	for _, accessPoints := range accessPoints {
		accessPointMap[accessPoints.ID] = accessPoints
	}

	items := make([]*NetworkUnit, len(networkUnits))
	for idx, networkUnit := range networkUnits {
		item := newEmptyNetworkUnit()
		*item.TenantId = networkUnit.TenantID
		*item.BkNetworkunitId = networkUnit.ID
		*item.BkNetworkunitName = networkUnit.Name
		*item.BkNetworkareaId = networkUnit.NetworkAreaID
		item.Links = convertLinksFromTypes(networkUnit.Links)
		*item.IsDirect = networkUnit.IsDirect
		item.DirectEndpoints = convertEndpointFromTypes(networkUnit.DirectEndpoints)

		// convert accesspoints
		for _, accessPointID := range networkUnit.AccessPoints {
			if accessPoint, ok := accessPointMap[accessPointID]; ok {
				item.Accesspoints = append(item.Accesspoints, convertAccessPointFromTypes(accessPoint))
			}
		}

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
		ids := make([]int64, len(item.GetAccesspoints()))
		for subIdx, accessPoint := range item.GetAccesspoints() {
			ids[subIdx] = accessPoint.GetAccesspointId()
		}

		result[idx] = &types.NetworkUnit{
			TenantID:        item.GetTenantId(),
			NetworkAreaID:   item.GetBkNetworkareaId(),
			ID:              item.GetBkNetworkunitId(),
			Name:            item.GetBkNetworkunitName(),
			AccessPoints:    ids,
			Links:           convertLinksToTypes(item.GetLinks()),
			IsDirect:        item.GetIsDirect(),
			DirectEndpoints: convertEndpointToTypes(item.GetDirectEndpoints()),
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

// AutoConvert auto convert.
func (x *TopoNetworkUnitDeleteReq) AutoConvert() {
}

// ConvertNetworkUnitFromTypes convert networkunit from types to proto.
func (x *TopoNetworkUnitDeleteResp) ConvertNetworkUnitFromTypes(networkUnitID int64) {
	data := &TopoNetworkUnitDeleteResp_Data{BkNetworkunitId: new(int64)}
	*data.BkNetworkunitId = networkUnitID

	x.Data = data
}

func (ap *AccessPoint) autoConvert() {
	if ap.AccesspointId == nil {
		ap.AccesspointId = new(int64)
		*ap.AccesspointId = -1
	}

	if ap.BkNetworkareaId == nil {
		ap.BkNetworkareaId = new(int64)
		*ap.BkNetworkareaId = -1
	}
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
		IsDirect: new(bool),
		DirectEndpoints: &Endpoints{
			Cluster: make([]string, 0),
			File:    make([]string, 0),
			Data:    make([]string, 0),
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
		IsDirect: new(bool),
		DirectEndpoints: &Endpoints{
			Cluster: make([]string, 0),
			File:    make([]string, 0),
			Data:    make([]string, 0),
		},
	}
}

func newEmptyLink() *Link {
	return &Link{AccesspointId: -1}
}

func newEmptyAccessPoint() *AccessPoint {
	return &AccessPoint{
		TenantId:        new(string),
		AccesspointId:   new(int64),
		AccesspointName: new(string),
		BkNetworkareaId: new(int64),
		Endpoints: &Endpoints{
			Cluster: make([]string, 0),
			File:    make([]string, 0),
			Data:    make([]string, 0),
		},
	}
}

func convertAccesspointsToTypes(
	tenantID string, networkAreaID int64, accessPoints []*AccessPoint) []*types.AccessPoint {

	data := make([]*types.AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		data[idx] = convertAccessPointToTypes(accessPoint)
		data[idx].TenantID = tenantID
		data[idx].NetworkAreaID = networkAreaID
	}

	return data
}

func convertAccesspointsFromTypes(accessPoints []*types.AccessPoint) []*AccessPoint {
	data := make([]*AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		data[idx] = convertAccessPointFromTypes(accessPoint)
	}

	return data
}

func convertAccessPointFromTypes(accessPoint *types.AccessPoint) *AccessPoint {
	data := newEmptyAccessPoint()
	*data.TenantId = accessPoint.TenantID
	*data.BkNetworkareaId = accessPoint.NetworkAreaID
	*data.AccesspointId = accessPoint.ID
	*data.AccesspointName = accessPoint.Name
	data.Endpoints = &Endpoints{
		Cluster: accessPoint.Endpoints.Cluster,
		File:    accessPoint.Endpoints.File,
		Data:    accessPoint.Endpoints.Data,
	}

	return data
}

func convertAccessPointToTypes(accessPoint *AccessPoint) *types.AccessPoint {
	data := &types.AccessPoint{
		TenantID:      accessPoint.GetTenantId(),
		NetworkAreaID: accessPoint.GetBkNetworkareaId(),
		ID:            accessPoint.GetAccesspointId(),
		Name:          accessPoint.GetAccesspointName(),
	}

	if endpoints := accessPoint.GetEndpoints(); endpoints != nil {
		data.Endpoints = types.Endpoints{
			Cluster: endpoints.GetCluster(),
			File:    endpoints.GetFile(),
			Data:    endpoints.GetData(),
		}
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

func convertEndpointToTypes(endpoint *Endpoints) *types.Endpoints {
	if endpoint == nil {
		return nil
	}

	return &types.Endpoints{
		Cluster: endpoint.Cluster,
		File:    endpoint.File,
		Data:    endpoint.Data,
	}
}

func convertEndpointFromTypes(endpoint *types.Endpoints) *Endpoints {
	if endpoint == nil {
		return nil
	}

	return &Endpoints{
		Cluster: endpoint.Cluster,
		File:    endpoint.File,
		Data:    endpoint.Data,
	}
}
