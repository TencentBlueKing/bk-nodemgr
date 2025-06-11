/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package operinstdata

const (
	// FieldKeyOperationID is the field name for operation id.
	FieldKeyOperationID = "data.operation_id"

	// FieldKeyOperInstID is the field name for oper inst id.
	FieldKeyOperInstID = "data.oper_inst_id"

	// FieldKeyLifeCycle is the field name for life cycle.
	FieldKeyLifeCycle = "data.life_cycle"

	// FieldKeyTriggerID is the field name for trigger id.
	FieldKeyTriggerID = "data.trigger_id"

	// FieldKeyState is the field name for state.
	FieldKeyState = "data.life_cycle.state"

	// FieldOfActionData is the field name for action data.
	FieldOfActionData = "action_data"
)

// FieldKeyActionInstState is the field name for action instance state.
func FieldKeyActionInstState(actionName string) string {
	return "data.action_data." + actionName + ".life_cycle.state"
}

// FieldKeyActInstPrivateData is the field name for action instance private data.
func FieldKeyActInstPrivateData(actionName string) string {
	return "data.action_data." + actionName + ".private_data"
}

// FieldKeyActionInstLifeCycle is the field name for action instance life cycle.
func FieldKeyActionInstLifeCycle(actionName string) string {
	return "data.action_data." + actionName + ".life_cycle"
}

// FieldKeyActionInstMessages is the field name for action instance messages.
func FieldKeyActionInstMessages(actionName string) string {
	return "data.action_data." + actionName + ".messages"
}

// FieldKeyActionInstContent is the field name for action instance content.
func FieldKeyActionInstContent(actionName string) string {
	return "data.action_data." + actionName + ".content"
}

// FieldKeyActionInstData is the field name for action instance data.
func FieldKeyActionInstData(actionName string) string {
	return "data.action_data." + actionName
}
