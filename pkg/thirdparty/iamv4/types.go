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

// Package iamv4 provides the IAM v4 client implementation.
package iamv4

import (
	"fmt"

	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
)

// RespCommon describes the common part of IAM v4 response.
type RespCommon struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// BaseBroker describes the base broker for IAM v4 API response.
type BaseBroker[T any] struct {
	RespCommon
	Data  T          `json:"data"`
	Error *RespError `json:"error,omitempty"`
}

// IsFailed checks if the response indicates an error.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Code != 0 {
		return fmt.Errorf("code(%d), msg(%s)", resp.Code, resp.Message)
	}

	if resp.Error != nil {
		return fmt.Errorf("code(%s), msg(%s)", resp.Error.Code, resp.Error.Message)
	}

	return nil
}

// RespError describes IAM v4 error envelope details.
type RespError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Config is the config of IAM v4.
type Config struct {
	VirtualUserConfig apigwclient.VirtualUserConfig
	// SystemID is the system identifier registered in IAM.
	SystemID string
	// CallbackPath is the callback path for IAM resource provider.
	CallbackPath string
}

// Validate validates the config.
func (conf *Config) Validate() error {
	if err := conf.VirtualUserConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate IAM v4 client config: %w", err)
	}

	if conf.SystemID == "" {
		return fmt.Errorf("system ID is required for IAM v4")
	}

	if conf.CallbackPath == "" {
		return fmt.Errorf("callback path is required for IAM v4")
	}

	return nil
}

// Subject is the object of permission.
type Subject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Resource is the resource instance used by IAM v4 authorization APIs.
type Resource struct {
	ID string `json:"id"`
}

// DirectAuthRequest is the request for single action authorization.
type DirectAuthRequest struct {
	Subject  Subject   `json:"subject"`
	ActionID string    `json:"action_id"`
	Resource *Resource `json:"resource,omitempty"`
}

// DirectAuthResponse is the response for single action authorization.
type DirectAuthResponse struct {
	Allowed bool `json:"allowed"`
}

// AuthByResourcesRequest is the request for one action over multiple resources.
type AuthByResourcesRequest struct {
	Subject   Subject    `json:"subject"`
	ActionID  string     `json:"action_id"`
	Resources []Resource `json:"resources"`
}

// AuthByResourcesResponse is the response for one action over multiple resources.
type AuthByResourcesResponse struct {
	ResourceID string `json:"resource_id"`
	Allowed    bool   `json:"allowed"`
}

// AuthByActionsRequest is the request for multiple actions over one resource.
type AuthByActionsRequest struct {
	Subject   Subject   `json:"subject"`
	ActionIDs []string  `json:"action_ids"`
	Resource  *Resource `json:"resource,omitempty"`
}

// AuthByActionsResponse is the response for multiple actions over one resource.
type AuthByActionsResponse struct {
	ActionID string `json:"action_id"`
	Allowed  bool   `json:"allowed"`
}

// AuthorizedResourceRequest is the request for authorized resource scope.
type AuthorizedResourceRequest struct {
	Subject  Subject `json:"subject"`
	ActionID string  `json:"action_id"`
}

// AuthorizedResourceResponse is one authorized resource scope entry.
type AuthorizedResourceResponse struct {
	Type string   `json:"type"`
	IDs  []string `json:"ids"`
}

// ApplyResource is the resource topology used for permission apply URL.
type ApplyResource struct {
	ID        string              `json:"id"`
	Type      string              `json:"type"`
	Ancestors []ApplyAncestorNode `json:"ancestors,omitempty"`
}

// ApplyAncestorNode is one ancestor node in a permission apply resource path.
type ApplyAncestorNode struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ApplyPermission is one permission item in a permission apply URL request.
type ApplyPermission struct {
	ActionID  string          `json:"action_id"`
	Resources []ApplyResource `json:"resources,omitempty"`
}

// ApplyURLRequest is the request for generating a permission apply URL.
type ApplyURLRequest struct {
	SystemID    string            `json:"system_id"`
	Permissions []ApplyPermission `json:"permissions"`
}

// ApplyURLResponse is the response for generating a permission apply URL.
type ApplyURLResponse struct {
	URL string `json:"url"`
}

// TokenResponse is the response for retrieving IAM v4 system auth token.
type TokenResponse struct {
	AuthToken string `json:"auth_token"`
}
