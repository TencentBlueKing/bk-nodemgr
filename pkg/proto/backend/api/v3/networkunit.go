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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
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

	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return fmt.Errorf("invalid generation, generation(%d): %w", x.GetGeneration(), err)
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

// ConvertCustomDeployConfigToTypes convert custom deploy config from proto to types.
func (x *TopoNetworkUnitCreateReq) ConvertCustomDeployConfigToTypes() map[criteria.OSType]types.CustomDeployConfig {
	return convertCustomDeployConfigToTypes(x.GetCustomDeployConfig())
}

// ConvertCustomDeployConfigFromTypes convert custom deploy config from types to proto.
func (x *TopoNetworkUnitCreateReq) ConvertCustomDeployConfigFromTypes(deployConfig map[criteria.OSType]types.CustomDeployConfig) {
	x.CustomDeployConfig = convertCustomDeployConfigFromTypes(deployConfig)
}

// ConvertNetworkUnitFromTypes convert networkunit from types to proto.
func (x *TopoNetworkUnitCreateReq) ConvertNetworkUnitFromTypes(networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) {
	x.BkNetworkunitName = networkUnit.Name
	x.BkNetworkareaId = networkUnit.NetworkAreaID
	x.IsDirect = networkUnit.IsDirect
	x.ConvertAccesspointsFromTypes(accessPoints)
	x.ConvertLinksFromTypes(networkUnit.Links)
	x.ConvertDirectEndpointsFromTypes(networkUnit.DirectEndpoints)
	x.Generation = int64(networkUnit.Generation)
	x.ConvertCustomDeployConfigFromTypes(networkUnit.CustomDeployConfig)
}

// ConvertNetworkUnitFromTypes convert networkunit from types to proto.
func (x *TopoNetworkUnitCreateResp) ConvertNetworkUnitFromTypes(networkUnitID int64) {
	data := &TopoNetworkUnitCreateResp_Data{BkNetworkunitId: new(int64)}
	*data.BkNetworkunitId = networkUnitID

	x.Data = data
}

// Validate check body.
func (x *TopoNetworkUnitUpdateReq) Validate() error {
	if x.GetFields() == nil {
		return errors.New("fields is required")
	}

	if x.GetNetworkunit() == nil {
		return errors.New("networkunit is required")
	}

	if x.GetNetworkunit().GetBkNetworkunitId() < 0 {
		return errors.New("bk_networkunit_id is invalid")
	}

	if x.GetNetworkunit().GetBkNetworkareaId() < 0 {
		return errors.New("bk_networkarea_id is invalid")
	}

	return nil
}

// AutoConvert auto convert.
func (x *TopoNetworkUnitUpdateReq) AutoConvert() {
	for _, accesspoint := range x.GetNetworkunit().GetAccesspoints() {
		accesspoint.autoConvert()
	}
}

// ConvertFieldsToTypes convert fields to types.
func (x *TopoNetworkUnitUpdateReq) ConvertFieldsToTypes() types.NetworkUnitUpdateFields {
	return convertFieldsToTypes(x.GetFields())
}

// ConvertFieldsFromTypes convert fields from types to proto.
func (x *TopoNetworkUnitUpdateReq) ConvertFieldsFromTypes(fields types.NetworkUnitUpdateFields) {
	x.Fields = convertFieldsFromTypes(fields)
}

// ConvertAccssPointsToTypes convert access points from proto to types.
func (x *TopoNetworkUnitUpdateReq) ConvertAccssPointsToTypes(
	tenantID string,
	networkAreaID int64) []*types.AccessPoint {

	return convertAccesspointsToTypes(tenantID, networkAreaID, x.GetNetworkunit().GetAccesspoints())
}

// ConvertAccesspointsFromTypes convert access points from types to proto.
func (x *TopoNetworkUnitUpdateReq) ConvertAccesspointsFromTypes(accessPoints []*types.AccessPoint) {
	x.Networkunit.Accesspoints = convertAccesspointsFromTypes(accessPoints)
}

// ConvertLinksToTypes convert links from proto to types.
func (x *TopoNetworkUnitUpdateReq) ConvertLinksToTypes() types.Links {
	return convertLinksToTypes(x.GetNetworkunit().GetLinks())
}

// ConvertLinksFromTypes convert links from types to proto.
func (x *TopoNetworkUnitUpdateReq) ConvertLinksFromTypes(links types.Links) {
	x.Networkunit.Links = convertLinksFromTypes(links)
}

// ConvertDirectEndpointsToTypes convert endpoints from proto to types.
func (x *TopoNetworkUnitUpdateReq) ConvertDirectEndpointsToTypes() *types.Endpoints {
	return convertEndpointToTypes(x.GetNetworkunit().GetDirectEndpoints())
}

// ConvertDirectEndpointsFromTypes convert endpoints from types to proto.
func (x *TopoNetworkUnitUpdateReq) ConvertDirectEndpointsFromTypes(endpoints *types.Endpoints) {
	x.Networkunit.DirectEndpoints = convertEndpointFromTypes(endpoints)
}

// ConvertCustomDeployConfigToTypes convert custom deploy config from proto to types.
func (x *TopoNetworkUnitUpdateReq) ConvertCustomDeployConfigToTypes() map[criteria.OSType]types.CustomDeployConfig {
	return convertCustomDeployConfigToTypes(x.GetNetworkunit().GetCustomDeployConfig())
}

// ConvertCustomDeployConfigFromTypes convert custom deploy config from types to proto.
func (x *TopoNetworkUnitUpdateReq) ConvertCustomDeployConfigFromTypes(deployConfig map[criteria.OSType]types.CustomDeployConfig) {
	x.Networkunit.CustomDeployConfig = convertCustomDeployConfigFromTypes(deployConfig)
}

// ConvertNetworkUnitFromTypes convert networkunit from types to proto.
func (x *TopoNetworkUnitUpdateReq) ConvertNetworkUnitFromTypes(
	fields types.NetworkUnitUpdateFields,
	networkUnit *types.NetworkUnit,
	accessPoints ...*types.AccessPoint,
) {
	x.Fields = convertFieldsFromTypes(fields)

	generation := int64(networkUnit.Generation)
	x.Networkunit = &NetworkUnit{
		BkNetworkunitId:    &networkUnit.ID,
		BkNetworkunitName:  &networkUnit.Name,
		BkNetworkareaId:    &networkUnit.NetworkAreaID,
		IsDirect:           &networkUnit.IsDirect,
		Accesspoints:       convertAccesspointsFromTypes(accessPoints),
		Links:              convertLinksFromTypes(networkUnit.Links),
		DirectEndpoints:    convertEndpointFromTypes(networkUnit.DirectEndpoints),
		Generation:         &generation,
		CustomDeployConfig: convertCustomDeployConfigFromTypes(networkUnit.CustomDeployConfig),
	}
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
	*data.Generation = int64(networkUnit.Generation)
	data.CustomDeployConfig = convertCustomDeployConfigFromTypes(networkUnit.CustomDeployConfig)

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
		TenantID:           data.GetTenantId(),
		NetworkAreaID:      data.GetBkNetworkareaId(),
		ID:                 data.GetBkNetworkunitId(),
		Name:               data.GetBkNetworkunitName(),
		AccessPoints:       accessPointIDs,
		Links:              convertLinksToTypes(data.GetLinks()),
		IsDirect:           data.GetIsDirect(),
		DirectEndpoints:    convertEndpointToTypes(data.GetDirectEndpoints()),
		Generation:         types.Generation(data.GetGeneration()),
		CustomDeployConfig: convertCustomDeployConfigToTypes(data.GetCustomDeployConfig()),
	}, accessPointsMap
}

// Validate check body.
func (x *TopoNetworkUnitListReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert auto convert.
func (x *TopoNetworkUnitListReq) AutoConvert() {
}

const (
	// network unit list max limit
	maxNetworkUnitLimit = 1000
)

// PageTimeout return page timeout.
func (x *TopoNetworkUnitListReq) PageTimeout() time.Duration {
	return backendPagingListTimeout
}

// PageLimit return page limit.
func (x *TopoNetworkUnitListReq) PageLimit() int {
	return maxNetworkUnitLimit
}

// ConvertPageToTypes convert page to types.
func (x *TopoNetworkUnitListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
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
			Generation:    exactCond.GetGeneration(),
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
			IsDirect:        condition.ExactInclude.IsDirect,
			Generation:      condition.ExactInclude.Generation,
		}
	}

	if condition.FuzzyInclude != nil || condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return fmt.Errorf("fuzzy-include, exact-exclude and fuzzy-exclude not supported")
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
		*item.IsDirect = networkUnit.IsDirect
		item.DirectEndpoints = convertEndpointFromTypes(networkUnit.DirectEndpoints)
		*item.Generation = int64(networkUnit.Generation)
		item.CustomDeployConfig = convertCustomDeployConfigFromTypes(networkUnit.CustomDeployConfig)

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
			TenantID:           item.GetTenantId(),
			NetworkAreaID:      item.GetBkNetworkareaId(),
			ID:                 item.GetBkNetworkunitId(),
			Name:               item.GetBkNetworkunitName(),
			AccessPoints:       item.GetAccesspoints(),
			Links:              convertLinksToTypes(item.GetLinks()),
			IsDirect:           item.GetIsDirect(),
			DirectEndpoints:    convertEndpointToTypes(item.GetDirectEndpoints()),
			Generation:         types.Generation(item.GetGeneration()),
			CustomDeployConfig: convertCustomDeployConfigToTypes(item.GetCustomDeployConfig()),
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
		Generation:         new(int64),
		CustomDeployConfig: make(map[string]*CustomDeployConfig, 0),
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
		Generation:         new(int64),
		CustomDeployConfig: make(map[string]*CustomDeployConfig, 0),
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
		Endpoints: &Endpoints{
			Cluster: make([]string, 0),
			File:    make([]string, 0),
			Data:    make([]string, 0),
		},
	}
}

func convertFieldsToTypes(fields *NetworkUnitUpdateFields) types.NetworkUnitUpdateFields {
	if fields == nil {
		return types.NetworkUnitUpdateFields{}
	}

	return types.NetworkUnitUpdateFields{
		Name:               fields.GetBkNetworkunitName(),
		AccessPoints:       fields.GetAccesspoints(),
		Links:              fields.GetLinks(),
		DirectEndpoints:    fields.GetDirectEndpoints(),
		CustomDeployConfig: fields.GetCustomDeployConfig(),
	}
}

func convertFieldsFromTypes(fields types.NetworkUnitUpdateFields) *NetworkUnitUpdateFields {
	return &NetworkUnitUpdateFields{
		BkNetworkunitName:  fields.Name,
		Accesspoints:       fields.AccessPoints,
		Links:              fields.Links,
		DirectEndpoints:    fields.DirectEndpoints,
		CustomDeployConfig: fields.CustomDeployConfig,
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

func convertCustomDeployConfigToTypes(deployConfig map[string]*CustomDeployConfig) map[criteria.OSType]types.CustomDeployConfig {
	if deployConfig == nil {
		return nil
	}

	data := make(map[criteria.OSType]types.CustomDeployConfig, len(deployConfig))
	for osType, config := range deployConfig {
		data[criteria.OSType(osType)] = types.CustomDeployConfig{
			InstallerRuntime: types.InstallerRuntime{
				BaseWorkDir: config.GetInstallerRuntime().GetBaseWorkDir(),
			},
			NodeRuntime: types.NodeRuntime{
				BaseDeployDir: config.GetNodeRuntime().GetBaseDeployDir(),
				DataIPC:       config.GetNodeRuntime().GetDataIpc(),
				PluginIPC:     config.GetNodeRuntime().GetPluginIpc(),
				LogDir:        config.GetNodeRuntime().GetLogDir(),
			},
			PluginRuntime: types.PluginRuntime{
				BaseDeployDir: config.GetPluginRuntime().GetBaseDeployDir(),
				LogDir:        config.GetPluginRuntime().GetLogDir(),
			},
		}
	}

	return data
}

func convertCustomDeployConfigFromTypes(deployConfig map[criteria.OSType]types.CustomDeployConfig) map[string]*CustomDeployConfig {
	if deployConfig == nil {
		return nil
	}

	data := make(map[string]*CustomDeployConfig, len(deployConfig))
	for osType, config := range deployConfig {
		data[osType.String()] = &CustomDeployConfig{
			InstallerRuntime: &InstallerRuntime{
				BaseWorkDir: &config.InstallerRuntime.BaseWorkDir,
			},
			NodeRuntime: &NodeRuntime{
				BaseDeployDir: &config.NodeRuntime.BaseDeployDir,
				DataIpc:       &config.NodeRuntime.DataIPC,
				PluginIpc:     &config.NodeRuntime.PluginIPC,
				LogDir:        &config.NodeRuntime.LogDir,
			},
			PluginRuntime: &PluginRuntime{
				BaseDeployDir: &config.PluginRuntime.BaseDeployDir,
				LogDir:        &config.PluginRuntime.LogDir,
			},
		}
	}

	return data
}
