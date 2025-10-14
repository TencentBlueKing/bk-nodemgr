/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package scheduledworkflow provides storage for schedule workflow.
package scheduledworkflow

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

// WithoutWorkflowID filters by no contains workflow ID.
func WithoutWorkflowID(workflowIDs ...string) OptFn {
	return base.WithoutValues(FieldKeyWorkflowID, workflowIDs...)
}

// WithTriggerID filters by trigger ID.
func WithTriggerID(triggerIDs ...string) OptFn {
	return base.WithValues(string(FieldKeyTriggerID), triggerIDs...)
}

// WithoutTriggerID filters by no contains trigger ID.
func WithoutTriggerID(triggerIDs ...string) OptFn {
	return base.WithoutValues(string(FieldKeyTriggerID), triggerIDs...)
}

// WithWorkflowName filters by workflow name.
func WithWorkflowName(scheduleWorkflowName ...string) OptFn {
	return base.WithValues(FieldKeyWorkflowName, scheduleWorkflowName...)
}

// WithoutWorkflowName filters by not contains workflow name.
func WithoutWorkflowName(scheduleWorkflowName ...string) OptFn {
	return base.WithoutValues(FieldKeyWorkflowName, scheduleWorkflowName...)
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
