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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// RequestMethod represents an IAM V4 callback method.
type RequestMethod string

// IAM V4 callback methods.
const (
	RequestMethodListInstance      RequestMethod = "list_instance"
	RequestMethodFetchInstanceInfo RequestMethod = "fetch_instance_info"
)

// Handler validates typed queries and routes them to registered V4 providers.
type Handler struct{ providers map[string]IProvider }

// NewHandler creates a V4 provider registry.
func NewHandler() *Handler { return &Handler{providers: make(map[string]IProvider)} }

var _ IHandler = (*Handler)(nil)

// RegisterProvider registers a resource type during service initialization.
func (h *Handler) RegisterProvider(resourceType string, provider IProvider) {
	h.providers[resourceType] = provider
}

// GetProvider looks up a registered resource type.
func (h *Handler) GetProvider(resourceType string) (IProvider, bool) {
	provider, ok := h.providers[resourceType]
	return provider, ok
}

// DispatchMethod converts callback filters and delegates to typed resource queries.
func (h *Handler) DispatchMethod(
	ctx contextx.IContext, resourceType string, method RequestMethod,
	filterMap map[string]interface{}, page types.Page, requires []string,
) (interface{}, error) {

	switch method {
	case RequestMethodListInstance:
		var filter ListInstanceFilter
		if err := conv.MapToStruct(filterMap, &filter); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
		}

		return h.ListInstance(ctx, resourceType, &Request[ListInstanceFilter]{Filter: filter, Page: page})
	case RequestMethodFetchInstanceInfo:
		var filter FetchInstanceFilter
		if err := conv.MapToStruct(filterMap, &filter); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
		}

		return h.FetchInstanceInfo(ctx, resourceType, &Request[FetchInstanceFilter]{Filter: filter, Requires: requires})
	default:
		return nil, ErrNotFound
	}
}

// ListInstance validates pagination and lists candidates in the request tenant.
func (h *Handler) ListInstance(ctx contextx.IContext, resourceType string, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	provider, ok := h.GetProvider(resourceType)
	if !ok {
		return nil, ErrNotFound
	}
	if req == nil {
		return nil, ErrInvalidArgument
	}
	if err := req.Page.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}
	if req.Page.Limit > MaxListInstancePageSize {
		return nil, ErrInvalidArgument
	}

	return provider.ListInstance(ctx, req)
}

// FetchInstanceInfo validates IDs and selects supported attributes using requires.
func (h *Handler) FetchInstanceInfo(ctx contextx.IContext, resourceType string, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
	provider, ok := h.GetProvider(resourceType)
	if !ok {
		return nil, ErrNotFound
	}
	if req == nil || req.Filter.IDs == nil || len(req.Filter.IDs) > MaxFetchInstanceIDs {
		return nil, ErrInvalidArgument
	}
	for _, id := range req.Filter.IDs {
		if id == "" {
			return nil, ErrInvalidArgument
		}
	}
	if len(req.Filter.IDs) == 0 {
		data := FetchInstanceInfoData{}
		return &data, nil
	}
	data, err := provider.FetchInstanceInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch instance info: %w", err)
	}
	if data == nil {
		return nil, fmt.Errorf("provider returned nil instance info")
	}
	for index := range *data {
		info := &(*data)[index]
		attributes := make(map[string]interface{})
		for key, value := range info.Attributes {
			if len(req.Requires) == 0 || slices.Contains(req.Requires, key) {
				attributes[key] = value
			}
		}
		if len(req.Requires) == 0 || slices.Contains(req.Requires, "display_name") {
			attributes["display_name"] = info.DisplayName
		} else {
			info.DisplayName = ""
		}
		info.Attributes = attributes
	}

	return data, nil
}

// FetchResourceAttributes preserves runtime path arrays and supports more than 1000 IDs.
func (h *Handler) FetchResourceAttributes(
	ctx contextx.IContext, resourceType string, resourceIDs []string,
) (map[string]map[string]interface{}, error) {

	result := make(map[string]map[string]interface{})
	if _, ok := h.GetProvider(resourceType); !ok {
		return result, nil
	}
	for batch := range slices.Chunk(resourceIDs, MaxFetchInstanceIDs) {
		data, err := h.FetchInstanceInfo(ctx, resourceType, &Request[FetchInstanceFilter]{Filter: FetchInstanceFilter{IDs: batch}})
		if err != nil {
			return nil, fmt.Errorf("failed to enrich %s: %w", resourceType, err)
		}
		for _, info := range *data {
			delete(info.Attributes, "display_name")
			if len(info.Attributes) > 0 {
				result[info.ID] = info.Attributes
			}
		}
	}

	return result, nil
}
