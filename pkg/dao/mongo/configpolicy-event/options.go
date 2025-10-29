/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package configpolicyevent

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

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
