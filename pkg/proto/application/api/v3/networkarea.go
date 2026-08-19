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

package v3

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *TopoNetworkAreaCreateReq) Validate() error {
	if x.GetBkNetworkareaName() == "" {
		return errors.New("bk_networkarea_name is required")
	}

	if x.GetCloudVendor() == "" {
		return errors.New("cloud_vendor is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *TopoNetworkAreaCreateReq) AutoConvert() {
}

// ConvertNetworkAreaToTypes convert networkarea from proto types.
func (x *TopoNetworkAreaCreateReq) ConvertNetworkAreaToTypes(tenantID string, networkAreaID int64) *types.NetworkArea {
	return &types.NetworkArea{
		TenantID:    tenantID,
		ID:          networkAreaID,
		Name:        x.GetBkNetworkareaName(),
		CloudVendor: x.GetCloudVendor(),
	}
}

// ConvertNetworkAreaFromTypes convert networkarea from types to proto.
func (x *TopoNetworkAreaCreateResp) ConvertNetworkAreaFromTypes(networkAreaID int64) {
	data := &TopoNetworkAreaCreateResp_Data{BkNetworkareaId: new(int64)}
	*data.BkNetworkareaId = networkAreaID

	x.Data = data
}

// Validate check body.
func (x *TopoNetworkAreaUpdateReq) Validate() error {
	if x.GetBkNetworkareaId() < 0 {
		return errors.New("bk_networkarea_id is invalid")
	}

	if x.GetBkNetworkareaName() == "" {
		return errors.New("bk_networkarea_name is required")
	}

	if x.GetCloudVendor() == "" {
		return errors.New("cloud_vendor is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *TopoNetworkAreaUpdateReq) AutoConvert() {
}

// ConvertNetworkAreaToTypes convert networkarea from proto to types.
func (x *TopoNetworkAreaUpdateReq) ConvertNetworkAreaToTypes(tenantID string) *types.NetworkArea {
	return &types.NetworkArea{
		TenantID:    tenantID,
		ID:          x.GetBkNetworkareaId(),
		Name:        x.GetBkNetworkareaName(),
		CloudVendor: x.GetCloudVendor(),
	}
}

// ConvertNetworkAreaFromTypes convert networkarea from types to proto.
func (x *TopoNetworkAreaUpdateResp) ConvertNetworkAreaFromTypes(networkAreaID int64) {
	data := &TopoNetworkAreaUpdateResp_Data{BkNetworkareaId: new(int64)}
	*data.BkNetworkareaId = networkAreaID

	x.Data = data
}

// Validate check body.
func (x *TopoNetworkAreaGetReq) Validate() error {
	if x.GetBkNetworkareaId() < 0 {
		return errors.New("bk_networkarea_id is invalid")
	}

	return nil
}

// AutoConvert auto convert.
func (x *TopoNetworkAreaGetReq) AutoConvert() {
}

// ConvertNetworkAreaFromTypes convert networkarea from types to proto.
func (x *TopoNetworkAreaGetResp) ConvertNetworkAreaFromTypes(networkArea *types.NetworkArea) {
	data := newEmptyNetworkArea()
	*data.TenantId = networkArea.TenantID
	*data.BkNetworkareaId = networkArea.ID
	*data.BkNetworkareaName = networkArea.Name
	*data.CloudVendor = networkArea.CloudVendor

	x.Data = data
}

// Validate check body.
func (x *TopoNetworkAreaListReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert auto convert.
func (x *TopoNetworkAreaListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *TopoNetworkAreaListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoNetworkAreaListReq) ConvertConditionsToTypes() *types.NetworkAreaCondition {
	condition := &types.NetworkAreaCondition{}

	// exact conditions.
	if exactCond := x.GetExactIncludeConditions(); exactCond != nil {
		condition.ExactInclude = &types.NetworkAreaExactFields{
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			CloudVendor:   exactCond.GetCloudVendor(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond := x.GetFuzzyIncludeConditions(); fuzzyCond != nil {
		condition.FuzzyInclude = &types.NetworkAreaFuzzyFields{
			NetworkAreaName: fuzzyCond.GetBkNetworkareaName(),
		}
	}

	return condition
}

// ConvertConditionsFromTypes convert conditions from types to proto.
func (x *TopoNetworkAreaListReq) ConvertConditionsFromTypes(condition *types.NetworkAreaCondition) error {
	if condition == nil {
		return nil
	}

	if condition.ExactInclude != nil {
		x.ExactIncludeConditions = &TopoNetworkAreaListReq_ExactConditions{
			BkNetworkareaId: condition.ExactInclude.NetworkAreaID,
			CloudVendor:     condition.ExactInclude.CloudVendor,
		}
	}

	if condition.FuzzyInclude != nil {
		x.FuzzyIncludeConditions = &TopoNetworkAreaListReq_FuzzyConditions{
			BkNetworkareaName: condition.FuzzyInclude.NetworkAreaName,
		}
	}

	if condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return errors.New("exact-exclude and fuzzy-exclude not supported")
	}

	return nil
}

// ConvertNetworkAreasFromTypes convert networkareas from types to proto.
func (x *TopoNetworkAreaListResp) ConvertNetworkAreasFromTypes(total int64, networkAreas []*types.NetworkArea) {
	items := make([]*NetworkArea, len(networkAreas))
	for idx, networkArea := range networkAreas {
		item := newEmptyNetworkArea()
		*item.TenantId = networkArea.TenantID
		*item.BkNetworkareaId = networkArea.ID
		*item.BkNetworkareaName = networkArea.Name
		*item.CloudVendor = networkArea.CloudVendor

		items[idx] = item
	}

	x.Data = &TopoNetworkAreaListResp_Data{
		Total: total,
		Items: items,
	}
}

// NetworkAreaStatistics describes the networkarea statics.
type NetworkAreaStatistics struct {
	NetworkAreaID    int64
	NetworkUnitCount int64
	AgentCount       int64
	ProxyCount       int64
	LastOperator     string
	LastOperateTime  time.Time
}

// Validate check body.
func (x *TopoNetworkAreaStatisticsReq) Validate() error {
	if len(x.GetBkNetworkareaId()) == 0 {
		return errors.New("bk_networkarea_id is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *TopoNetworkAreaStatisticsReq) AutoConvert() {
}

// ConvertNetworkUnitConditionToTypes convert networkunit condition to types.
func (x *TopoNetworkAreaStatisticsReq) ConvertNetworkUnitConditionToTypes() *types.NetworkUnitCondition {
	return &types.NetworkUnitCondition{
		ExactInclude: &types.NetworkUnitExactFields{
			NetworkAreaID: x.GetBkNetworkareaId(),
		},
	}
}

// ConvertNetworkAreaStatisticsFromResult convert networkarea statics from result.
func (x *TopoNetworkAreaStatisticsResp) ConvertNetworkAreaStatisticsFromResult(
	result map[int64]*NetworkAreaStatistics) {

	items := make([]*TopoNetworkAreaStatisticsResp_StatisticsInfo, len(result))
	idx := 0
	for networkAreaID, networkAreaStatics := range result {
		info := &TopoNetworkAreaStatisticsResp_StatisticsInfo{
			BkNetworkareaId:  new(int64),
			NetworkunitCount: new(int64),
			ProxyCount:       new(int64),
			AgentCount:       new(int64),
		}

		*info.BkNetworkareaId = networkAreaID
		*info.NetworkunitCount = networkAreaStatics.NetworkUnitCount
		*info.ProxyCount = networkAreaStatics.ProxyCount
		*info.AgentCount = networkAreaStatics.AgentCount

		items[idx] = info
		idx++
	}

	x.Data = &TopoNetworkAreaStatisticsResp_Data{
		Items: items,
	}
}

// Validate check body.
func (x *TopoNetworkAreaDeleteReq) Validate() error {
	if x.GetBkNetworkareaId() < 0 {
		return errors.New("bk_networkarea_id is invalid")
	}

	return nil
}

// AutoConvert auto convert.
func (x *TopoNetworkAreaDeleteReq) AutoConvert() {
}

// ConvertNetworkUnitFromTypes convert networkunit from types to proto.
func (x *TopoNetworkAreaDeleteResp) ConvertNetworkUnitFromTypes(networkUnitID int64) {
	data := &TopoNetworkAreaDeleteResp_Data{BkNetworkareaId: new(int64)}
	*data.BkNetworkareaId = networkUnitID

	x.Data = data
}

func newEmptyNetworkArea() *NetworkArea {
	return &NetworkArea{
		TenantId:          new(string),
		BkNetworkareaId:   new(int64),
		BkNetworkareaName: new(string),
		CloudVendor:       new(string),
	}
}
