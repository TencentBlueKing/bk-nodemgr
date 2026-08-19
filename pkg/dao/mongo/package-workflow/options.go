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

package packageworkflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithWorkflowID filters by workflow ID.
func WithWorkflowID(workflowIDs ...string) OptFn {
	return base.WithValues(FieldKeyWorkflowID, workflowIDs...)
}

// WithStatus filters by status.
func WithStatus(statuses ...types.PackageWorkflowStatus) OptFn {
	return base.WithValues(FieldKeyStatus, statuses...)
}

// WithType filters by type.
func WithType(packageWorkflowTypes ...types.PackageWorkflowType) OptFn {
	return base.WithValues(FieldKeyType, packageWorkflowTypes...)
}

// WithOperator filters by operator.
func WithOperator(operators ...string) OptFn {
	return base.WithValues(FieldKeyOperator, operators...)
}

// WithOperateTimeRange filters by operate-time.
func WithOperateTimeRange(timeRange types.TimeRange) OptFn {
	return base.WithTimeRange(FieldKeyOperateTime, timeRange.StartTime, timeRange.EndTime)
}

// WithTriggerID filters by trigger ID.
func WithTriggerID(triggerIDs ...string) OptFn {
	return base.WithValues(FieldKeyTriggerID, triggerIDs...)
}
