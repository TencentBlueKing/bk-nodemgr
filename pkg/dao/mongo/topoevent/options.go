/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topoevent

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithNetworkAreaID filters by networkarea-id.
func WithNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithValues(FieldKeyNetworkAreaID, networkAreaIDs...)
}

// WithoutNetworkAreaID filters by not contains networkarea-id.
func WithoutNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyNetworkAreaID, networkAreaIDs...)
}

// WithNetworkUnitID filters by networkunit-id.
func WithNetworkUnitID(networkUnitIDs ...int64) OptFn {
	return base.WithValues(FieldKeyNetworkUnitID, networkUnitIDs...)
}

// WithoutNetworkUnitID filters by not contains networkunit-id.
func WithoutNetworkUnitID(networkUnitIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyNetworkUnitID, networkUnitIDs...)
}

// WithAccessPointID filters by accesspoint-id.
func WithAccessPointID(accessPointIDs ...int64) OptFn {
	return base.WithValues(FieldKeyAccessPointID, accessPointIDs...)
}

// WithoutAccessPointID filters by not contains accesspoint-id.
func WithoutAccessPointID(accessPointIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyAccessPointID, accessPointIDs...)
}

// WithType filters by event-type.
func WithType(eventTypes ...types.TopoEventType) OptFn {
	return base.WithValues(FieldKeyType, types.TopoEventTypeListToStringList(eventTypes)...)
}

// WithoutType filters by not contains event-type.
func WithoutType(eventTypes ...types.TopoEventType) OptFn {
	return base.WithoutValues(FieldKeyType, types.TopoEventTypeListToStringList(eventTypes)...)
}

// WithOperator filters by operator.
func WithOperator(operator ...string) OptFn {
	return base.WithValues(FieldKeyOperator, operator...)
}

// WithoutOperator filters by not contains operator.
func WithoutOperator(operator ...string) OptFn {
	return base.WithoutValues(FieldKeyOperator, operator...)
}

// WithFuzzyNetworkAreaName filters by networkarea-name.
func WithFuzzyNetworkAreaName(networkAreaNames ...string) OptFn {
	return base.WithFuzzyValues("data.networkarea_name", networkAreaNames...)
}

// WithoutFuzzyNetworkAreaName filters by not contains networkarea-name.
func WithoutFuzzyNetworkAreaName(networkAreaNames ...string) OptFn {
	return base.WithoutFuzzyValues("data.networkarea_name", networkAreaNames...)
}

// WithFuzzyNetworkUnitName filters by networkunit-name.
func WithFuzzyNetworkUnitName(networkUnitNames ...string) OptFn {
	return base.WithFuzzyValues("data.networkunit_name", networkUnitNames...)
}

// WithoutFuzzyNetworkUnitName filters by not contains networkunit-name.
func WithoutFuzzyNetworkUnitName(networkUnitNames ...string) OptFn {
	return base.WithoutFuzzyValues("data.networkunit_name", networkUnitNames...)
}

// WithOperateTimeRange filters by operate-time.
func WithOperateTimeRange(timeRange types.TimeRange) OptFn {
	return base.WithTimeRange(FieldKeyOperateTime, timeRange.StartTime, timeRange.EndTime)
}
