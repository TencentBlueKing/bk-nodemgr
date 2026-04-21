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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv3/policy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
	"github.com/mitchellh/mapstructure"
)

// IHandler is the handler interface for IAM v3 permission checks and token management.
type IHandler interface {
	// IsAllowed checks if a user is allowed to perform an action on resources.
	IsAllowed(ctx contextx.IContext, req types.IAMCheckRequest) (bool, error)

	// IsAllowedWithCache checks if a user is allowed to perform an action on resources with caching.
	IsAllowedWithCache(ctx contextx.IContext, req types.IAMCheckRequest, ttl time.Duration) (bool, error)

	// BatchIsAllowed checks if a user is allowed to perform an action on multiple resource sets.
	BatchIsAllowed(ctx contextx.IContext, req types.IAMCheckRequest, resourcesList [][]types.IAMResource) (map[string]bool, error)

	// ResourceMultiActionsAllowed checks if a user is allowed to perform multiple actions on resources.
	ResourceMultiActionsAllowed(ctx contextx.IContext, req types.IAMMultiActionCheckRequest) (map[string]bool, error)

	// BatchResourceMultiActionsAllowed checks if a user is allowed to perform multiple actions on multiple resource sets.
	BatchResourceMultiActionsAllowed(
		ctx contextx.IContext,
		req types.IAMMultiActionCheckRequest,
		resourcesList [][]types.IAMResource,
	) (map[string]map[string]bool, error)

	// GetToken retrieves the IAM token for the current context.
	GetToken(ctx contextx.IContext) (string, error)

	// IsBasicAuthAllowed checks if basic authentication credentials are valid.
	IsBasicAuthAllowed(ctx contextx.IContext, username, password string) error

	// GetApplyURL generates a permission apply URL for the given request.
	GetApplyURL(ctx contextx.IContext, req types.IAMApplyRequest) (string, error)

	// GetPolicyExpression retrieves the raw IAM policy expression for a single action.
	GetPolicyExpression(
		ctx contextx.IContext,
		req types.IAMAuthorizedInstancesRequest,
	) (*expression.ExprCell, error)

	// ListAuthorizedInstances queries IAM policy and resolves the authorized.
	ListAuthorizedInstances(
		ctx contextx.IContext,
		req types.IAMAuthorizedInstancesRequest,
	) (bool, []types.IAMResource, error)
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

func toWireResources(resources []types.IAMResource) Resources {
	wireResources := make(Resources, 0, len(resources))
	for _, resource := range resources {
		wireResources = append(wireResources, ResourceNode{
			System:    resource.SystemID,
			Type:      resource.Type,
			ID:        resource.ID,
			Attribute: resource.Attributes,
		})
	}

	return wireResources
}

func toWireRequest(req types.IAMCheckRequest) Request {
	return Request{
		System: req.SystemID,
		Subject: Subject{
			Type: "user",
			ID:   req.Username,
		},
		Action:    Action{ID: req.ActionID},
		Resources: toWireResources(req.Resources),
	}
}

func toAuthorizedInstancesWireRequest(req types.IAMAuthorizedInstancesRequest) Request {
	return Request{
		System: req.SystemID,
		Subject: Subject{
			Type: "user",
			ID:   req.Username,
		},
		Action: Action{ID: req.ActionID},
	}
}

// IsAllowed checks if the user has permission for the given action.
func (h *Handler) IsAllowed(ctx contextx.IContext, req types.IAMCheckRequest) (bool, error) {
	request := toWireRequest(req)
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

	// Evaluate the expression against the resources
	objSet := request.GenObjectSet()

	return policyData.Eval(objSet), nil
}

// IsAllowedWithCache checks permission with caching support.
func (h *Handler) IsAllowedWithCache(ctx contextx.IContext, req types.IAMCheckRequest, ttl time.Duration) (bool, error) {
	// Generate cache key
	wireReq := toWireRequest(req)
	cacheKey, err := wireReq.CacheKey()
	if err != nil {
		// If cache key generation fails, fall back to non-cached check
		return h.IsAllowed(ctx, req)
	}

	// Check cache first
	if result, found := h.cache.Get(cacheKey); found {
		return result, nil
	}

	// Not in cache, perform actual check
	result, err := h.IsAllowed(ctx, req)
	if err != nil {
		return false, err
	}

	// Cache the result
	h.cache.Set(cacheKey, result, ttl)

	return result, nil
}

// BatchIsAllowed checks permissions for multiple resource sets.
func (h *Handler) BatchIsAllowed(ctx contextx.IContext, req types.IAMCheckRequest,
	resourcesList [][]types.IAMResource) (map[string]bool, error) {

	request := toWireRequest(req)

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
			wireResources := toWireResources(resources)
			key := buildResourceID(wireResources)
			results[key] = false
		}

		return results, nil
	}

	// Evaluate for each resource set
	for _, resources := range resourcesList {
		wireResources := toWireResources(resources)
		objSet := NewObjectSet(wireResources)
		key := buildResourceID(wireResources)
		results[key] = policyData.Eval(objSet)
	}

	return results, nil
}

// ResourceMultiActionsAllowed checks multiple actions for a single resource.
func (h *Handler) ResourceMultiActionsAllowed(ctx contextx.IContext,
	req types.IAMMultiActionCheckRequest) (map[string]bool, error) {

	request := MultiActionRequest{
		System: req.SystemID,
		Subject: Subject{
			Type: "user",
			ID:   req.Username,
		},
		Actions: func() []Action {
			actions := make([]Action, 0, len(req.ActionIDs))
			for _, id := range req.ActionIDs {
				actions = append(actions, Action{ID: id})
			}

			return actions
		}(),
		Resources: toWireResources(req.Resources),
	}

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
	req types.IAMMultiActionCheckRequest, resourcesList [][]types.IAMResource) (map[string]map[string]bool, error) {

	request := MultiActionRequest{
		System: req.SystemID,
		Subject: Subject{
			Type: "user",
			ID:   req.Username,
		},
		Actions: func() []Action {
			actions := make([]Action, 0, len(req.ActionIDs))
			for _, id := range req.ActionIDs {
				actions = append(actions, Action{ID: id})
			}

			return actions
		}(),
	}

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
		wireResources := toWireResources(resources)
		resourceKey := buildResourceID(wireResources)
		actionResults := make(map[string]bool, len(request.Actions))
		objSet := NewObjectSet(wireResources)

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

// IsBasicAuthAllowed validates basic auth credentials per the upstream iam-go-sdk
// convention: username must equal "bk_iam" and password must match the IAM system
// token. Both comparisons use constant-time operations to mitigate timing
// side-channel attacks.
func (h *Handler) IsBasicAuthAllowed(ctx contextx.IContext, username, password string) error {
	if subtle.ConstantTimeCompare([]byte(username), []byte("bk_iam")) != 1 {
		return fmt.Errorf("invalid credentials")
	}

	token, err := h.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	if subtle.ConstantTimeCompare([]byte(password), []byte(token)) != 1 {
		return fmt.Errorf("invalid credentials")
	}

	return nil
}

// GetApplyURL retrieves the permission apply URL from IAM.
func (h *Handler) GetApplyURL(ctx contextx.IContext, req types.IAMApplyRequest) (string, error) {
	actions := make([]ApplicationAction, 0, len(req.Actions))
	for _, action := range req.Actions {
		appAction := ApplicationAction{
			ID: action.ID,
		}

		if len(action.RelatedResourceTypes) > 0 {
			appAction.RelatedResourceTypes = make([]ApplicationRelatedResourceType, 0, len(action.RelatedResourceTypes))
			for _, related := range action.RelatedResourceTypes {
				instances := make([]ApplicationResourceInstance, 0, len(related.Instances))
				for _, instance := range related.Instances {
					nodes := make(ApplicationResourceInstance, 0, len(instance))
					for _, node := range instance {
						nodes = append(nodes, ApplicationResourceNode{
							Type: node.Type,
							ID:   node.ID,
						})
					}
					instances = append(instances, nodes)
				}

				appAction.RelatedResourceTypes = append(appAction.RelatedResourceTypes, ApplicationRelatedResourceType{
					SystemID:  related.SystemID,
					Type:      related.Type,
					Instances: instances,
				})
			}
		}

		actions = append(actions, appAction)
	}

	application := Application{
		SystemID: req.SystemID,
		Actions:  actions,
	}
	if err := application.Validate(); err != nil {
		return "", err
	}

	return h.cli.getApplyURL(ctx, &application)
}

// GetPolicyExpression retrieves the raw IAM policy expression without parsing.
func (h *Handler) GetPolicyExpression(
	ctx contextx.IContext,
	req types.IAMAuthorizedInstancesRequest,
) (*expression.ExprCell, error) {

	request := toAuthorizedInstancesWireRequest(req)

	if err := request.Validate(); err != nil {
		return nil, err
	}

	input := &PolicyQueryInput{
		System:  request.System,
		Subject: request.Subject,
		Action:  request.Action,
	}

	return h.cli.v2PolicyQuery(ctx, input)
}

// ListAuthorizedInstances queries IAM policy and resolves the authorized
// resource scope for a single action.
func (h *Handler) ListAuthorizedInstances(
	ctx contextx.IContext,
	req types.IAMAuthorizedInstancesRequest,
) (bool, []types.IAMResource, error) {

	request := toAuthorizedInstancesWireRequest(req)

	if err := request.Validate(); err != nil {
		return false, nil, err
	}

	input := &PolicyQueryInput{
		System:  request.System,
		Subject: request.Subject,
		Action:  request.Action,
	}

	policyData, err := h.cli.v2PolicyQuery(ctx, input)
	if err != nil {
		return false, nil, err
	}

	return policy.Parse(policyData, req.SystemID, req.ResourceType)
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
