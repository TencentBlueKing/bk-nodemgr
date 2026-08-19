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

package configpolicy

import (
	"strconv"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

const tableNamePrefix = "configpolicy"
const counterKeyPriority = tableNamePrefix + "_priority"

// TableName config policy table name.
func TableName(tenantID string) string {
	return tableNamePrefix + "_" + tenantID
}

var _ base.IData = &ConfigPolicy{}

// RawData presents a config policy without version.
type RawData struct {
	TenantID         string         `json:"tenant_id" bson:"tenant_id"`
	ConfigPolicyID   int64          `json:"configpolicy_id" bson:"configpolicy_id"`
	ConfigPolicyName string         `json:"configpolicy_name" bson:"configpolicy_name"`
	ConfigPolicyType string         `json:"configpolicy_type" bson:"configpolicy_type"`
	BizID            int64          `json:"biz_id" bson:"biz_id"`
	Remark           string         `json:"remark" bson:"remark"`
	Scopes           []Scope        `json:"scopes" bson:"scopes"`
	TargetHostIDs    []int64        `json:"target_host_ids" bson:"target_host_ids"`
	TargetPluginName string         `json:"target_plugin_name" bson:"target_plugin_name"`
	Configs          map[string]any `json:"configs" bson:"configs"`
	Enabled          bool           `json:"enabled" bson:"enabled"`
	Priority         int64          `json:"priority" bson:"priority"`
	UpdatedAt        time.Time      `json:"updated_at" bson:"updated_at"`
	Operator         string         `json:"operator" bson:"operator"`
}

// ConfigPolicy presents a config policy.
type ConfigPolicy struct {
	Raw RawData `json:"raw" bson:"raw"`

	Version int `json:"version" bson:"version"`
}

// UniqueFields unique fields of the table.
func (r *ConfigPolicy) UniqueFields() []string {
	return []string{FieldKeyConfigPolicyID}
}

// Scope presents a single scope for config policy.
type Scope struct {
	NetworkAreaID int64  `json:"networkarea_id" bson:"networkarea_id"`
	NetworkUnitID int64  `json:"networkunit_id" bson:"networkunit_id"`
	NodeOsType    string `json:"node_os_type" bson:"node_os_type"`
	NodeCPUArch   string `json:"node_cpu_arch" bson:"node_cpu_arch"`
}

// UniqueKey unique key of the table.
func (r *ConfigPolicy) UniqueKey() string {
	return strconv.FormatInt(r.Raw.ConfigPolicyID, 10)
}

// TableConfigPolicy represents the complete db structures.
type TableConfigPolicy base.TableBroker[*ConfigPolicy]
