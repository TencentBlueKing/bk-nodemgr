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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
)

// RequestMethod represents the method of the request.
type RequestMethod string

// IAM callback methods.
const (
	RequestMethodListAttr                RequestMethod = "list_attr"
	RequestMethodListAttrValue           RequestMethod = "list_attr_value"
	RequestMethodListInstance            RequestMethod = "list_instance"
	RequestMethodFetchInstanceInfo       RequestMethod = "fetch_instance_info"
	RequestMethodListInstanceByPolicy    RequestMethod = "list_instance_by_policy"
	RequestMethodSearchInstance          RequestMethod = "search_instance"
	RequestMethodFetchInstanceList       RequestMethod = "fetch_instance_list"
	RequestMethodFetchResourceTypeSchema RequestMethod = "fetch_resource_type_schema"
)

// Validate validates the request method.
// nolint:varnamelen
func (m RequestMethod) Validate() error {
	switch m {
	case RequestMethodListAttr, RequestMethodListAttrValue, RequestMethodListInstance, RequestMethodFetchInstanceInfo,
		RequestMethodListInstanceByPolicy, RequestMethodSearchInstance, RequestMethodFetchInstanceList,
		RequestMethodFetchResourceTypeSchema:
		return nil
	default:
		return fmt.Errorf("invalid request method: %s", m)
	}
}

// Handler is the unified implementation that provides both IDispatcher and IResolver capabilities.
// It maintains a single provider registry and exposes different interfaces for different use cases:
//   - IDispatcher: for HTTP callback routing (used by router layer)
//   - IResolver: for policy-based instance listing (used by auth layer)
type Handler struct {
	providers map[string]IProvider
}

// NewHandler creates a unified handler that implements both IDispatcher and IResolver interfaces.
func NewHandler() *Handler {
	return &Handler{
		providers: make(map[string]IProvider),
	}
}

// Ensure Handler implements both interfaces at compile time.
var (
	_ IDispatcher = (*Handler)(nil)
	_ IResolver   = (*Handler)(nil)
)

// RegisterProvider registers a provider.
func (h *Handler) RegisterProvider(_type string, provider IProvider) {
	h.providers[_type] = provider
}

// GetProvider gets the provider by type.
func (h *Handler) GetProvider(_type string) (IProvider, bool) {
	provider, exist := h.providers[_type]
	return provider, exist
}

// dispatchToProvider routes the callback method to the appropriate provider method.
func (h *Handler) dispatchToProvider(
	ctx contextx.IContext,
	method RequestMethod,
	provider IProvider,
	filterMap map[string]interface{},
	page types.Page,
) (interface{}, error) {

	switch method {
	case RequestMethodListAttr:
		return h.dispatchListAttr(ctx, provider, page)
	case RequestMethodListAttrValue:
		return h.dispatchListAttrValue(ctx, provider, filterMap, page)
	case RequestMethodListInstance:
		return h.dispatchListInstance(ctx, provider, filterMap, page)
	case RequestMethodFetchInstanceInfo:
		return h.dispatchFetchInstanceInfo(ctx, provider, filterMap, page)
	case RequestMethodListInstanceByPolicy:
		return h.dispatchListInstanceByPolicy(ctx, provider, filterMap, page)
	case RequestMethodSearchInstance:
		return h.dispatchSearchInstance(ctx, provider, filterMap, page)
	case RequestMethodFetchInstanceList:
		return h.dispatchFetchInstanceList(ctx, provider, filterMap, page)
	case RequestMethodFetchResourceTypeSchema:
		return h.dispatchFetchResourceTypeSchema(ctx, provider, page)
	default:
		return nil, fmt.Errorf("method %s not supported", method)
	}
}

// DispatchMethod handles IAM callback method dispatch with pre-parsed parameters.
func (h *Handler) DispatchMethod(
	ctx contextx.IContext,
	resourceType string,
	method RequestMethod,
	filterMap map[string]interface{},
	page types.Page,
) (interface{}, error) {
	// Get the provider via resourceType
	provider, exist := h.GetProvider(resourceType)
	if !exist {
		return nil, fmt.Errorf("resource type %s not supported or the provider not registered", resourceType)
	}

	// Dispatch to provider
	return h.dispatchToProvider(ctx, method, provider, filterMap, page)
}

// ListInstancesByPolicy implements IResolver interface by delegating to the provider.
func (h *Handler) ListInstancesByPolicy(
	ctx contextx.IContext,
	resourceType string,
	filterMap map[string]interface{},
	page types.Page,
) (*ListInstanceData, error) {

	provider, exist := h.GetProvider(resourceType)
	if !exist {
		return nil, fmt.Errorf("resource type %s not supported or the provider not registered", resourceType)
	}

	var filter ListInstanceByPolicyFilter
	if err := conv.MapToStruct(filterMap, &filter); err != nil {
		return nil, fmt.Errorf("failed to parse ListInstanceByPolicyFilter: %w", err)
	}

	req := &Request[ListInstanceByPolicyFilter]{
		Filter: filter,
		Page:   page,
	}

	return provider.ListInstanceByPolicy(ctx, req)
}

// ListInstancesByExpression implements IResolver interface.
// It evaluates IAM policy expression (ExprCell format) against provider instances.
// This method is used by auth layer fallback when discrete policy parsing fails.
func (h *Handler) ListInstancesByExpression(
	ctx contextx.IContext,
	resourceType string,
	expr *expression.ExprCell,
	page types.Page,
) (*ListInstanceData, error) {

	provider, exist := h.GetProvider(resourceType)
	if !exist {
		return nil, fmt.Errorf("resource type %s not supported or the provider not registered", resourceType)
	}

	// Convert ExprCell to map for provider compatibility
	// Note: This involves a serialization round-trip (ExprCell -> map -> ExprCell in evalExpressionFilter)
	// but it's the minimal-change approach that reuses existing provider logic.
	// The overhead is acceptable compared to database queries and expression evaluation.
	expressionMap := make(map[string]interface{})
	if expr != nil {
		// Serialize ExprCell to JSON then deserialize to map
		exprBytes, err := json.Marshal(expr)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal expression: %w", err)
		}
		if err := json.Unmarshal(exprBytes, &expressionMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal expression to map: %w", err)
		}
	}

	// Call provider's ListInstanceByPolicy with the expression map
	req := &Request[ListInstanceByPolicyFilter]{
		Filter: ListInstanceByPolicyFilter{
			Expression: expressionMap,
		},
		Page: page,
	}

	return provider.ListInstanceByPolicy(ctx, req)
}

// dispatchListAttr dispatches the list_attr method.
func (h *Handler) dispatchListAttr(
	ctx contextx.IContext,
	provider IProvider,
	page types.Page,
) (*ListAttrData, error) {

	req := &Request[EmptyFilter]{
		Filter: EmptyFilter{},
		Page:   page,
	}

	return provider.ListAttr(ctx, req)
}

// dispatchListAttrValue dispatches the list_attr_value method.
func (h *Handler) dispatchListAttrValue(
	ctx contextx.IContext,
	provider IProvider,
	filterMap map[string]interface{},
	page types.Page,
) (*ListAttrValueData, error) {

	var filter ListAttrValueFilter
	if err := conv.MapToStruct(filterMap, &filter); err != nil {
		return nil, fmt.Errorf("failed to parse ListAttrValueFilter: %w", err)
	}

	req := &Request[ListAttrValueFilter]{
		Filter: filter,
		Page:   page,
	}

	return provider.ListAttrValue(ctx, req)
}

// dispatchListInstance dispatches the list_instance method.
func (h *Handler) dispatchListInstance(
	ctx contextx.IContext,
	provider IProvider,
	filterMap map[string]interface{},
	page types.Page,
) (*ListInstanceData, error) {

	var filter ListInstanceFilter
	if err := conv.MapToStruct(filterMap, &filter); err != nil {
		return nil, fmt.Errorf("failed to parse ListInstanceFilter: %w", err)
	}

	req := &Request[ListInstanceFilter]{
		Filter: filter,
		Page:   page,
	}

	return provider.ListInstance(ctx, req)
}

// dispatchFetchInstanceInfo dispatches the fetch_instance_info method.
func (h *Handler) dispatchFetchInstanceInfo(
	ctx contextx.IContext,
	provider IProvider,
	filterMap map[string]interface{},
	page types.Page,
) (*FetchInstanceInfoData, error) {

	var filter FetchInstanceFilter
	if err := conv.MapToStruct(filterMap, &filter); err != nil {
		return nil, fmt.Errorf("failed to parse FetchInstanceFilter: %w", err)
	}

	// Validate IDs count limit according to IAM specification
	if len(filter.IDs) > MaxFetchInstanceIDs {
		return nil, fmt.Errorf("IDs count exceeds maximum limit of %d", MaxFetchInstanceIDs)
	}

	req := &Request[FetchInstanceFilter]{
		Filter: filter,
		Page:   page,
	}

	return provider.FetchInstanceInfo(ctx, req)
}

// dispatchListInstanceByPolicy dispatches the list_instance_by_policy method.
func (h *Handler) dispatchListInstanceByPolicy(
	ctx contextx.IContext,
	provider IProvider,
	filterMap map[string]interface{},
	page types.Page,
) (*ListInstanceData, error) {

	var filter ListInstanceByPolicyFilter
	if err := conv.MapToStruct(filterMap, &filter); err != nil {
		return nil, fmt.Errorf("failed to parse ListInstanceByPolicyFilter: %w", err)
	}

	req := &Request[ListInstanceByPolicyFilter]{
		Filter: filter,
		Page:   page,
	}

	return provider.ListInstanceByPolicy(ctx, req)
}

// dispatchSearchInstance dispatches the search_instance method.
func (h *Handler) dispatchSearchInstance(
	ctx contextx.IContext,
	provider IProvider,
	filterMap map[string]interface{},
	page types.Page,
) (*ListInstanceData, error) {

	var filter SearchInstanceFilter
	if err := conv.MapToStruct(filterMap, &filter); err != nil {
		return nil, fmt.Errorf("failed to parse SearchInstanceFilter: %w", err)
	}

	// Validate keyword is not empty
	if strings.TrimSpace(filter.Keyword) == "" {
		return nil, fmt.Errorf("keyword is required and cannot be empty")
	}

	req := &Request[SearchInstanceFilter]{
		Filter: filter,
		Page:   page,
	}

	return provider.SearchInstance(ctx, req)
}

// dispatchFetchInstanceList dispatches the fetch_instance_list method.
func (h *Handler) dispatchFetchInstanceList(
	ctx contextx.IContext,
	provider IProvider,
	filterMap map[string]interface{},
	page types.Page,
) (*ListInstanceData, error) {

	var filter FetchInstanceListFilter
	if err := conv.MapToStruct(filterMap, &filter); err != nil {
		return nil, fmt.Errorf("failed to parse FetchInstanceListFilter: %w", err)
	}

	req := &Request[FetchInstanceListFilter]{
		Filter: filter,
		Page:   page,
	}

	return provider.FetchInstanceList(ctx, req)
}

// dispatchFetchResourceTypeSchema dispatches the fetch_resource_type_schema method.
func (h *Handler) dispatchFetchResourceTypeSchema(
	ctx contextx.IContext,
	provider IProvider,
	page types.Page,
) (*ListInstanceData, error) {

	req := &Request[EmptyFilter]{
		Filter: EmptyFilter{},
		Page:   page,
	}

	return provider.FetchResourceTypeSchema(ctx, req)
}

// FetchResourceAttributes implements IAttributeEnricher interface.
// It fetches resource attributes by calling the provider's FetchInstanceInfo method.
func (h *Handler) FetchResourceAttributes(
	ctx contextx.IContext,
	resourceType string,
	resourceIDs []string,
) (map[string]map[string]interface{}, error) {

	provider, exists := h.GetProvider(resourceType)
	if !exists {
		// If provider not found, return empty map (no attributes to enrich)
		return make(map[string]map[string]interface{}), nil
	}

	// Call FetchInstanceInfo to get resource details
	req := &Request[FetchInstanceFilter]{
		Filter: FetchInstanceFilter{
			IDs: resourceIDs,
		},
	}

	data, err := provider.FetchInstanceInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch instance info for %s: %w", resourceType, err)
	}

	// Build result map: resourceID -> attributes
	result := make(map[string]map[string]interface{}, len(*data))
	for _, instance := range *data {
		if len(instance.Attributes) > 0 {
			result[instance.ID] = instance.Attributes
		}
	}

	return result, nil
}
