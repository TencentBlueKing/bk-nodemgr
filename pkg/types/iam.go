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

// IAMResource represents a resource for IAM permission checks.
type IAMResource struct {
	SystemID   string
	Type       string
	ID         string
	Attributes map[string]interface{}
}

// IAMCheckRequest is the business-semantic request for a single-action permission check.
// It describes the subject, the action, and the resources to check against.
type IAMCheckRequest struct {
	System    string
	Username  string
	ActionID  string
	Resources []IAMResource
}

// IAMMultiActionCheckRequest is the business-semantic request for multi-action permission check.
// It allows checking multiple actions against the same set of resources in a single call.
type IAMMultiActionCheckRequest struct {
	System    string
	Username  string
	ActionIDs []string
	Resources []IAMResource
}

// IAMApplyResourceType is a deduplicated resource type for apply URL generation.
// It identifies a resource type within a specific system.
type IAMApplyResourceType struct {
	SystemID string
	Type     string
}

// IAMApplyAction is a single action with its related resource types for apply URL generation.
// RelatedResourceTypes lists the resource types that the action operates on.
type IAMApplyAction struct {
	ID                   string
	RelatedResourceTypes []IAMApplyResourceType
}

// IAMApplyRequest is the business-semantic request for getting a permission apply URL.
// Supports multiple actions to avoid capability regression vs the current Application wire type,
// which also carries a slice of actions (Application.Actions []ApplicationAction).
type IAMApplyRequest struct {
	SystemID string
	Actions  []IAMApplyAction
}
