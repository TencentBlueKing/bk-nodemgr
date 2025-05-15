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

// HostDistinctRequest describes the wanted distinct fields.
type HostDistinctRequest struct {
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
