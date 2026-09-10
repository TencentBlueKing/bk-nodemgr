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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ResourceTypePackageType is the IAM resource type for package types.
const ResourceTypePackageType = "package_type"

// PackageTypeProvider exposes the supported package types.
type PackageTypeProvider struct{}

// NewPackageTypeProvider creates a package type provider.
func NewPackageTypeProvider() *PackageTypeProvider { return &PackageTypeProvider{} }

// ListInstance filters and paginates supported package types.
func (p *PackageTypeProvider) ListInstance(_ contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	if req.Filter.Parent != nil {
		return nil, ErrInvalidArgument
	}
	names := packageTypeDisplayNames()
	results := make([]ResourceInstance, 0, len(names))
	for _, releaseType := range allPackageTypes() {
		results = append(results, ResourceInstance{ID: string(releaseType), DisplayName: names[releaseType]})
	}

	return paginateInstances(results, req.Filter.Keyword, req.Page), nil
}

// FetchInstanceInfo returns existing package type names.
func (p *PackageTypeProvider) FetchInstanceInfo(_ contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
	names := packageTypeDisplayNames()
	data := FetchInstanceInfoData{}
	for _, id := range req.Filter.IDs {
		if name, ok := names[types.ReleaseType(id)]; ok {
			data = append(data, InstanceInfo{ID: id, DisplayName: name})
		}
	}

	return &data, nil
}
