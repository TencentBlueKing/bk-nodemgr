/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package types ...
package types

// CycleTime describes the IP and round-trip time of a proxy connection.
type CycleTime struct {
	HostID    int64
	InnerIP   []string
	InnerIPV6 []string
	AgentID   string
	Time      string
}

// GraphNodeInfo graph node info.
type GraphNodeInfo struct {
	NetworkUnitID int64
	RunningProxy  int64
	TotalProxy    int64
	RunningAgent  int64
	TotalAgent    int64
	IsHealthy     bool
	CycleTimes    []CycleTime
}
