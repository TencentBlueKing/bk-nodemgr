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

// IAMResource represents a single resource for IAM permission checks. It maps
// to the IAM v3 wire type ResourceNode as follows:
//
//	SystemID   → ResourceNode.System    (the system that owns this resource)
//	Type       → ResourceNode.Type
//	ID         → ResourceNode.ID
//	Attributes → ResourceNode.Attribute
//
// Note: SystemID here identifies the resource-owning system, which may differ
// from the request-level SystemID (e.g. NodeMgr checking CMDB host permissions).
type IAMResource struct {
	SystemID   string
	Type       string
	ID         string
	Attributes map[string]interface{}
}

// IAMCheckRequest is the business-semantic request for a single-action
// permission check. SystemID is the target IAM system ID (the same concept as
// iamv3.Config.SystemID and IAMApplyRequest.SystemID).
type IAMCheckRequest struct {
	SystemID  string
	Username  string
	ActionID  string
	Resources []IAMResource
}

// IAMMultiActionCheckRequest is the business-semantic request for a
// multi-action permission check. SystemID carries the same target IAM system ID
// semantics as IAMCheckRequest.SystemID.
type IAMMultiActionCheckRequest struct {
	SystemID  string
	Username  string
	ActionIDs []string
	Resources []IAMResource
}

// IAMApplyResourceType identifies a resource type within a specific system for
// apply URL generation. SystemID is the resource-owning system, matching the
// convention used in IAMResource.SystemID.
type IAMApplyResourceType struct {
	SystemID  string
	Type      string
	Instances []IAMApplyResourceInstance
}

// IAMApplyResourceNode identifies one node on a resource instance path used in
// IAM apply URL generation.
type IAMApplyResourceNode struct {
	Type string
	ID   string
}

// IAMApplyResourceInstance represents one resource instance path. For
// topology-style resources, this may contain multiple nodes ordered by the
// instance view chain.
type IAMApplyResourceInstance []IAMApplyResourceNode

// IAMApplyAction describes a single action and its related resource types for
// apply URL generation. RelatedResourceTypes lists the resource types the
// action operates on.
type IAMApplyAction struct {
	ID                   string
	RelatedResourceTypes []IAMApplyResourceType
}

// IAMApplyRequest is the business-semantic input for obtaining a permission
// apply URL from IAM. SystemID is the target IAM system ID (same concept as
// IAMCheckRequest.SystemID and iamv3.Config.SystemID). Actions may contain
// multiple entries to request permissions for several actions at once.
type IAMApplyRequest struct {
	SystemID string
	Actions  []IAMApplyAction
}
