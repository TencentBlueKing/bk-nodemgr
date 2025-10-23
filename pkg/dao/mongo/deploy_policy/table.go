/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deploy_policy

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

const tableNamePrefix = "deploy_policy"

// TableName deploy_policy table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("%s_%s", tableNamePrefix, tenantID)
}

var _ base.IData = &DeployPolicy{}

// DeployPolicy represents the table of deploy policy deployment.
// DeployPolicyID should be the unique key.
type DeployPolicy struct {
	TenantID       string `json:"tenant_id" bson:"tenant_id"`
	DeployPolicyID int64  `json:"deploy_policy_id" bson:"deploy_policy_id"`
}

// UniqueFields unique fields of the table.
func (deploy *DeployPolicy) UniqueFields() []string {
	return []string{FieldKeyDeployPolicyID}
}

// UniqueKey unique key of the table.
func (deploy *DeployPolicy) UniqueKey() string {
	return fmt.Sprintf("%d", deploy.DeployPolicyID)
}

// Table represent the complete db structures of deploy policy deployment.
type Table base.TableBroker[*DeployPolicy]
