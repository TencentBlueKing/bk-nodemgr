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

// SystemID constants for IAM resource ownership.
const (
	// SystemIDCMDB is the IAM system ID for BlueKing CMDB.
	SystemIDCMDB = "bk_cmdb"
	// SystemIDNodeMgr is the IAM system ID for bk-nodemgr itself.
	SystemIDNodeMgr = "bk_nodemgr"
)

// AuthResourceType represents the IAM resource type identifier.
type AuthResourceType string

const (
	// AuthResourceTypeBiz represents a CMDB business resource (system: bk_cmdb).
	// Used for agent, proxy, and plugin operations whose related_resource_types point to cmdb.biz.
	AuthResourceTypeBiz AuthResourceType = "biz"

	// AuthResourceTypeNetworkArea represents a network area resource registered in bk-nodemgr.
	AuthResourceTypeNetworkArea AuthResourceType = "networkarea"

	// AuthResourceTypeNetworkUnit represents a network unit resource registered in bk-nodemgr.
	AuthResourceTypeNetworkUnit AuthResourceType = "networkunit"

	// AuthResourceTypePackageType represents a package type resource registered in bk-nodemgr.
	AuthResourceTypePackageType AuthResourceType = "package_type"

	// AuthResourceTypePackage represents a package resource registered in bk-nodemgr.
	AuthResourceTypePackage AuthResourceType = "package"
)

// AuthResource describes a single resource instance used for authorization checks.
type AuthResource struct {
	// SystemID is the IAM system that owns this resource (e.g. "bk_cmdb" or "bk_nodemgr").
	SystemID string
	// Type is the resource type identifier.
	Type AuthResourceType
	// ID is the resource instance ID.
	ID string
	// Attributes stores resource attributes for IAM checks (e.g. _bk_iam_path_).
	Attributes map[string]interface{}
}

// AuthResourceTypeDisplayName returns the human-readable display name for the given resource type.
// Returns an empty string for unknown resource types.
func AuthResourceTypeDisplayName(art AuthResourceType) string {
	switch art {
	case AuthResourceTypeBiz:
		return "业务"
	case AuthResourceTypeNetworkArea:
		return "管控区域"
	case AuthResourceTypeNetworkUnit:
		return "管控单元"
	case AuthResourceTypePackageType:
		return "资源包类型"
	case AuthResourceTypePackage:
		return "资源包"
	default:
		return ""
	}
}

// AuthResourceTypeToSystemID returns the IAM system ID that owns the given resource type.
func AuthResourceTypeToSystemID(art AuthResourceType) string {
	switch art {
	case AuthResourceTypeBiz:
		return SystemIDCMDB
	default:
		return SystemIDNodeMgr
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

// AuthVerifyItem represents a single action and its associated resources for authorization verification.
type AuthVerifyItem struct {
	Action    string
	Resources []AuthResource
}

// AuthVerifyResult represents the authorization check result for a single action.
type AuthVerifyResult struct {
	Action     string
	Authorized bool
}

// AuthorizedItem represents a single action-resource type pair for authorized scope querying.
type AuthorizedItem struct {
	Action       string
	ResourceType AuthResourceType
}

// AuthorizedResult represents the authorized scope for a single action-resource type pair.
type AuthorizedResult struct {
	Action       string
	ResourceType AuthResourceType
	IsAny        bool
	Resources    []AuthResource
}
