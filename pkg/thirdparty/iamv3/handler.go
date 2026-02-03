/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package iamv3

import (
	"crypto/subtle"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
	"github.com/mitchellh/mapstructure"
)

// IHandler the Handler of IAM v3.
type IHandler interface {
	// Permission checks

	IsAllowed(ctx contextx.IContext, request Request) (bool, error)
	IsAllowedWithCache(ctx contextx.IContext, request Request, ttl time.Duration) (bool, error)
	BatchIsAllowed(ctx contextx.IContext, request Request, resourcesList []Resources) (map[string]bool, error)

	// Multi-action permission checks
	ResourceMultiActionsAllowed(ctx contextx.IContext, request MultiActionRequest) (map[string]bool, error)
	BatchResourceMultiActionsAllowed(ctx contextx.IContext, request MultiActionRequest,
		resourcesList []Resources) (map[string]map[string]bool, error)

	// Token and authentication
	GetToken(ctx contextx.IContext) (string, error)
	IsBasicAuthAllowed(ctx contextx.IContext, username, password string) error

	// Apply URL
	GetApplyURL(ctx contextx.IContext, application Application) (string, error)
	GenPermissionApplyData(a ApplicationActionListForApply) (map[string]interface{}, error)
}

// Handler the Handler of IAM v3.
type Handler struct {
	cli   *cli
	cache *Cache
}

// Verify that Handler implements IHandler interface.
var _ IHandler = (*Handler)(nil)

// New initialize a new IAM v3 Handler.
func New(c *restclient.Capability, conf *Config) (*Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		cli:   cli,
		cache: NewCache(),
	}

	return h, nil
}

// IsAllowed checks if the user has permission for the given action.
func (h *Handler) IsAllowed(ctx contextx.IContext, request Request) (bool, error) {
	if err := request.Validate(); err != nil {
		return false, err
	}

	// Build policy query input
	input := &PolicyQueryInput{
		System:    request.System,
		Subject:   request.Subject,
		Action:    request.Action,
		Resources: request.Resources,
	}

	// Query policy from IAM (use v2 API)
	policyData, err := h.cli.v2PolicyQuery(ctx, input)
	if err != nil {
		return false, err
	}

	// If no policy data, deny access
	if policyData == nil {
		return false, nil
	}

	// Decode policy data to expression
	expr := expression.ExprCell{}
	if err := mapstructure.Decode(policyData, &expr); err != nil {
		return false, fmt.Errorf("failed to decode expression: %w", err)
	}

	// Evaluate the expression against the resources
	objSet := request.GenObjectSet()

	return expr.Eval(objSet), nil
}

// IsAllowedWithCache checks permission with caching support.
func (h *Handler) IsAllowedWithCache(ctx contextx.IContext, request Request, ttl time.Duration) (bool, error) {
	// Generate cache key
	cacheKey, err := request.CacheKey()
	if err != nil {
		// If cache key generation fails, fall back to non-cached check
		return h.IsAllowed(ctx, request)
	}

	// Check cache first
	if result, found := h.cache.Get(cacheKey); found {
		return result, nil
	}

	// Not in cache, perform actual check
	result, err := h.IsAllowed(ctx, request)
	if err != nil {
		return false, err
	}

	// Cache the result
	h.cache.Set(cacheKey, result, ttl)

	return result, nil
}

// BatchIsAllowed checks permissions for multiple resource sets.
func (h *Handler) BatchIsAllowed(ctx contextx.IContext, request Request,
	resourcesList []Resources) (map[string]bool, error) {

	if err := request.Validate(); err != nil {
		return nil, err
	}

	results := make(map[string]bool, len(resourcesList))

	// Build policy query input (without resources, to get the policy)
	// Clear resources as we'll evaluate for each resource set separately
	input := &PolicyQueryInput{
		System:  request.System,
		Subject: request.Subject,
		Action:  request.Action,
	}

	// Query policy from IAM (use v2 API)
	policyData, err := h.cli.v2PolicyQuery(ctx, input)
	if err != nil {
		return nil, err
	}

	// If no policy data, deny all
	if policyData == nil {
		for _, resources := range resourcesList {
			key := buildResourceID(resources)
			results[key] = false
		}

		return results, nil
	}

	// Decode expression once
	expr := expression.ExprCell{}
	if err := mapstructure.Decode(policyData, &expr); err != nil {
		return nil, fmt.Errorf("failed to decode expression: %w", err)
	}

	// Evaluate for each resource set
	for _, resources := range resourcesList {
		objSet := NewObjectSet(resources)
		key := buildResourceID(resources)
		results[key] = expr.Eval(objSet)
	}

	return results, nil
}

// ResourceMultiActionsAllowed checks multiple actions for a single resource.
func (h *Handler) ResourceMultiActionsAllowed(ctx contextx.IContext,
	request MultiActionRequest) (map[string]bool, error) {

	if err := request.Validate(); err != nil {
		return nil, err
	}

	results := make(map[string]bool, len(request.Actions))

	// Build policy query by actions input
	input := &PolicyQueryByActionsInput{
		System:    request.System,
		Subject:   request.Subject,
		Actions:   request.Actions,
		Resources: request.Resources,
	}

	// Query policies from IAM (use v2 API)
	policiesData, err := h.cli.v2PolicyQueryByActions(ctx, input)
	if err != nil {
		return nil, err
	}

	// Create object set for evaluation
	objSet := NewObjectSet(request.Resources)

	// Decode to action policies
	var actionPolicies []ActionPolicy
	if err := mapstructure.Decode(policiesData, &actionPolicies); err != nil {
		return nil, fmt.Errorf("failed to decode action policies: %w", err)
	}

	// Process each action's policy
	for _, actionPolicy := range actionPolicies {
		results[actionPolicy.Action.ID] = actionPolicy.Condition.Eval(objSet)
	}

	return results, nil
}

// BatchResourceMultiActionsAllowed checks multiple actions for multiple resources.
func (h *Handler) BatchResourceMultiActionsAllowed(ctx contextx.IContext,
	request MultiActionRequest, resourcesList []Resources) (map[string]map[string]bool, error) {

	if err := request.Validate(); err != nil {
		return nil, err
	}

	results := make(map[string]map[string]bool, len(resourcesList))

	// Build policy query by actions input (without resources)
	// Clear resources as we'll evaluate for each resource set separately
	input := &PolicyQueryByActionsInput{
		System:  request.System,
		Subject: request.Subject,
		Actions: request.Actions,
	}

	// Query policies from IAM (use v2 API)
	policiesData, err := h.cli.v2PolicyQueryByActions(ctx, input)
	if err != nil {
		return nil, err
	}

	// Decode to action policies once
	var actionPolicies []ActionPolicy
	if err := mapstructure.Decode(policiesData, &actionPolicies); err != nil {
		return nil, fmt.Errorf("failed to decode action policies: %w", err)
	}

	// Evaluate for each resource set
	for _, resources := range resourcesList {
		resourceKey := buildResourceID(resources)
		actionResults := make(map[string]bool, len(request.Actions))
		objSet := NewObjectSet(resources)

		for _, actionPolicy := range actionPolicies {
			actionResults[actionPolicy.Action.ID] = actionPolicy.Condition.Eval(objSet)
		}

		results[resourceKey] = actionResults
	}

	return results, nil
}

// GetToken retrieves the system token from IAM.
func (h *Handler) GetToken(ctx contextx.IContext) (string, error) {
	return h.cli.getToken(ctx, h.cli.config.SystemID)
}

// IsBasicAuthAllowed validates basic auth credentials against IAM system token.
func (h *Handler) IsBasicAuthAllowed(ctx contextx.IContext, username, password string) error {
	// Get system token
	token, err := h.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Validate password matches token using constant-time comparison
	if subtle.ConstantTimeCompare([]byte(password), []byte(token)) != 1 {
		return fmt.Errorf("invalid password")
	}

	return nil
}

// GetApplyURL retrieves the permission apply URL from IAM.
func (h *Handler) GetApplyURL(ctx contextx.IContext, application Application) (string, error) {
	if err := application.Validate(); err != nil {
		return "", err
	}

	return h.cli.getApplyURL(ctx, &application)
}

// GenPermissionApplyData generates permission apply data for frontend display.
func (h *Handler) GenPermissionApplyData(a ApplicationActionListForApply) (map[string]interface{}, error) {
	return map[string]interface{}{
		"system_id":   a.SystemID,
		"system_name": a.SystemName,
		"actions":     a.Actions,
	}, nil
}

// buildResourceID generates a unique key for a resource set.
// This matches the original iam-go-sdk behavior.
func buildResourceID(resources Resources) string {
	if len(resources) == 0 {
		return ""
	}

	if len(resources) == 1 {
		return resources[0].ID
	}

	nodeIDs := make([]string, 0, len(resources))
	for _, node := range resources {
		nodeIDs = append(nodeIDs, fmt.Sprintf("%s,%s", node.Type, node.ID))
	}

	return strings.Join(nodeIDs, "/")
}
