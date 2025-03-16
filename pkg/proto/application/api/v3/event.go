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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *TopoEventListReq) Validate() error {
	return validateTopoPage(x.GetPage())
}

// ConvertPageToTypes convert page to types.
func (x *TopoEventListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoEventListReq) ConvertConditionsToTypes() *types.TopoEventCondition {
	// exact conditions.
	if exactCond := x.GetExactIncludeConditions(); exactCond != nil {
		conditions := &types.TopoEventCondition{
			Type: types.ConditionTypeExactInclude,
		}
		conditions.Exact = &types.TopoEventExactFields{
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			NetworkUnitID: exactCond.GetBkNetworkunitId(),
			AccessPointID: exactCond.GetAccesspointId(),
			Type:          types.TopoEventTypeToNodeRoleList(exactCond.GetType()),
			Operator:      exactCond.GetOperator(),
		}

		return conditions
	}

	// default empty conditions.
	return &types.TopoEventCondition{
		Type: types.ConditionTypeExactInclude,
	}
}

// ConvertConditionsFromTypes convert types to conditions.
func (x *TopoEventListReq) ConvertConditionsFromTypes(condition *types.TopoEventCondition) error {
	if condition == nil {
		return nil
	}

	switch condition.Type {
	case types.ConditionTypeExactInclude:
		if condition.Exact != nil {
			x.ExactIncludeConditions = &TopoEventListReq_ExactConditions{
				BkNetworkareaId: condition.Exact.NetworkAreaID,
				BkNetworkunitId: condition.Exact.NetworkUnitID,
				AccesspointId:   condition.Exact.AccessPointID,
				Type:            types.TopoEventTypeListToStringList(condition.Exact.Type),
				Operator:        condition.Exact.Operator,
			}
		}

	default:
		return fmt.Errorf("unknown condition type: %s", condition.Type)
	}

	return nil
}

// ConvertTopoEventsToTypes convert topo events to types.
func (x *TopoEventListResp) ConvertTopoEventsToTypes() (int64, []*types.TopoEvent) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.TopoEvent, len(items))
	for idx, item := range items {
		result[idx] = &types.TopoEvent{
			TenantID:        item.GetTenantId(),
			Type:            types.TopoEventType(item.GetType()),
			NetworkAreaID:   item.GetBkNetworkareaId(),
			NetworkAreaName: item.GetBkNetworkareaName(),
			NetworkUnitID:   item.GetBkNetworkunitId(),
			NetworkUnitName: item.GetBkNetworkunitName(),
			AccessPointID:   item.GetAccesspointId(),
			AccessPointName: item.GetAccesspointName(),
			OperateTime:     time.UnixMilli(item.GetOperateTime()),
			Operator:        item.GetOperator(),
		}
	}

	return data.GetTotal(), result
}

// ConvertTopoEventsFromTypes convert types to topo events.
func (x *TopoEventListResp) ConvertTopoEventsFromTypes(total int64, events []*types.TopoEvent) {
	items := make([]*TopoEvent, len(events))
	for idx, event := range events {
		item := newEmptyTopoEvent()
		*item.TenantId = event.TenantID
		*item.Type = string(event.Type)
		*item.BkNetworkareaId = event.NetworkAreaID
		*item.BkNetworkareaName = event.NetworkAreaName
		*item.BkNetworkunitId = event.NetworkUnitID
		*item.BkNetworkunitName = event.NetworkUnitName
		*item.AccesspointId = event.AccessPointID
		*item.AccesspointName = event.AccessPointName
		*item.OperateTime = event.OperateTime.UnixMilli()
		*item.Operator = event.Operator

		items[idx] = item
	}

	x.Data = &TopoEventListResp_Data{
		Total: total,
		Items: items,
	}
}

func newEmptyTopoEvent() *TopoEvent {
	return &TopoEvent{
		TenantId:          new(string),
		Type:              new(string),
		BkNetworkareaId:   new(int64),
		BkNetworkareaName: new(string),
		BkNetworkunitId:   new(int64),
		BkNetworkunitName: new(string),
		AccesspointId:     new(int64),
		AccesspointName:   new(string),
		OperateTime:       new(int64),
		Operator:          new(string),
	}
}
