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

// Set represents the set structure of cmdb.
type Set struct {
	SetID              int64
	SetName            string
	SetDesc            string
	SetEnv             string
	BizID              int64
	Capacity           int64
	ParentID           int64
	Description        string
	SetTemplateID      int64
	SetTemplateVersion int64
	ServiceStatus      string
}

// Module represents the module structure of cmdb.
type Module struct {
	ModuleID          int64
	ModuleName        string
	SetID             int64
	BakOperator       string
	BizID             int64
	ModuleType        string
	ParentID          int64
	HostApplyEnabled  bool
	ServiceCategoryID int64
	ServiceTemplateID int64
	SetTemplateID     int64
}

// BusinessInstanceTopo represents the topology of a business instance.
type BusinessInstanceTopo struct {
	// belongs to
	TenantID string

	InstID   int64
	InstName string
	ObjID    string
	ObjName  string
	Children []*BusinessInstanceTopo
}

// BusinessInternalModule represents the module information of idle, fault, and recycle under the business.
type BusinessInternalModule struct {
	// belongs to
	TenantID string

	SetID   int64
	SetName string
	Module  []*Module
}

// TopoNode represents the topology node defined by cmdb.
type TopoNode struct {
	ObjID    string
	InstID   int64
	InstName string
}

// TopoNodePath represents the topology node path defined by cmdb.
type TopoNodePath struct {
	ObjID    string
	InstID   int64
	InstName string
	Paths    [][]*TopoNode
}

// CloudVendor represents a cloud vendor option.
type CloudVendor struct {
	Key  string
	Name string
}

// OsType represents a os type option.
type OsType struct {
	Key  string
	Name string
}

// AddressingType represents an addressing type.
type AddressingType string

const (
	// DynamicAddressing means the addressing type is dynamic.
	DynamicAddressing AddressingType = "dynamic"

	// StaticAddressing means the addressing type is static.
	StaticAddressing AddressingType = "static"
)

// CreateHostInfo represents the host information to be created.
type CreateHostInfo struct {
	NetworkAreaID int64
	InnerIP       string
	InnerIPV6     string
	OuterIP       string
	OuterIPV6     string
	OSType        string
	Arch          string
	Addressing    AddressingType
}

// UpdateHostProperties represents the properties of host to be updated.
type UpdateHostProperties struct {
	HostName string
	Operator string
	Comment  string
	IspName  string
	HostID   int64
}

// ServiceTemplate represents a service template option.
type ServiceTemplate struct {
	ID                  int64
	BizID               int64
	ServiceTemplateName string
	ServiceCategoryID   int64
	HostApplyEnabled    bool
}

// ServiceInstance represents a service instance option.
type ServiceInstance struct {
	ID                  int64
	BizID               int64
	ServiceInstanceName string
	HostID              int64
	ModuleID            int64
}

// ServiceInstanceDetail represents a service instance detail option.
type ServiceInstanceDetail struct {
	ID                  int64
	BizID               int64
	ServiceTemplateName string
	ServiceTemplateID   int64
	HostID              int64
	ModuleID            int64
	ServiceCategoryID   int64
	ProcessInstances    []*ProcessInstance
}

// BindInfo represents a bind info option.
type BindInfo struct {
	Enable        bool
	IP            string
	Port          string
	Protocol      string
	TemplateRowID int64
}

// ProcessProperty represents a process property option.
type ProcessProperty struct {
	AutoStart       bool
	BizID           int64
	FuncName        string
	ProcessID       int64
	ProcessName     string
	StartParamRegex string
	Description     string
	FaceStopCMD     string
	PidFile         string
	Priority        int64
	ProcNum         int64
	ReloadCMD       string
	RestartCMD      string
	StartCMD        string
	StopCMD         string
	Timeout         int64
	User            string
	WorkPath        string
	BindInfo        []*BindInfo
}

// ProcessRelation represents a process relation option.
type ProcessRelation struct {
	BizID             int64
	ProcessID         int64
	ServiceInstanceID int64
	ProcessTemplateID int64
	HostID            int64
}

// ProcessInstance represents a process instance option.
type ProcessInstance struct {
	Property *ProcessProperty
	Relation *ProcessRelation
}

// ProcessTemplate represents a process template option.
type ProcessTemplate struct {
	ID                  int64
	BizID               int64
	ProcessTemplateName string
	ServiceTemplateID   int64
	Property            *ProcessProperty
}

// HostTopoRelation represents a host topology relation option.
type HostTopoRelation struct {
	BizID    int64
	HostID   int64
	SetID    int64
	ModuleID int64
}

// MainlineObjectTopo represents a mainline object topology option.
type MainlineObjectTopo struct {
	ObjID      string
	ObjName    string
	NextObj    string
	NextName   string
	PreObjID   string
	PreObjName string
}

// HostTopo represents a host topology option.
type HostTopo struct {
	SetID   int64
	SetName string
	Module  []*struct {
		ModuleID   int64
		ModuleName string
	}
}

// HostWithHostTopo represents a host info with topology option.
type HostWithHostTopo struct {
	Host     *Host
	HostTopo []*HostTopo
}

// HostAgentID represents the agent id of host.
type HostAgentID struct {
	HostID  int64
	AgentID string
}

// HostIdentification represents the identification of host.
type HostIdentification struct {
	HostID         int64
	Identification string
}

// PushHostIdentifierTaskResult represents the result of push host identifier task.
type PushHostIdentifierTaskResult struct {
	TaskID   string
	HostInfo []*HostIdentification
}

// ResourceWatchEventType represents the event type of watch event.
type ResourceWatchEventType string

const (
	// ResourceWatchEventTypeCreate represents the create event type.
	ResourceWatchEventTypeCreate ResourceWatchEventType = "create"

	// ResourceWatchEventTypeUpdate represents the update event type.
	ResourceWatchEventTypeUpdate ResourceWatchEventType = "update"

	// ResourceWatchEventTypeDelete represents the delete event type.
	ResourceWatchEventTypeDelete ResourceWatchEventType = "delete"
)

// ResourceWatchEventResource represents the resource type of watch event.
type ResourceWatchEventResource string

const (
	// ResourceWatchEventResourceHost represents the host resource type.
	ResourceWatchEventResourceHost ResourceWatchEventResource = "host"

	// ResourceWatchEventResourceHostRelation represents the host relation resource type.
	ResourceWatchEventResourceHostRelation ResourceWatchEventResource = "host_relation"

	// ResourceWatchEventResourceBiz represents the biz resource type.
	ResourceWatchEventResourceBiz ResourceWatchEventResource = "biz"

	// ResourceWatchEventResourceSet represents the set resource type.
	ResourceWatchEventResourceSet ResourceWatchEventResource = "set"

	// ResourceWatchEventResourceModule represents the module resource type.
	ResourceWatchEventResourceModule ResourceWatchEventResource = "module"

	// ResourceWatchEventResourceProcess represents the process resource type.
	ResourceWatchEventResourceProcess ResourceWatchEventResource = "process"

	// ResourceWatchEventResourceObjectInstance represents the object instance resource type.
	ResourceWatchEventResourceObjectInstance ResourceWatchEventResource = "object_instance"

	// ResourceWatchEventResourceMainlineInstance represents the mainline instance resource type.
	ResourceWatchEventResourceMainlineInstance ResourceWatchEventResource = "mainline_instance"

	// ResourceWatchEventResourceBizSet represents the biz set resource type.
	ResourceWatchEventResourceBizSet ResourceWatchEventResource = "biz_set"

	// ResourceWatchEventResourceBizSetRelation represents the biz set relation resource type.
	ResourceWatchEventResourceBizSetRelation ResourceWatchEventResource = "biz_set_relation"

	// ResourceWatchEventResourcePlat represents the plat resource type.
	ResourceWatchEventResourcePlat ResourceWatchEventResource = "plat"

	// ResourceWatchEventResourceProject represents the project resource type.
	ResourceWatchEventResourceProject ResourceWatchEventResource = "project"
)

// ResourceWatchEvent represents the watch event of resource.
type ResourceWatchEvent struct {
	EventType []ResourceWatchEventType
	Cursor    string
	Resource  ResourceWatchEventResource
	Fields    []string
	// StartFrom is the start position of the watch event (UNIX time).
	StartFrom int64
}

// HostResourceWatchEvent represents the watch event of host.
type HostResourceWatchEvent struct {
	EventType ResourceWatchEventType
	Cursor    string
	Resource  ResourceWatchEventResource
	Detail    *Host
}

// HostRelationResourceWatchEvent represents the watch event of host relation.
type HostRelationResourceWatchEvent struct {
	EventType ResourceWatchEventType
	Cursor    string
	Resource  ResourceWatchEventResource
	Detail    *HostTopoRelation
}

// ProcessResourceWatchEvent represents the watch event of process.
type ProcessResourceWatchEvent struct {
	EventType ResourceWatchEventType
	Cursor    string
	Resource  ResourceWatchEventResource
	Detail    *ProcessProperty
}

// FieldCondition represents the field condition of cmdb.
type FieldCondition struct {
	Field string
	// Opreator can be $in $nin $eq $ne
	Operator string
	// Value's type is related to the operator, $in and $nin is []any other is any
	Value any
}

// DynamicGroupCondition represents the dynamic group condition of cmdb.
type DynamicGroupCondition struct {
	ObjID     string
	Condition []*FieldCondition
}

// DynamicGroup represents the dynamic group of cmdb.
type DynamicGroup struct {
	ID                string
	BizID             int64
	ObjID             string
	Name              string
	Condition         []*DynamicGroupCondition
	VariableCondition []*DynamicGroupCondition
}
