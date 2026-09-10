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

package provider

import (
	"fmt"
	"slices"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ResourceTypePackage is the IAM resource type for packages.
const ResourceTypePackage = "package"

type packageStorage interface {
	DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)
	DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error)
}

// PackageProvider queries canonical package names without release-version truncation.
type PackageProvider struct{ storage packageStorage }

// NewPackageProvider creates a package provider.
func NewPackageProvider(storage packageStorage) *PackageProvider {
	return &PackageProvider{storage: storage}
}

// ListInstance enumerates packages under an optional package type.
func (p *PackageProvider) ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	releaseTypes := allPackageTypes()
	if parent := req.Filter.Parent; parent != nil {
		if parent.Type != ResourceTypePackageType || parent.ID == "" || strings.ContainsAny(parent.ID, "/, \t\r\n") {
			return nil, ErrInvalidArgument
		}
		if !slices.Contains(releaseTypes, types.ReleaseType(parent.ID)) {
			return newEmptyListInstanceData(), nil
		}
		releaseTypes = []types.ReleaseType{types.ReleaseType(parent.ID)}
	}
	names := make([]string, 0)
	seen := make(map[string]types.ReleaseType)
	for _, releaseType := range releaseTypes {
		typeNames, err := p.packageNames(ctx, releaseType)
		if err != nil {
			return nil, err
		}
		for _, name := range typeNames {
			if previous, ok := seen[name]; ok && previous != releaseType {
				return nil, fmt.Errorf("ambiguous canonical package ID %q across package types", name)
			}
			seen[name] = releaseType
			names = append(names, name)
		}
	}
	filtered := make([]string, 0, len(names))
	for _, name := range names {
		if strings.Contains(strings.ToLower(name), strings.ToLower(req.Filter.Keyword)) {
			filtered = append(filtered, name)
		}
	}
	pageNames, total := distinctNamesWithPagination(filtered, req.Page)
	results := make([]ResourceInstance, 0, len(pageNames))
	for _, name := range pageNames {
		results = append(results, ResourceInstance{ID: name, DisplayName: name})
	}

	return &ListInstanceData{Count: total, Results: results}, nil
}

func distinctNamesWithPagination(names []string, page types.Page) ([]string, int64) {
	slices.Sort(names)
	total := int64(len(names))
	if page.Offset >= len(names) {
		return []string{}, total
	}
	end := page.Offset + min(page.Limit, len(names)-page.Offset)

	return names[page.Offset:end], total
}

func (p *PackageProvider) packageNames(
	ctx contextx.IContext, releaseType types.ReleaseType, conditions ...*types.ReleaseCondition,
) ([]string, error) {

	var names []string
	var err error
	switch releaseType {
	case types.ReleaseTypePlugin:
		names, err = p.storage.DistinctNameReleasePlugin(ctx, conditions...)
	case types.ReleaseTypePluginBinTool:
		names, err = p.storage.DistinctNameReleasePluginBinTool(ctx, conditions...)
	default:
		return []string{string(releaseType)}, nil
	}
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to get distinct package names")
		return nil, fmt.Errorf("failed to get distinct %s names: %w", releaseType, err)
	}

	return names, nil
}

// FetchInstanceInfo preserves plugin, pluginbintool, then fixed-type precedence.
func (p *PackageProvider) FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
	data := FetchInstanceInfoData{}
	if len(req.Filter.IDs) == 0 {
		return &data, nil
	}
	condition := &types.ReleaseCondition{ExactInclude: &types.ReleaseExactFields{Name: req.Filter.IDs}}
	parents := make(map[string]types.ReleaseType)
	for _, releaseType := range []types.ReleaseType{types.ReleaseTypePlugin, types.ReleaseTypePluginBinTool} {
		names, err := p.packageNames(ctx, releaseType, condition)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			if _, ok := parents[name]; !ok {
				parents[name] = releaseType
			}
		}
	}
	for _, releaseType := range []types.ReleaseType{types.ReleaseTypeAgent, types.ReleaseTypeProxy, types.ReleaseTypeCert, types.ReleaseTypeBinTool} {
		if _, ok := parents[string(releaseType)]; !ok {
			parents[string(releaseType)] = releaseType
		}
	}
	for _, id := range req.Filter.IDs {
		parent, ok := parents[id]
		if !ok {
			continue
		}
		data = append(data, InstanceInfo{ID: id, DisplayName: id, Attributes: buildInstanceAttributes(id, ResourceTypePackageType, string(parent))})
	}

	return &data, nil
}
