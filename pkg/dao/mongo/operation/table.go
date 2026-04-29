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

// TableName table name.
const TableName = "operation"

var _ base.IData = &Operation{}

// Operation represents an operation.
// OperationID should be the unique key.
type Operation struct {
	OperationID         string         `json:"operation_id" bson:"operation_id"`
	TriggerID           string         `json:"trigger_id" bson:"trigger_id"`
	OperInstIDs         []string       `json:"oper_inst_ids" bson:"oper_inst_ids"`
	DefSnapshot         DefSnapshot    `json:"def_snapshot" bson:"def_snapshot"`
	Parameters          Parameters     `json:"parameters" bson:"parameters"`
	RetryFlags          []RetryFlag    `json:"retry_flags" bson:"retry_flags"`
	Instantiated        bool           `json:"instantiated" bson:"instantiated"`
	LatestInstBriefData *InstBriefData `json:"latest_inst_brief_data" bson:"latest_inst_brief_data"`
	CreateTime          time.Time      `json:"create_time" bson:"create_time"`
}

// DefSnapshot represents the snapshot of the operation definition.
type DefSnapshot struct {
	OperDefName        string     `json:"oper_def_name" bson:"oper_def_name"`
	ActionNames        []string   `json:"action_names" bson:"action_names"`
	DefaultParameters  Parameters `json:"default_parameters" bson:"default_parameters"`
	ExtraExecutionName string     `json:"extra_execution_name" bson:"extra_execution_name"`
}

// Parameters represents the snapshot of the operation definition.
type Parameters struct {
	ParentOperationID string          `json:"parent_operation_id" bson:"parent_operation_id"`
	ParentOperInstID  string          `json:"parent_oper_inst_id" bson:"parent_oper_inst_id"`
	Timeout           time.Duration   `json:"timeout" bson:"timeout"`
	InitContent       map[string]any  `json:"init_content" bson:"init_content"`
	RetryStartPoint   map[string]bool `json:"retry_start_point" bson:"retry_start_point" `
}

// RetryFlag represents the retry flag of the operation.
type RetryFlag struct {
	Mode             string `json:"mode" bson:"mode"`
	SourceInstanceID string `json:"source_instance_id" bson:"source_instance_id"`
	RetryInstanceID  string `json:"retry_instance_id" bson:"retry_instance_id"`
}

// InstBriefData represents the brief data of an operation instance.
type InstBriefData struct {
	OperInstID                string               `json:"oper_inst_id" bson:"oper_inst_id"`
	LifeCycle                 *LifeCycle           `json:"life_cycle" bson:"life_cycle"`
	LatestActionInstBriefData *ActionInstBriefData `json:"latest_action_inst_brief_data" bson:"latest_action_inst_brief_data"`
}

// ActionInstBriefData represents the brief data of an action instance.
type ActionInstBriefData struct {
	Name string   `json:"name" bson:"name"`
	Tags []string `json:"tags" bson:"tags"`
}

// LifeCycle is the lifecycle of an operation instance.
type LifeCycle struct {
	State     string    `json:"state" bson:"state"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	StartedAt time.Time `json:"started_at" bson:"started_at"`
	EndedAt   time.Time `json:"ended_at" bson:"ended_at"`
	StoppedAt time.Time `json:"stopped_at" bson:"stopped_at"`
}

// UniqueFields unique fields of the table.
func (oper *Operation) UniqueFields() []string {
	return []string{FieldKeyOperationID}
}

// UniqueKey unique key of the table.
func (oper *Operation) UniqueKey() string {
	return oper.OperationID
}

// TableOperation represents the complete db structures of an operation.
type TableOperation base.TableBroker[*Operation]
