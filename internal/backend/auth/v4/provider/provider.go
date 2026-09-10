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

// Package provider implements IAM V4 resource queries and authorization enrichment.
package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ErrInvalidArgument identifies invalid provider query parameters.
var ErrInvalidArgument = errors.New("invalid resource query")

// ErrNotFound identifies an unregistered resource type.
var ErrNotFound = errors.New("resource type not found")

// MaxFetchInstanceIDs bounds each callback fetch and enrichment batch.
const MaxFetchInstanceIDs = 1000

// MaxListInstancePageSize is the IAM V4 callback page size limit.
const MaxListInstancePageSize = 1000

// ParentFilter identifies a direct parent instance.
type ParentFilter struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// ListInstanceFilter selects candidates, not authorization grants.
type ListInstanceFilter struct {
	Parent  *ParentFilter `json:"parent,omitempty"`
	Keyword string        `json:"keyword,omitempty"`
}

// FetchInstanceFilter selects instances by their canonical IDs.
type FetchInstanceFilter struct {
	IDs []string `json:"ids"`
}

// Request carries a typed query within the V4 provider boundary.
type Request[F any] struct {
	Filter   F
	Page     types.Page
	Requires []string
}

// IProvider implements the two IAM V4 resource callback operations.
type IProvider interface {
	ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error)
	FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error)
}

// ListInstanceData contains the filtered total and the requested page.
type ListInstanceData struct {
	Count   int64              `json:"count"`
	Results []ResourceInstance `json:"results"`
}

// ResourceInstance is an IAM resource candidate.
type ResourceInstance struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// FetchInstanceInfoData contains the existing requested instances.
type FetchInstanceInfoData []InstanceInfo

// InstanceInfo contains provider-owned attributes selected by requires.
type InstanceInfo struct {
	ID          string                 `json:"id"`
	DisplayName string                 `json:"display_name,omitempty"`
	Attributes  map[string]interface{} `json:"-"`
}

// MarshalJSON flattens selected attributes and converts runtime paths to V4 strings.
func (info InstanceInfo) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{}, len(info.Attributes)+1)
	for key, value := range info.Attributes {
		result[key] = value
	}
	result[AttrID] = info.ID
	if info.DisplayName != "" {
		result["display_name"] = info.DisplayName
	}
	if paths, ok := result[AttrIAMPath].([]string); ok {
		if len(paths) != 1 {
			return nil, fmt.Errorf("expected one IAM V4 ancestor path")
		}
		result[AttrIAMPath] = paths[0]
	}

	return json.Marshal(result)
}

// UnmarshalJSON reads flattened instance attributes.
func (info *InstanceInfo) UnmarshalJSON(data []byte) error {
	attributes := make(map[string]interface{})
	if err := json.Unmarshal(data, &attributes); err != nil {
		return fmt.Errorf("failed to decode instance info: %w", err)
	}
	info.ID, _ = attributes[AttrID].(string)
	info.DisplayName, _ = attributes["display_name"].(string)
	delete(attributes, AttrID)
	delete(attributes, "display_name")
	info.Attributes = attributes

	return nil
}

func newEmptyListInstanceData() *ListInstanceData {
	return &ListInstanceData{Results: []ResourceInstance{}}
}

func paginateInstances(instances []ResourceInstance, keyword string, page types.Page) *ListInstanceData {
	results := make([]ResourceInstance, 0, len(instances))
	for _, instance := range instances {
		if strings.Contains(strings.ToLower(instance.DisplayName), strings.ToLower(keyword)) {
			results = append(results, instance)
		}
	}
	slices.SortFunc(results, func(left, right ResourceInstance) int { return strings.Compare(left.ID, right.ID) })
	data := &ListInstanceData{Count: int64(len(results)), Results: []ResourceInstance{}}
	if page.Offset >= len(results) {
		return data
	}
	end := page.Offset + min(page.Limit, len(results)-page.Offset)
	data.Results = results[page.Offset:end]

	return data
}

// BuildIAMPath preserves runtime authorization and apply-request path semantics.
func BuildIAMPath(resourceType string, resourceID string) []string {
	return []string{fmt.Sprintf("/%s,%s/", resourceType, resourceID)}
}

// ParseParentFromIAMPath extracts the direct ancestor used in permission applications.
func ParseParentFromIAMPath(attributes map[string]interface{}) *types.IAMApplyResourceNode {
	var path string
	switch value := attributes[AttrIAMPath].(type) {
	case []string:
		if len(value) > 0 {
			path = value[0]
		}
	case []interface{}:
		if len(value) > 0 {
			path, _ = value[0].(string)
		}
	case string:
		path = value
	}
	if len(path) < 3 || path[0] != '/' || path[len(path)-1] != '/' {
		return nil
	}
	parts := strings.Split(path[1:len(path)-1], ",")
	const expectedParts = 2
	if len(parts) != expectedParts {
		return nil
	}

	return &types.IAMApplyResourceNode{Type: parts[0], ID: parts[1]}
}

func buildInstanceAttributes(instanceID string, parentType string, parentID string) map[string]interface{} {
	return map[string]interface{}{AttrID: instanceID, AttrIAMPath: BuildIAMPath(parentType, parentID)}
}
