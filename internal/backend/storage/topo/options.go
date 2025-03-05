/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import "github.com/TencentBlueKing/bk-nodemgr/pkg/types"

// BusinessCondition defines the business condition.
// in this condition, fields are generated with AND expr.
type BusinessCondition struct {
	Type types.ConditionType

	BizID   []int64
	BizName []string
}

// HostExactFields defines the host exact fields.
// support includes and excludes.
type HostExactFields struct {
	HostID        []int64
	BizID         []int64
	NetworkAreaID []int64
	OSType        []string
	NodeRole      []types.NodeRole
	NodeStatus    []types.NodeStatus
	NodeVersion   []string
	AgentID       []string
}

// HostFuzzyFields defines the host fuzzy fields.
type HostFuzzyFields struct {
	HostName  []string
	DeptName  []string
	InnerIP   []string
	InnerIPV6 []string
	OuterIP   []string
	OuterIPV6 []string
}

// HostCondition defines the host condition.
// in this condition, fields are generated with AND expr.
type HostCondition struct {
	Type types.ConditionType

	// will be used when condition type is in exact mode.
	Exact *HostExactFields

	// will be used when condition type is in fuzzy mode.
	Fuzzy *HostFuzzyFields
}

// NetworkAreaExactFields defines the network area exact fields.
type NetworkAreaExactFields struct {
	NetworkAreaID []int64
	CloudVendor   []string
}

// NetworkAreaFuzzyFields defines the network area fuzzy fields.
type NetworkAreaFuzzyFields struct {
	NetworkAreaName []string
}

// NetworkAreaCondition defines the network area condition.
// in this condition, fields are generated with AND expr.
type NetworkAreaCondition struct {
	Type types.ConditionType

	// will be used when condition type is in exact mode.
	Exact *NetworkAreaExactFields

	// will be used when condition type is in fuzzy mode.
	Fuzzy *NetworkAreaFuzzyFields
}

// NetworkUnitExactFields defines the network unit exact fields.
type NetworkUnitExactFields struct {
	NetworkUnitID []int64
	NetworkAreaID []int64
}

// NetworkUnitFuzzyFields defines the network unit fuzzy fields.
type NetworkUnitFuzzyFields struct {
}

// NetworkUnitCondition defines the network unit condition.
// in this condition, fields are generated with AND expr.
type NetworkUnitCondition struct {
	Type types.ConditionType

	// will be used when condition type is in exact mode.
	Exact *NetworkUnitExactFields

	// will be used when condition type is in fuzzy mode.
	Fuzzy *NetworkUnitFuzzyFields
}

// AccessPointExactFields defines the access point exact fields.
type AccessPointExactFields struct {
	AccessPointID []int64
	NetworkAreaID []int64
}

// AccessPointFuzzyFields defines the access point fuzzy fields.
type AccessPointFuzzyFields struct {
}

// AccessPointCondition defines the access point condition.
// in this condition, fields are generated with AND expr.
type AccessPointCondition struct {
	Type types.ConditionType

	// will be used when condition type is in exact mode.
	Exact *AccessPointExactFields

	// will be used when condition type is in fuzzy mode.
	Fuzzy *AccessPointFuzzyFields
}
