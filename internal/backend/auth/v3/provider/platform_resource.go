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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ResourceTypePlatformResource is the IAM resource type for platform capabilities.
const ResourceTypePlatformResource = "platform_resource"

// PlatformResourceProvider exposes the fixed platform capabilities.
type PlatformResourceProvider struct{}

// NewPlatformResourceProvider creates a platform resource provider.
func NewPlatformResourceProvider() *PlatformResourceProvider { return &PlatformResourceProvider{} }

// ListAttr returns no custom attributes for platform capabilities.
func (p *PlatformResourceProvider) ListAttr(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListAttrData, error) {
	data := ListAttrData([]ResourceAttribute{})
	return &data, nil
}

// ListAttrValue returns no custom attribute values for platform capabilities.
func (p *PlatformResourceProvider) ListAttrValue(_ contextx.IContext, _ *Request[ListAttrValueFilter]) (*ListAttrValueData, error) {
	return &ListAttrValueData{Count: 0, Results: []AttributeValue{}}, nil
}

// ListInstance lists the fixed platform capabilities.
func (p *PlatformResourceProvider) ListInstance(_ contextx.IContext, _ *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	resources := allPlatformResources()
	names := platformResourceDisplayNames()
	results := make([]ResourceInstance, 0, len(resources))
	for _, resource := range resources {
		results = append(results, ResourceInstance{ID: string(resource), DisplayName: names[resource]})
	}

	return &ListInstanceData{Count: int64(len(results)), Results: results}, nil
}

// FetchInstanceInfo returns existing platform capability names.
func (p *PlatformResourceProvider) FetchInstanceInfo(_ contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
	names := platformResourceDisplayNames()
	data := FetchInstanceInfoData{}
	for _, id := range req.Filter.IDs {
		if name, ok := names[types.PlatformResource(id)]; ok {
			data = append(data, InstanceInfo{ID: id, DisplayName: name, Attributes: map[string]interface{}{}})
		}
	}

	return &data, nil
}

// ListInstanceByPolicy filters platform capabilities by the IAM policy expression.
func (p *PlatformResourceProvider) ListInstanceByPolicy(_ contextx.IContext, req *Request[ListInstanceByPolicyFilter]) (*ListInstanceData, error) {
	resources := allPlatformResources()
	names := platformResourceDisplayNames()
	instances := make([]InstanceForEval, 0, len(resources))
	for _, resource := range resources {
		instances = append(instances, InstanceForEval{
			Instance:   ResourceInstance{ID: string(resource), DisplayName: names[resource]},
			Attributes: map[string]interface{}{},
		})
	}

	return evalExpressionFilter(req.Filter.Expression, ResourceTypePlatformResource, instances, req.Page)
}

// SearchInstance searches platform capabilities by ID or display name.
func (p *PlatformResourceProvider) SearchInstance(_ contextx.IContext, req *Request[SearchInstanceFilter]) (*ListInstanceData, error) {
	keyword := strings.ToLower(strings.TrimSpace(req.Filter.Keyword))
	names := platformResourceDisplayNames()
	results := make([]ResourceInstance, 0)
	for _, resource := range allPlatformResources() {
		id := string(resource)
		name := names[resource]
		if !strings.Contains(strings.ToLower(id), keyword) && !strings.Contains(strings.ToLower(name), keyword) {
			continue
		}
		results = append(results, ResourceInstance{ID: id, DisplayName: name})
	}

	return &ListInstanceData{Count: int64(len(results)), Results: results}, nil
}

// FetchInstanceList returns no audit-center instances.
func (p *PlatformResourceProvider) FetchInstanceList(_ contextx.IContext, _ *Request[FetchInstanceListFilter]) (*ListInstanceData, error) {
	return newEmptyListInstanceData(), nil
}

// FetchResourceTypeSchema returns no custom schema for platform capabilities.
func (p *PlatformResourceProvider) FetchResourceTypeSchema(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListInstanceData, error) {
	return newEmptyListInstanceData(), nil
}
