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
	"crypto/md5" //nolint:gosec // MD5 is used for cache key generation, not security
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
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

// defaultAuthCacheExpiration is only the container default of the auth result cache; entries are
// always written with the caller-provided TTL via SetWithExpiration.
const defaultAuthCacheExpiration = 5 * time.Minute

// defaultAuthCacheCleanupInterval is the interval at which expired auth cache entries are purged;
// auth results accumulate per user/action/resource and expired entries are rarely read again, so
// lookup-time expiration alone would never reclaim their memory.
const defaultAuthCacheCleanupInterval = 10 * time.Minute

// authCacheAllowedValue is the cached representation of an allowed auth result.
const authCacheAllowedValue = "1"

// ErrInvalidCredentials identifies callback credential mismatches, not token-service failures.
var ErrInvalidCredentials = errors.New("invalid IAM v4 callback credentials")

// IHandler combines all IAM v4 runtime operations without HTTP details.
type IHandler interface {
	IHandlerAuth
	IHandlerToken
	IHandlerRole
}

// IHandlerAuth checks permissions, queries authorized scopes, and generates apply URLs.
type IHandlerAuth interface {
	// IsAllowed checks if a user is allowed to perform an action on an optional resource.
	IsAllowed(ctx contextx.IContext, req types.IAMCheckRequest) (bool, error)
	// IsAllowedWithCache checks permission like IsAllowed, caching allowed results for ttl.
	// Denied results and IAM errors are never cached, so newly granted permissions take effect
	// immediately; the deny caching policy is tracked separately.
	IsAllowedWithCache(ctx contextx.IContext, req types.IAMCheckRequest, ttl time.Duration) (bool, error)
	// ResourcesAllowed checks if a user is allowed to perform one action on multiple resources.
	ResourcesAllowed(ctx contextx.IContext, req types.IAMCheckRequest) (map[string]bool, error)
	// ActionsAllowed checks if a user is allowed to perform multiple actions on an optional resource.
	ActionsAllowed(ctx contextx.IContext, req types.IAMMultiActionCheckRequest) (map[string]bool, error)
	// ListAuthorizedResources queries IAM v4 authorized resource scopes.
	ListAuthorizedResources(
		ctx contextx.IContext,
		req types.IAMAuthorizedInstancesRequest,
	) ([]AuthorizedResourceResponse, error)
	// GetApplyURL generates a permission apply URL for the given request.
	GetApplyURL(ctx contextx.IContext, req types.IAMApplyRequest) (string, error)
}

// IHandlerToken retrieves system tokens and validates callback credentials.
type IHandlerToken interface {
	// GetToken retrieves the IAM v4 callback token for the current context.
	GetToken(ctx contextx.IContext) (string, error)
	// IsBasicAuthAllowed checks if basic authentication credentials are valid.
	IsBasicAuthAllowed(ctx contextx.IContext, username, password string) error
}

// IHandlerRole grants IAM v4 roles independently of read-only permission checks.
type IHandlerRole interface {
	GrantRole(ctx contextx.IContext, req types.IAMRoleGrantRequest) error
}

// Handler is the Handler of IAM v4.
type Handler struct {
	cli        *cli
	tokenCache cache.ICache
	authCache  cache.ICache
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
		authCache:  cache.NewMemoryCacheWithConfig(defaultAuthCacheExpiration, defaultAuthCacheCleanupInterval),
	}, nil
}

func toSubject(req types.IAMCheckRequest) Subject {
	return Subject{Type: "user", ID: req.Username}
}

func toMultiActionSubject(req types.IAMMultiActionCheckRequest) Subject {
	return Subject{Type: "user", ID: req.Username}
}

func toResource(resource types.IAMResource) Resource {
	return Resource{ID: resource.ID, Attributes: resource.Attributes}
}

func firstResource(resources []types.IAMResource) *Resource {
	if len(resources) == 0 {
		return nil
	}

	resource := toResource(resources[0])

	return &resource
}

// IsAllowed checks if a user is allowed to perform an action on an optional resource.
func (h *Handler) IsAllowed(ctx contextx.IContext, req types.IAMCheckRequest) (bool, error) {
	return h.cli.directAuth(ctx, req.SystemID, DirectAuthRequest{
		Subject:  toSubject(req),
		ActionID: req.ActionID,
		Resource: firstResource(req.Resources),
	})
}

// IsAllowedWithCache checks permission like IsAllowed, caching allowed results for ttl.
// Denied results and IAM errors are never cached, so newly granted permissions take effect
// immediately; the deny caching policy is tracked separately.
func (h *Handler) IsAllowedWithCache(ctx contextx.IContext, req types.IAMCheckRequest, ttl time.Duration) (bool, error) {
	cacheKey, err := h.authCacheKey(ctx, req)
	if err != nil {
		// If cache key generation fails, fall back to non-cached check
		return h.IsAllowed(ctx, req)
	}

	data, err := h.authCache.Get(ctx, cacheKey)
	if err == nil && string(data) == authCacheAllowedValue {
		return true, nil
	}

	allowed, err := h.IsAllowed(ctx, req)
	if err != nil {
		return false, err
	}
	if !allowed {
		return false, nil
	}

	// Caching is best-effort: a cache write failure must not fail an IAM-allowed request;
	// the result is cached on the next allowed check.
	_ = h.authCache.SetWithExpiration(ctx, cacheKey, []byte(authCacheAllowedValue), ttl)

	return true, nil
}

// authCacheKey builds the auth result cache key covering tenant, user, system, action, and
// resource attributes, so entries are never shared across different requests.
func (h *Handler) authCacheKey(ctx contextx.IContext, req types.IAMCheckRequest) (string, error) {
	key := struct {
		SystemID string    `json:"system_id"`
		TenantID string    `json:"tenant_id"`
		Username string    `json:"username"`
		ActionID string    `json:"action_id"`
		Resource *Resource `json:"resource,omitempty"`
	}{
		SystemID: req.SystemID,
		TenantID: ctx.TenantID(),
		Username: req.Username,
		ActionID: req.ActionID,
		Resource: firstResource(req.Resources),
	}

	// encoding/json sorts map keys, so identical resource attributes always marshal to the
	// same bytes and produce a deterministic cache key.
	b, err := json.Marshal(key)
	if err != nil {
		return "", err
	}

	sum := md5.Sum(b) //nolint:gosec // MD5 is used for cache key generation, not security

	return "iamv4:auth:" + hex.EncodeToString(sum[:]), nil
}

// GrantRole grants a role to req.Username, using the context's operator and tenant.
func (h *Handler) GrantRole(ctx contextx.IContext, req types.IAMRoleGrantRequest) error {
	return h.cli.addAuthorization(ctx, h.cli.config.SystemID, AuthorizationRequest{
		Subject:               Subject{Type: "user", ID: req.Username},
		RoleID:                req.RoleID,
		RelatedResourceTypeID: string(req.Resource.Type),
		Resources: []AuthorizationResource{{
			Type: string(req.Resource.Type),
			ID:   req.Resource.ID,
		}},
		ExpiredAt: req.ExpiresAt.Unix(),
	})
}

// ResourcesAllowed checks if a user is allowed to perform one action on multiple resources.
func (h *Handler) ResourcesAllowed(ctx contextx.IContext, req types.IAMCheckRequest) (map[string]bool, error) {
	resources := make([]Resource, 0, len(req.Resources))
	for _, resource := range req.Resources {
		resources = append(resources, toResource(resource))
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
