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

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// HostDistinctRequest describes the wanted distinct fields.
type HostDistinctRequest struct {
	BizID         bool
	NodeRole      bool
	NodeStatus    bool
	NodeVersion   bool
	DeptName      bool
	OSType        bool
	Arch          bool
	Addressing    bool
	NetworkAreaID bool
	NetworkUnitID bool
}

// NewHostDistinctRequestAllSet creates a HostDistinctRequest with all fields set to true.
func NewHostDistinctRequestAllSet() HostDistinctRequest {
	return HostDistinctRequest{
		BizID:         true,
		NodeRole:      true,
		NodeStatus:    true,
		NodeVersion:   true,
		DeptName:      true,
		OSType:        true,
		Arch:          true,
		Addressing:    true,
		NetworkAreaID: true,
		NetworkUnitID: true,
	}
}

// HostDistinctResult describes the result of distinct.
type HostDistinctResult struct {
	BizID         []int64
	NodeRole      []NodeRole
	NodeStatus    []NodeStatus
	NodeVersion   []string
	DeptName      []string
	OSType        []string
	Arch          []string
	Addressing    []string
	NetworkAreaID []int64
	NetworkUnitID []int64
}

// TopoEventDistinctRequest describes the wanted distinct fields.
type TopoEventDistinctRequest struct {
	Type          bool
	NetworkAreaID bool
	NetworkUnitID bool
	AccessPointID bool
	Operator      bool
}

// NewTopoEventDistinctRequestAllSet creates a TopoEventDistinctRequest with all fields set to true.
func NewTopoEventDistinctRequestAllSet() TopoEventDistinctRequest {
	return TopoEventDistinctRequest{
		Type:          true,
		NetworkAreaID: true,
		NetworkUnitID: true,
		AccessPointID: true,
		Operator:      true,
	}
}

// TopoEventDistinctResult describes the result of distinct.
type TopoEventDistinctResult struct {
	Type          []TopoEventType
	NetworkAreaID []int64
	NetworkUnitID []int64
	AccessPointID []int64
	Operator      []string
}

// NodeWorkflowDistinctRequest describes the wanted distinct fields.
type NodeWorkflowDistinctRequest struct {
	Type     bool
	BizID    bool
	Status   bool
	Operator bool
}

// NewNodeWorkflowDistinctRequestAllSet creates a NodeWorkflowDistinctRequest with all fields set to true.
func NewNodeWorkflowDistinctRequestAllSet() NodeWorkflowDistinctRequest {
	return NodeWorkflowDistinctRequest{
		Type:     true,
		BizID:    true,
		Status:   true,
		Operator: true,
	}
}

// NodeWorkflowDistinctResult describes the result of distinct.
type NodeWorkflowDistinctResult struct {
	Type     []NodeWorkflowType
	BizID    []int64
	Status   []NodeWorkflowStatus
	Operator []string
}

// PluginWorkflowDistinctRequest describes the wanted distinct fields.
type PluginWorkflowDistinctRequest struct {
	Type     bool
	HostID   bool
	Status   bool
	Operator bool
}

// NewPluginWorkflowDistinctRequestAllSet creates a PluginWorkflowDistinctRequest with all fields set to true.
func NewPluginWorkflowDistinctRequestAllSet() PluginWorkflowDistinctRequest {
	return PluginWorkflowDistinctRequest{
		Type:     true,
		HostID:   true,
		Status:   true,
		Operator: true,
	}
}

// PluginWorkflowDistinctResult describes the result of distinct.
type PluginWorkflowDistinctResult struct {
	Type     []PluginWorkflowType
	HostID   []int64
	Status   []PluginWorkflowStatus
	Operator []string
}

// ReleaseDistinctField describes the wanted distinct fields.
type ReleaseDistinctField struct {
	OSType  bool
	CPUArch bool
	Name    bool
	Version bool
}

// ReleaseDistinctResult describes the result of distinct.
type ReleaseDistinctResult struct {
	OSType  []string
	CPUArch []string
	Name    []string
	Version []string
}

// ReleaseDistinctFieldAllSet creates a ReleaseDistinctField with all fields set to true.
func ReleaseDistinctFieldAllSet() ReleaseDistinctField {
	return ReleaseDistinctField{
		OSType:  true,
		CPUArch: true,
		Name:    true,
		Version: true,
	}
}

// PackageEventDistinctRequest describes the wanted distinct fields.
type PackageEventDistinctRequest struct {
	EventType   bool
	ReleaseType bool
	OSType      bool
	CPUArch     bool
	Version     bool
	Operator    bool
}

// NewpckageEventDistinctRequestAllSet creates a PackageEventDistinctRequest with all fields set to true.
func NewpckageEventDistinctRequestAllSet() PackageEventDistinctRequest {
	return PackageEventDistinctRequest{
		EventType:   true,
		ReleaseType: true,
		OSType:      true,
		CPUArch:     true,
		Version:     true,
		Operator:    true,
	}
}

// PackageEventDistinctResult describes the result of distinct.
type PackageEventDistinctResult struct {
	EventType   []PackageEventType
	ReleaseType []ReleaseType
	OSType      []criteria.OSType
	CPUArch     []criteria.CPUArch
	Version     []string
	Operator    []string
}

// NewPackageEventDistinctRequestAllSet creates a PackageEventDistinctRequest with all fields set to true.
func NewPackageEventDistinctRequestAllSet() PackageEventDistinctRequest {
	return PackageEventDistinctRequest{
		EventType:   true,
		ReleaseType: true,
		OSType:      true,
		CPUArch:     true,
		Version:     true,
		Operator:    true,
	}
}

// ConfigPolicyEventDistinctRequest describes the wanted distinct fields.
type ConfigPolicyEventDistinctRequest struct {
	BizID            bool
	Type             bool
	ConfigPolicyType bool
	ConfigPolicyID   bool
	ConfigPolicyName bool
	Version          bool
	Operator         bool
}

// NewConfigPolicyEventDistinctRequestAllSet creates a policy event distinct request with all fields set to true.
func NewConfigPolicyEventDistinctRequestAllSet() ConfigPolicyEventDistinctRequest {
	return ConfigPolicyEventDistinctRequest{
		BizID:            true,
		Type:             true,
		ConfigPolicyType: true,
		ConfigPolicyID:   true,
		ConfigPolicyName: true,
		Version:          true,
		Operator:         true,
	}
}

// ConfigPolicyEventDistinctResult describes the result of distinct.
type ConfigPolicyEventDistinctResult struct {
	BizID            []int64
	Type             []ConfigPolicyEventType
	ConfigPolicyType []ConfigPolicyType
	ConfigPolicyID   []int64
	ConfigPolicyName []string
	Version          []int64
	Operator         []string
}

// ===============================================================================
// ProcessDistinct
// ===============================================================================

// NewProcessDistinctSelectorAllSet creates a ProcessDistinctSelector with all fields set to true.
func NewProcessDistinctSelectorAllSet() ProcessDistinctSelector {
	return ProcessDistinctSelector{
		OSType:        true,
		CPUArch:       true,
		Version:       true,
		Status:        true,
		PluginName:    true,
		PluginGroup:   true,
		PluginPkgName: true,
	}
}

// ProcessDistinctSelector describes the wanted distinct fields.
type ProcessDistinctSelector struct {
	OSType        bool
	CPUArch       bool
	Version       bool
	Status        bool
	PluginName    bool
	PluginGroup   bool
	PluginPkgName bool
}

// ProcessDistinctResult describes the result of distinct.
type ProcessDistinctResult struct {
	OSType        []criteria.OSType
	CPUArch       []criteria.CPUArch
	Version       []string
	Status        []ProcessStatus
	PluginName    []string
	PluginGroup   []string
	PluginPkgName []string
}

// ===============================================================================
// Workflow Operation Distinct (Common for Node and Plugin)
// ===============================================================================

// NewWorkflowOperationDistinctSelectorAllSet creates a WorkflowOperationDistinctSelector with all fields set to true.
func NewWorkflowOperationDistinctSelectorAllSet() WorkflowOperationDistinctSelector {
	return WorkflowOperationDistinctSelector{
		State: true,
	}
}

// WorkflowOperationDistinctSelector describes the wanted distinct fields for workflow operations.
type WorkflowOperationDistinctSelector struct {
	State bool
}

// WorkflowOperationDistinctResult describes the result of distinct for workflow operations.
type WorkflowOperationDistinctResult struct {
	State []operation.State
}

// NewNodeWorkflowOperationDistinctSelectorAllSet creates a WorkflowOperationDistinctSelector for node workflow.
func NewNodeWorkflowOperationDistinctSelectorAllSet() WorkflowOperationDistinctSelector {
	return NewWorkflowOperationDistinctSelectorAllSet()
}

// NewPluginWorkflowOperationDistinctSelectorAllSet creates a WorkflowOperationDistinctSelector for plugin workflow.
func NewPluginWorkflowOperationDistinctSelectorAllSet() WorkflowOperationDistinctSelector {
	return NewWorkflowOperationDistinctSelectorAllSet()
}
