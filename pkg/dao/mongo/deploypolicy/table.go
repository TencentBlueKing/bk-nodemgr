/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deploypolicy

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

const tableNamePrefix = "deploypolicy"

// TableName deploy policy table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("%s_%s", tableNamePrefix, tenantID)
}

var _ base.IData = &DeployPolicy{}

// DeployPolicy represents the table of deploy policy deployment.
// DeployPolicyID should be the unique key.
type DeployPolicy struct {
	TenantID       string    `json:"tenant_id" bson:"tenant_id"`
	DeployPolicyID int64     `json:"deploy_policy_id" bson:"deploy_policy_id"`
	Meta           Meta      `json:"meta" bson:"meta"`
	Specs          []*Spec   `json:"specs" bson:"specs"`
	Scopes         []*Scope  `json:"scopes" bson:"scopes"`
	Enabled        bool      `json:"enabled" bson:"enabled"`
	LifeCycle      LifeCycle `json:"life_cycle" bson:"life_cycle"`
	Operator       string    `json:"operator" bson:"operator"`
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

// Meta represents the meta of deploy policy.
type Meta struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
}

// Spec represents the spec of deploy policy.
type Spec struct {
	Type  string         `json:"type" bson:"type"`
	Param map[string]any `json:"param" bson:"param"`
}

// Scope represents the scope of deploy policy.
type Scope struct {
	BizID       int64            `json:"biz_id" bson:"biz_id"`
	Type        string           `json:"type" bson:"type"`
	Granularity string           `json:"granularity" bson:"granularity"`
	Filter      *TargetFilter    `json:"filter" bson:"filter"`
	Items       []map[string]any `json:"items" bson:"items"`
}

// LifeCycle represents the life cycle of deploy policy.
type LifeCycle struct {
	CreateAt time.Time `json:"create_at" bson:"create_at"`
	UpdateAt time.Time `json:"update_at" bson:"update_at"`
}

// TargetFilter represents the target filter of deploy policy.
type TargetFilter struct {
}
