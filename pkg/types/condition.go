/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
)

// TimeRange defines the time range.
type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

// RecentTimeRange returns the time range.
func RecentTimeRange(duration time.Duration) TimeRange {
	return TimeRange{
		StartTime: time.Now().Add(-duration),
		EndTime:   time.Now(),
	}
}

// BusinessExactFields defines the business exact fields.
// support includes and excludes.
type BusinessExactFields struct {
	BizID []int64
}

// BusinessFuzzyFields defines the business fuzzy fields.
// support includes and excludes.
type BusinessFuzzyFields struct {
	BizName []string
}

// BusinessCondition defines the business condition.
// in this condition, fields are generated with AND expr.
type BusinessCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *BusinessExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *BusinessFuzzyFields

	// will be used when condition type is excluded in exclude mode.
	ExactExclude *BusinessExactFields

	// will be used when condition type is excluded in exclude mode.
	FuzzyExclude *BusinessFuzzyFields
}

// HostExactFields defines the host exact fields.
// support includes and excludes.
type HostExactFields struct {
	HostID         []int64
	BizID          []int64
	NetworkAreaID  []int64
	NetworkUnitID  []int64
	OSType         []string
	Arch           []string
	NodeRole       []NodeRole
	NodeStatus     []NodeStatus
	NodeVersion    []string
	NodeGeneration []int64
	AgentID        []string
	InnerIP        []string
	Addressing     []Addressing
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
	// will be used when condition type is included in exact mode.
	ExactInclude *HostExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *HostFuzzyFields

	// will be used when condition type is excluded in exclude mode.
	ExactExclude *HostExactFields

	// will be used when condition type is excluded in exclude mode.
	FuzzyExclude *HostFuzzyFields
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
	// will be used when condition type is included in exact mode.
	ExactInclude *NetworkAreaExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *NetworkAreaFuzzyFields

	// will be used when condition type is excluded in exclude mode.
	ExactExclude *NetworkAreaExactFields

	// will be used when condition type is excluded in exclude mode.
	FuzzyExclude *NetworkAreaFuzzyFields
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
	// will be used when condition type is included in exact mode.
	ExactInclude *NetworkUnitExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *NetworkUnitFuzzyFields

	// will be used when condition type is excluded in exclude mode.
	ExactExclude *NetworkUnitExactFields

	// will be used when condition type is excluded in exclude mode.
	FuzzyExclude *NetworkUnitFuzzyFields
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
	// will be used when condition type is included in exact mode.
	ExactInclude *AccessPointExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *AccessPointFuzzyFields

	// will be used when condition type is excluded in exclude mode.
	ExactExclude *AccessPointExactFields

	// will be used when condition type is excluded in exclude mode.
	FuzzyExclude *AccessPointFuzzyFields
}

// TopoEventExactFields defines the topo event exact fields.
type TopoEventExactFields struct {
	NetworkAreaID []int64
	NetworkUnitID []int64
	AccessPointID []int64
	Type          []TopoEventType
	Operator      []string
}

// TopoEventFuzzyFields defines the topo event fuzzy fields.
type TopoEventFuzzyFields struct {
	NetworkAreaName []string
	NetworkUnitName []string
}

// TopoEventCondition defines the topo event condition.
type TopoEventCondition struct {
	// operate time range will be used whatever condition type is.
	OperateTimeRange *TimeRange

	// will be used when condition type is included in exact mode.
	ExactInclude *TopoEventExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *TopoEventFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *TopoEventExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *TopoEventFuzzyFields
}

// TopoConstantFields defines the topo constant fields.
type TopoConstantFields struct {
	CloudVendor bool
	OSType      bool
}

// NodeWorkflowExactFields defines the node workflow exact fields.
type NodeWorkflowExactFields struct {
	WorkflowID []string
	Type       []NodeWorkflowType
	BizID      []int64
	Status     []NodeWorkflowStatus
	Operator   []string
}

// NodeWorkflowFuzzyFields defines the node workflow fuzzy fields.
type NodeWorkflowFuzzyFields struct {
}

// NodeWorkflowCondition defines the node workflow condition.
type NodeWorkflowCondition struct {
	// operate time range will be used whatever condition type is.
	OperateTimeRange *TimeRange

	// will be used when condition type is included in exact mode.
	ExactInclude *NodeWorkflowExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *NodeWorkflowFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *NodeWorkflowExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *NodeWorkflowFuzzyFields
}

// NodeWorkflowOperInstanceStatusExactFields defines the node workflow instance status exact fields.
type NodeWorkflowOperInstanceStatusExactFields struct {
	TriggerID []string
}

// NodeWorkflowOperInstanceStatusFuzzyFields defines the node workflow  instance status fuzzy fields.
type NodeWorkflowOperInstanceStatusFuzzyFields struct {
}

// NodeWorkflowOperInstanceStatusCondition defines the node workflow instance status condition.
type NodeWorkflowOperInstanceStatusCondition struct {

	// will be used when condition type is included in exact mode.
	ExactInclude *NodeWorkflowOperInstanceStatusExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *NodeWorkflowOperInstanceStatusFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *NodeWorkflowOperInstanceStatusExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *NodeWorkflowOperInstanceStatusFuzzyFields
}

// ReleaseExactFields defines the release exact fields.
type ReleaseExactFields struct {
	Generation []Generation
	Platform   []platform.Platform
	Type       []ReleaseType
	Version    []string
	AsDefault  []bool
	Enabled    []bool
}

// ReleaseFuzzyFields defines the release fuzzy fields.
type ReleaseFuzzyFields struct {
}

// ReleaseCondition defines the release condition.
type ReleaseCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *ReleaseExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *ReleaseFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *ReleaseExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *ReleaseFuzzyFields
}

// NodeWorkflowOperationExactFields defines the condition of list operation.
type NodeWorkflowOperationExactFields struct {
	TriggerID     string
	WorkflowID    string
	State         []OperationState
	InnerIP       []string
	InnerIPv6     []string
	BizID         []int64
	NetworkAreaID []int64
	NodeVersion   []string
}

// NodeWorkflowoperationFuzzyFields defines the node workflow operation fuzzy fields.
type NodeWorkflowoperationFuzzyFields struct {
}

// NodeWorkflowOperationCondition defines the node workflow operation condition.
type NodeWorkflowOperationCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *NodeWorkflowOperationExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *NodeWorkflowoperationFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *NodeWorkflowOperationExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *NodeWorkflowoperationFuzzyFields
}

// ScheduleWorkflowExactFields defines the schedule workflow exact fields.
type ScheduleWorkflowExactFields struct {
	WorkflowID   []string
	WorkflowName []string
	Operator     []string
}

// ScheduleWorkflowFuzzyFields defines the schedule workflow fuzzy fields.
type ScheduleWorkflowFuzzyFields struct {
}

// ScheduleWorkflowCondition defines the schedule workflow condition.
type ScheduleWorkflowCondition struct {
	// operate time range will be used whatever condition type is.
	OperateTimeRange *TimeRange

	// will be used when condition type is included in exact mode.
	ExactInclude *ScheduleWorkflowExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *ScheduleWorkflowFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *ScheduleWorkflowExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *ScheduleWorkflowFuzzyFields
}

// ScheduleWorkflowOperInstanceStatusExactFields defines the schedule workflow instance status exact fields.
type ScheduleWorkflowOperInstanceStatusExactFields struct {
	TriggerID []string
}

// ScheduleWorkflowOperInstanceStatusFuzzyFields defines the schedule workflow  instance status fuzzy fields.
type ScheduleWorkflowOperInstanceStatusFuzzyFields struct {
}

// ScheduleWorkflowOperInstanceStatusCondition defines the schedule workflow instance status condition.
type ScheduleWorkflowOperInstanceStatusCondition struct {

	// will be used when condition type is included in exact mode.
	ExactInclude *ScheduleWorkflowOperInstanceStatusExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *ScheduleWorkflowOperInstanceStatusFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *ScheduleWorkflowOperInstanceStatusExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *ScheduleWorkflowOperInstanceStatusFuzzyFields
}

// GlobalSettingsExactFields defines the global settings exact fields.
type GlobalSettingsExactFields struct {
	SettingName []string
}

// GlobalSettingsFuzzyFields defines the global settings fuzzy fields.
type GlobalSettingsFuzzyFields struct {
	SettingName []string
}

// GlobalSettingsCondition defines the global settings condition.
type GlobalSettingsCondition struct {
	// operate time range will be used whatever condition type is.
	OperateTimeRange *TimeRange

	// will be used when condition type is included in exact mode.
	ExactInclude *GlobalSettingsExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *GlobalSettingsFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *GlobalSettingsExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *GlobalSettingsFuzzyFields
}
