/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package schedule provides the operation definition for scheduling host synchronization.
package schedule

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameScheduleOnceTriggerOperation defines the operation definition name.
const OperDefNameScheduleOnceTriggerOperation = "schedule_once_trigger_operation"

// NewOperScheduleOnceTriggerOperation new an operation.
func NewOperScheduleOnceTriggerOperation(
	param OperParamScheduleOnceTriggerOperation, actName string) operation.Definition {

	return &operScheduleOnceTriggerOperation{
		param:   param,
		actName: actName,
	}
}

// operScheduleOnceTriggerOperation implements the operation.Definition interface.
type operScheduleOnceTriggerOperation struct {
	param   OperParamScheduleOnceTriggerOperation
	actName string
}

// OperParamScheduleOnceTriggerOperation defines the parameters for operScheduleOnceTriggerOperation.
type OperParamScheduleOnceTriggerOperation struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operScheduleOnceTriggerOperation) Name() string {
	return OperDefNameScheduleOnceTriggerOperation
}

// ActionDefNames returns the action def names.
func (oper *operScheduleOnceTriggerOperation) ActionDefNames() []string {
	return []string{
		fmt.Sprintf(ActionNameGenScheduleOnceTrigger, oper.actName),
	}
}

// DefaultParameters returns the default parameters.
func (oper *operScheduleOnceTriggerOperation) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operScheduleOnceTriggerOperation) ExtraExecutionName() string {
	return ""
}
