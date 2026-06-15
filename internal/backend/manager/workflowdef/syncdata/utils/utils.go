/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package utils use to provide some common utils for sync data actions.
package utils

import "github.com/TencentBlueKing/bk-nodemgr/pkg/types"

// FillHostDynamicAgentID fill the dynamic agent ID of the host with the synced agent ID if it exists.
func FillHostDynamicAgentID(host *types.Host) {
	if host.Static.SyncedAgentID != "" {
		host.Dynamic.AgentID = host.Static.SyncedAgentID
	}
}

// FillHostDynamicOpsInfo fill the dynamic ops fields of the host with the synced ops fields if they exist.
func FillHostDynamicOpsInfo(host *types.Host) {
	if host.Static.SyncedOpsConsoleHostID > 0 {
		host.Dynamic.OpsConsoleHostID = host.Static.SyncedOpsConsoleHostID
	}
	if host.Static.SyncedOpsOutBandType != "" {
		host.Dynamic.OpsOutBandType = host.Static.SyncedOpsOutBandType
	}
	if host.Static.SyncedOpsOutBandProtocol != "" {
		host.Dynamic.OpsOutBandProtocol = host.Static.SyncedOpsOutBandProtocol
	}
	if host.Static.SyncedOpsBMCIP != "" {
		host.Dynamic.OpsBMCIP = host.Static.SyncedOpsBMCIP
	}
	if host.Static.SyncedOpsBMCPort > 0 {
		host.Dynamic.OpsBMCPort = host.Static.SyncedOpsBMCPort
	}
}
