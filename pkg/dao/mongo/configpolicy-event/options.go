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

package configpolicyevent

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithBizID provides filtering by business id.
func WithBizID(bizIDs ...int64) OptFn {
	return base.WithValues(FieldKeyBizID, bizIDs...)
}

// WithType provides filtering by event type.
func WithType(eventType ...types.ConfigPolicyEventType) OptFn {
	return base.WithValues(FieldKeyType, types.ConfigPolicyEventTypeListToStringList(eventType)...)
}

// WithVersion provides filtering by version.
func WithVersion(version ...int64) OptFn {
	return base.WithValues(FieldKeyVersion, version...)
}

// WithConfigPolicyType filters by config policy type.
func WithConfigPolicyType(configPolicyType ...types.ConfigPolicyType) OptFn {
	return base.WithValues(FieldKeyConfigPolicyType, types.ConfigPolicyTypeListToStringList(configPolicyType)...)
}

// WithConfigPolicyID filters by config policy id.
func WithConfigPolicyID(configPolicyIDs ...int64) OptFn {
	return base.WithValues(FieldKeyConfigPolicyID, configPolicyIDs...)
}

// WithConfigPolicyName filters by config policy name.
func WithConfigPolicyName(configPolicyNames ...string) OptFn {
	return base.WithValues(FieldKeyConfigPolicyName, configPolicyNames...)
}

// WithOperator filters by operator.
func WithOperator(operator ...string) OptFn {
	return base.WithValues(FieldKeyOperator, operator...)
}

// WithOperateTimeRange filters by operate-time.
func WithOperateTimeRange(timeRange types.TimeRange) OptFn {
	return base.WithTimeRange(FieldKeyOperateTime, timeRange.StartTime, timeRange.EndTime)
}
