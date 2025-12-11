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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
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

// Scope defines the scope of the deploy policy.
type Scope struct {
	TenantID    string
	BizID       int64
	Type        ScopeType
	Granularity TargetGranularity
	Filter      TargetFilter
	Items       []map[string]any
}

type scopeItemServiceTemplate struct {
	ServiceTemplateID int64
	ServiceID         int64
}

// ToScopeServiceTemplate converts the common scope to scope service template.
func (scope *Scope) ToScopeServiceTemplate() (*ScopeServiceTemplate, error) {
	if scope.Type != ScopeTypeServiceTemplate {
		return nil, fmt.Errorf("failed to convert scope type to service template, scope-type(%s)", scope.Type)
	}

	items := make([]scopeItemServiceTemplate, len(scope.Items))
	for idx := range scope.Items {
		err := conv.MapToStruct(scope.Items[idx], &items[idx])
		if err != nil {
			return nil, fmt.Errorf("failed to convert scope to scope service template, scope(%+v): %w", scope, err)
		}
	}

	scopeServiceTemplate := &ScopeServiceTemplate{
		Granularity:        scope.Granularity,
		BizID:              scope.BizID,
		TenantID:           scope.TenantID,
		Filter:             scope.Filter,
		ServiceTemplateIDs: make([]int64, len(items)),
		ModuleIDs:          make([]int64, len(items)),
	}

	for idx := range items {
		scopeServiceTemplate.ServiceTemplateIDs[idx] = items[idx].ServiceTemplateID
		scopeServiceTemplate.ModuleIDs[idx] = items[idx].ServiceID
	}

	return scopeServiceTemplate, nil
}

type scopeItemSetTemplate struct {
	SetTemplateID int64
	SetID         int64
}

// ToScopeSetTemplate converts the common scope to scope set template.
func (scope *Scope) ToScopeSetTemplate() (*ScopeSetTemplate, error) {
	if scope.Type != ScopeTypeSetTemplate {
		return nil, fmt.Errorf("failed to convert scope type to set template, scope-type(%s)", scope.Type)
	}

	items := make([]scopeItemSetTemplate, len(scope.Items))
	for idx := range scope.Items {
		err := conv.MapToStruct(scope.Items[idx], &items[idx])
		if err != nil {
			return nil, fmt.Errorf("failed to convert scope to scope set template, scope(%+v): %w", scope, err)
		}
	}

	scopeSetTemplate := &ScopeSetTemplate{
		Granularity:    scope.Granularity,
		TenantID:       scope.TenantID,
		BizID:          scope.BizID,
		SetTemplateIDs: make([]int64, len(items)),
		SetIDs:         make([]int64, len(items)),
		Filter:         scope.Filter,
	}

	for idx := range items {
		scopeSetTemplate.SetTemplateIDs[idx] = items[idx].SetTemplateID
		scopeSetTemplate.SetIDs[idx] = items[idx].SetID
	}

	return scopeSetTemplate, nil
}

type scopeItemInstance struct {
	InstanceID int64
}

// ToScopeInstance converts the common scope to scope instance.
func (scope *Scope) ToScopeInstance() (*ScopeInstance, error) {
	if scope.Type != ScopeTypeInstance {
		return nil, fmt.Errorf("failed to convert scope type to instance, scope-type(%s)", scope.Type)
	}

	items := make([]scopeItemInstance, len(scope.Items))
	for idx := range scope.Items {
		err := conv.MapToStruct(scope.Items[idx], &items[idx])
		if err != nil {
			return nil, fmt.Errorf("failed to convert scope to scope instance, scope(%+v): %w", scope, err)
		}
	}

	scopeInstance := &ScopeInstance{
		Granularity: scope.Granularity,
		BizID:       scope.BizID,
		TenantID:    scope.TenantID,
		InstanceIDs: make([]int64, len(items)),
		Filter:      scope.Filter,
	}

	for idx := range items {
		scopeInstance.InstanceIDs[idx] = items[idx].InstanceID
	}

	return scopeInstance, nil
}

type scopeItemTopo struct {
	TopoObjID  string
	TopoInstID int64
}

// ToScopeTopo converts the common scope to scope topo.
func (scope *Scope) ToScopeTopo() (*ScopeTopo, error) {
	if scope.Type != ScopeTypeTopo {
		return nil, fmt.Errorf("failed to convert scope type to topo, scope-type(%s)", scope.Type)
	}

	items := make([]scopeItemTopo, len(scope.Items))
	for idx := range scope.Items {
		err := conv.MapToStruct(scope.Items[idx], &items[idx])
		if err != nil {
			return nil, fmt.Errorf("failed to convert scope to scope topo, scope(%+v): %w", scope, err)
		}
	}

	scopeTopo := &ScopeTopo{
		Granularity: scope.Granularity,
		BizID:       scope.BizID,
		TenantID:    scope.TenantID,
		Paths:       make([]ScopeTopoNode, len(items)),
		Filter:      scope.Filter,
	}

	for idx := range items {
		scopeTopo.Paths[idx] = ScopeTopoNode{
			TopoObjID:  items[idx].TopoObjID,
			TopoInstID: items[idx].TopoInstID,
		}
	}

	return scopeTopo, nil
}

type scopeItemDynamicGroup struct {
	DynamicGroupID string
}

// ToScopeDynamicGroup converts the common scope to scope dynamic group.
func (scope *Scope) ToScopeDynamicGroup() (*ScopeDynamicGroup, error) {
	if scope.Type != ScopeTypeDynamicGroup {
		return nil, fmt.Errorf("failed to convert scope type to dynamic group, scope-type(%s)", scope.Type)
	}

	items := make([]scopeItemDynamicGroup, len(scope.Items))
	for idx := range scope.Items {
		err := conv.MapToStruct(scope.Items[idx], &items[idx])
		if err != nil {
			return nil, fmt.Errorf("failed to convert scope to scope dynamic group, scope(%+v): %w", scope, err)
		}
	}

	scopeDynamicGroup := &ScopeDynamicGroup{
		Granularity:     scope.Granularity,
		BizID:           scope.BizID,
		TenantID:        scope.TenantID,
		DynamicGroupIDs: make([]string, len(items)),
		Filter:          scope.Filter,
	}

	for idx := range items {
		scopeDynamicGroup.DynamicGroupIDs[idx] = items[idx].DynamicGroupID
	}

	return scopeDynamicGroup, nil
}

// ScopeServiceTemplate defines the scope of service template.
type ScopeServiceTemplate struct {
	Granularity        TargetGranularity
	BizID              int64
	TenantID           string
	Filter             TargetFilter
	ServiceTemplateIDs []int64
	ModuleIDs          []int64
}

// Validate validates the service template scope.
func (scope *ScopeServiceTemplate) Validate() error {
	if err := scope.Granularity.Validate(); err != nil {
		return fmt.Errorf("failed to validate scope service template: %w", err)
	}

	if scope.TenantID == "" {
		return fmt.Errorf("failed to validate scope service template, tenantID is empty")
	}

	if scope.BizID == 0 {
		return fmt.Errorf("failed to validate scope service template, bizID is 0")
	}

	return nil
}

// ToScope converts the service template scope to common scope.
func (scope *ScopeServiceTemplate) ToScope() (*Scope, error) {
	if len(scope.ServiceTemplateIDs) != len(scope.ModuleIDs) {
		return nil, fmt.Errorf("service template IDs and service IDs length mismatch, template-len(%d), service-len(%d)",
			len(scope.ServiceTemplateIDs), len(scope.ModuleIDs))
	}

	scopeItems := make([]map[string]any, len(scope.ServiceTemplateIDs))
	for idx := range scope.ServiceTemplateIDs {
		scopeItem := scopeItemServiceTemplate{
			ServiceTemplateID: scope.ServiceTemplateIDs[idx],
			ServiceID:         scope.ModuleIDs[idx],
		}

		var err error
		scopeItems[idx], err = conv.StructToMap(scopeItem)
		if err != nil {
			return nil, fmt.Errorf("failed to convert scope service template to scope, scope-service-template(%+v): %w", scope, err)
		}
	}

	return &Scope{
		TenantID:    scope.TenantID,
		BizID:       scope.BizID,
		Type:        ScopeTypeServiceTemplate,
		Granularity: scope.Granularity,
		Filter:      scope.Filter,
		Items:       scopeItems,
	}, nil
}

// ScopeSetTemplate defines the scope of set template.
type ScopeSetTemplate struct {
	Granularity    TargetGranularity
	TenantID       string
	BizID          int64
	SetTemplateIDs []int64
	SetIDs         []int64
	Filter         TargetFilter
}

// Validate validates the set template scope.
func (scope *ScopeSetTemplate) Validate() error {
	if err := scope.Granularity.Validate(); err != nil {
		return fmt.Errorf("failed to validate scope set template: %w", err)
	}

	if scope.TenantID == "" {
		return fmt.Errorf("failed to validate scope set template, tenantID is empty")
	}

	if scope.BizID == 0 {
		return fmt.Errorf("failed to validate scope set template, bizID is 0")
	}

	return nil
}

// ToScope converts the set template scope to common scope.
func (scope *ScopeSetTemplate) ToScope() (*Scope, error) {
	if len(scope.SetTemplateIDs) != len(scope.SetIDs) {
		return nil, fmt.Errorf("set template IDs and set IDs length mismatch, template-len(%d), set-len(%d)",
			len(scope.SetTemplateIDs), len(scope.SetIDs))
	}

	scopeItems := make([]map[string]any, len(scope.SetTemplateIDs))
	for idx := range scope.SetTemplateIDs {
		scopeItem := scopeItemSetTemplate{
			SetTemplateID: scope.SetTemplateIDs[idx],
			SetID:         scope.SetIDs[idx],
		}

		var err error
		scopeItems[idx], err = conv.StructToMap(scopeItem)
		if err != nil {
			return nil, fmt.Errorf("failed to convert scope set template to scope, scope-set-template(%+v): %w", scope, err)
		}
	}

	return &Scope{
		TenantID:    scope.TenantID,
		BizID:       scope.BizID,
		Type:        ScopeTypeSetTemplate,
		Granularity: scope.Granularity,
		Filter:      scope.Filter,
		Items:       scopeItems,
	}, nil
}

// ScopeTopo defines the scope of topo.
type ScopeTopo struct {
	Granularity TargetGranularity
	BizID       int64
	TenantID    string
	Paths       []*ScopeTopoNode
	Filter      TargetFilter
}

// ScopeTopoNode defines the scope of topo path.
type ScopeTopoNode struct {
	TopoObjID  string
	TopoInstID int64
}

// Validate validates the scope topo.
func (scope *ScopeTopo) Validate() error {
	if err := scope.Granularity.Validate(); err != nil {
		return fmt.Errorf("failed to validate scope topo: %w", err)
	}

	if scope.BizID == 0 {
		return fmt.Errorf("failed to validate scope topo, bizID is 0")
	}

	if scope.TenantID == "" {
		return fmt.Errorf("failed to validate scope topo, tenantID is empty")
	}

	if len(scope.Paths) == 0 {
		return fmt.Errorf("failed to validate scope topo, paths is empty")
	}

	return nil
}

// ToScope converts the common scope to scope topo.
func (scope *ScopeTopo) ToScope() (*Scope, error) {
	scopeItems := make([]map[string]any, len(scope.Paths))
	for idx := range scope.Paths {
		scopeItem := scopeItemTopo{
			TopoObjID:  scope.Paths[idx].TopoObjID,
			TopoInstID: scope.Paths[idx].TopoInstID,
		}

		var err error
		scopeItems[idx], err = conv.StructToMap(scopeItem)
		if err != nil {
			return nil, fmt.Errorf("failed to convert scope topo to scope, scope-topo(%+v): %w", scope, err)
		}
	}

	return &Scope{
		TenantID:    scope.TenantID,
		BizID:       scope.BizID,
		Type:        ScopeTypeTopo,
		Granularity: scope.Granularity,
		Filter:      scope.Filter,
		Items:       scopeItems,
	}, nil
}

// TargetFilter defines the filter of the target.
type TargetFilter struct {
}

// ScopeInstance defines the scope of instance.
type ScopeInstance struct {
	Granularity TargetGranularity
	BizID       int64
	TenantID    string
	InstanceIDs []int64
	Filter      TargetFilter
}

// Validate validates the instance scope.
func (scope *ScopeInstance) Validate() error {
	if err := scope.Granularity.Validate(); err != nil {
		return fmt.Errorf("failed to validate scope instance: %w", err)
	}

	if scope.BizID == 0 {
		return fmt.Errorf("failed to validate scope instance, bizID is 0")
	}

	if scope.TenantID == "" {
		return fmt.Errorf("failed to validate scope instance, tenantID is empty")
	}

	if len(scope.InstanceIDs) == 0 {
		return fmt.Errorf("failed to validate scope instance, instanceIDs is empty")
	}

	return nil
}

// ToScope converts the instance scope to common scope.
func (scope *ScopeInstance) ToScope() (*Scope, error) {
	scopeItems := make([]map[string]any, len(scope.InstanceIDs))
	for idx := range scope.InstanceIDs {
		scopeItem := scopeItemInstance{
			InstanceID: scope.InstanceIDs[idx],
		}

		var err error
		scopeItems[idx], err = conv.StructToMap(scopeItem)
		if err != nil {
			return nil, fmt.Errorf("failed to convert scope instance to scope, scope-instance(%+v): %w", scope, err)
		}
	}

	return &Scope{
		TenantID:    scope.TenantID,
		BizID:       scope.BizID,
		Type:        ScopeTypeInstance,
		Granularity: scope.Granularity,
		Filter:      scope.Filter,
		Items:       scopeItems,
	}, nil
}

// ScopeDynamicGroup defines the scope of dynamic group.
type ScopeDynamicGroup struct {
	Granularity     TargetGranularity
	BizID           int64
	TenantID        string
	DynamicGroupIDs []string
	Filter          TargetFilter
}

// Validate validates the dynamic group scope.
func (scope *ScopeDynamicGroup) Validate() error {
	if err := scope.Granularity.Validate(); err != nil {
		return fmt.Errorf("failed to validate scope dynamic group: %w", err)
	}

	if scope.BizID == 0 {
		return fmt.Errorf("failed to validate scope dynamic group, bizID is 0")
	}

	if scope.TenantID == "" {
		return fmt.Errorf("failed to validate scope dynamic group, tenantID is empty")
	}

	if len(scope.DynamicGroupIDs) == 0 {
		return fmt.Errorf("failed to validate scope dynamic group, dynamicGroupIDs is empty")
	}

	return nil
}

// ToScope converts the dynamic group scope to common scope.
func (scope *ScopeDynamicGroup) ToScope() (*Scope, error) {
	scopeItems := make([]map[string]any, len(scope.DynamicGroupIDs))
	for idx := range scope.DynamicGroupIDs {
		scopeItem := scopeItemDynamicGroup{
			DynamicGroupID: scope.DynamicGroupIDs[idx],
		}

		var err error
		scopeItems[idx], err = conv.StructToMap(scopeItem)
		if err != nil {
			return nil, fmt.Errorf("failed to convert scope dynamic group to scope, scope-dynamic-group(%+v): %w", scope, err)
		}
	}

	return &Scope{
		TenantID:    scope.TenantID,
		BizID:       scope.BizID,
		Type:        ScopeTypeDynamicGroup,
		Granularity: scope.Granularity,
		Filter:      scope.Filter,
		Items:       scopeItems,
	}, nil
}
