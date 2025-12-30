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
	"fmt"
)

// ScopeType defines the type of the scope.
type ScopeType string

const (
	// ScopeTypeTopo means the scope is topo.
	ScopeTypeTopo ScopeType = "topo"

	// ScopeTypeServiceTemplate means the scope is service template.
	ScopeTypeServiceTemplate ScopeType = "service_template"

	// ScopeTypeSetTemplate means the scope is set template.
	ScopeTypeSetTemplate ScopeType = "set_template"

	// ScopeTypeInstance means the scope is instance.
	ScopeTypeInstance ScopeType = "instance"

	// ScopeTypeDynamicGroup means the scope is dynamic group.
	ScopeTypeDynamicGroup ScopeType = "dynamic_group"
)

// Validate validates the scope type.
func (scopeType ScopeType) Validate() error {
	switch scopeType {
	case
		ScopeTypeTopo,
		ScopeTypeServiceTemplate,
		ScopeTypeSetTemplate,
		ScopeTypeInstance,
		ScopeTypeDynamicGroup:
		return nil
	default:
		return fmt.Errorf("invalid scope type, scope-type(%s)", scopeType)
	}
}

// TargetGranularity defines the granularity of the scope.
type TargetGranularity string

const (
	// TargetGranularityHost means the scope is host.
	TargetGranularityHost TargetGranularity = "host"

	// TargetGranularityServiceInstance means the scope is service instance.
	TargetGranularityServiceInstance TargetGranularity = "service_instance"
)

// Validate validates the scope granularity.
func (granularity TargetGranularity) Validate() error {
	switch granularity {
	case TargetGranularityHost, TargetGranularityServiceInstance:
		return nil
	default:
		return fmt.Errorf("invalid target granularity, target-granularity(%s)", granularity)
	}
}

// Target define the target.
type Target struct {
	Host            Host
	ServiceInstance ServiceInstance
}

// UniqueID returns the unique id of the target.
func (target *Target) UniqueID() string {
	return fmt.Sprintf("%d:%d", target.Host.HostID, target.ServiceInstance.ModuleID)
}

// ScopeTopoNode defines the scope of topo path.
type ScopeTopoNode struct {
	TopoObjID  string
	TopoInstID int64
}

// ScopeServiceTemplate defines the scope of service template.
type ScopeServiceTemplate struct {
	Granularity        TargetGranularity
	BizID              int64
	Filter             *TargetFilter
	ServiceTemplateIDs []int64
	ModuleIDs          []int64
}

// Validate validates the service template scope.
func (scope *ScopeServiceTemplate) Validate() error {
	if err := scope.Granularity.Validate(); err != nil {
		return fmt.Errorf("failed to validate scope service template: %w", err)
	}

	if scope.BizID == 0 {
		return fmt.Errorf("failed to validate scope service template, bizID is 0")
	}

	return nil
}

// ScopeSetTemplate defines the scope of set template.
type ScopeSetTemplate struct {
	Granularity    TargetGranularity
	BizID          int64
	Filter         *TargetFilter
	SetTemplateIDs []int64
	SetIDs         []int64
}

// Validate validates the set template scope.
func (scope *ScopeSetTemplate) Validate() error {
	if err := scope.Granularity.Validate(); err != nil {
		return fmt.Errorf("failed to validate scope set template: %w", err)
	}

	if scope.BizID == 0 {
		return fmt.Errorf("failed to validate scope set template, bizID is 0")
	}

	return nil
}

// ScopeInstance defines the scope of instance.
type ScopeInstance struct {
	Granularity TargetGranularity
	BizID       int64
	Filter      *TargetFilter
	InstanceIDs []int64
}

// Validate validates the instance scope.
func (scope *ScopeInstance) Validate() error {
	if err := scope.Granularity.Validate(); err != nil {
		return fmt.Errorf("failed to validate scope instance: %w", err)
	}

	if scope.BizID == 0 {
		return fmt.Errorf("failed to validate scope instance, bizID is 0")
	}

	if len(scope.InstanceIDs) == 0 {
		return fmt.Errorf("failed to validate scope instance, instanceIDs is empty")
	}

	return nil
}

// ScopeTopo defines the scope of topo.
type ScopeTopo struct {
	Granularity TargetGranularity
	BizID       int64
	Filter      *TargetFilter
	Paths       []*ScopeTopoNode
}

// Validate validates the scope topo.
func (scope *ScopeTopo) Validate() error {
	if err := scope.Granularity.Validate(); err != nil {
		return fmt.Errorf("failed to validate scope topo: %w", err)
	}

	if scope.BizID == 0 {
		return fmt.Errorf("failed to validate scope topo, bizID is 0")
	}

	if len(scope.Paths) == 0 {
		return fmt.Errorf("failed to validate scope topo, paths is empty")
	}

	return nil
}

// ScopeDynamicGroup defines the scope of dynamic group.
type ScopeDynamicGroup struct {
	Granularity     TargetGranularity
	BizID           int64
	Filter          *TargetFilter
	DynamicGroupIDs []string
}

// Validate validates the dynamic group scope.
func (scope *ScopeDynamicGroup) Validate() error {
	if err := scope.Granularity.Validate(); err != nil {
		return fmt.Errorf("failed to validate scope dynamic group: %w", err)
	}

	if scope.BizID == 0 {
		return fmt.Errorf("failed to validate scope dynamic group, bizID is 0")
	}

	if len(scope.DynamicGroupIDs) == 0 {
		return fmt.Errorf("failed to validate scope dynamic group, dynamicGroupIDs is empty")
	}

	return nil
}

// TargetFilter defines the filter of the target.
type TargetFilter struct {
}

// Scope defines the scope of the deploy policy.
type Scope struct {
	scopeType            ScopeType
	scopeServiceTemplate *ScopeServiceTemplate
	scopeSetTemplate     *ScopeSetTemplate
	scopeInstance        *ScopeInstance
	scopeTopo            *ScopeTopo
	scopeDynamicGroup    *ScopeDynamicGroup
}

// Type returns the scope type.
func (scope *Scope) Type() ScopeType {
	return scope.scopeType
}

// GetServiceTemplateScope returns the service template scope.
func (scope *Scope) GetServiceTemplateScope() (*ScopeServiceTemplate, error) {
	if scope.scopeServiceTemplate == nil {
		return nil, fmt.Errorf("scope_service_template is nil")
	}

	if err := scope.scopeServiceTemplate.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate scope service template: %w", err)
	}

	return scope.scopeServiceTemplate, nil
}

// GetSetTemplateScope returns the set template scope.
func (scope *Scope) GetSetTemplateScope() (*ScopeSetTemplate, error) {
	if scope.scopeSetTemplate == nil {
		return nil, fmt.Errorf("scope_set_template is nil")
	}

	if err := scope.scopeSetTemplate.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate scope set template: %w", err)
	}

	return scope.scopeSetTemplate, nil
}

// GetInstanceScope returns the instance scope.
func (scope *Scope) GetInstanceScope() (*ScopeInstance, error) {
	if scope.scopeInstance == nil {
		return nil, fmt.Errorf("scope_instance is nil")
	}

	if err := scope.scopeInstance.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate scope instance: %w", err)
	}

	return scope.scopeInstance, nil
}

// GetTopoScope returns the topo scope.
func (scope *Scope) GetTopoScope() (*ScopeTopo, error) {
	if scope.scopeTopo == nil {
		return nil, fmt.Errorf("scope_topo is nil")
	}

	if err := scope.scopeTopo.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate scope topo: %w", err)
	}

	return scope.scopeTopo, nil
}

// GetDynamicGroupScope returns the dynamic group scope.
func (scope *Scope) GetDynamicGroupScope() (*ScopeDynamicGroup, error) {
	if scope.scopeDynamicGroup == nil {
		return nil, fmt.Errorf("scope_dynamic_group is nil")
	}

	if err := scope.scopeDynamicGroup.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate scope dynamic group: %w", err)
	}

	return scope.scopeDynamicGroup, nil
}

// NewScopeWithServiceTemplate creates a new scope with service template.
func NewScopeWithServiceTemplate(scope *ScopeServiceTemplate) (*Scope, error) {
	if scope == nil {
		return nil, fmt.Errorf("service template scope is nil")
	}

	return &Scope{
		scopeType:            ScopeTypeServiceTemplate,
		scopeServiceTemplate: scope,
	}, nil
}

// NewScopeWithSetTemplate creates a new scope with set template.
func NewScopeWithSetTemplate(scope *ScopeSetTemplate) (*Scope, error) {
	if scope == nil {
		return nil, fmt.Errorf("set template scope is nil")
	}

	return &Scope{
		scopeType:        ScopeTypeSetTemplate,
		scopeSetTemplate: scope,
	}, nil
}

// NewScopeWithInstance creates a new scope with instance.
func NewScopeWithInstance(scope *ScopeInstance) (*Scope, error) {
	if scope == nil {
		return nil, fmt.Errorf("instance scope is nil")
	}

	return &Scope{
		scopeType:     ScopeTypeInstance,
		scopeInstance: scope,
	}, nil
}

// NewScopeWithTopo creates a new scope with topo.
func NewScopeWithTopo(scope *ScopeTopo) (*Scope, error) {
	if scope == nil {
		return nil, fmt.Errorf("topo scope is nil")
	}

	return &Scope{
		scopeType: ScopeTypeTopo,
		scopeTopo: scope,
	}, nil
}

// NewScopeWithDynamicGroup creates a new scope with dynamic group.
func NewScopeWithDynamicGroup(scope *ScopeDynamicGroup) (*Scope, error) {
	if scope == nil {
		return nil, fmt.Errorf("dynamic group scope is nil")
	}

	return &Scope{
		scopeType:         ScopeTypeDynamicGroup,
		scopeDynamicGroup: scope,
	}, nil
}
