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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *TopoAccessPointListReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert auto convert.
func (x *TopoAccessPointListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *TopoAccessPointListReq) ConvertPageToTypes(maxLimit int) (types.Page, error) {
	return convPageToTypes(x.GetPage(), maxLimit)
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoAccessPointListReq) ConvertConditionsToTypes() *types.AccessPointCondition {
	condition := &types.AccessPointCondition{}

	// exact conditions.
	if exactCond := x.GetExactIncludeConditions(); exactCond != nil {
		condition.ExactInclude = &types.AccessPointExactFields{
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			AccessPointID: exactCond.GetAccesspointId(),
		}
	}

	return condition
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *TopoAccessPointListReq) ConvertConditionsFromTypes(condition *types.AccessPointCondition) error {
	if condition == nil {
		return nil
	}

	if condition.ExactInclude != nil {
		x.ExactIncludeConditions = &TopoAccessPointListReq_ExactConditions{
			AccesspointId:   condition.ExactInclude.AccessPointID,
			BkNetworkareaId: condition.ExactInclude.NetworkAreaID,
		}
	}

	if condition.FuzzyInclude != nil || condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return fmt.Errorf("fuzzy-include, exact-exclude and fuzzy-exclude not supported")
	}

	return nil
}

// ConvertAccessPointsFromTypes convert accesspoints from types.
func (x *TopoAccessPointListResp) ConvertAccessPointsFromTypes(total int64, accessPoints []*types.AccessPoint) {
	items := make([]*AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		items[idx] = convertAccessPointFromTypes(accessPoint)
	}

	x.Data = &TopoAccessPointListResp_Data{
		Total: total,
		Items: items,
	}
}

// ConvertAccessPointsToTypes convert accesspoints to types.
func (x *TopoAccessPointListResp) ConvertAccessPointsToTypes() (int64, []*types.AccessPoint) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.AccessPoint, len(items))
	for idx, item := range items {
		result[idx] = convertAccessPointToTypes(item)
	}

	return data.GetTotal(), result
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
