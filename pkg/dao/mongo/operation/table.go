/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation ...
package operation

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName  tenantID table name.
func TableName(tenantID string) string {
	return "operation_" + tenantID
}

// Operation represents an operation under a tenant.
// OperationID should be the unique key.
type Operation struct {
	OperationID string      `json:"operation_id" bson:"operation_id"`
	TriggerID   string      `json:"trigger_id" bson:"trigger_id"`
	OperInstIDs []string    `json:"oper_inst_ids" bson:"oper_inst_ids"`
	DefSnapshot DefSnapshot `json:"def_snapshot" bson:"def_snapshot"`

	Parameters Parameters `json:"parameters" bson:"parameters"`
}

// DefSnapshot represents the snapshot of the operation definition.
type DefSnapshot struct {
	OperDefName       string     `json:"oper_def_name" bson:"oper_def_name"`
	ActionNames       []string   `json:"action_names" bson:"action_names"`
	DefaultParameters Parameters `json:"default_parameters" bson:"default_parameters"`
}

// Parameters represents the snapshot of the operation definition.
type Parameters struct {
	ParentOperationID string         `json:"parent_operation_id" bson:"parent_operation_id"`
	Timeout           time.Duration  `json:"timeout" bson:"timeout"`
	InitContent       map[string]any `json:"init_content" bson:"init_content"`
}

// UniqueKey unique key of the table.
func (oper *Operation) UniqueKey() string {
	return oper.OperationID
}

// TableOperation represents the complete db structures of an operation.
type TableOperation base.TableBroker[*Operation]
