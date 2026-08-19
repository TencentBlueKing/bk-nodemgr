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

package nodeworkflow

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName node workflow table name.
func TableName(tenantID string) string {
	return "node_workflow_" + tenantID
}

var _ base.IData = &Data{}

// Data represents the table of node workflow.
// Token should be the unique key.
type Data struct {
	TenantID        string    `json:"tenant_id" bson:"tenant_id"`
	WorkflowID      string    `json:"workflow_id" bson:"workflow_id"`
	TriggerID       string    `json:"trigger_id" bson:"trigger_id"`
	Type            string    `json:"type" bson:"type"`
	BizIDs          []int64   `json:"biz_ids" bson:"biz_ids"`
	NetworkAreaIDs  []int64   `json:"networkarea_ids" bson:"networkarea_ids"`
	NetworkUnitIDs  []int64   `json:"networkunit_ids" bson:"networkunit_ids"`
	NodeRoles       []string  `json:"node_roles" bson:"node_roles"`
	Operator        string    `json:"operator" bson:"operator"`
	DeployPolicyIDs []int64   `json:"deploy_policy_ids" bson:"deploy_policy_ids"`
	OperateTime     time.Time `json:"operate_time" bson:"operate_time"`
	FinishTime      time.Time `json:"finish_time" bson:"finish_time"`
	Status          string    `json:"status" bson:"status"`
}

// UniqueFields unique fields of the table.
func (workflow *Data) UniqueFields() []string {
	return []string{FieldKeyWorkflowID}
}

// UniqueKey unique key of the table.
func (workflow *Data) UniqueKey() string {
	return workflow.WorkflowID
}

// Table represent the complete db structures of node workflow.
type Table base.TableBroker[*Data]
