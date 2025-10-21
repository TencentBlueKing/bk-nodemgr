/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed on the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tenant

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName tenant table name.
func TableName() string {
	return "tenant"
}

var _ base.IData = &Tenant{}

// Tenant represents the table of tenant.
// ID should be the unique key.
type Tenant struct {
	ID      string `json:"id" bson:"id"`
	Name    string `json:"name" bson:"name"`
	Enabled bool   `json:"enabled" bson:"enabled"`
}

// UniqueFields unique fields of the table.
func (t *Tenant) UniqueFields() []string {
	return []string{FieldKeyID}
}

// UniqueKey unique key of the table.
func (t *Tenant) UniqueKey() string {
	return t.ID
}

// Table represent the complete db structures of tenant.
type Table base.TableBroker[*Tenant]
