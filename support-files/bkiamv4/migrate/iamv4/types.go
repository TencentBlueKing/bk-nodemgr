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

package iamv4

import (
	"encoding/json"
	"fmt"

	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
)

// Config configures application-authenticated requests to the IAM V4 gateway.
type Config struct {
	BaseURL string
	apigwclient.AppConfig
}

// Validate checks the application-only gateway configuration.
func (c *Config) Validate() error {
	if err := c.AppConfig.Validate(); err != nil {
		return fmt.Errorf("invalid app config: %w", err)
	}
	// The shared auth helper formats JSON rather than marshaling credentials.
	if !json.Valid([]byte(c.GetAuthHeader())) {
		return fmt.Errorf("app-code or app-secret cannot be encoded by the APIGW auth helper")
	}

	return nil
}

// SystemFields preserves omitted fields separately from explicit empty values.
// Nil preserves the remote value; a pointer to an empty value clears it.
type SystemFields struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Clients     *[]string `json:"clients,omitempty"`
	Managers    *[]string `json:"managers,omitempty"`
	CallbackURL *string   `json:"callback_url,omitempty"`
}

// RetrieveSystemReq identifies the system to retrieve.
type RetrieveSystemReq struct {
	SystemID string `json:"-"`
}

// CreateSystemReq is the system registration payload.
type CreateSystemReq struct {
	ID string `json:"id"`
	SystemFields
}

// UpdateSystemReq identifies a system and its supplied mutable fields.
type UpdateSystemReq struct {
	SystemID string `json:"-"`
	SystemFields
}

// RetrieveSystemResp contains the system identity returned by IAM.
type RetrieveSystemResp struct {
	ID string `json:"id"`
}

// CreateSystemResp contains the registered system identity.
type CreateSystemResp struct {
	ID string `json:"id"`
}

// ResourceType describes an IAM resource type and its ordered ancestor chain.
type ResourceType struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Ancestors []string `json:"ancestors"`
}

// ResourceTypeFields preserves omitted fields separately from explicit updates.
// Nil leaves a field unchanged; an empty ancestor slice clears the ancestor chain.
type ResourceTypeFields struct {
	Name      *string   `json:"name,omitempty"`
	Ancestors *[]string `json:"ancestors,omitempty"`
}

// ListResourceTypesReq identifies one page of a system's resource types.
type ListResourceTypesReq struct {
	SystemID string `json:"-"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// ListResourceTypesResp contains one page and the total resource type count.
type ListResourceTypesResp struct {
	// A pointer distinguishes an omitted count from a valid empty collection.
	Count   *int           `json:"count"`
	Results []ResourceType `json:"results"`
}

// BatchCreateResourceTypeReq identifies the system and the array sent as the body.
type BatchCreateResourceTypeReq struct {
	SystemID  string         `json:"-"`
	Resources []ResourceType `json:"-"`
}

// BatchCreateResourceTypeResp contains the registered resource type IDs.
type BatchCreateResourceTypeResp []string

// UpdateResourceTypeReq identifies a resource type and its supplied mutable fields.
type UpdateResourceTypeReq struct {
	SystemID       string `json:"-"`
	ResourceTypeID string `json:"-"`
	ResourceTypeFields
}

// UpdateResourceTypeResp represents the absence of data on a successful HTTP 204.
type UpdateResourceTypeResp struct{}

// Action describes an IAM action; an empty resource type ID means no resource binding.
type Action struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ResourceTypeID string `json:"resource_type_id"`
}

// ListActionsReq identifies one page of a system's actions.
type ListActionsReq struct {
	SystemID string `json:"-"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// ListActionsResp contains one page and the total action count.
type ListActionsResp struct {
	// A pointer distinguishes an omitted count from a valid empty collection.
	Count   *int     `json:"count"`
	Results []Action `json:"results"`
}

// BatchCreateActionReq identifies the system and the array sent as the body.
type BatchCreateActionReq struct {
	SystemID string   `json:"-"`
	Actions  []Action `json:"-"`
}

// BatchCreateActionResp contains the registered action IDs.
type BatchCreateActionResp []string

// UpdateActionReq identifies an action and its new name; binding updates are not supported.
type UpdateActionReq struct {
	SystemID string `json:"-"`
	ActionID string `json:"-"`
	Name     string `json:"name"`
}

// UpdateActionResp represents the absence of data on a successful HTTP 204.
type UpdateActionResp struct{}

// RespError describes an IAM API error.
type RespError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespCommon contains the common IAM response fields.
type RespCommon struct {
	Error     *RespError `json:"error"`
	RequestID string     `json:"request_id"`
}

// BaseBroker describes an IAM response envelope.
type BaseBroker[T any] struct {
	RespCommon
	Data T `json:"data"`
}

// IsFailed checks the IAM error envelope independently of the HTTP status.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Error != nil {
		// Remote messages can contain echoed request data; do not print them.
		return fmt.Errorf("IAM error code(%s), request-id(%s)", resp.Error.Code, resp.RequestID)
	}

	return nil
}
