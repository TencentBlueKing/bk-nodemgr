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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
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

// Dispatcher is the interface of dispatcher, for callback.
type Dispatcher interface {
	RegisterProvider(_type string, provider IProvider)
	GetProvider(_type string) (IProvider, bool)
	Dispatch(rCtx restserver.IContext) (interface{}, error)
}

// NewDispatcher creates a dispatcher.
func NewDispatcher() Dispatcher {
	return &dispatcher{
		providers: make(map[string]IProvider),
	}
}

type dispatcher struct {
	providers map[string]IProvider
}

// RegisterProvider registers a provider.
func (d *dispatcher) RegisterProvider(_type string, provider IProvider) {
	d.providers[_type] = provider
}

// GetProvider gets the provider by type.
func (d *dispatcher) GetProvider(_type string) (IProvider, bool) {
	provider, exist := d.providers[_type]
	return provider, exist
}

// Dispatch handles IAM resource callback requests.
// nolint:varnamelen
func (d *dispatcher) Dispatch(rCtx restserver.IContext) (interface{}, error) {
	// Bind request body to proto message
	req := new(protoBackend.IAMResourceCallbackReq)
	if err := rCtx.GContext().ShouldBindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to dispatch IAM callback, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// Validate request method
	if err := RequestMethod(req.GetMethod()).Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("method", req.GetMethod()).Error("failed to dispatch IAM callback, invalid method")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// Get the provider via resourceType
	provider, exist := d.GetProvider(req.GetType())
	if !exist {
		err := fmt.Errorf("resource type %s not supported or the provider not registered", req.GetType())
		logger.G.Biz(rCtx).WithErr(err).With("type", req.GetType()).Error("failed to dispatch IAM callback, provider not found")

		return nil, resterrf.ErrWrap(resterrf.RecordNotFound, err)
	}

	// Extract filter as map
	var filterMap map[string]interface{}
	if req.GetFilter() != nil {
		filterMap = req.GetFilter().AsMap()
	}

	// Convert and validate page parameters
	page, err := protoBackend.ConvIAMCallbackPageToTypes(req.GetPage())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to dispatch IAM callback, invalid page parameters")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// Get context
	ctx := contextx.FromContext(rCtx)

	// Dispatch the method
	var result any
	var providerErr error

	switch RequestMethod(req.GetMethod()) {
	case RequestMethodListAttr:
		result, providerErr = d.dispatchListAttr(ctx, provider, page)
	case RequestMethodListAttrValue:
		result, providerErr = d.dispatchListAttrValue(ctx, provider, filterMap, page)
	case RequestMethodListInstance:
		result, providerErr = d.dispatchListInstance(ctx, provider, filterMap, page)
	case RequestMethodFetchInstanceInfo:
		result, providerErr = d.dispatchFetchInstanceInfo(ctx, provider, filterMap, page)
	case RequestMethodListInstanceByPolicy:
		result, providerErr = d.dispatchListInstanceByPolicy(ctx, provider, filterMap, page)
	case RequestMethodSearchInstance:
		result, providerErr = d.dispatchSearchInstance(ctx, provider, filterMap, page)
	case RequestMethodFetchInstanceList:
		result, providerErr = d.dispatchFetchInstanceList(ctx, provider, filterMap, page)
	case RequestMethodFetchResourceTypeSchema:
		result, providerErr = d.dispatchFetchResourceTypeSchema(ctx, provider, page)
	default:
		providerErr = fmt.Errorf("method %s not supported", req.GetMethod())
		logger.G.Biz(rCtx).WithErr(providerErr).With("method", req.GetMethod()).Error("failed to dispatch IAM callback, unsupported method")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, providerErr)
	}

	if providerErr != nil {
		logger.G.Biz(rCtx).WithErr(providerErr).With("type", req.GetType(), "method", req.GetMethod()).
			Error("failed to dispatch IAM callback, provider execution failed")

		return nil, resterrf.ErrWrap(resterrf.Unknown, providerErr)
	}

	return result, nil
}

// dispatchListAttr dispatches the list_attr method.
func (d *dispatcher) dispatchListAttr(
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
func (d *dispatcher) dispatchListAttrValue(
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
func (d *dispatcher) dispatchListInstance(
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
func (d *dispatcher) dispatchFetchInstanceInfo(
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
func (d *dispatcher) dispatchListInstanceByPolicy(
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
func (d *dispatcher) dispatchSearchInstance(
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
		err := fmt.Errorf("keyword is required and cannot be empty")
		return nil, resterrf.ErrWrap(resterrf.InvalidKeyword, err)
	}

	req := &Request[SearchInstanceFilter]{
		Filter: filter,
		Page:   page,
	}

	return provider.SearchInstance(ctx, req)
}

// dispatchFetchInstanceList dispatches the fetch_instance_list method.
func (d *dispatcher) dispatchFetchInstanceList(
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
func (d *dispatcher) dispatchFetchResourceTypeSchema(
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
