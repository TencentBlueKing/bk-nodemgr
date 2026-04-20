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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AttributeValueID represents a resource attribute value identifier.
// According to IAM list_attr_value API, attribute value IDs support string/int/bool types.
// This type provides automatic JSON type conversion.
type AttributeValueID string

var _ json.Unmarshaler = (*AttributeValueID)(nil)

// UnmarshalJSON implements custom JSON deserialization to support string/int/bool types.
func (id *AttributeValueID) UnmarshalJSON(data []byte) error {
	// Try to parse as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*id = AttributeValueID(s)
		return nil
	}

	// Try to parse as int64
	var i int64
	if err := json.Unmarshal(data, &i); err == nil {
		*id = AttributeValueID(strconv.FormatInt(i, 10))
		return nil
	}

	// Try to parse as bool
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*id = AttributeValueID(strconv.FormatBool(b))
		return nil
	}

	return fmt.Errorf("invalid attribute value ID type: %s", string(data))
}

// EmptyFilter is used for IAM callback methods that do not require filter parameters.
type EmptyFilter struct{}

// ParentFilter represents parent resource information in IAM callback requests.
type ParentFilter struct {
	Type string `json:"type"` // Parent resource type
	ID   string `json:"id"`   // Parent resource instance ID
}

// ListAttrValueFilter is used for list_attr_value API to filter attribute values.
type ListAttrValueFilter struct {
	Attr    string             `json:"attr"`              // Required: attribute ID
	Keyword string             `json:"keyword,omitempty"` // Optional: search keyword
	IDs     []AttributeValueID `json:"ids,omitempty"`     // Optional: attribute value ID list (supports string/int/bool)
}

// ListInstanceFilter is used for list_instance API to filter instances by parent.
type ListInstanceFilter struct {
	Parent *ParentFilter `json:"parent,omitempty"` // Optional: direct parent resource
}

// MaxFetchInstanceIDs is the maximum number of instance IDs allowed in a single fetch_instance_info request.
// The ids field supports a maximum of 1000 items.
const MaxFetchInstanceIDs = 1000

// MaxListInstanceByPolicyLimit is the maximum page limit for loading all instances
// during ListInstanceByPolicy expression evaluation and FetchInstanceInfo batch queries.
const MaxListInstanceByPolicyLimit = 10000

// FetchInstanceFilter is used for fetch_instance_info API to fetch instance details.
// Note: According to IAM spec, filter.ids only supports string type (array(string)).
type FetchInstanceFilter struct {
	IDs   []string `json:"ids"`             // Required: resource instance ID list (max 1000, only string type)
	Attrs []string `json:"attrs,omitempty"` // Optional: attributes to query, empty means all attributes
}

// ListInstanceByPolicyFilter is used for list_instance_by_policy API (policy expression based).
type ListInstanceByPolicyFilter struct {
	Expression map[string]interface{} `json:"expression"` // Required: policy expression (dynamic structure)
}

// SearchInstanceFilter is used for search_instance API to search instances by keyword.
type SearchInstanceFilter struct {
	Keyword string        `json:"keyword"`          // Required: search keyword
	Parent  *ParentFilter `json:"parent,omitempty"` // Optional: parent filter
}

// FetchInstanceListFilter is used for fetch_instance_list API (audit center).
type FetchInstanceListFilter struct {
	StartTime int64 `json:"start_time"` // Required: start time (milliseconds)
	EndTime   int64 `json:"end_time"`   // Required: end time (milliseconds)
}

// Request is a generic IAM callback request with type-safe filter.
type Request[F any] struct {
	Type   string
	Method string
	Filter F
	Page   types.Page
}

// IProvider is the interface for IAM resource provider.
// Each method corresponds to a specific IAM callback API as defined in:
// https://github.com/TencentBlueKing/BKDocs/tree/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback
// nolint: lll
type IProvider interface {
	// ListAttr lists resource attributes that can be used for permission configuration.
	// Doc: https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/10-list_attr.md
	// Performance requirement: < 50ms
	ListAttr(ctx contextx.IContext, req *Request[EmptyFilter]) (*ListAttrData, error)

	// ListAttrValue lists values for a specific resource attribute, supports keyword search and batch ID filtering.
	// Doc: https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/11-list_attr_value.md
	// Performance requirement:
	//   - No filter: < 50ms
	//   - Keyword search: < 100ms
	//   - Batch ID filter (≤10): < 100ms
	//   - Batch ID filter (>10): < 200ms
	ListAttrValue(ctx contextx.IContext, req *Request[ListAttrValueFilter]) (*ListAttrValueData, error)

	// ListInstance lists resource instances with optional parent filtering and pagination.
	// Doc: https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/12-list_instance.md
	// Performance requirement: < 50ms (with parent filter)
	ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error)

	// FetchInstanceInfo fetches detailed information (attributes) of specific instances by IDs.
	// This is a performance-critical API used for authorization.
	// Doc: https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/13-fetch_instance_info.md
	// Performance requirement:
	//   - Single instance: < 20ms
	//   - Batch instances: < 100ms
	FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error)

	// ListInstanceByPolicy lists instances matching the policy expression.
	// Note: Currently not used by IAM, can be left unimplemented or return empty result.
	// Performance requirement: < 500ms
	ListInstanceByPolicy(ctx contextx.IContext, req *Request[ListInstanceByPolicyFilter]) (*ListInstanceData, error)

	// SearchInstance searches instances by keyword with optional parent filtering.
	// Doc: https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/15-search_instance.md
	// Performance requirement: < 100ms
	// IMPORTANT: Search should be case-insensitive and support display_name search at minimum.
	// Should return code=422 if scan size is too large, code=406 if keyword is invalid.
	SearchInstance(ctx contextx.IContext, req *Request[SearchInstanceFilter]) (*ListInstanceData, error)

	// FetchInstanceList is deprecated or custom method, consider removing if not used.
	FetchInstanceList(ctx contextx.IContext, req *Request[FetchInstanceListFilter]) (*ListInstanceData, error)

	// FetchResourceTypeSchema is a custom method, not part of standard IAM callback APIs.
	FetchResourceTypeSchema(ctx contextx.IContext, req *Request[EmptyFilter]) (*ListInstanceData, error)
}

// ListAttrData represents the response data for ListAttr API.
// Returns a list of resource attributes that can be used for permission configuration.
// According to IAM official documentation, the data field should be an array directly.
type ListAttrData []ResourceAttribute

// ResourceAttribute represents a resource attribute definition.
type ResourceAttribute struct {
	// ID is the unique identifier of the attribute (e.g., "os", "country")
	ID string `json:"id"`
	// DisplayName is the human-readable name of the attribute (e.g., "操作系统", "国家")
	DisplayName string `json:"display_name"`
}

// ListAttrValueData represents the response data for ListAttrValue API.
// Returns paginated attribute values with total count.
type ListAttrValueData struct {
	// Count is the total number of attribute values matching the filter
	Count int64 `json:"count"`
	// Results contains the list of attribute values
	Results []AttributeValue `json:"results"`
}

// AttributeValue represents a value of a resource attribute.
type AttributeValue struct {
	// ID is the unique identifier of the attribute value
	// Can be string, int, or bool depending on the attribute type
	ID interface{} `json:"id"`
	// DisplayName is the human-readable name of the attribute value
	DisplayName string `json:"display_name"`
}

// ListInstanceData represents the response data for ListInstance, SearchInstance, and ListInstanceByPolicy APIs.
// Returns paginated resource instances with total count.
type ListInstanceData struct {
	// Count is the total number of instances matching the filter
	Count int64 `json:"count"`
	// Results contains the list of resource instances
	Results []ResourceInstance `json:"results"`
}

// ResourceInstance represents a resource instance in the system.
type ResourceInstance struct {
	// ID is the unique identifier of the resource instance
	ID string `json:"id"`
	// DisplayName is the human-readable name of the resource instance
	DisplayName string `json:"display_name"`
	// ChildType indicates the next level resource type (optional, only for dynamic hierarchical resources)
	// Empty value means no next level. This is a special field for "user management" system.
	ChildType string `json:"child_type,omitempty"`
}

// FetchInstanceInfoData represents the response data for FetchInstanceInfo API.
// Returns detailed attribute information for each requested instance.
type FetchInstanceInfoData []InstanceInfo

// InstanceInfo represents detailed information of a resource instance.
// Contains the instance ID and all its attributes (dynamic fields based on attrs filter in request).
type InstanceInfo struct {
	// ID is the unique identifier of the resource instance
	ID string `json:"id"`
	// DisplayName is the human-readable name of the resource instance (required)
	DisplayName string `json:"display_name,omitempty"`
	// Attributes stores all other attributes dynamically
	// Special attributes:
	//   - "_bk_iam_path_": []string - resource topology paths
	//   - "_bk_iam_approver_": []string - resource approvers
	//   - Other custom attributes based on resource type
	Attributes map[string]interface{} `json:"-"`
}

// MarshalJSON implements custom JSON marshaling to flatten the Attributes map.
func (info *InstanceInfo) MarshalJSON() ([]byte, error) {
	// Create a map combining ID and all attributes
	result := make(map[string]interface{})
	result["id"] = info.ID

	// Add DisplayName if present
	if info.DisplayName != "" {
		result["display_name"] = info.DisplayName
	}

	// Flatten all attributes into the result
	for k, v := range info.Attributes {
		result[k] = v
	}

	// Marshal the flattened map
	return json.Marshal(result)
}

// UnmarshalJSON implements custom JSON unmarshaling to populate the Attributes map.
func (info *InstanceInfo) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary map
	temp := make(map[string]interface{})
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Extract ID
	if id, ok := temp["id"].(string); ok {
		info.ID = id
		delete(temp, "id")
	}

	// Extract DisplayName if present
	if displayName, ok := temp["display_name"].(string); ok {
		info.DisplayName = displayName
		delete(temp, "display_name")
	}

	// Store remaining fields in Attributes
	info.Attributes = temp

	return nil
}

// BuildIAMPath constructs IAM resource path attribute value.
//
// The path format follows IAM specification: "/parent_type,parent_id/"
// This attribute is used by IAM for hierarchical authorization checks.
//
// Parameters:
//   - resourceType: Parent resource type (e.g., "networkarea", "package_type")
//   - resourceID: Parent resource ID(s), will be joined with comma if multiple
//
// Returns:
//   - []string: Path array (IAM supports multiple paths, we return single-element array)
//
// Example:
//
//	BuildIAMPath("networkarea", "123") → ["/networkarea,123/"]
//	BuildIAMPath("package_type", "plugin") → ["/package_type,plugin/"]
//
// Integration with iam-go-sdk:
//
//	This value is set as the _bk_iam_path_ attribute in resource instance info.
//	During policy evaluation, expression.ExprCell.Eval() uses this attribute when
//	matching IAM policy expressions with starts_with operator.
//	The SDK automatically strips ",*/" suffix for wildcard path matching.
//
// See also:
//   - expression.KeywordBKIAMPath: The attribute key constant
//   - expression.ExprCell.Eval(): Policy expression evaluation logic
func BuildIAMPath(resourceType string, resourceID ...string) []string {
	return []string{
		fmt.Sprintf("/%s,%s/", resourceType, strings.Join(resourceID, ",")),
	}
}
