/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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
	TenantID      string `json:"tenant_id" bson:"tenant_id"`
	PluginID      string `json:"plugin_id" bson:"plugin_id"`
	Name          string `json:"name" bson:"name"`
	Group         string `json:"group" bson:"group"`
	PluginPkgName string `json:"plugin_pkg_name" bson:"plugin_pkg_name"`
}

// Platform defines the Platform.
type Platform struct {
	OS   string `json:"os" bson:"os"`
	Arch string `json:"arch" bson:"arch"`
}

// UniqueFields unique fields of the table.
func (deploy *Plugin) UniqueFields() []string {
	return []string{FieldKeyPluginID}
}

// UniqueKey unique key of the table.
func (deploy *Plugin) UniqueKey() string {
	return deploy.PluginID
}

// Table represent the complete db structures of plugin deployment.
type Table base.TableBroker[*Plugin]
