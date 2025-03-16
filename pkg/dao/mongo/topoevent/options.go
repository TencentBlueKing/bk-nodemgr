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
	return base.WithInt64Values("data.networkarea_id", networkAreaIDs...)
}

// WithoutNetworkAreaID filters by not contains networkarea-id.
func WithoutNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithoutInt64Values("data.networkarea_id", networkAreaIDs...)
}

// WithNetworkUnitID filters by networkunit-id.
func WithNetworkUnitID(networkUnitIDs ...int64) OptFn {
	return base.WithInt64Values("data.networkunit_id", networkUnitIDs...)
}

// WithoutNetworkUnitID filters by not contains networkunit-id.
func WithoutNetworkUnitID(networkUnitIDs ...int64) OptFn {
	return base.WithoutInt64Values("data.networkunit_id", networkUnitIDs...)
}

// WithAccessPointID filters by accesspoint-id.
func WithAccessPointID(accessPointIDs ...int64) OptFn {
	return base.WithInt64Values("data.accesspoint_id", accessPointIDs...)
}

// WithoutAccessPointID filters by not contains accesspoint-id.
func WithoutAccessPointID(accessPointIDs ...int64) OptFn {
	return base.WithoutInt64Values("data.accesspoint_id", accessPointIDs...)
}

// WithType filters by event-type.
func WithType(eventTypes ...types.TopoEventType) OptFn {
	return base.WithStringValues("type", types.TopoEventTypeListToStringList(eventTypes)...)
}

// WithoutType filters by not contains event-type.
func WithoutType(eventTypes ...types.TopoEventType) OptFn {
	return base.WithoutStringValues("type", types.TopoEventTypeListToStringList(eventTypes)...)
}

// WithOperator filters by operator.
func WithOperator(operator ...string) OptFn {
	return base.WithStringValues("operator", operator...)
}

// WithoutOperator filters by not contains operator.
func WithoutOperator(operator ...string) OptFn {
	return base.WithoutStringValues("operator", operator...)
}
