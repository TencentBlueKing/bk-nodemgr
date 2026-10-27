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

package nodeopshistory

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// TableName returns the tenant-specific operation history collection.
func TableName(tenantID string) string {
	return "node_ops_history_" + tenantID
}

var _ base.IData = &Data{}

// Data stores one operation and its host snapshots in a single document.
type Data struct {
	TenantID    string         `json:"tenant_id" bson:"tenant_id"`
	OperateID   string         `json:"operate_id" bson:"operate_id"`
	Operator    string         `json:"operator" bson:"operator"`
	OperateTime time.Time      `json:"operate_time" bson:"operate_time"`
	NodeRole    string         `json:"node_role" bson:"node_role"`
	BizIDs      []int64        `json:"biz_ids" bson:"biz_ids"`
	Snapshots   []HostSnapshot `json:"snapshots" bson:"snapshots"`
}

// HostSnapshot stores the host identity and submitted settings.
// Operation metadata belongs to the containing Data document.
type HostSnapshot struct {
	HostID                  int64    `json:"bk_host_id" bson:"bk_host_id"`
	OpsConsoleHostID        int64    `json:"ops_console_host_id" bson:"ops_console_host_id"`
	OpsConsoleHostInnerIP   []string `json:"ops_console_host_innerip_list" bson:"ops_console_host_innerip_list"`
	OpsConsoleHostInnerIPV6 []string `json:"ops_console_host_innerip_v6_list" bson:"ops_console_host_innerip_v6_list"`
	OpsOutBandIP            string   `json:"ops_out_band_ip" bson:"ops_out_band_ip"`
	OpsOutBandPort          int64    `json:"ops_out_band_port" bson:"ops_out_band_port"`
	OpsOutBandProtocol      string   `json:"ops_out_band_protocol" bson:"ops_out_band_protocol"`
	OpsOutBandType          string   `json:"ops_out_band_type" bson:"ops_out_band_type"`
}

// UniqueFields identifies an operation within its tenant collection.
func (d *Data) UniqueFields() []string {
	return []string{FieldKeyOperateID}
}

// UniqueKey identifies an operation within its tenant collection.
func (d *Data) UniqueKey() string {
	return d.OperateID
}

// Table represents the complete database structure of an operation history entry.
type Table base.TableBroker[*Data]
