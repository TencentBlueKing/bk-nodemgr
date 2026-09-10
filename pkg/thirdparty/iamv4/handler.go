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
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IAM docs recommend caching queried system token, but token query protocol exposes no expiration;
// this 1min value is a conservative bk-nodemgr local TTL, not official IAM token lifetime.
const systemTokenCacheExpiration = time.Minute

// ErrInvalidCredentials identifies callback credential mismatches, not token-service failures.
var ErrInvalidCredentials = errors.New("invalid IAM v4 callback credentials")

// IHandler is the handler interface for IAM v4 permission checks and token management.
type IHandler interface {
	// IsAllowed checks if a user is allowed to perform an action on an optional resource.
	IsAllowed(ctx contextx.IContext, req types.IAMCheckRequest) (bool, error)
	// ResourcesAllowed checks if a user is allowed to perform one action on multiple resources.
	ResourcesAllowed(ctx contextx.IContext, req types.IAMCheckRequest) (map[string]bool, error)
	// ActionsAllowed checks if a user is allowed to perform multiple actions on an optional resource.
	ActionsAllowed(ctx contextx.IContext, req types.IAMMultiActionCheckRequest) (map[string]bool, error)
	// ListAuthorizedResources queries IAM v4 authorized resource scopes.
	ListAuthorizedResources(
		ctx contextx.IContext,
		req types.IAMAuthorizedInstancesRequest,
	) ([]AuthorizedResourceResponse, error)
	// GetToken retrieves the IAM v4 callback token for the current context.
	GetToken(ctx contextx.IContext) (string, error)
	// IsBasicAuthAllowed checks if basic authentication credentials are valid.
	IsBasicAuthAllowed(ctx contextx.IContext, username, password string) error
	// GetApplyURL generates a permission apply URL for the given request.
	GetApplyURL(ctx contextx.IContext, req types.IAMApplyRequest) (string, error)
}

// Handler is the Handler of IAM v4.
type Handler struct {
	cli        *cli
	tokenCache cache.ICache
}

var _ IHandler = (*Handler)(nil)

// New initializes a new IAM v4 Handler.
func New(c *restclient.Capability, conf *Config) (*Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &Handler{
		cli:        cli,
		tokenCache: cache.NewMemoryCache(systemTokenCacheExpiration),
	}, nil
}

func toSubject(req types.IAMCheckRequest) Subject {
	return Subject{Type: "user", ID: req.Username}
}

func toMultiActionSubject(req types.IAMMultiActionCheckRequest) Subject {
	return Subject{Type: "user", ID: req.Username}
}

func firstResource(resources []types.IAMResource) *Resource {
	if len(resources) == 0 {
		return nil
	}

	return &Resource{ID: resources[0].ID}
}

// IsAllowed checks if a user is allowed to perform an action on an optional resource.
func (h *Handler) IsAllowed(ctx contextx.IContext, req types.IAMCheckRequest) (bool, error) {
	return h.cli.directAuth(ctx, req.SystemID, DirectAuthRequest{
		Subject:  toSubject(req),
		ActionID: req.ActionID,
		Resource: firstResource(req.Resources),
	})
}

// ResourcesAllowed checks if a user is allowed to perform one action on multiple resources.
func (h *Handler) ResourcesAllowed(ctx contextx.IContext, req types.IAMCheckRequest) (map[string]bool, error) {
	resources := make([]Resource, 0, len(req.Resources))
	for _, resource := range req.Resources {
		resources = append(resources, Resource{ID: resource.ID})
	}

	results, err := h.cli.authByResources(ctx, req.SystemID, AuthByResourcesRequest{
		Subject:   toSubject(req),
		ActionID:  req.ActionID,
		Resources: resources,
	})
	if err != nil {
		return nil, err
	}

	allowedByResource := make(map[string]bool, len(results))
	for _, result := range results {
		allowedByResource[result.ResourceID] = result.Allowed
	}

	return allowedByResource, nil
}

// ActionsAllowed checks if a user is allowed to perform multiple actions on an optional resource.
func (h *Handler) ActionsAllowed(ctx contextx.IContext, req types.IAMMultiActionCheckRequest) (map[string]bool, error) {
	results, err := h.cli.authByActions(ctx, req.SystemID, AuthByActionsRequest{
		Subject:   toMultiActionSubject(req),
		ActionIDs: req.ActionIDs,
		Resource:  firstResource(req.Resources),
	})
	if err != nil {
		return nil, err
	}

	allowedByAction := make(map[string]bool, len(results))
	for _, result := range results {
		allowedByAction[result.ActionID] = result.Allowed
	}

	return allowedByAction, nil
}

// ListAuthorizedResources queries IAM v4 authorized resource scopes.
func (h *Handler) ListAuthorizedResources(
	ctx contextx.IContext,
	req types.IAMAuthorizedInstancesRequest,
) ([]AuthorizedResourceResponse, error) {

	return h.cli.listAuthorizedResources(ctx, req.SystemID, AuthorizedResourceRequest{
		Subject:  Subject{Type: "user", ID: req.Username},
		ActionID: req.ActionID,
	})
}

// GetApplyURL generates a permission apply URL for the given request.
func (h *Handler) GetApplyURL(ctx contextx.IContext, req types.IAMApplyRequest) (string, error) {
	permissions := make([]ApplyPermission, 0, len(req.Actions))
	for _, action := range req.Actions {
		resources := make([]ApplyResource, 0)
		for _, rt := range action.RelatedResourceTypes {
			resources = append(resources, toApplyResources(rt.Instances)...)
		}

		permissions = append(permissions, ApplyPermission{
			ActionID:  action.ID,
			Resources: resources,
		})
	}

	return h.cli.getApplyURL(ctx, ApplyURLRequest{
		SystemID:    req.SystemID,
		Permissions: permissions,
	})
}

func toApplyResources(instances []types.IAMApplyResourceInstance) []ApplyResource {
	resources := make([]ApplyResource, 0, len(instances))
	for _, instance := range instances {
		if len(instance) == 0 {
			continue
		}

		leaf := instance[len(instance)-1]
		resource := ApplyResource{ID: leaf.ID, Type: leaf.Type}
		if len(instance) > 1 {
			resource.Ancestors = make([]ApplyAncestorNode, 0, len(instance)-1)
			for _, ancestor := range instance[:len(instance)-1] {
				resource.Ancestors = append(resource.Ancestors, ApplyAncestorNode{ID: ancestor.ID, Type: ancestor.Type})
			}
		}

		resources = append(resources, resource)
	}

	return resources
}

// GetToken retrieves the IAM token for the current context.
func (h *Handler) GetToken(ctx contextx.IContext) (string, error) {
	cacheKey := fmt.Sprintf("iamv4:system-token:%s:%s", h.cli.config.SystemID, ctx.TenantID())
	data, err := h.tokenCache.Get(ctx, cacheKey)
	if err != nil {
		return h.refreshSystemToken(ctx, cacheKey)
	}

	token := string(data)
	if token == "" {
		return h.refreshSystemToken(ctx, cacheKey)
	}

	return token, nil
}

func (h *Handler) refreshSystemToken(ctx contextx.IContext, cacheKey string) (string, error) {
	token, err := h.cli.getToken(ctx, h.cli.config.SystemID)
	if err != nil {
		return "", fmt.Errorf("failed to refresh IAM v4 system token: %w", err)
	}
	if token == "" {
		return "", fmt.Errorf("IAM v4 system token is empty")
	}

	if err := h.tokenCache.SetWithExpiration(ctx, cacheKey, []byte(token), systemTokenCacheExpiration); err != nil {
		return "", err
	}

	return token, nil
}

// IsBasicAuthAllowed checks if basic authentication credentials are valid.
func (h *Handler) IsBasicAuthAllowed(ctx contextx.IContext, username, password string) error {
	if subtle.ConstantTimeCompare([]byte(username), []byte("bk_iam")) != 1 {
		return ErrInvalidCredentials
	}

	token, err := h.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get IAM v4 system token: %w", err)
	}

	if subtle.ConstantTimeCompare([]byte(password), []byte(token)) != 1 {
		return ErrInvalidCredentials
	}

	return nil
}
