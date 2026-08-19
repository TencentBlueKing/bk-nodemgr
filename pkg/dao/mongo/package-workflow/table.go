/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package packageworkflow

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName returns the package workflow table name for a tenant.
func TableName(tenantID string) string {
	return "package_workflow_" + tenantID
}

var _ base.IData = &Data{}

// Data represents the table of package workflow.
type Data struct {
	TenantID    string    `json:"tenant_id" bson:"tenant_id"`
	WorkflowID  string    `json:"workflow_id" bson:"workflow_id"`
	TriggerID   string    `json:"trigger_id" bson:"trigger_id"`
	Type        string    `json:"type" bson:"type"`
	Operator    string    `json:"operator" bson:"operator"`
	OperateTime time.Time `json:"operate_time" bson:"operate_time"`
	FinishTime  time.Time `json:"finish_time" bson:"finish_time"`
	Status      string    `json:"status" bson:"status"`
}

// UniqueFields returns unique fields of the table.
func (workflow *Data) UniqueFields() []string {
	return []string{FieldKeyWorkflowID}
}

// UniqueKey returns the unique key of the table.
func (workflow *Data) UniqueKey() string {
	return workflow.WorkflowID
}

// Table represents the complete db structures of package workflow.
type Table base.TableBroker[*Data]
