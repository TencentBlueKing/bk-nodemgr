/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package provider

import (
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ResourceTypePackageType is the IAM resource type for package type.
const ResourceTypePackageType = "package_type"

// PackageTypeProvider implements resource.Provider interface for package type resources.
type PackageTypeProvider struct{}

// NewPackageTypeProvider creates a new PackageTypeProvider.
func NewPackageTypeProvider() *PackageTypeProvider {
	return &PackageTypeProvider{}
}

// ListAttr returns empty result as package type has no attributes.
func (p *PackageTypeProvider) ListAttr(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListAttrData, error) {
	data := &ListAttrData{
		Results: []ResourceAttribute{},
	}

	return data, nil
}

// ListAttrValue returns empty result as package type has no attribute values.
func (p *PackageTypeProvider) ListAttrValue(_ contextx.IContext, _ *Request[ListAttrValueFilter]) (*ListAttrValueData, error) {
	data := &ListAttrValueData{
		Count:   0,
		Results: []AttributeValue{},
	}

	return data, nil
}

// ListInstance lists package type instances.
func (p *PackageTypeProvider) ListInstance(_ contextx.IContext, _ *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	// Return predefined package types (excluding origin_* types)
	allTypes := allPackageTypes()
	results := make([]ResourceInstance, 0, len(allTypes))
	packageTypeDisplayNameMap := packageTypeDisplayNames()
	for _, rt := range allTypes {
		results = append(results, ResourceInstance{
			ID:          string(rt),
			DisplayName: packageTypeDisplayNameMap[rt],
		})
	}

	data := &ListInstanceData{
		Count:   int64(len(results)),
		Results: results,
	}

	return data, nil
}

// FetchInstanceInfo fetches package type details by IDs.
func (p *PackageTypeProvider) FetchInstanceInfo(_ contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
	// ids are already []string
	ids := req.Filter.IDs

	if len(ids) == 0 {
		data := FetchInstanceInfoData(nil)
		return &data, nil
	}

	// Convert to IAM response format
	packageTypeDisplayNameMap := packageTypeDisplayNames()
	results := make([]InstanceInfo, 0, len(ids))
	for _, id := range ids {
		if displayName, ok := packageTypeDisplayNameMap[types.ReleaseType(id)]; ok {
			results = append(results, InstanceInfo{
				ID:          id,
				DisplayName: displayName,
				Attributes:  make(map[string]interface{}),
			})
		}
	}

	data := FetchInstanceInfoData(results)

	return &data, nil
}

// ListInstanceByPolicy lists package type instances filtered by IAM policy expression.
func (p *PackageTypeProvider) ListInstanceByPolicy(_ contextx.IContext, req *Request[ListInstanceByPolicyFilter]) (*ListInstanceData, error) {
	// Build all package type instances
	allTypesSlice := allPackageTypes()
	packageTypeDisplayNameMap := packageTypeDisplayNames()
	allTypes := make([]InstanceForEval, 0, len(allTypesSlice))
	for _, rt := range allTypesSlice {
		allTypes = append(allTypes, InstanceForEval{
			Instance:   ResourceInstance{ID: string(rt), DisplayName: packageTypeDisplayNameMap[rt]},
			Attributes: map[string]interface{}{},
		})
	}

	// Evaluate expression filter and apply pagination
	return evalExpressionFilter(req.Filter.Expression, ResourceTypePackageType, allTypes, req.Page)
}

// SearchInstance searches package types by keyword.
func (p *PackageTypeProvider) SearchInstance(_ contextx.IContext, req *Request[SearchInstanceFilter]) (*ListInstanceData, error) {
	keyword := strings.TrimSpace(req.Filter.Keyword)

	// All package types
	allTypesSlice := allPackageTypes()
	packageTypeDisplayNameMap := packageTypeDisplayNames()
	allTypes := make([]ResourceInstance, 0, len(allTypesSlice))
	for _, rt := range allTypesSlice {
		allTypes = append(allTypes, ResourceInstance{
			ID:          string(rt),
			DisplayName: packageTypeDisplayNameMap[rt],
		})
	}

	// Filter by keyword if provided
	results := make([]ResourceInstance, 0)
	if keyword == "" {
		results = allTypes
	} else {
		keywordLower := strings.ToLower(keyword)

		for _, t := range allTypes {
			nameMatch := strings.Contains(strings.ToLower(t.ID), keywordLower) ||
				strings.Contains(strings.ToLower(t.DisplayName), keywordLower)
			if nameMatch {
				results = append(results, t)
			}
		}
	}

	data := &ListInstanceData{
		Count:   int64(len(results)),
		Results: results,
	}

	return data, nil
}

// FetchInstanceList returns empty result as this is for audit center.
func (p *PackageTypeProvider) FetchInstanceList(_ contextx.IContext, _ *Request[FetchInstanceListFilter]) (*ListInstanceData, error) {
	data := &ListInstanceData{
		Count:   0,
		Results: []ResourceInstance{},
	}

	return data, nil
}

// FetchResourceTypeSchema returns empty schema as package type has no custom schema.
func (p *PackageTypeProvider) FetchResourceTypeSchema(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListInstanceData, error) {
	data := &ListInstanceData{
		Count:   0,
		Results: []ResourceInstance{},
	}

	return data, nil
}
