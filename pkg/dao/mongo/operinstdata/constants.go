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

import "fmt"

// Field is the field name in the oper inst data
type Field string

// String returns the string representation of the field
func (field Field) String() string {
	return string(field)
}

// Validate checks if the field is valid
func (field Field) Validate() error {
	switch field {
	case FieldOperInstID:
		return nil

	default:
		return fmt.Errorf("invalid field, field(%s)", field)
	}
}

const (
	// FieldOperInstID is the field name for oper inst id
	FieldOperInstID   Field = "data.oper_inst_id"
	FieldKeyLifeCycle Field = "data.life_cycle"
)

// FieldKeyActionInstState is the field name for action instance state
func FieldKeyActionInstState(actionName string) Field {
	return Field("data.action_data." + actionName + ".life_cycle.state")
}

// FieldKeyActInstPrivateData is the field name for action instance private data
func FieldKeyActInstPrivateData(actionName string) Field {
	return Field("data.action_data." + actionName + ".private_data")
}

// FieldKeyActionInstLifeCycle is the field name for action instance life cycle
func FieldKeyActionInstLifeCycle(actionName string) Field {
	return Field("data.action_data." + actionName + ".life_cycle")
}

// FieldKeyActionInstMessages is the field name for action instance messages
func FieldKeyActionInstMessages(actionName string) Field {
	return Field("data.action_data." + actionName + ".messages")
}

// FieldKeyActionInstContent is the field name for action instance content.
func FieldKeyActionInstContent(actionName string) Field {
	return Field("data.action_data." + actionName + ".content")
}

// FieldKeyActionInstData is the field name for action instance data
func FieldKeyActionInstData(actionName string) Field {
	return Field("data.action_data." + actionName)
}
