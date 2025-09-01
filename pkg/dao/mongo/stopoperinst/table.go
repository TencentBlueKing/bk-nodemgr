/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package stopoperinst ...
package stopoperinst

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName host table name.
const TableName = "stopping_operation_inst"

var _ base.IData = &StopOperInst{}

// StopOperInst represents a task engine stopping task.
type StopOperInst struct {
	OperInstID string    `json:"oper_inst_id" bson:"oper_inst_id"`
	ExpireAt   time.Time `json:"expire_at" bson:"expire_at"`
}

// UniqueFields unique fields of the table.
func (inst *StopOperInst) UniqueFields() []string {
	return []string{FieldKeyOperInstID}
}

// UniqueKey unique key of the table.
func (inst *StopOperInst) UniqueKey() string {
	return inst.OperInstID
}

// TableStopOperInst represents the complete db structures of a stopping task.
type TableStopOperInst base.TableBroker[*StopOperInst]

// TableStopOperInstChangeEvent represents the complete db structures of a task data change event.
type TableStopOperInstChangeEvent base.TableChangeEventBroker[*StopOperInst]
