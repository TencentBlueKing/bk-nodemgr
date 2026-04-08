/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package auth provider the interface to check permission.
package auth

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AuthorizedScope describes the auth-domain authorized instance scope for one action/resource type query.
type AuthorizedScope struct {
	// IsAny indicates whether the caller has full access to the target resource type.
	IsAny bool
	// Resources contains the authorized resource instances when IsAny is false.
	Resources []types.AuthResource
}

// IAuthorizer defines the contract for permission checking.
type IAuthorizer interface {
	// Check verifies whether the current request context can perform one action on the given resources.
	// Empty resources trigger an action-level IAM check, while non-empty resources are evaluated
	// per resource and aggregated into one permission error.
	Check(ctx contextx.IContext, action Action, resources []types.AuthResource) error

	// CheckMany verifies multiple actions over their corresponding resources and aggregates denied actions.
	// Each action follows the same action-level versus per-resource evaluation rules as Check.
	CheckMany(ctx contextx.IContext, actionResources map[Action][]types.AuthResource) error

	// ListAuthorizedInstances returns the auth-domain authorized scope for one action and resource type.
	ListAuthorizedInstances(ctx contextx.IContext, action Action, resourceType types.AuthResourceType) (AuthorizedScope, error)
}
