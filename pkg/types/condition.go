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

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// ===============================================================================
// Time Related Conditions
// ===============================================================================

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

// BeforeTimeRange before the end time.
func BeforeTimeRange(endTime time.Time) TimeRange {
	return TimeRange{
		StartTime: time.Time{},
		EndTime:   endTime,
	}
}

// ===============================================================================
// Business Related Conditions
// ===============================================================================

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

// ===============================================================================
// Host Related Conditions
// ===============================================================================

// HostDynamicExactFields defines the host exact fields.
// support includes and excludes.
type HostDynamicExactFields struct {
	NetworkUnitID  []int64
	OSType         []string
	Arch           []string
	NodeRole       []NodeRole
	NodeStatus     []NodeStatus
	NodeVersion    []string
	NodeGeneration []int64
	AgentID        []string
}

// HostDynamicFuzzyFields defines the host fuzzy fields.
type HostDynamicFuzzyFields struct {
}

// HostStaticExactFields defines the host exact fields.
// support includes and excludes.
type HostStaticExactFields struct {
	HostID        []int64
	BizID         []int64
	NetworkAreaID []int64
	InnerIP       []string
	InnerIPV6     []string
	Addressing    []Addressing
}

// HostStaticFuzzyFields defines the host fuzzy fields.
type HostStaticFuzzyFields struct {
	HostID    []int64
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
	// will be used when condition type is included in static exact mode.
	StaticExactInclude *HostStaticExactFields

	// will be used when condition type is included in static fuzzy mode.
	StaticFuzzyInclude *HostStaticFuzzyFields

	// will be used when condition type is excluded in static exclude mode.
	StaticExactExclude *HostStaticExactFields

	// will be used when condition type is excluded in static exclude mode.
	StaticFuzzyExclude *HostStaticFuzzyFields

	// will be used when condition type is included in dynamic exact mode.
	DynamicExactInclude *HostDynamicExactFields

	// will be used when condition type is included in dynamic fuzzy mode.
	DynamicFuzzyInclude *HostDynamicFuzzyFields

	// will be used when condition type is excluded in dynamic exclude mode.
	DynamicExactExclude *HostDynamicExactFields

	// will be used when condition type is excluded in dynamic exclude mode.
	DynamicFuzzyExclude *HostDynamicFuzzyFields
}

// HostStaticCondition defines the host condition.
// in this condition, fields are generated with AND expr.
type HostStaticCondition struct {
	HostStaticExactCondition
	HostStaticFuzzyCondition
}

// HostStaticExactCondition defines the host condition.
// in this condition, fields are generated with AND expr.
type HostStaticExactCondition struct {
	// will be used when condition type is included in static exact mode.
	StaticExactInclude *HostStaticExactFields

	// will be used when condition type is excluded in static exclude mode.
	StaticExactExclude *HostStaticExactFields
}

// HostStaticFuzzyCondition defines the host condition.
// in this condition, fields are generated with AND expr.
type HostStaticFuzzyCondition struct {
	// will be used when condition type is included in static fuzzy mode.
	StaticFuzzyInclude *HostStaticFuzzyFields

	// will be used when condition type is excluded in static exclude mode.
	StaticFuzzyExclude *HostStaticFuzzyFields
}

// ===============================================================================
// NetworkArea Related Conditions
// ===============================================================================

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

// ===============================================================================
// NetworkUnit Related Conditions
// ===============================================================================

// NetworkUnitExactFields defines the network unit exact fields.
type NetworkUnitExactFields struct {
	NetworkUnitID []int64
	NetworkAreaID []int64
	IsDirect      []bool
}

// NetworkUnitFuzzyFields defines the network unit fuzzy fields.
type NetworkUnitFuzzyFields struct {
	NetworkUnitName []string
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

// ===============================================================================
// AccessPoint Related Conditions
// ===============================================================================

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

// ===============================================================================
// Event Related Conditions
// ===============================================================================

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

// PackageEventExactFields defines the package event exact fields.
type PackageEventExactFields struct {
	Generation  []Generation
	ReleaseType []ReleaseType
	OSType      []criteria.OSType
	CPUArch     []criteria.CPUArch
	Version     []string
	Operator    []string
	EventType   []PackageEventType
}

// PackageEventFuzzyFields defines the package event fuzzy fields.
type PackageEventFuzzyFields struct {
}

// PackageEventCondition defines the package event condition.
type PackageEventCondition struct {
	// operate time range will be used whatever condition type is.
	OperateTimeRange *TimeRange

	// will be used when condition type is included in exact mode.
	ExactInclude *PackageEventExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *PackageEventFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *PackageEventExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *PackageEventFuzzyFields
}

// ConfigPolicyEventExactFields defines the policy event exact fields.
type ConfigPolicyEventExactFields struct {
	ConfigPolicyID   []int64
	ConfigPolicyType []ConfigPolicyType
	Version          []int64
	Type             []ConfigPolicyEventType
}

// ConfigPolicyEventFuzzyFields defines the policy event fuzzy fields.
type ConfigPolicyEventFuzzyFields struct {
	ConfigPolicyName []string
	Operator         []string
}

// ConfigPolicyEventCondition defines the policy event condition.
type ConfigPolicyEventCondition struct {
	// operate time range will be used whatever condition type is.
	OperateTimeRange *TimeRange

	// will be used when condition type is included in exact mode.
	ExactInclude *ConfigPolicyEventExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *ConfigPolicyEventFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *ConfigPolicyEventExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *ConfigPolicyEventFuzzyFields
}

// ===============================================================================
// Workflow Related Conditions
// ===============================================================================

// OperInstDataExactFields defines the workflow operation instance data condition exact fields.
type OperInstDataExactFields struct {
	TriggerID   []string
	OperationID []string
	OperInstID  []string
	State       []operation.State
}

// OperInstDataFuzzyFields defines the workflow operation instance data fuzzy fields.
type OperInstDataFuzzyFields struct{}

// OperInstDataCondition defines the workflow operation instance data condition.
type OperInstDataCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *OperInstDataExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *OperInstDataFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *OperInstDataExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *OperInstDataFuzzyFields
}

// OperationExactFields defines the condition exact fields.
type OperationExactFields struct {
	TriggerID []string
	State     []operation.State
}

// OperationFuzzyFields defines the operation fuzzy fields.
type OperationFuzzyFields struct {
}

// OperationCondition defines the operation condition.
type OperationCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *OperationExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *OperationFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *OperationExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *OperationFuzzyFields
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

// ApplicationNodeOperationListExactFields defines the exact fields for node workflow operation list.
type ApplicationNodeOperationListExactFields struct {
	BizID         []int64
	HostInnerIP   []string
	HostInnerIPV6 []string
	NetworkAreaID []int64
	NetworkUnitID []int64
	NodeVersion   []string
	State         []operation.State
}

// ApplicationNodeOperationListFuzzyFields defines the fuzzy fields for node operation list.
type ApplicationNodeOperationListFuzzyFields struct {
}

// ApplicationNodeOperationListCondition defines the node workflow operation list condition.
type ApplicationNodeOperationListCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *ApplicationNodeOperationListExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *ApplicationNodeOperationListFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *ApplicationNodeOperationListExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *ApplicationNodeOperationListFuzzyFields
}

// PluginWorkflowExactFields defines the plugin workflow exact fields.
type PluginWorkflowExactFields struct {
	WorkflowID []string
	Type       []PluginWorkflowType
	HostID     []int64
	Status     []PluginWorkflowStatus
	Operator   []string
}

// PluginWorkflowFuzzyFields defines the plugin workflow fuzzy fields.
type PluginWorkflowFuzzyFields struct {
}

// PluginWorkflowCondition defines the plugin workflow condition.
type PluginWorkflowCondition struct {
	// operate time range will be used whatever condition type is.
	OperateTimeRange *TimeRange

	// will be used when condition type is included in exact mode.
	ExactInclude *PluginWorkflowExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *PluginWorkflowFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *PluginWorkflowExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *PluginWorkflowFuzzyFields
}

// ApplicationPluginOperationListExactFields defines the exact fields for plugin workflow operation list.
type ApplicationPluginOperationListExactFields struct {
	HostID        []int64
	PluginName    []string
	PluginVersion []string
	State         []operation.State
}

// ApplicationPluginOperationListFuzzyFields defines the fuzzy fields for plugin workflow operation list.
type ApplicationPluginOperationListFuzzyFields struct{}

// ApplicationPluginOperationListCondition defines the application plugin workflow operation list condition.
type ApplicationPluginOperationListCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *ApplicationPluginOperationListExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *ApplicationPluginOperationListFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *ApplicationPluginOperationListExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *ApplicationPluginOperationListFuzzyFields
}

// ScheduledWorkflowExactFields defines the scheduled workflow exact fields.
type ScheduledWorkflowExactFields struct {
	WorkflowID   []string
	WorkflowName []string
	Operator     []string
}

// ScheduledWorkflowFuzzyFields defines the scheduled workflow fuzzy fields.
type ScheduledWorkflowFuzzyFields struct {
}

// ScheduledWorkflowCondition defines the schedule workflow condition.
type ScheduledWorkflowCondition struct {
	// operate time range will be used whatever condition type is.
	OperateTimeRange *TimeRange

	// will be used when condition type is included in exact mode.
	ExactInclude *ScheduledWorkflowExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *ScheduledWorkflowFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *ScheduledWorkflowExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *ScheduledWorkflowFuzzyFields
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

// ===============================================================================
// Release Related Conditions
// ===============================================================================

// ReleaseExactFields defines the release exact fields.
type ReleaseExactFields struct {
	Name       []string
	FileName   []string
	Generation []Generation
	Platform   []platfmt.Platform
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

// ===============================================================================
// GlobalSettings Related Conditions
// ===============================================================================

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

// ===============================================================================
// ConfigPolicy Related Conditions
// ===============================================================================

// ConfigPolicyExactFields defines the config policy exact fields.
type ConfigPolicyExactFields struct {
	ConfigPolicyID []int64
	BizID          []int64
	Type           []ConfigPolicyType
	Enabled        []bool
}

// ConfigPolicyFuzzyFields defines the config policy fuzzy fields.
type ConfigPolicyFuzzyFields struct {
	ConfigPolicyName []string
	Operator         []string
}

// ConfigPolicyCondition defines the config policy condition.
type ConfigPolicyCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *ConfigPolicyExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *ConfigPolicyFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *ConfigPolicyExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *ConfigPolicyFuzzyFields
}

// ===============================================================================
// Process Related Conditions
// ===============================================================================

// ProcessExactFields defines the process exact fields.
// support includes and excludes.
type ProcessExactFields struct {
	HostID         []int64
	PluginGroup    []string
	NodeGeneration []string
	PlatformOS     []string
	PlatformArch   []string
	InfoStatus     []ProcessStatus
	InfoAgentID    []string
	InfoVersion    []string
	PluginName     []string
	PluginPkgName  []string
}

// ProcessFuzzyFields defines the process fuzzy fields.
// support includes and excludes.
type ProcessFuzzyFields struct {
	Name    []string
	PkgName []string
}

// ProcessCondition defines the process condition.
// in this condition, fields are generated with AND expr.
type ProcessCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *ProcessExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *ProcessFuzzyFields

	// will be used when condition type is excluded in exclude mode.
	ExactExclude *ProcessExactFields

	// will be used when condition type is excluded in exclude mode.
	FuzzyExclude *ProcessFuzzyFields
}

// ProcessConfigExactFields defines the process config exact fields.
// support includes and excludes.
type ProcessConfigExactFields struct {
	Name         []string
	ProcessName  []string
	HostID       []int64
	IsMainConfig []bool
}

// ProcessConfigFuzzyFields defines the process config fuzzy fields.
// support includes and excludes.
type ProcessConfigFuzzyFields struct{}

// ProcessConfigCondition defines the process config condition.
type ProcessConfigCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *ProcessConfigExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *ProcessConfigFuzzyFields

	// will be used when condition type is excluded in exclude mode.
	ExactExclude *ProcessConfigExactFields

	// will be used when condition type is excluded in exclude mode.
	FuzzyExclude *ProcessConfigFuzzyFields
}

// ===============================================================================
// Plugin Related Conditions
// ===============================================================================

// PluginCondition defines the plugin condition.
type PluginCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *PluginExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *PluginFuzzyFields

	// will be used when condition type is excluded in exclude mode.
	ExactExclude *PluginExactFields

	// will be used when condition type is excluded in exclude mode.
	FuzzyExclude *PluginFuzzyFields
}

// PluginExactFields defines the plugin exact fields.
type PluginExactFields struct {
	Name  []string
	Group []string
}

// PluginFuzzyFields defines the plugin fuzzy fields.
type PluginFuzzyFields struct {
	Name    []string
	PkgName []string
}

// ===============================================================================
// Deployment Related Conditions
// ===============================================================================

// NodeDeploymentExactFields defines the node deployment exact fields.
type NodeDeploymentExactFields struct {
	Token         []string
	BizID         []int64
	NodeVersion   []string
	InnerIP       []string
	InnerIPv6     []string
	NetworkAreaID []int64
	NetworkUnitID []int64
}

// NodeDeploymentFuzzyFields defines the node deployment fuzzy fields.
type NodeDeploymentFuzzyFields struct {
}

// NodeDeploymentCondition defines the node deployment condition.
type NodeDeploymentCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *NodeDeploymentExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *NodeDeploymentFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *NodeDeploymentExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *NodeDeploymentFuzzyFields
}

// PluginDeploymentExactFields defines the plugin deployment exact fields.
type PluginDeploymentExactFields struct {
	Token         []string
	HostID        []int64
	PluginName    []string
	PluginVersion []string
}

// PluginDeploymentFuzzyFields defines the plugin deployment fuzzy fields.
type PluginDeploymentFuzzyFields struct {
}

// PluginDeploymentCondition defines the plugin deployment condition.
type PluginDeploymentCondition struct {
	// will be used when condition type is included in exact mode.
	ExactInclude *PluginDeploymentExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *PluginDeploymentFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *PluginDeploymentExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *PluginDeploymentFuzzyFields
}

// ===============================================================================
// DeployPolicy Related Conditions
// ===============================================================================

// DeployPolicyCondition defines the deploy policy condition.
type DeployPolicyCondition struct {
	// executed time range will be used whatever condition type is.
	ExecutedTimeRange *TimeRange

	// will be used when condition type is included in exact mode.
	ExactInclude *DeployPolicyExactFields

	// will be used when condition type is included in fuzzy mode.
	FuzzyInclude *DeployPolicyFuzzyFields

	// will be used when condition type is excluded in exact mode.
	ExactExclude *DeployPolicyExactFields

	// will be used when condition type is excluded in fuzzy mode.
	FuzzyExclude *DeployPolicyFuzzyFields
}

// DeployPolicyExactFields defines the deploy policy exact fields.
type DeployPolicyExactFields struct {
	DeployPolicyID   []int64
	DsuID            []int64
	Enabled          []bool
	DeployPolicyName []string
	Operator         []string
}

// DeployPolicyFuzzyFields defines the deploy policy fuzzy fields.
type DeployPolicyFuzzyFields struct {
	DeployPolicyName []string
	Operator         []string
}
