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

// WithTaskID filters by task-id.
func WithTaskID(taskIDs ...int64) OptFn {
	return base.WithValues(FieldKeyTaskID, taskIDs...)
}

// WithStatus filters by status.
func WithStatus(statuses ...types.NodeWorkflowStatus) OptFn {
	strs := make([]string, len(statuses))
	for idx, status := range statuses {
		strs[idx] = string(status)
	}

	return base.WithValues(FieldKeyStatus, strs...)
}

// WithOperType filters by oper-type.
func WithOperType(operTypes ...types.NodeWorkflowOperType) OptFn {
	strs := make([]string, len(operTypes))
	for idx, operType := range operTypes {
		strs[idx] = string(operType)
	}

	return base.WithValues(FieldKeyOperType, strs...)
}

// WithTaskType filters by task-type.
func WithTaskType(taskTypes ...types.NodeWorkflowTaskType) OptFn {
	strs := make([]string, len(taskTypes))
	for idx, taskType := range taskTypes {
		strs[idx] = string(taskType)
	}

	return base.WithValues(FieldKeyTaskType, strs...)
}

// WithBizID filters by biz-id.
func WithBizID(bizIDs ...int64) OptFn {
	return base.WithValues(FieldKeyBizID, bizIDs...)
}

// WithExecuteUser filters by execute-user.
func WithExecuteUser(executeUsers ...string) OptFn {
	return base.WithValues(FieldKeyExecuteUser, executeUsers...)
}
