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

package deploypolicyworkflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithWorkflowID filters by workflow ID.
func WithWorkflowID(workflowIDs ...string) OptFn {
	return base.WithValues(FieldKeyWorkflowID, workflowIDs...)
}

// WithOperationID filters by operation ID.
func WithOperationID(operationIDs ...string) OptFn {
	return base.WithValues(FieldKeyOperationID, operationIDs...)
}

// WithDeployPolicyID filters by deploy policy ID.
func WithDeployPolicyID(deployPolicyIDs ...int64) OptFn {
	return base.WithValues(FieldKeyDeployPolicyID, deployPolicyIDs...)
}

// WithoutWorkflowID excludes workflow IDs.
func WithoutWorkflowID(workflowIDs ...string) OptFn {
	return base.WithoutValues(FieldKeyWorkflowID, workflowIDs...)
}

// WithoutDeployPolicyID excludes deploy policy IDs.
func WithoutDeployPolicyID(deployPolicyIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyDeployPolicyID, deployPolicyIDs...)
}

// WithStatus filters by workflow status.
func WithStatus(statuses ...types.DeployPolicyWorkflowStatus) OptFn {
	return base.WithValues(FieldKeyStatus, statuses...)
}

// WithoutStatus excludes workflow statuses.
func WithoutStatus(statuses ...types.DeployPolicyWorkflowStatus) OptFn {
	return base.WithoutValues(FieldKeyStatus, statuses...)
}

// WithOperator filters by operator.
func WithOperator(operators ...string) OptFn {
	return base.WithValues(FieldKeyOperator, operators...)
}

// WithoutOperator excludes operators.
func WithoutOperator(operators ...string) OptFn {
	return base.WithoutValues(FieldKeyOperator, operators...)
}

// WithOperateTimeRange filters by the inclusive operation time range.
func WithOperateTimeRange(timeRange types.TimeRange) OptFn {
	return base.WithTimeRange(FieldKeyOperateTime, timeRange.StartTime, timeRange.EndTime)
}

// WithRunningOrMissingStatus includes legacy records for status backfill by the monitor.
func WithRunningOrMissingStatus() OptFn {
	return func(filter bson.D) bson.D {
		return append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.M{FieldKeyStatus: types.DeployPolicyWorkflowStatusRunning},
			bson.M{FieldKeyStatus: ""},
			bson.M{FieldKeyStatus: bson.M{"$exists": false}},
		}})
	}
}
