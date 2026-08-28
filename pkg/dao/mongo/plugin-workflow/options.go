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

package pluginworkflow

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

// WithoutWorkflowID filters by not contains workflow ID.
func WithoutWorkflowID(workflowIDs ...string) OptFn {
	return base.WithoutValues(FieldKeyWorkflowID, workflowIDs...)
}

// WithStatus filters by status.
func WithStatus(statuses ...types.PluginWorkflowStatus) OptFn {
	return base.WithValues(FieldKeyStatus, statuses...)
}

// WithoutStatus filters by not contains status.
func WithoutStatus(statuses ...types.PluginWorkflowStatus) OptFn {
	return base.WithoutValues(FieldKeyStatus, statuses...)
}

// WithType filters by type.
func WithType(pluginWorkflowTypes ...types.PluginWorkflowType) OptFn {
	return base.WithValues(FieldKeyType, pluginWorkflowTypes...)
}

// WithoutType filters by not contains type.
func WithoutType(pluginWorkflowTypes ...types.PluginWorkflowType) OptFn {
	return base.WithoutValues(FieldKeyType, pluginWorkflowTypes...)
}

// WithHostIDs filters by host-id.
func WithHostIDs(hostIDs ...int64) OptFn {
	return base.WithValues(FieldKeyHostIDs, hostIDs...)
}

// WithoutHostIDs filters by not contains host-id.
func WithoutHostIDs(hostID ...int64) OptFn {
	return base.WithoutValues(FieldKeyHostIDs, hostID...)
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

// WithTriggerID filters by trigger ID.
func WithTriggerID(triggerIDs ...string) OptFn {
	return base.WithValues(FieldKeyTriggerID, triggerIDs...)
}

// WithBizIDs filters by biz-id.
func WithBizIDs(bizIDs ...int64) OptFn {
	return base.WithValues(FieldKeyBizIDs, bizIDs...)
}

// WithoutBizIDs filters by not contains biz-id.
func WithoutBizIDs(bizIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyBizIDs, bizIDs...)
}

// WithDeployPolicyID filters by deploy policy ID.
func WithDeployPolicyID(deployPolicyIDs ...int64) OptFn {
	return base.WithValues(FieldKeyDeployPolicyIDs, deployPolicyIDs...)
}

// WithoutDeployPolicyID filters by not contains deploy policy ID.
func WithoutDeployPolicyID(deployPolicyIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyDeployPolicyIDs, deployPolicyIDs...)
}
