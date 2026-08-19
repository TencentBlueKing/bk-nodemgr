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

package configpolicyevent

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

const tableNamePrefix = "configpolicyevent"

// TableName policy event table name.
func TableName(tenantID string) string {
	return fmt.Sprintf("%s_%s", tableNamePrefix, tenantID)
}

var _ base.IData = &ConfigPolicyEvent{}

// ConfigPolicyEvent represents config policy event table.
type ConfigPolicyEvent struct {
	EventID          int64     `json:"event_id" bson:"event_id"`
	TenantID         string    `json:"tenant_id" bson:"tenant_id"`
	BizID            int64     `json:"biz_id" bson:"biz_id"`
	ConfigpolicyID   int64     `json:"configpolicy_id" bson:"configpolicy_id"`
	ConfigpolicyName string    `json:"configpolicy_name" bson:"configpolicy_name"`
	ConfigpolicyType string    `json:"configpolicy_type" bson:"configpolicy_type"`
	Type             string    `json:"type" bson:"type"`
	Version          int64     `json:"version" bson:"version"`
	OperateTime      time.Time `json:"operate_time" bson:"operate_time"`
	Operator         string    `json:"operator" bson:"operator"`
}

// UniqueFields unique fields of the table.
func (event *ConfigPolicyEvent) UniqueFields() []string {
	return []string{
		FieldKeyEventID,
	}
}

// UniqueKey unique key of the table.
func (event *ConfigPolicyEvent) UniqueKey() string {
	return fmt.Sprintf("%d", event.EventID)
}

// TableConfigPolicyEvent represents the complete db structures.
type TableConfigPolicyEvent base.TableBroker[*ConfigPolicyEvent]
