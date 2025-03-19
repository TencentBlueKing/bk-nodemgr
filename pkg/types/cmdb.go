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

// ServiceTemplate represents a service template option.
type ServiceTemplate struct {
	ID                  int64
	BizID               int64
	ServiceTemplateName string
	ServiceCategoryID   int64
	HostApplyEnabled    bool
}

// SetTemplate represents a set template option.
type SetTemplate struct {
	ID              int64
	BizID           int64
	SetTemplateName string
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

// PushHostIdentifierTaskInfo represents the result of push host identifier task.
type PushHostIdentifierTaskInfo struct {
	TaskID   string
	HostInfo []*HostIdentification
}

// PushHostIdentifiersTaskResult represents the result of push identifiers.
type PushHostIdentifiersTaskResult struct {
	SuccessHostIDList []int64
	FailedHostIDList  []int64
	PendingHostIDList []int64
}

// DynamicGroup represents the dynamic group of cmdb.
type DynamicGroup struct {
	ID    string
	BizID int64
	ObjID string
	Name  string
}
