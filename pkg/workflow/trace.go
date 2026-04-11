/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

const (
	scopeNameTrigger      = "trigger"
	spanNamePrefixTrigger = "trigger"

	scopeNameOperationInstance       = "operation_instance"
	scopeNamePrefixOperationInstance = "operation_instance"

	scopeNameAction       = "action"
	scopeNamePrefixAction = "action"

	attributeKeyTriggerID       = "trigger_id"
	attributeKeyTriggerCategory = "trigger_category"

	attributeKeyOperationID = "operation_id"

	attributeKeyOperationInstanceID = "operation_instance_id"
	attributeKeyOperationDefName    = "operation_def_name"

	attributeKeyActionName = "action_name"

	attributeKeySkipReason = "skip_reason"
	attributeKeyError      = "error"
	attributeKeyRetryCount = "retry_count"
	attributeKeyIsFirst    = "is_first"
	attributeKeyIsLast     = "is_last"
	attributeKeyFinalState = "final_state"

	spanEventActionReceived  = "action.received"
	spanEventActionSkipped   = "action.skipped"
	spanEventActionStarted   = "action.started"
	spanEventActionFailed    = "action.failed"
	spanEventActionSucceeded = "action.succeeded"
	spanEventActionCompleted = "action.completed"
)
