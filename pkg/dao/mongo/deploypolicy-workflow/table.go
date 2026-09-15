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

package deploypolicyworkflow

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName returns the tenant's deploy policy workflow collection name.
func TableName(tenantID string) string {
	return "deploypolicy_workflow_" + tenantID
}

// Data is the persisted deploy policy workflow record.
type Data struct {
	TenantID       string    `json:"tenant_id" bson:"tenant_id"`
	WorkflowID     string    `json:"workflow_id" bson:"workflow_id"`
	OperationID    string    `json:"operation_id" bson:"operation_id"`
	TriggerID      string    `json:"trigger_id" bson:"trigger_id"`
	DeployPolicyID int64     `json:"deploy_policy_id" bson:"deploy_policy_id"`
	Operator       string    `json:"operator" bson:"operator"`
	OperateTime    time.Time `json:"operate_time" bson:"operate_time"`
	FinishTime     time.Time `json:"finish_time" bson:"finish_time"`
	Status         string    `json:"status" bson:"status"`
	Children       []Child   `json:"children" bson:"children"`
}

// Child is an embedded child workflow reference.
type Child struct {
	WorkflowID     string `json:"workflow_id" bson:"workflow_id"`
	WorkflowDomain string `json:"workflow_domain" bson:"workflow_domain"`
}

// UniqueFields returns the stable workflow identity fields.
func (*Data) UniqueFields() []string {
	return []string{FieldKeyWorkflowID}
}

// UniqueKey returns the stable workflow identity.
func (data *Data) UniqueKey() string {
	return data.WorkflowID
}

// Table wraps the persisted record in the common document envelope.
type Table base.TableBroker[*Data]
