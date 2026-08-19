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

// Package scheduledworkflow provides storage for scheduled workflow.
package scheduledworkflow

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName scheduled workflow table name.
func TableName() string {
	return "scheduled_workflow"
}

var _ base.IData = &ScheduledWorkflow{}

// ScheduledWorkflow represents the table of scheduled workflow.
// Token should be the unique key.
type ScheduledWorkflow struct {
	WorkflowID   string         `json:"workflow_id" bson:"workflow_id"`
	WorkflowName string         `json:"workflow_name" bson:"workflow_name"`
	TenantID     string         `json:"tenant_id" bson:"tenant_id"`
	TriggerID    string         `json:"trigger_id" bson:"trigger_id"`
	Enabled      bool           `json:"enabled" bson:"enabled"`
	Interval     string         `json:"interval" bson:"interval"`
	PrivateData  map[string]any `json:"private_data" bson:"private_data"`
	Operator     string         `json:"operator" bson:"operator"`
	OperateTime  time.Time      `json:"operate_time" bson:"operate_time"`
}

// UniqueFields unique fields of the table.
func (workflow *ScheduledWorkflow) UniqueFields() []string {
	return []string{FieldKeyWorkflowID}
}

// UniqueKey unique key of the table.
func (workflow *ScheduledWorkflow) UniqueKey() string {
	return workflow.WorkflowID
}

// TableScheduledWorkflow represent the complete db structures of scheduled workflow.
type TableScheduledWorkflow base.TableBroker[*ScheduledWorkflow]
