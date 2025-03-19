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

import "time"

// Module represents the module structure of cmdb
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
	SupplierAccount   string
	CreatedBy         string
	Operator          string
	LastTime          time.Time
	CreateTime        time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// BusinessInstanceTopo represents the topology of a business instance
type BusinessInstanceTopo struct {
	// belongs to
	TenantID string

	InstID   int64
	InstName string
	ObjID    string
	ObjName  string
	Children []*BusinessInstanceTopo
}

// BusinessInternalModule represents the module information of idle host, fault host, and recycle host under the business
type BusinessInternalModule struct {
	// belongs to
	TenantID string

	SetID   int64
	SetName string
	Module  []*Module
}

// TopoNode represents the topology node defined by CMDB
type TopoNode struct {
	ObjID    string
	InstID   int64
	InstName string
}

// TopoNodePath represents the topology node path defined by CMDB
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
