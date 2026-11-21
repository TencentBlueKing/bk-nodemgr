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

// ResourceTypePlatformResource is the IAM resource type for platform capabilities.
const ResourceTypePlatformResource = "platform_resource"

// PlatformResourceProvider exposes the fixed platform capabilities.
type PlatformResourceProvider struct{}

// NewPlatformResourceProvider creates a platform resource provider.
func NewPlatformResourceProvider() *PlatformResourceProvider { return &PlatformResourceProvider{} }

// ListInstance filters and paginates platform capabilities.
func (p *PlatformResourceProvider) ListInstance(_ contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	if req.Filter.Parent != nil {
		return nil, ErrInvalidArgument
	}
	names := platformResourceDisplayNames()
	results := make([]ResourceInstance, 0, len(names))
	for _, resource := range allPlatformResources() {
		results = append(results, ResourceInstance{ID: string(resource), DisplayName: names[resource]})
	}

	return paginateInstances(results, req.Filter.Keyword, req.Page), nil
}

// FetchInstanceInfo returns existing platform capability names.
func (p *PlatformResourceProvider) FetchInstanceInfo(_ contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
	names := platformResourceDisplayNames()
	data := FetchInstanceInfoData{}
	for _, id := range req.Filter.IDs {
		if name, ok := names[types.PlatformResource(id)]; ok {
			data = append(data, InstanceInfo{ID: id, DisplayName: name})
		}
	}

	return &data, nil
}
