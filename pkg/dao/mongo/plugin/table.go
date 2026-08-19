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

package plugin

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName plugin table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("plugin_%s", tenantID)
}

var _ base.IData = &Plugin{}

// Plugin represents the table of plugin deployment.
// Token should be the unique key.
type Plugin struct {
	TenantID      string  `json:"tenant_id" bson:"tenant_id"`
	Name          string  `json:"name" bson:"name"`
	Group         string  `json:"group" bson:"group"`
	PkgName       string  `json:"pkg_name" bson:"pkg_name"`
	Memo          string  `json:"memo" bson:"memo"`
	VisibleBizIDs []int64 `json:"visible_biz_ids" bson:"visible_biz_ids"`
}

// UniqueFields unique fields of the table.
func (deploy *Plugin) UniqueFields() []string {
	return []string{FieldKeyName}
}

// UniqueKey unique key of the table.
func (deploy *Plugin) UniqueKey() string {
	return deploy.Name
}

// Table represent the complete db structures of plugin deployment.
type Table base.TableBroker[*Plugin]
