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

const (
	// 30 days for max operate time range.
	maxOperateTimeRangeDuration = 365 * 24 * time.Hour
)

// Validate check body.
func (x *TopoEventListReq) Validate() error {
	if err := validateTopoPage(x.GetPage()); err != nil {
		return err
	}

	if err := validateTimeRange(x.GetOperateTimeRange(), maxOperateTimeRangeDuration); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *TopoEventListReq) AutoConvert() {
	if x.GetOperateTimeRange() == nil {
		x.OperateTimeRange = &TimeRange{
			StartTimestampSec: time.Now().Add(-1 * maxOperateTimeRangeDuration).Unix(),
			EndTimestampSec:   time.Now().Unix(),
		}
	}
}

// ConvertPageToTypes convert page to types.
func (x *TopoEventListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoEventListReq) ConvertConditionsToTypes() *types.TopoEventCondition {
	return convertTopoEventConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetOperateTimeRange())
}

// ConvertConditionsFromTypes convert types to conditions.
func (x *TopoEventListReq) ConvertConditionsFromTypes(condition *types.TopoEventCondition) error {
	exactCond, fuzzyCond, timeRange, err := convertTopoEventConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.OperateTimeRange = timeRange
	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

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

// Validate check body.
func (x *TopoEventDistinctReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoEventDistinctReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoEventDistinctReq) ConvertConditionsToTypes() *types.TopoEventCondition {
	return convertTopoEventConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetOperateTimeRange())
}

// ConvertConditionsFromTypes convert types to conditions.
func (x *TopoEventDistinctReq) ConvertConditionsFromTypes(condition *types.TopoEventCondition) error {
	exactCond, fuzzyCond, timeRange, err := convertTopoEventConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.OperateTimeRange = timeRange
	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertFiltersFromTypes convert result from types.
func (x *TopoEventDistinctResp) ConvertResultFromTypes(result *types.TopoEventDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &TopoEventDistinctResp_Data{
		Type:            formatRespSlice(types.TopoEventTypeListToStringList(result.Type)),
		BkNetworkareaId: formatRespSlice(result.NetworkAreaID),
		BkNetworkunitId: formatRespSlice(result.NetworkUnitID),
		AccesspointId:   formatRespSlice(result.AccessPointID),
		Operator:        formatRespSlice(result.Operator),
	}
}

// ConvertResultToTypes convert result to types.
func (x *TopoEventDistinctResp) ConvertResultToTypes() *types.TopoEventDistinctResult {
	if x.GetData() == nil {
		return &types.TopoEventDistinctResult{}
	}

	data := x.GetData()
	return &types.TopoEventDistinctResult{
		Type:          types.StringListToTopoEventTypeList(data.GetType()),
		NetworkAreaID: data.GetBkNetworkareaId(),
		NetworkUnitID: data.GetBkNetworkunitId(),
		AccessPointID: data.GetAccesspointId(),
		Operator:      data.GetOperator(),
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

func convertTopoEventConditionsToTypes(
	exactCond *TopoEventExactConditions, fuzzyCond *TopoEventFuzzyConditions, timeRange *TimeRange) *types.TopoEventCondition {

	condition := &types.TopoEventCondition{}

	if timeRange != nil {
		condition.OperateTimeRange = &types.TimeRange{
			StartTime: time.Unix(timeRange.GetStartTimestampSec(), 0),
			EndTime:   time.Unix(timeRange.GetEndTimestampSec(), 0),
		}
	}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.TopoEventExactFields{
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			NetworkUnitID: exactCond.GetBkNetworkunitId(),
			AccessPointID: exactCond.GetAccesspointId(),
			Type:          types.StringListToTopoEventTypeList(exactCond.GetType()),
			Operator:      exactCond.GetOperator(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond != nil {
		condition.FuzzyInclude = &types.TopoEventFuzzyFields{
			NetworkAreaName: fuzzyCond.GetBkNetworkareaName(),
			NetworkUnitName: fuzzyCond.GetBkNetworkunitName(),
		}
	}

	return condition
}

func convertTopoEventConditionsFromTypes(condition *types.TopoEventCondition) (
	*TopoEventExactConditions, *TopoEventFuzzyConditions, *TimeRange, error) {

	if condition == nil {
		return nil, nil, nil, nil
	}

	var timeRange *TimeRange
	var exactCond *TopoEventExactConditions
	var fuzzyCond *TopoEventFuzzyConditions

	if condition.OperateTimeRange != nil {
		timeRange = &TimeRange{
			StartTimestampSec: condition.OperateTimeRange.StartTime.Unix(),
			EndTimestampSec:   condition.OperateTimeRange.EndTime.Unix(),
		}
	}

	if condition.ExactInclude != nil {
		exactCond = &TopoEventExactConditions{
			BkNetworkareaId: condition.ExactInclude.NetworkAreaID,
			BkNetworkunitId: condition.ExactInclude.NetworkUnitID,
			AccesspointId:   condition.ExactInclude.AccessPointID,
			Type:            types.TopoEventTypeListToStringList(condition.ExactInclude.Type),
			Operator:        condition.ExactInclude.Operator,
		}
	}

	if condition.FuzzyInclude != nil {
		fuzzyCond = &TopoEventFuzzyConditions{
			BkNetworkareaName: condition.FuzzyInclude.NetworkAreaName,
			BkNetworkunitName: condition.FuzzyInclude.NetworkUnitName,
		}
	}

	if condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, nil, nil, fmt.Errorf("exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, timeRange, nil
}
