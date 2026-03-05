/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package auth

// SystemID constants for IAM resource ownership.
const (
	// SystemIDCMDB is the IAM system ID for BlueKing CMDB.
	SystemIDCMDB = "bk_cmdb"
	// SystemIDNodeMgr is the IAM system ID for bk-nodemgr itself.
	SystemIDNodeMgr = "bk_nodemgr"
)

// ResourceType represents the IAM resource type identifier.
type ResourceType string

const (
	// ResourceTypeBiz represents a CMDB business resource (system: bk_cmdb).
	// Used for agent, proxy, and plugin operations whose related_resource_types point to cmdb.biz.
	ResourceTypeBiz ResourceType = "biz"

	// ResourceTypeNetworkArea represents a network area resource registered in bk-nodemgr.
	ResourceTypeNetworkArea ResourceType = "networkarea"

	// ResourceTypeNetworkUnit represents a network unit resource registered in bk-nodemgr.
	ResourceTypeNetworkUnit ResourceType = "networkunit"

	// ResourceTypePackageType represents a package type resource registered in bk-nodemgr.
	ResourceTypePackageType ResourceType = "package_type"

	// ResourceTypePackage represents a package resource registered in bk-nodemgr.
	ResourceTypePackage ResourceType = "package"
)

// Resource describes a single resource instance used for authorization checks.
type Resource struct {
	// SystemID is the IAM system that owns this resource (e.g. "bk_cmdb" or "bk_nodemgr").
	SystemID string
	// Type is the resource type identifier.
	Type ResourceType
	// ID is the resource instance ID.
	ID string
}

// ResourceTypeDisplayName returns the human-readable display name for the given resource type.
// Returns an empty string for unknown resource types.
func ResourceTypeDisplayName(t ResourceType) string {
	switch t {
	case ResourceTypeBiz:
		return "业务"
	case ResourceTypeNetworkArea:
		return "云区域"
	case ResourceTypeNetworkUnit:
		return "网络单元"
	case ResourceTypePackageType:
		return "安装包类型"
	case ResourceTypePackage:
		return "安装包"
	default:
		return ""
	}
}

// SystemDisplayName returns the human-readable display name for the given IAM system ID.
func SystemDisplayName(systemID string) string {
	switch systemID {
	case SystemIDNodeMgr:
		return "节点管理"
	case SystemIDCMDB:
		return "配置平台"
	default:
		return systemID
	}
}
