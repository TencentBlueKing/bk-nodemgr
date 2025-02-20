/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operinstdata ...
package operinstdata

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName operation instance data table name.
const TableName = "oper_inst_data"

// ActionInstData represents a action data.
type ActionInstData struct {
	TriggerID  string            `json:"trigger_id" bson:"trigger_id"`
	OperInstID string            `json:"oper_inst_id" bson:"oper_inst_id"`
	Name       string            `json:"name" bson:"name"`
	Index      int               `json:"index" bson:"index"`
	Lifecycle  *ActInstLifeCycle `json:"life_cycle" bson:"life_cycle"`
	Messages   []Message         `json:"messages" bson:"messages"`
	Content    string            `json:"content" bson:"content"`
}

// Message represents a message.
type Message struct {
	Time time.Time `json:"time" bson:"time"`
	Text string    `json:"text" bson:"text"`
}

// ActInstLifeCycle is the lifecycle of an action instance.
type ActInstLifeCycle struct {
	State     string    `json:"state" bson:"state"`
	StartedAt time.Time `json:"started_at" bson:"started_at"`
	EndedAt   time.Time `json:"ended_at" bson:"ended_at"`
	StoppedAt time.Time `json:"stopped_at" bson:"stopped_at"`
}

// OperInstData represents a operation instance data.
type OperInstData struct {
	TriggerID         string                     `json:"trigger_id" bson:"trigger_id"`
	OperInstID        string                     `json:"oper_inst_id" bson:"oper_inst_id"`
	ActionNames       []string                   `json:"actions" bson:"actions"`
	ActionInstDataMap map[string]*ActionInstData `json:"action_data" bson:"action_data"`
	OperDefName       string                     `json:"oper_def_name" bson:"oper_def_name"`
	ParentOperInstID  string                     `json:"parent_oper_inst_id" bson:"parent_oper_inst_id"`
	Timeout           time.Duration              `json:"timeout" bson:"timeout"`
	InitContent       string                     `json:"init_content" bson:"init_content"`
	Lifecycle         *Lifecycle                 `json:"lifecycle" bson:"lifecycle"`
}

// Lifecycle is the lifecycle of an operation instance.
type Lifecycle struct {
	State     string    `json:"state" bson:"state"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	StartedAt time.Time `json:"started_at" bson:"started_at"`
	EndedAt   time.Time `json:"ended_at" bson:"ended_at"`
	StoppedAt time.Time `json:"stopped_at" bson:"stopped_at"`
}

// UniqueKey unique key of the table.
func (data *OperInstData) UniqueKey() string {
	return data.OperInstID
}

// TableOperInstData represents the complete db structures of an operation instance data.
type TableOperInstData base.TableBroker[*OperInstData]
