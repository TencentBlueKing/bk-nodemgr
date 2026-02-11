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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
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

// PackageProvider implements resource.Provider interface for package resources.
type PackageProvider struct {
	storage release.IStorage
}

// NewPackageProvider creates a new PackageProvider.
func NewPackageProvider(storage release.IStorage) *PackageProvider {
	return &PackageProvider{
		storage: storage,
	}
}

// ListAttr returns empty result as package has no attributes.
func (p *PackageProvider) ListAttr(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListAttrData, error) {
	data := &ListAttrData{
		Results: []ResourceAttribute{},
	}

	return data, nil
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
		data := &ListInstanceData{
			Count:   0,
			Results: []ResourceInstance{},
		}

		return data, nil
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
		return &ListInstanceData{
			Count:   0,
			Results: []ResourceInstance{},
		}, nil
	}
}

// listPluginBinToolInstances lists plugin bintool instances, distinct by Name.
func (p *PackageProvider) listPluginBinToolInstances(ctx contextx.IContext, page types.Page) (*ListInstanceData, error) {
	// Query all plugin bintool releases (no filter by Enabled/Generation)
	releases, _, err := p.storage.ListReleasePluginBinTool(ctx, page)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list plugin bintool releases")
		return nil, fmt.Errorf("failed to list plugin bintool releases: %w", err)
	}

	// Extract unique names
	nameSet := make(map[string]bool)
	results := make([]ResourceInstance, 0)
	for _, r := range releases {
		if !nameSet[r.Name] {
			nameSet[r.Name] = true
			results = append(results, ResourceInstance{
				ID:          r.Name,
				DisplayName: r.Name,
			})
		}
	}

	data := &ListInstanceData{
		Count:   int64(len(nameSet)), // Use distinct count
		Results: results,
	}

	return data, nil
}

// listPluginInstances lists plugin instances, distinct by Name.
func (p *PackageProvider) listPluginInstances(ctx contextx.IContext, page types.Page) (*ListInstanceData, error) {
	// Query all plugin releases (no filter by Enabled/Generation)
	releases, _, err := p.storage.ListReleasePlugin(ctx, page)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list plugin releases")
		return nil, fmt.Errorf("failed to list plugin releases: %w", err)
	}

	// Extract unique names
	nameSet := make(map[string]bool)
	results := make([]ResourceInstance, 0)
	for _, r := range releases {
		if !nameSet[r.Name] {
			nameSet[r.Name] = true
			results = append(results, ResourceInstance{
				ID:          r.Name,
				DisplayName: r.Name,
			})
		}
	}

	data := &ListInstanceData{
		Count:   int64(len(nameSet)), // Use distinct count
		Results: results,
	}

	return data, nil
}

// FetchInstanceInfo fetches package details by IDs (Release.Name).
func (p *PackageProvider) FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
	// ids are already []string
	ids := req.Filter.IDs

	if len(ids) == 0 {
		data := &FetchInstanceInfoData{
			Results: []InstanceInfo{},
		}

		return data, nil
	}

	// Build name set for quick lookup
	nameSet := make(map[string]bool)
	for _, id := range ids {
		nameSet[id] = true
	}

	results := make([]InstanceInfo, 0)
	addedSet := make(map[string]bool)

	// Check plugin releases
	p.addPluginReleases(ctx, nameSet, addedSet, &results)

	// Check plugin bintool releases
	p.addPluginBinToolReleases(ctx, nameSet, addedSet, &results)

	// Check fixed types (agent, proxy, cert, bintool)
	p.addFixedTypePackages(nameSet, addedSet, &results)

	data := &FetchInstanceInfoData{
		Results: results,
	}

	return data, nil
}

// addPluginReleases adds matching plugin releases to results.
func (p *PackageProvider) addPluginReleases(ctx contextx.IContext, nameSet map[string]bool, addedSet map[string]bool, results *[]InstanceInfo) {
	// Calculate page limit with multiplier to handle multiple versions per name.
	// Cap at MaxListInstanceByPolicyLimit to avoid excessive database load.
	pageLimit := len(nameSet) * fetchInstanceInfoLimitMultiplier
	if pageLimit > MaxListInstanceByPolicyLimit {
		pageLimit = MaxListInstanceByPolicyLimit
	}

	pluginReleases, _, err := p.storage.ListReleasePlugin(ctx, types.Page{Limit: pageLimit})
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Warn("failed to list plugin releases for FetchInstanceInfo")
		return
	}

	for _, r := range pluginReleases {
		if nameSet[r.Name] && !addedSet[r.Name] {
			*results = append(*results, InstanceInfo{
				ID:          r.Name,
				DisplayName: r.Name,
				Attributes:  make(map[string]interface{}),
			})
			addedSet[r.Name] = true
		}
	}
}

// addPluginBinToolReleases adds matching plugin bintool releases to results.
func (p *PackageProvider) addPluginBinToolReleases(
	ctx contextx.IContext, nameSet map[string]bool, addedSet map[string]bool, results *[]InstanceInfo,
) {
	// Calculate page limit with multiplier to handle multiple versions per name.
	// Cap at MaxListInstanceByPolicyLimit to avoid excessive database load.
	pageLimit := len(nameSet) * fetchInstanceInfoLimitMultiplier
	if pageLimit > MaxListInstanceByPolicyLimit {
		pageLimit = MaxListInstanceByPolicyLimit
	}

	pluginBinToolReleases, _, err := p.storage.ListReleasePluginBinTool(ctx, types.Page{Limit: pageLimit})
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Warn("failed to list plugin bintool releases for FetchInstanceInfo")
		return
	}

	for _, r := range pluginBinToolReleases {
		if nameSet[r.Name] && !addedSet[r.Name] {
			*results = append(*results, InstanceInfo{
				ID:          r.Name,
				DisplayName: r.Name,
				Attributes:  make(map[string]interface{}),
			})
			addedSet[r.Name] = true
		}
	}
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
		if !addedSet[name] {
			if _, ok := fixedTypes[name]; ok {
				*results = append(*results, InstanceInfo{
					ID:          name,
					DisplayName: name,
					Attributes:  make(map[string]interface{}),
				})
				addedSet[name] = true
			}
		}
	}
}

// ListInstanceByPolicy lists package instances filtered by IAM policy expression.
func (p *PackageProvider) ListInstanceByPolicy(ctx contextx.IContext, req *Request[ListInstanceByPolicyFilter]) (*ListInstanceData, error) {
	// Load all releases across all package types (use large limit for in-memory evaluation)
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
			Attributes: map[string]interface{}{},
		})
	}

	// Plugin bintool releases
	pluginBinTools, _, err := p.storage.ListReleasePluginBinTool(ctx, types.Page{Offset: 0, Limit: MaxListInstanceByPolicyLimit})
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list plugin bintool releases")
		return nil, fmt.Errorf("failed to list plugin bintool releases: %w", err)
	}

	nameSet := make(map[string]bool)
	for _, r := range pluginBinTools {
		if !nameSet[r.Name] {
			nameSet[r.Name] = true
			instances = append(instances, InstanceForEval{
				Instance: ResourceInstance{
					ID:          r.Name,
					DisplayName: r.Name,
				},
				Attributes: map[string]interface{}{},
			})
		}
	}

	// Plugin releases
	plugins, _, err := p.storage.ListReleasePlugin(ctx, types.Page{Offset: 0, Limit: MaxListInstanceByPolicyLimit})
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list plugin releases")
		return nil, fmt.Errorf("failed to list plugin releases: %w", err)
	}

	nameSet = make(map[string]bool)
	for _, r := range plugins {
		if !nameSet[r.Name] {
			nameSet[r.Name] = true
			instances = append(instances, InstanceForEval{
				Instance: ResourceInstance{
					ID:          r.Name,
					DisplayName: r.Name,
				},
				Attributes: map[string]interface{}{},
			})
		}
	}

	// Evaluate expression filter and apply pagination
	return evalExpressionFilter(req.Filter.Expression, ResourceTypePackage, instances, req.Page)
}

// SearchInstance searches packages by keyword with optional parent filtering.
func (p *PackageProvider) SearchInstance(ctx contextx.IContext, req *Request[SearchInstanceFilter]) (*ListInstanceData, error) {
	// Check if parent is specified
	if req.Filter.Parent == nil {
		data := &ListInstanceData{
			Count:   0,
			Results: []ResourceInstance{},
		}

		return data, nil
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
		return &ListInstanceData{
			Count:   0,
			Results: []ResourceInstance{},
		}, nil
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

	return &ListInstanceData{
		Count:   0,
		Results: []ResourceInstance{},
	}, nil
}

// searchPluginBinToolInstances searches plugin bintool instances by keyword.
func (p *PackageProvider) searchPluginBinToolInstances(ctx contextx.IContext, keyword string, page types.Page) (*ListInstanceData, error) {
	releases, _, err := p.storage.ListReleasePluginBinTool(ctx, page)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to search plugin bintool releases")
		return nil, fmt.Errorf("failed to search plugin bintool releases: %w", err)
	}

	// Filter by keyword and extract unique names
	nameSet := make(map[string]bool)
	results := make([]ResourceInstance, 0)
	keywordLower := strings.ToLower(keyword)

	for _, r := range releases {
		if !nameSet[r.Name] {
			nameSet[r.Name] = true
			if keyword == "" || strings.Contains(strings.ToLower(r.Name), keywordLower) {
				results = append(results, ResourceInstance{
					ID:          r.Name,
					DisplayName: r.Name,
				})
			}
		}
	}

	data := &ListInstanceData{
		Count:   int64(len(results)),
		Results: results,
	}

	return data, nil
}

// searchPluginInstances searches plugin instances by keyword.
func (p *PackageProvider) searchPluginInstances(ctx contextx.IContext, keyword string, page types.Page) (*ListInstanceData, error) {
	releases, _, err := p.storage.ListReleasePlugin(ctx, page)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to search plugin releases")
		return nil, fmt.Errorf("failed to search plugin releases: %w", err)
	}

	// Filter by keyword and extract unique names
	nameSet := make(map[string]bool)
	results := make([]ResourceInstance, 0)
	keywordLower := strings.ToLower(keyword)

	for _, r := range releases {
		if !nameSet[r.Name] {
			nameSet[r.Name] = true
			if keyword == "" || strings.Contains(strings.ToLower(r.Name), keywordLower) {
				results = append(results, ResourceInstance{
					ID:          r.Name,
					DisplayName: r.Name,
				})
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
func (p *PackageProvider) FetchInstanceList(_ contextx.IContext, _ *Request[FetchInstanceListFilter]) (*ListInstanceData, error) {
	data := &ListInstanceData{
		Count:   0,
		Results: []ResourceInstance{},
	}

	return data, nil
}

// FetchResourceTypeSchema returns empty schema as package has no custom schema.
func (p *PackageProvider) FetchResourceTypeSchema(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListInstanceData, error) {
	data := &ListInstanceData{
		Count:   0,
		Results: []ResourceInstance{},
	}

	return data, nil
}
