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
	"fmt"
	"sort"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// fetchInstanceInfoLimitMultiplier is used to calculate page limit when fetching releases.
	// We multiply by this factor to ensure we get enough results for distinct name matching.
	fetchInstanceInfoLimitMultiplier = 10
)

// ResourceTypePackage is the IAM resource type for package.
const ResourceTypePackage = "package"

type packageStorage interface {
	DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)
	ListReleasePluginWithoutCount(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleasePlugin, error)
	DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)
	ListReleasePluginBinToolWithoutCount(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
		[]*types.ReleasePluginBinTool, error)
}

// PackageProvider implements resource.Provider interface for package resources.
type PackageProvider struct {
	storage packageStorage
}

// NewPackageProvider creates a new PackageProvider.
func NewPackageProvider(storage packageStorage) *PackageProvider {
	return &PackageProvider{
		storage: storage,
	}
}

func (p *PackageProvider) listInstancesForPolicy(ctx contextx.IContext) ([]InstanceForEval, error) {
	instances := make([]InstanceForEval, 0)

	// Fixed types: agent, proxy, cert, bintool
	fixedTypes := []struct {
		releaseType types.ReleaseType
		id          string
		displayName string
	}{
		{types.ReleaseTypeAgent, string(types.ReleaseTypeAgent), string(types.ReleaseTypeAgent)},
		{types.ReleaseTypeProxy, string(types.ReleaseTypeProxy), string(types.ReleaseTypeProxy)},
		{types.ReleaseTypeCert, string(types.ReleaseTypeCert), string(types.ReleaseTypeCert)},
		{types.ReleaseTypeBinTool, string(types.ReleaseTypeBinTool), string(types.ReleaseTypeBinTool)},
	}

	for _, ft := range fixedTypes {
		instances = append(instances, InstanceForEval{
			Instance: ResourceInstance{
				ID:          ft.id,
				DisplayName: ft.displayName,
			},
			Attributes: buildInstanceAttributes(ft.id, ResourceTypePackageType, ft.id),
		})
	}

	pluginBinToolNames, err := p.storage.DistinctNameReleasePluginBinTool(ctx)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to get distinct plugin bintool names")
		return nil, fmt.Errorf("failed to get distinct plugin bintool names: %w", err)
	}

	for _, name := range pluginBinToolNames {
		instances = append(instances, InstanceForEval{
			Instance: ResourceInstance{
				ID:          name,
				DisplayName: name,
			},
			Attributes: buildInstanceAttributes(name, ResourceTypePackageType, string(types.ReleaseTypePluginBinTool)),
		})
	}

	pluginNames, err := p.storage.DistinctNameReleasePlugin(ctx)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to get distinct plugin names")
		return nil, fmt.Errorf("failed to get distinct plugin names: %w", err)
	}

	for _, name := range pluginNames {
		instances = append(instances, InstanceForEval{
			Instance: ResourceInstance{
				ID:          name,
				DisplayName: name,
			},
			Attributes: buildInstanceAttributes(name, ResourceTypePackageType, string(types.ReleaseTypePlugin)),
		})
	}

	return instances, nil
}

// distinctNamesWithPagination gets distinct names, sorts them, and applies pagination.
func distinctNamesWithPagination(names []string, page types.Page) ([]string, int64) {
	// Sort for stable pagination
	sort.Strings(names)

	total := int64(len(names))

	// Apply pagination
	start := page.Offset
	if start > len(names) {
		return []string{}, total
	}

	end := start + page.Limit
	if end > len(names) {
		end = len(names)
	}

	return names[start:end], total
}

// ListAttr returns empty result as package has no attributes.
func (p *PackageProvider) ListAttr(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListAttrData, error) {
	data := ListAttrData([]ResourceAttribute{})

	return &data, nil
}

// ListAttrValue returns empty result as package has no attribute values.
func (p *PackageProvider) ListAttrValue(_ contextx.IContext, _ *Request[ListAttrValueFilter]) (*ListAttrValueData, error) {
	data := &ListAttrValueData{
		Count:   0,
		Results: []AttributeValue{},
	}

	return data, nil
}

// ListInstance lists package instances with pagination and parent filtering.
// NOTE: Package resource has a required parent (package_type) in IAM permission model
// (see support-files/bkiamv3/templates/0002_bk_nodemgr_resource_type.json.tpl lines 55-69).
// When parent=nil, return empty result because IAM will never send list_instance requests
// without specifying the parent package_type. This differs from networkarea which has no parent.
func (p *PackageProvider) ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	// Check if parent is specified
	if req.Filter.Parent == nil {
		return newEmptyListInstanceData(), nil
	}

	// Validate parent type
	if req.Filter.Parent.Type != ResourceTypePackageType {
		err := fmt.Errorf("invalid parent type: expected %s, got %s",
			ResourceTypePackageType, req.Filter.Parent.Type)
		logger.G.Biz(ctx).WithErr(err).Error("invalid parent type in package provider")

		return nil, err
	}

	releaseType := types.ReleaseType(req.Filter.Parent.ID)

	// Handle fixed types (agent, proxy, cert, bintool)
	switch releaseType {
	case types.ReleaseTypeAgent:
		return &ListInstanceData{
			Count:   1,
			Results: []ResourceInstance{{ID: string(types.ReleaseTypeAgent), DisplayName: string(types.ReleaseTypeAgent)}},
		}, nil
	case types.ReleaseTypeProxy:
		return &ListInstanceData{
			Count:   1,
			Results: []ResourceInstance{{ID: string(types.ReleaseTypeProxy), DisplayName: string(types.ReleaseTypeProxy)}},
		}, nil
	case types.ReleaseTypeCert:
		return &ListInstanceData{
			Count:   1,
			Results: []ResourceInstance{{ID: string(types.ReleaseTypeCert), DisplayName: string(types.ReleaseTypeCert)}},
		}, nil
	case types.ReleaseTypeBinTool:
		return &ListInstanceData{
			Count:   1,
			Results: []ResourceInstance{{ID: string(types.ReleaseTypeBinTool), DisplayName: string(types.ReleaseTypeBinTool)}},
		}, nil
	case types.ReleaseTypePluginBinTool:
		return p.listPluginBinToolInstances(ctx, req.Page)
	case types.ReleaseTypePlugin:
		return p.listPluginInstances(ctx, req.Page)
	default:
		// Unknown type, return empty
		return newEmptyListInstanceData(), nil
	}
}

// listPluginBinToolInstances lists plugin bintool instances, distinct by Name.
func (p *PackageProvider) listPluginBinToolInstances(ctx contextx.IContext, page types.Page) (*ListInstanceData, error) {
	// Use DistinctName to get all unique plugin bintool names
	names, err := p.storage.DistinctNameReleasePluginBinTool(ctx)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to get distinct plugin bintool names")
		return nil, fmt.Errorf("failed to get distinct plugin bintool names: %w", err)
	}

	// Sort and paginate
	paginatedNames, total := distinctNamesWithPagination(names, page)

	// Build results
	results := make([]ResourceInstance, 0, len(paginatedNames))
	for _, name := range paginatedNames {
		results = append(results, ResourceInstance{
			ID:          name,
			DisplayName: name,
		})
	}

	data := &ListInstanceData{
		Count:   total,
		Results: results,
	}

	return data, nil
}

// listPluginInstances lists plugin instances, distinct by Name.
func (p *PackageProvider) listPluginInstances(ctx contextx.IContext, page types.Page) (*ListInstanceData, error) {
	// Use DistinctName to get all unique plugin names
	names, err := p.storage.DistinctNameReleasePlugin(ctx)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to get distinct plugin names")
		return nil, fmt.Errorf("failed to get distinct plugin names: %w", err)
	}

	// Sort and paginate
	paginatedNames, total := distinctNamesWithPagination(names, page)

	// Build results
	results := make([]ResourceInstance, 0, len(paginatedNames))
	for _, name := range paginatedNames {
		results = append(results, ResourceInstance{
			ID:          name,
			DisplayName: name,
		})
	}

	data := &ListInstanceData{
		Count:   total,
		Results: results,
	}

	return data, nil
}

// FetchInstanceInfo fetches package details by IDs (Release.Name).
func (p *PackageProvider) FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
	// ids are already []string
	ids := req.Filter.IDs

	if len(ids) == 0 {
		data := FetchInstanceInfoData(nil)
		return &data, nil
	}

	// Build name set for quick lookup
	nameSet := make(map[string]bool)
	for _, id := range ids {
		nameSet[id] = true
	}

	results := make([]InstanceInfo, 0)
	addedSet := make(map[string]bool)

	// Check plugin releases
	if err := p.addPluginReleases(ctx, nameSet, addedSet, &results); err != nil {
		return nil, err
	}

	// Check plugin bintool releases
	if err := p.addPluginBinToolReleases(ctx, nameSet, addedSet, &results); err != nil {
		return nil, err
	}

	// Check fixed types (agent, proxy, cert, bintool)
	p.addFixedTypePackages(nameSet, addedSet, &results)

	data := FetchInstanceInfoData(results)

	return &data, nil
}

// addPluginReleases adds matching plugin releases to results.
func (p *PackageProvider) addPluginReleases(ctx contextx.IContext, nameSet map[string]bool, addedSet map[string]bool, results *[]InstanceInfo) error {
	// Calculate page limit with multiplier to handle multiple versions per name.
	// Cap at MaxListInstanceByPolicyLimit to avoid excessive database load.
	pageLimit := len(nameSet) * fetchInstanceInfoLimitMultiplier
	if pageLimit > MaxListInstanceByPolicyLimit {
		pageLimit = MaxListInstanceByPolicyLimit
	}

	pluginReleases, err := p.storage.ListReleasePluginWithoutCount(ctx, types.Page{Limit: pageLimit})
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list plugin releases for FetchInstanceInfo")
		return fmt.Errorf("failed to list plugin releases: %w", err)
	}

	for _, r := range pluginReleases {
		if nameSet[r.Name] && !addedSet[r.Name] {
			*results = append(*results, InstanceInfo{
				ID:          r.Name,
				DisplayName: r.Name,
				Attributes:  buildInstanceAttributes(r.Name, ResourceTypePackageType, string(types.ReleaseTypePlugin)),
			})
			addedSet[r.Name] = true
		}
	}

	return nil
}

// addPluginBinToolReleases adds matching plugin bintool releases to results.
func (p *PackageProvider) addPluginBinToolReleases(
	ctx contextx.IContext, nameSet map[string]bool, addedSet map[string]bool, results *[]InstanceInfo,
) error {
	// Calculate page limit with multiplier to handle multiple versions per name.
	// Cap at MaxListInstanceByPolicyLimit to avoid excessive database load.
	pageLimit := len(nameSet) * fetchInstanceInfoLimitMultiplier
	if pageLimit > MaxListInstanceByPolicyLimit {
		pageLimit = MaxListInstanceByPolicyLimit
	}

	pluginBinToolReleases, err := p.storage.ListReleasePluginBinToolWithoutCount(ctx, types.Page{Limit: pageLimit})
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list plugin bintool releases for FetchInstanceInfo")
		return fmt.Errorf("failed to list plugin bintool releases: %w", err)
	}

	for _, r := range pluginBinToolReleases {
		if nameSet[r.Name] && !addedSet[r.Name] {
			*results = append(*results, InstanceInfo{
				ID:          r.Name,
				DisplayName: r.Name,
				Attributes:  buildInstanceAttributes(r.Name, ResourceTypePackageType, string(types.ReleaseTypePluginBinTool)),
			})
			addedSet[r.Name] = true
		}
	}

	return nil
}

// addFixedTypePackages adds fixed type packages to results.
func (p *PackageProvider) addFixedTypePackages(nameSet map[string]bool, addedSet map[string]bool, results *[]InstanceInfo) {
	fixedTypes := map[string]string{
		string(types.ReleaseTypeAgent):   string(types.ReleaseTypeAgent),
		string(types.ReleaseTypeProxy):   string(types.ReleaseTypeProxy),
		string(types.ReleaseTypeCert):    string(types.ReleaseTypeCert),
		string(types.ReleaseTypeBinTool): string(types.ReleaseTypeBinTool),
	}

	for name := range nameSet {
		if addedSet[name] {
			continue
		}

		_, ok := fixedTypes[name]
		if !ok {
			continue
		}

		*results = append(*results, InstanceInfo{
			ID:          name,
			DisplayName: name,
			Attributes:  buildInstanceAttributes(name, ResourceTypePackageType, name),
		})
		addedSet[name] = true
	}
}

// ListInstanceByPolicy lists package instances filtered by IAM policy expression.
func (p *PackageProvider) ListInstanceByPolicy(ctx contextx.IContext, req *Request[ListInstanceByPolicyFilter]) (*ListInstanceData, error) {
	instances, err := p.listInstancesForPolicy(ctx)
	if err != nil {
		return nil, err
	}

	// Evaluate expression filter and apply pagination
	return evalExpressionFilter(req.Filter.Expression, ResourceTypePackage, instances, req.Page)
}

// SearchInstance searches packages by keyword with optional parent filtering.
func (p *PackageProvider) SearchInstance(ctx contextx.IContext, req *Request[SearchInstanceFilter]) (*ListInstanceData, error) {
	// Check if parent is specified
	if req.Filter.Parent == nil {
		return newEmptyListInstanceData(), nil
	}

	// Validate parent type
	if req.Filter.Parent.Type != ResourceTypePackageType {
		err := fmt.Errorf("invalid parent type: expected %s, got %s",
			ResourceTypePackageType, req.Filter.Parent.Type)
		logger.G.Biz(ctx).WithErr(err).Error("invalid parent type in package provider")

		return nil, err
	}

	keyword := strings.TrimSpace(req.Filter.Keyword)
	releaseType := types.ReleaseType(req.Filter.Parent.ID)

	// Handle fixed types
	switch releaseType {
	case types.ReleaseTypeAgent:
		return p.searchFixedType(string(types.ReleaseTypeAgent), keyword)
	case types.ReleaseTypeProxy:
		return p.searchFixedType(string(types.ReleaseTypeProxy), keyword)
	case types.ReleaseTypeCert:
		return p.searchFixedType(string(types.ReleaseTypeCert), keyword)
	case types.ReleaseTypeBinTool:
		return p.searchFixedType(string(types.ReleaseTypeBinTool), keyword)
	case types.ReleaseTypePluginBinTool:
		return p.searchPluginBinToolInstances(ctx, keyword, req.Page)
	case types.ReleaseTypePlugin:
		return p.searchPluginInstances(ctx, keyword, req.Page)
	default:
		return newEmptyListInstanceData(), nil
	}
}

// searchFixedType searches fixed type packages by keyword.
// Uses simple case-insensitive string matching (not regex), so no need for QuoteMeta.
func (p *PackageProvider) searchFixedType(name, keyword string) (*ListInstanceData, error) {
	if keyword == "" {
		return &ListInstanceData{
			Count:   1,
			Results: []ResourceInstance{{ID: name, DisplayName: name}},
		}, nil
	}

	keywordLower := strings.ToLower(keyword)
	if strings.Contains(strings.ToLower(name), keywordLower) {
		return &ListInstanceData{
			Count:   1,
			Results: []ResourceInstance{{ID: name, DisplayName: name}},
		}, nil
	}

	return newEmptyListInstanceData(), nil
}

// searchPluginBinToolInstances searches plugin bintool instances by keyword.
func (p *PackageProvider) searchPluginBinToolInstances(ctx contextx.IContext, keyword string, page types.Page) (*ListInstanceData, error) {
	// Get all distinct names
	allNames, err := p.storage.DistinctNameReleasePluginBinTool(ctx)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to get distinct plugin bintool names")
		return nil, fmt.Errorf("failed to get distinct plugin bintool names: %w", err)
	}

	// Filter by keyword (case-insensitive)
	lowerKeyword := strings.ToLower(keyword)
	filteredNames := make([]string, 0)
	for _, name := range allNames {
		if strings.Contains(strings.ToLower(name), lowerKeyword) {
			filteredNames = append(filteredNames, name)
		}
	}

	// Sort and paginate
	paginatedNames, total := distinctNamesWithPagination(filteredNames, page)

	// Build results
	results := make([]ResourceInstance, 0, len(paginatedNames))
	for _, name := range paginatedNames {
		results = append(results, ResourceInstance{
			ID:          name,
			DisplayName: name,
		})
	}

	data := &ListInstanceData{
		Count:   total,
		Results: results,
	}

	return data, nil
}

// searchPluginInstances searches plugin instances by keyword.
func (p *PackageProvider) searchPluginInstances(ctx contextx.IContext, keyword string, page types.Page) (*ListInstanceData, error) {
	// Get all distinct names
	allNames, err := p.storage.DistinctNameReleasePlugin(ctx)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to get distinct plugin names")
		return nil, fmt.Errorf("failed to get distinct plugin names: %w", err)
	}

	// Filter by keyword (case-insensitive)
	lowerKeyword := strings.ToLower(keyword)
	filteredNames := make([]string, 0)
	for _, name := range allNames {
		if strings.Contains(strings.ToLower(name), lowerKeyword) {
			filteredNames = append(filteredNames, name)
		}
	}

	// Sort and paginate
	paginatedNames, total := distinctNamesWithPagination(filteredNames, page)

	// Build results
	results := make([]ResourceInstance, 0, len(paginatedNames))
	for _, name := range paginatedNames {
		results = append(results, ResourceInstance{
			ID:          name,
			DisplayName: name,
		})
	}

	data := &ListInstanceData{
		Count:   total,
		Results: results,
	}

	return data, nil
}

// FetchInstanceList returns empty result as this is for audit center.
func (p *PackageProvider) FetchInstanceList(_ contextx.IContext, _ *Request[FetchInstanceListFilter]) (*ListInstanceData, error) {
	return newEmptyListInstanceData(), nil
}

// FetchResourceTypeSchema returns empty schema as package has no custom schema.
func (p *PackageProvider) FetchResourceTypeSchema(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListInstanceData, error) {
	return newEmptyListInstanceData(), nil
}

// BuildPackageResources constructs package resources for permission checks.
// For agent/proxy/cert/bintool types, the resource ID is the release type itself.
// For plugin types, the resource ID is the plugin name.
// Supports multiple release types for batch permission checks.
// This is the canonical helper used by router handlers to construct types.AuthResource
// matching the resource ID format returned by this provider's IAM callbacks.
func BuildPackageResources(releaseType ...string) []types.AuthResource {
	resources := make([]types.AuthResource, 0, len(releaseType))
	for _, rt := range releaseType {
		resources = append(resources, types.AuthResource{
			SystemID: types.SystemIDNodeMgr,
			Type:     types.AuthResourceTypePackage,
			ID:       rt,
		})
	}

	return resources
}
