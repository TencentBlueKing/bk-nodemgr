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

// Package iamv3 provides the IAM v3 client implementation.
package iamv3

import (
	"crypto/md5" //nolint:gosec // MD5 is used for cache key generation, not security
	"encoding/hex"
	"fmt"
	"time"

	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
	jsoniter "github.com/json-iterator/go" //nolint:importas // standard alias for jsoniter
)

// RespCommon describe the common part of IAM response.
type RespCommon struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// BaseBroker describe the base broker for IAM API response.
type BaseBroker[T any] struct {
	RespCommon
	Data T `json:"data"`
}

// IsFailed check if the response indicates an error.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Code == 0 {
		return nil
	}

	return fmt.Errorf("code(%d), msg(%s)", resp.Code, resp.Message)
}

// Config the config of IAM v3.
type Config struct {
	APIGWUserConfig apigwclient.UserConfig
	// SystemID is the system identifier registered in IAM.
	SystemID string
	// CallbackPath is the callback path for IAM resource provider.
	// This field is required.
	CallbackPath string
}

// Validate validates the config.
func (conf *Config) Validate() error {
	if err := conf.APIGWUserConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate IAM v3 client config: %w", err)
	}

	if conf.SystemID == "" {
		return fmt.Errorf("system ID is required for IAM v3")
	}

	// Validate callback path is configured.
	if conf.CallbackPath == "" {
		return fmt.Errorf("callback path is required for IAM v3")
	}

	return nil
}

// Subject is the object of permission.
type Subject struct {
	Type string `json:"type" binding:"required,oneof=user"`
	ID   string `json:"id" binding:"required"`
}

// Action is the action of permission.
type Action struct {
	ID string `json:"id" binding:"required" mapstructure:"id"`
}

// ResourceNode is the mini unit of a resource.
type ResourceNode struct {
	System    string                 `json:"system" binding:"required"`
	Type      string                 `json:"type" binding:"required"`
	ID        string                 `json:"id" binding:"required"`
	Attribute map[string]interface{} `json:"attribute" binding:"required"`
}

// Resources is a slice of ResourceNode.
type Resources []ResourceNode

// Request is the policy query request body.
type Request struct {
	System    string    `json:"system" binding:"required"`
	Subject   Subject   `json:"subject" binding:"required"`
	Action    Action    `json:"action" binding:"required"`
	Resources Resources `json:"resources" binding:"omitempty"`
}

// MultiActionRequest is the request object for Multi Actions Request.
type MultiActionRequest struct {
	System    string    `json:"system" binding:"required"`
	Subject   Subject   `json:"subject" binding:"required"`
	Actions   []Action  `json:"actions" binding:"required"`
	Resources Resources `json:"resources" binding:"omitempty"`
}

// ActionPolicy is the response struct for action policy.
type ActionPolicy struct {
	Action    Action              `json:"action" mapstructure:"action"`
	Condition expression.ExprCell `json:"condition" mapstructure:"condition"`
}

// ApplicationResourceNode is the resource node struct for application.
type ApplicationResourceNode struct {
	Type string `json:"type" binding:"required"`
	ID   string `json:"id" binding:"required"`
}

// ApplicationResourceInstance is a slice of ApplicationResourceNode.
type ApplicationResourceInstance []ApplicationResourceNode

// ApplicationRelatedResourceType is the related resource type for application.
type ApplicationRelatedResourceType struct {
	SystemID  string                        `json:"system_id"`
	Type      string                        `json:"type"`
	Instances []ApplicationResourceInstance `json:"instances"`
}

// ApplicationAction is the action for application.
type ApplicationAction struct {
	ID                   string                           `json:"id"`
	RelatedResourceTypes []ApplicationRelatedResourceType `json:"related_resource_types"`
}

// Application is the application for permission.
type Application struct {
	SystemID string              `json:"system_id"`
	Actions  []ApplicationAction `json:"actions"`
}

// ApplicationResourceNodeWithName is the resource node struct for application with names.
type ApplicationResourceNodeWithName struct {
	Type     string `json:"type" binding:"required"`
	TypeName string `json:"type_name" binding:"required"`
	ID       string `json:"id" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

// ApplicationResourceInstanceWithName is a slice of ApplicationResourceNodeWithName.
type ApplicationResourceInstanceWithName []ApplicationResourceNodeWithName

// ApplicationRelatedResourceTypeWithName is the related resource type with names.
type ApplicationRelatedResourceTypeWithName struct {
	SystemID   string                                `json:"system_id" binding:"required"`
	SystemName string                                `json:"system_name" binding:"required"`
	Type       string                                `json:"type" binding:"required"`
	TypeName   string                                `json:"type_name" binding:"required"`
	Instances  []ApplicationResourceInstanceWithName `json:"instances" binding:"required"`
}

// ApplicationActionForApply is the action for apply.
type ApplicationActionForApply struct {
	ID                   string                                   `json:"id" binding:"required"`
	Name                 string                                   `json:"name" binding:"required"`
	RelatedResourceTypes []ApplicationRelatedResourceTypeWithName `json:"related_resource_types"`
}

// ApplicationActionListForApply is the action list for apply.
type ApplicationActionListForApply struct {
	SystemID   string                      `json:"system_id" binding:"required"`
	SystemName string                      `json:"system_name" binding:"required"`
	Actions    []ApplicationActionForApply `json:"actions" binding:"required"`
}

// PolicyQueryInput is the input for policy query API.
type PolicyQueryInput struct {
	System    string    `json:"system"`
	Subject   Subject   `json:"subject"`
	Action    Action    `json:"action"`
	Resources Resources `json:"resources,omitempty"`
}

// PolicyQueryByActionsInput is the input for policy query by actions API.
type PolicyQueryByActionsInput struct {
	System    string    `json:"system"`
	Subject   Subject   `json:"subject"`
	Actions   []Action  `json:"actions"`
	Resources Resources `json:"resources,omitempty"`
}

// PolicyAuthInput is the input for policy auth API.
type PolicyAuthInput struct {
	System    string    `json:"system"`
	Subject   Subject   `json:"subject"`
	Action    Action    `json:"action"`
	Resources Resources `json:"resources"`
}

// PolicyAuthByResourcesInput is the input for policy auth by resources API.
type PolicyAuthByResourcesInput struct {
	System        string      `json:"system"`
	Subject       Subject     `json:"subject"`
	Action        Action      `json:"action"`
	ResourcesList []Resources `json:"resources_list"`
}

// PolicyAuthByActionsInput is the input for policy auth by actions API.
type PolicyAuthByActionsInput struct {
	System    string    `json:"system"`
	Subject   Subject   `json:"subject"`
	Actions   []Action  `json:"actions"`
	Resources Resources `json:"resources"`
}

// TokenResponse is the response for get token API.
type TokenResponse struct {
	Token string `json:"token"`
}

// ApplyURLResponse is the response for get apply URL API.
type ApplyURLResponse struct {
	URL string `json:"url"`
}

// PolicySubjectsInput is the input for policy subjects API.
type PolicySubjectsInput struct {
	IDs string `json:"ids"`
}

// Validate validates the Request.
func (r *Request) Validate() error {
	if r.System == "" {
		return fmt.Errorf("system is required")
	}
	if r.Subject.Type == "" || r.Subject.ID == "" {
		return fmt.Errorf("subject is required")
	}
	if r.Action.ID == "" {
		return fmt.Errorf("action is required")
	}

	return nil
}

// GenObjectSet creates an ObjectSet from the resources of request.
func (r *Request) GenObjectSet() expression.ObjectSetInterface {
	return NewObjectSet(r.Resources)
}

// CacheKey makes the unique key of a request.
func (r *Request) CacheKey() (string, error) {
	b, err := jsoniter.Marshal(r)
	if err != nil {
		return "", err
	}

	h := md5.New() //nolint:gosec // MD5 is used for cache key generation, not security
	_, err = h.Write(b)
	if err != nil {
		return "", err
	}

	return "iam:" + hex.EncodeToString(h.Sum(nil)), nil
}

// NewObjectSet creates an ObjectSet from resources.
func NewObjectSet(resources Resources) expression.ObjectSetInterface {
	objSet := expression.NewObjectSet()

	if len(resources) == 0 {
		return objSet
	}

	for _, resource := range resources {
		attrs := make(map[string]interface{}, len(resource.Attribute)+1)
		attrs["id"] = resource.ID

		for key, value := range resource.Attribute {
			attrs[key] = value
		}

		objSet.Set(resource.Type, attrs)
	}

	return objSet
}

// Validate validates the MultiActionRequest.
func (mar *MultiActionRequest) Validate() error {
	if mar.System == "" {
		return fmt.Errorf("system is required")
	}
	if mar.Subject.Type == "" || mar.Subject.ID == "" {
		return fmt.Errorf("subject is required")
	}
	if len(mar.Actions) == 0 {
		return fmt.Errorf("actions is required")
	}

	return nil
}

// Validate validates the Application.
func (a *Application) Validate() error {
	if a.SystemID == "" {
		return fmt.Errorf("system_id is required")
	}
	if len(a.Actions) == 0 {
		return fmt.Errorf("actions is required")
	}

	return nil
}

// CacheEntry represents a cached permission check result.
type CacheEntry struct {
	Value     bool
	ExpiresAt time.Time
}
