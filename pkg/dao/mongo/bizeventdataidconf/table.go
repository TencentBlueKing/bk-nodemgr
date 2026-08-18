/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package bizeventdataidconf provides storage for business event data-id configs.
package bizeventdataidconf

import (
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName returns the business event data-id config table name for a tenant.
func TableName(tenantID string) string {
	return "biz_event_data_id_conf_" + tenantID
}

var _ base.IData = &BizEventDataIDConf{}

// BizEventDataIDConf represents the business event data-id config table.
type BizEventDataIDConf struct {
	BizID                     int64  `bson:"biz_id" json:"bk_biz_id"`
	AgentBaseAlarmEventDataID *int64 `bson:"agent_base_alarm_event_data_id,omitempty" json:"agent_base_alarm_event_data_id,omitempty"`
	TaskProcEventDataID       *int64 `bson:"task_proc_event_data_id,omitempty" json:"task_proc_event_data_id,omitempty"`
}

// UniqueFields returns unique fields of the table.
func (conf *BizEventDataIDConf) UniqueFields() []string {
	return []string{FieldKeyBizID}
}

// UniqueKey returns unique key of the table.
func (conf *BizEventDataIDConf) UniqueKey() string {
	return strconv.FormatInt(conf.BizID, 10)
}

// TableBizEventDataIDConf represents the complete db structure.
type TableBizEventDataIDConf base.TableBroker[*BizEventDataIDConf]
