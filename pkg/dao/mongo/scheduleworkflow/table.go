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

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName schedule workflow table name.
func TableName() string {
	return "schedule_workflow"
}

// ScheduleWorkflow represents the table of schedule workflow.
// Token should be the unique key.
type ScheduleWorkflow struct {
	WorkflowID   string    `json:"workflow_id" bson:"workflow_id"`
	WorkflowName string    `json:"workflow_name" bson:"workflow_name"`
	TriggerID    string    `json:"trigger_id" bson:"trigger_id"`
	Operator     string    `json:"operator" bson:"operator"`
	OperateTime  time.Time `json:"operate_time" bson:"operate_time"`
}

// UniqueKey unique key of the table.
func (workflow *ScheduleWorkflow) UniqueKey() string {
	return workflow.WorkflowID
}

// TableScheduleWorkflow represent the complete db structures of schedule workflow.
type TableScheduleWorkflow base.TableBroker[*ScheduleWorkflow]
