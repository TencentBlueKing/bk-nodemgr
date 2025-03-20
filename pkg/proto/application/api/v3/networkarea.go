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
func (x *TopoNetworkAreaCreateReq) Validate() error {
	if x.GetBkNetworkareaName() == "" {
		return errors.New("bk_networkarea_name is required")
	}

	if x.GetBkCloudVendor() == "" {
		return errors.New("bk_cloud_vendor is required")
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
		CloudVendor: x.GetBkCloudVendor(),
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

	if x.GetBkCloudVendor() == "" {
		return errors.New("bk_cloud_vendor is required")
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
		CloudVendor: x.GetBkCloudVendor(),
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
	*data.BkCloudVendor = networkArea.CloudVendor

	x.Data = data
}

// Validate check body.
func (x *TopoNetworkAreaListReq) Validate() error {
	return validateTopoPage(x.GetPage())
}

// AutoConvert auto convert.
func (x *TopoNetworkAreaListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *TopoNetworkAreaListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoNetworkAreaListReq) ConvertConditionsToTypes() *types.NetworkAreaCondition {
	// exact conditions.
	if exactCond := x.GetExactIncludeConditions(); exactCond != nil {
		conditions := &types.NetworkAreaCondition{
			Type: types.ConditionTypeExactInclude,
		}
		conditions.Exact = &types.NetworkAreaExactFields{
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			CloudVendor:   exactCond.GetBkCloudVendor(),
		}

		return conditions
	}

	// fuzzy conditions.
	if fuzzyCond := x.GetFuzzyIncludeConditions(); fuzzyCond != nil {
		conditions := &types.NetworkAreaCondition{
			Type: types.ConditionTypeFuzzyInclude,
		}
		conditions.Fuzzy = &types.NetworkAreaFuzzyFields{
			NetworkAreaName: fuzzyCond.GetBkNetworkareaName(),
		}

		return conditions
	}

	// default empty conditions.
	return &types.NetworkAreaCondition{
		Type: types.ConditionTypeExactInclude,
	}
}

// ConvertConditionsFromTypes convert conditions from types to proto.
func (x *TopoNetworkAreaListReq) ConvertConditionsFromTypes(condition *types.NetworkAreaCondition) error {
	if condition == nil {
		return nil
	}

	switch condition.Type {
	case types.ConditionTypeExactInclude:
		if condition.Exact != nil {
			x.ExactIncludeConditions = &TopoNetworkAreaListReq_ExactConditions{
				BkNetworkareaId: condition.Exact.NetworkAreaID,
				BkCloudVendor:   condition.Exact.CloudVendor,
			}
		}

	case types.ConditionTypeFuzzyInclude:
		if condition.Fuzzy != nil {
			x.FuzzyIncludeConditions = &TopoNetworkAreaListReq_FuzzyConditions{
				BkNetworkareaName: condition.Fuzzy.NetworkAreaName,
			}
		}

	default:
		return fmt.Errorf("unknown condition type: %s", condition.Type)
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
		*item.BkCloudVendor = networkArea.CloudVendor

		items[idx] = item
	}

	x.Data = &TopoNetworkAreaListResp_Data{
		Total: total,
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
		BkCloudVendor:     new(string),
	}
}
