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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
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

// Request contains the essential fields from IAM callback request.
// Page field has been converted and validated from proto message to types.Page.
type Request struct {
	Type   string
	Method string
	Filter map[string]interface{}
	Page   types.Page
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

	// Build request object
	callbackReq := &Request{
		Type:   req.GetType(),
		Method: req.GetMethod(),
		Filter: filterMap,
		Page:   page,
	}

	// Get context
	ctx := contextx.FromContext(rCtx)

	// Dispatch the method
	var result any
	var providerErr error

	switch RequestMethod(req.GetMethod()) {
	case RequestMethodListAttr:
		result, providerErr = provider.ListAttr(ctx, callbackReq)
	case RequestMethodListAttrValue:
		result, providerErr = provider.ListAttrValue(ctx, callbackReq)
	case RequestMethodListInstance:
		result, providerErr = provider.ListInstance(ctx, callbackReq)
	case RequestMethodFetchInstanceInfo:
		result, providerErr = provider.FetchInstanceInfo(ctx, callbackReq)
	case RequestMethodListInstanceByPolicy:
		result, providerErr = provider.ListInstanceByPolicy(ctx, callbackReq)
	case RequestMethodSearchInstance:
		result, providerErr = provider.SearchInstance(ctx, callbackReq)
	case RequestMethodFetchInstanceList:
		result, providerErr = provider.FetchInstanceList(ctx, callbackReq)
	case RequestMethodFetchResourceTypeSchema:
		result, providerErr = provider.FetchResourceTypeSchema(ctx, callbackReq)
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
