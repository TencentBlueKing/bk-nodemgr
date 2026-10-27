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

package types

import "time"

// NodeOpsHistoryInfo identifies an archived out-of-band settings operation.
type NodeOpsHistoryInfo struct {
	OperateID   string
	Operator    string
	OperateTime time.Time
	NodeRole    NodeRole
	BizIDs      []int64
}

// NodeOpsHistoryHostSnapshot preserves host attributes and submitted settings.
type NodeOpsHistoryHostSnapshot struct {
	HostID                  int64
	OpsConsoleHostID        int64
	OpsConsoleHostInnerIP   []string
	OpsConsoleHostInnerIPV6 []string
	OpsOutBandIP            string
	OpsOutBandPort          int64
	OpsOutBandProtocol      string
	OpsOutBandType          string
}

// NodeOpsHistory preserves one operation and its archived host snapshots.
type NodeOpsHistory struct {
	TenantID  string
	Info      NodeOpsHistoryInfo
	Snapshots []*NodeOpsHistoryHostSnapshot
}

// NodeOpsHistoryFieldSelection selects fields for operation history queries.
// Nil or an empty selection returns all fields. Otherwise, unselected scalar
// fields have zero values and unselected Snapshots is an empty slice.
// Snapshots selects the complete snapshot array, not individual host fields.
type NodeOpsHistoryFieldSelection struct {
	TenantID    bool
	OperateID   bool
	Operator    bool
	OperateTime bool
	NodeRole    bool
	BizIDs      bool
	Snapshots   bool
}
