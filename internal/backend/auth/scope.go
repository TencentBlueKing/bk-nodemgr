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

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

func resolveAuthorizedIDsResult[T comparable](
	isAny bool, requestedIDs []T, authorizedIDs []T,
) (narrowedIDs []T, fullAccess bool, hasAuthorized bool) {

	if isAny {
		return requestedIDs, true, true
	}

	if len(authorizedIDs) == 0 {
		return nil, false, false
	}

	if len(requestedIDs) == 0 {
		return authorizedIDs, false, true
	}

	return conv.SliceIntersect(requestedIDs, authorizedIDs), false, true
}

// ResolveAuthorizedResourceIDsInt64 narrows requested IDs by authorized scope for a specific resource type.
func ResolveAuthorizedResourceIDsInt64(
	scope AuthorizedScope, requestedIDs []int64, resourceType ResourceType,
) (narrowedIDs []int64, fullAccess bool, hasAuthorized bool, err error) {

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

	narrowedIDs, fullAccess, hasAuthorized = resolveAuthorizedIDsResult(scope.IsAny, requestedIDs, authorizedIDs)

	return narrowedIDs, fullAccess, hasAuthorized, nil
}
