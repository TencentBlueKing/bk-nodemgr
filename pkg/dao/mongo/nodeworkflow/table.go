/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeworkflow

import (
	"strconv"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName node workflow table name.
const TableName = "node_workflow"

// Data represents the table of node workflow.
// Token should be the unique key.
type Data struct {
	WorkflowID  int64     `json:"workflow_id" bson:"workflow_id"`
	TriggerID   string    `json:"trigger_id" bson:"trigger_id"`
	Type        string    `json:"type" bson:"type"`
	BizIDs      []int64   `json:"biz_ids" bson:"biz_ids"`
	ExecuteUser string    `json:"execute_user" bson:"execute_user"`
	ExecuteTime time.Time `json:"execute_time" bson:"execute_time"`
	Status      string    `json:"status" bson:"status"`
}

// UniqueKey unique key of the table.
func (workflow *Data) UniqueKey() string {
	return strconv.FormatInt(workflow.WorkflowID, 10)
}

// Table represent the complete db structures of node workflow.
type Table base.TableBroker[*Data]
