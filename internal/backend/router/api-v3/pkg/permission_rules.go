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

package pkg

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var (
	errPackageViewDeniedByEmptyScope = errors.New("package view permission denied: no authorized resources")
)

// narrowAuthorizedPackageNames narrows requested package names by authorized scope.
// For fixed types (agent/proxy/cert/bintool), requestedNames should be nil or empty.
// For plugin types, requestedNames contains the plugin names from query conditions.
func (h *handler) narrowAuthorizedPackageNames(
	rCtx restserver.IContext, requestedNames []string, releaseType types.ReleaseType,
) ([]string, bool, error) {

	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, auth.ActionPackageView, types.AuthResourceTypePackage)
	if err != nil {
		return nil, false, err
	}

	narrowedNames, scopeIsAny, err := auth.ResolveAuthorizedResourceIDsString(
		scope, requestedNames, types.AuthResourceTypePackage,
	)
	if err != nil {
		return nil, false, err
	}

	// For fixed types (agent/proxy/cert/bintool), the resource ID is the release type itself.
	// Check if the release type is in the authorized list.
	if isFixedType(releaseType) {
		releaseTypeStr := string(releaseType)
		if scopeIsAny {
			return nil, true, nil
		}
		// Check if the release type is authorized.
		for _, name := range narrowedNames {
			if name == releaseTypeStr {
				return nil, false, nil
			}
		}
		// Release type not authorized - trigger permission check to generate proper error.
		resources := auth.BuildPackageResources(releaseTypeStr)
		if checkErr := h.authorizer.Check(rCtx, auth.ActionPackageView, resources); checkErr != nil {
			return nil, false, checkErr
		}

		return nil, false, errPackageViewDeniedByEmptyScope
	}

	// For plugin types, return the narrowed plugin names.
	if scopeIsAny {
		return requestedNames, true, nil
	}

	if len(narrowedNames) == 0 {
		// User requested specific plugins but has no permission for any of them.
		resources := auth.BuildPackageResources(requestedNames...)
		if checkErr := h.authorizer.Check(rCtx, auth.ActionPackageView, resources); checkErr != nil {
			return nil, false, checkErr
		}
	}

	return narrowedNames, false, nil
}

// isFixedType returns true if the release type is a fixed type (agent/proxy/cert/bintool).
func isFixedType(releaseType types.ReleaseType) bool {
	switch releaseType {
	case types.ReleaseTypeAgent, types.ReleaseTypeProxy, types.ReleaseTypeCert, types.ReleaseTypeBinTool:
		return true
	default:
		return false
	}
}

// narrowReleaseCondition injects authorized package names into the release condition.
// For fixed types with scopeIsAny=true, returns the original condition.
// For plugin types, injects narrowedNames into ExactInclude.Name.
func narrowReleaseCondition(
	condition *types.ReleaseCondition, narrowedNames []string, scopeIsAny bool, releaseType types.ReleaseType,
) *types.ReleaseCondition {

	if scopeIsAny {
		return condition
	}

	// For fixed types, scopeIsAny=false means no permission (already handled in narrowAuthorizedPackageNames).
	// This function should not be called in that case.
	if isFixedType(releaseType) {
		return condition
	}

	// For plugin types, inject narrowed names.
	if condition == nil {
		condition = &types.ReleaseCondition{}
	}
	if condition.ExactInclude == nil {
		condition.ExactInclude = &types.ReleaseExactFields{}
	}
	condition.ExactInclude.Name = narrowedNames

	return condition
}
