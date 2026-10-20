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
	"slices"

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

	narrowedNames, scopeIsAny, scopeIsEmpty, err := auth.ResolveAuthorizedResourceIDsString(
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

	if scopeIsEmpty {
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

// narrowAuthorizedPackageNamesForHistory returns the fixed package types and plugin names the user is authorized to view history for.
func (h *handler) narrowAuthorizedPackageNamesForHistory(
	rCtx restserver.IContext,
) ([]types.ReleaseType, []string, bool, error) {
	scope, err := h.authorizer.ListAuthorizedInstances(rCtx, auth.ActionPackageHistoryView, types.AuthResourceTypePackage)
	if err != nil {
		return nil, nil, false, err
	}

	narrowedNames, scopeIsAny, _, err := auth.ResolveAuthorizedResourceIDsString(
		scope, nil, types.AuthResourceTypePackage,
	)
	if err != nil {
		return nil, nil, false, err
	}

	fixedTypes := make([]types.ReleaseType, 0, len(narrowedNames))
	pluginNames := make([]string, 0, len(narrowedNames))
	for _, name := range narrowedNames {
		releaseType := types.ReleaseType(name)
		if isFixedType(releaseType) {
			fixedTypes = append(fixedTypes, releaseType)
			continue
		}
		pluginNames = append(pluginNames, name)
	}

	return fixedTypes, pluginNames, scopeIsAny, nil
}

// narrowPackageEventConditions creates normal package event conditions for authorized fixed and plugin packages.
func narrowPackageEventConditions(
	condition *types.PackageEventCondition, fixedTypes []types.ReleaseType, pluginNames []string, scopeIsAny bool,
) (*types.PackageEventCondition, *types.PackageEventCondition, bool) {
	if scopeIsAny {
		if condition == nil {
			condition = &types.PackageEventCondition{}
		}
		return condition, nil, true
	}
	if len(fixedTypes) == 0 && len(pluginNames) == 0 {
		return nil, nil, false
	}

	requested := types.PackageEventExactFields{}
	if condition != nil && condition.ExactInclude != nil {
		requested = *condition.ExactInclude
	}
	fixedTypes = slices.Clone(fixedTypes)
	pluginNames = slices.Clone(pluginNames)
	pluginTypes := []types.ReleaseType{types.ReleaseTypePlugin, types.ReleaseTypePluginBinTool}
	if len(requested.ReleaseType) > 0 {
		fixedTypes = slices.DeleteFunc(fixedTypes, func(releaseType types.ReleaseType) bool {
			return !slices.Contains(requested.ReleaseType, releaseType)
		})
		pluginTypes = slices.DeleteFunc(pluginTypes, func(releaseType types.ReleaseType) bool {
			return !slices.Contains(requested.ReleaseType, releaseType)
		})
	}
	if len(requested.Name) > 0 {
		pluginNames = slices.DeleteFunc(pluginNames, func(name string) bool {
			return !slices.Contains(requested.Name, name)
		})
	}

	// Empty exact filters are unrestricted, so omit groups with no authorized matches.
	var fixedCondition, pluginCondition *types.PackageEventCondition
	if len(fixedTypes) > 0 {
		fixedCondition = clonePackageEventCondition(condition)
		fixedCondition.ExactInclude.ReleaseType = fixedTypes
	}
	if len(pluginTypes) > 0 && len(pluginNames) > 0 {
		pluginCondition = clonePackageEventCondition(condition)
		pluginCondition.ExactInclude.ReleaseType = pluginTypes
		pluginCondition.ExactInclude.Name = pluginNames
	}

	return fixedCondition, pluginCondition, fixedCondition != nil || pluginCondition != nil
}

func clonePackageEventCondition(condition *types.PackageEventCondition) *types.PackageEventCondition {
	if condition == nil {
		condition = &types.PackageEventCondition{}
	}
	clone := *condition
	fields := types.PackageEventExactFields{}
	if condition.ExactInclude != nil {
		fields = *condition.ExactInclude
	}
	clone.ExactInclude = &fields
	return &clone
}
