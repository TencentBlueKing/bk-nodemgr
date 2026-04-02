/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package errf

// Permission defines the permission struct.
type Permission struct {
	System     string   `json:"system"`
	SystemName string   `json:"system_name"`
	ApplyURL   string   `json:"apply_url,omitempty"`
	Actions    []Action `json:"actions,omitempty"`
}

// Action defines the action struct for IAM permission apply response.
type Action struct {
	// ID is the IAM action ID.
	ID string `json:"id"`
	// Name is the display name of the action.
	Name string `json:"name"`
	// RelatedResourceTypes lists the resource types this action operates on.
	RelatedResourceTypes []RelatedResourceType `json:"related_resource_types"`
}

// ResourceNode describes one resource node on an IAM resource instance path.
type ResourceNode struct {
	// Type is the resource type identifier.
	Type string `json:"type"`
	// TypeName is the display name of the resource type.
	TypeName string `json:"type_name"`
	// ID is the resource instance identifier.
	ID string `json:"id"`
	// Name is the display name of the resource instance.
	Name string `json:"name"`
}

// RelatedResourceType describes a resource type associated with an IAM action.
type RelatedResourceType struct {
	// SystemID is the IAM system that owns this resource type.
	SystemID string `json:"system_id"`
	// SystemName is the display name of the IAM system that owns this resource type.
	SystemName string `json:"system_name"`
	// Type is the resource type identifier.
	Type string `json:"type"`
	// TypeName is the display name of the resource type.
	TypeName string `json:"type_name"`
	// Instances lists the denied resource instances for this resource type.
	Instances []ResourceNode `json:"instances,omitempty"`
}

// PermissionError is implemented by errors that carry IAM permission-apply data.
type PermissionError interface {
	error
	// PermissionData returns the Permission payload to embed in the response.
	PermissionData() Permission
}
