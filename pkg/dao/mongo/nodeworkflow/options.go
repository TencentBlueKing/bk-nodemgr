/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeworkflow

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
func WithStatus(statuses ...types.NodeWorkflowStatus) OptFn {
	return base.WithValues(FieldKeyStatus, types.NodeWorkflowStatusListToStringList(statuses)...)
}

// WithoutStatus filters by not contains status.
func WithoutStatus(statuses ...types.NodeWorkflowStatus) OptFn {
	return base.WithoutValues(FieldKeyStatus, types.NodeWorkflowStatusListToStringList(statuses)...)
}

// WithType filters by type.
func WithType(nodeWorkflowTypes ...types.NodeWorkflowType) OptFn {
	return base.WithValues(FieldKeyType, types.NodeWorkflowTypeListToStringList(nodeWorkflowTypes)...)
}

// WithoutType filters by not contains type.
func WithoutType(nodeWorkflowTypes ...types.NodeWorkflowType) OptFn {
	return base.WithoutValues(FieldKeyType, types.NodeWorkflowTypeListToStringList(nodeWorkflowTypes)...)
}

// WithBizID filters by biz-id.
func WithBizID(bizIDs ...int64) OptFn {
	return base.WithValues(FieldKeyBizID, bizIDs...)
}

// WithoutBizID filters by not contains biz-id.
func WithoutBizID(bizID ...int64) OptFn {
	return base.WithoutValues(FieldKeyBizID, bizID...)
}

// WithOperator filters by operator.
func WithOperator(operator ...string) OptFn {
	return base.WithValues(FieldKeyOperator, operator...)
}

// WithoutOperator filters by not contains operator.
func WithoutOperator(operator ...string) OptFn {
	return base.WithoutValues(FieldKeyOperator, operator...)
}

// WithOperateTimeRange filters by operate-time.
func WithOperateTimeRange(timeRange types.TimeRange) OptFn {
	return base.WithTimeRange(FieldKeyOperateTime, timeRange.StartTime, timeRange.EndTime)
}
