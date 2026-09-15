/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package auth

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// nolint: nonamedreturns
func resolveAuthorizedIDsResult[T comparable](isAny bool, requestedIDs []T, authorizedIDs []T) (narrowedIDs []T, scopeIsEmpty bool) {
	if isAny {
		return requestedIDs, false
	}

	if len(authorizedIDs) == 0 {
		return nil, true
	}

	if len(requestedIDs) == 0 {
		return authorizedIDs, false
	}

	return conv.SliceIntersect(requestedIDs, authorizedIDs), false
}

// ResolveAuthorizedResourceIDsInt64 narrows requested IDs by authorized scope for a specific resource type.
// scopeIsEmpty reports an empty authorized set before intersection, never an unrestricted scope.
// Empty narrowedIDs with neither scopeIsAny nor scopeIsEmpty means the request has no authorized matches.
// nolint: nonamedreturns
func ResolveAuthorizedResourceIDsInt64(scope AuthorizedScope, requestedIDs []int64, resourceType types.AuthResourceType) (
	narrowedIDs []int64, scopeIsAny bool, scopeIsEmpty bool, err error) {

	authorizedIDs := make([]int64, 0, len(scope.Resources))

	for _, resource := range scope.Resources {
		if resource.Type != resourceType {
			continue
		}

		id, convErr := conv.ToInt64(resource.ID)
		if convErr != nil {
			return nil, false, false, fmt.Errorf("auth: convert authorized resource id %q: %w", resource.ID, convErr)
		}
		authorizedIDs = append(authorizedIDs, id)
	}
	authorizedIDs = conv.SliceUnique(authorizedIDs)

	narrowedIDs, scopeIsEmpty = resolveAuthorizedIDsResult(scope.IsAny, requestedIDs, authorizedIDs)

	return narrowedIDs, scope.IsAny, scopeIsEmpty, nil
}

// ResolveAuthorizedResourceIDsString narrows requested IDs by authorized scope for a specific resource type.
// scopeIsEmpty reports an empty authorized set before intersection, never an unrestricted scope.
// Empty narrowedIDs with neither scopeIsAny nor scopeIsEmpty means the request has no authorized matches.
// nolint: nonamedreturns
func ResolveAuthorizedResourceIDsString(scope AuthorizedScope, requestedIDs []string, resourceType types.AuthResourceType) (
	narrowedIDs []string, scopeIsAny bool, scopeIsEmpty bool, err error) {

	authorizedIDs := make([]string, 0, len(scope.Resources))

	for _, resource := range scope.Resources {
		if resource.Type != resourceType {
			continue
		}

		authorizedIDs = append(authorizedIDs, resource.ID)
	}
	authorizedIDs = conv.SliceUnique(authorizedIDs)

	narrowedIDs, scopeIsEmpty = resolveAuthorizedIDsResult(scope.IsAny, requestedIDs, authorizedIDs)

	return narrowedIDs, scope.IsAny, scopeIsEmpty, nil
}
