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
