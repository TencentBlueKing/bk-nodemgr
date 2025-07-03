/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package scheduleworkflow provides storage for schedule workflow.
package scheduleworkflow

const (
	// FieldKeyWorkflowID is the key for workflow ID.
	FieldKeyWorkflowID = "data.workflow_id"

	// FieldKeyWorkflowName is the key for workflow type.
	FieldKeyWorkflowName = "data.workflow_name"

	// FieldKeyTriggerID is the key for status.
	FieldKeyTriggerID = "data.trigger_id"

	// FieldKeyOperator is the key for operator.
	FieldKeyOperator = "data.operator"

	// FieldKeyOperateTime the operate time field key.
	FieldKeyOperateTime = "data.operate_time"
)
