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
	attributeKeyState      = "state"

	spanEventActionWorkerStarted       = "action.worker.started"
	spanEventActionDataFetching        = "action.data.fetching"
	spanEventActionDataFetched         = "action.data.fetched"
	spanEventActionSkipped             = "action.skipped"
	spanEventOperationInstanceStarting = "operation_instance.starting"
	spanEventOperationInstanceStarted  = "operation_instance.started"
	spanEventActionLifecycleStarting   = "action.lifecycle.starting"
	spanEventActionLifecycleStarted    = "action.lifecycle.started"
	spanEventActionExecuting           = "action.executing"
	spanEventActionExecutedSuccess     = "action.executed.success"
	spanEventActionExecutedFailed      = "action.executed.failed"
	spanEventActionLifecycleEnding     = "action.lifecycle.ending"
	spanEventActionLifecycleEnded      = "action.lifecycle.ended"
	spanEventOperationInstanceEnding   = "operation_instance.ending"
	spanEventOperationInstanceEnded    = "operation_instance.ended"
)
