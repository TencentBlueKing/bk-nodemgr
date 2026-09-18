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
	"fmt"
	"net/http"
	"regexp"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

// cli is the client for IAM v4.
type cli struct {
	client restclient.IClient
	config *Config
}

// newClient initializes a new IAM v4 client.
func newClient(c *restclient.Capability, conf *Config) (*cli, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	restCli, err := apigwclient.NewClient(c, "/api", conf.VirtualUserConfig)
	if err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
		config: conf,
	}, nil
}

// getHeader returns common HTTP headers for IAM v4 API requests.
func (c *cli) getHeader(ctx contextx.IContext) http.Header {
	header := http.Header{}
	header.Set(apigwheader.BKGWTenantIDKey, ctx.TenantID())
	header.Set(apigwheader.BKGWAuthKey, c.config.VirtualUserConfig.GetAuthHeader(ctx))

	return header
}

func (c *cli) addAuthorization(ctx contextx.IContext, systemID string, req AuthorizationRequest) error {
	if err := ctx.CheckTenantID(); err != nil {
		return fmt.Errorf("add authorization: %w", err)
	}
	if err := ctx.CheckBKUsername(); err != nil {
		return fmt.Errorf("add authorization: %w", err)
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`).MatchString(systemID) {
		return fmt.Errorf("add authorization: invalid system ID")
	}
	if err := req.Validate(); err != nil {
		return fmt.Errorf("add authorization: invalid request: %w", err)
	}

	resp := new(BaseBroker[struct{}])
	header := c.getHeader(ctx)
	header.Set("X-Bkiam-Operator", ctx.BKUsername())
	// IAM requires application and resource permission; send one grant as an array (maximum 20).
	result := c.client.Post().
		SubResourcef("/v1/open/rbac/mgmt/systems/%s/authorizations/", systemID).
		WithContext(ctx).
		WithHeaders(header).
		Body([]AuthorizationRequest{req}).
		EnableLogBody().
		Do()
	if err := result.Into(resp); err != nil {
		return fmt.Errorf("add authorization failed: %w", err)
	}
	if result.StatusCode != http.StatusCreated {
		return fmt.Errorf("add authorization: unexpected HTTP %d (request-id: %s)",
			result.StatusCode, result.Header.Get("X-Bkapi-Request-Id"))
	}
	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("add authorization failed: %w", err)
	}

	return nil
}

func (c *cli) directAuth(ctx contextx.IContext, systemID string, req DirectAuthRequest) (bool, error) {
	resp := new(BaseBroker[DirectAuthResponse])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v1/open/rbac/authorization/systems/%s/auth/", systemID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return false, err
	}

	if err := resp.IsFailed(); err != nil {
		return false, fmt.Errorf("direct auth failed: %w", err)
	}

	return resp.Data.Allowed, nil
}

func (c *cli) authByResources(
	ctx contextx.IContext, systemID string, req AuthByResourcesRequest,
) ([]AuthByResourcesResponse, error) {

	resp := new(BaseBroker[[]AuthByResourcesResponse])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v1/open/rbac/authorization/systems/%s/auth-by-resources/", systemID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("auth by resources failed: %w", err)
	}

	return resp.Data, nil
}

func (c *cli) authByActions(ctx contextx.IContext, systemID string, req AuthByActionsRequest) ([]AuthByActionsResponse, error) {
	resp := new(BaseBroker[[]AuthByActionsResponse])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v1/open/rbac/authorization/systems/%s/auth-by-actions/", systemID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("auth by actions failed: %w", err)
	}

	return resp.Data, nil
}

func (c *cli) listAuthorizedResources(
	ctx contextx.IContext, systemID string, req AuthorizedResourceRequest,
) ([]AuthorizedResourceResponse, error) {

	resp := new(BaseBroker[[]AuthorizedResourceResponse])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v1/open/rbac/authorization/systems/%s/relation/authorized-resources/", systemID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list authorized resources failed: %w", err)
	}

	return resp.Data, nil
}

func (c *cli) getApplyURL(ctx contextx.IContext, req ApplyURLRequest) (string, error) {
	resp := new(BaseBroker[ApplyURLResponse])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v1/open/application/permission-apply-urls/").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return "", err
	}

	if err := resp.IsFailed(); err != nil {
		return "", fmt.Errorf("get apply URL failed: %w", err)
	}

	return resp.Data.URL, nil
}

func (c *cli) getToken(ctx contextx.IContext, systemID string) (string, error) {
	resp := new(BaseBroker[TokenResponse])
	header := c.getHeader(ctx)

	err := c.client.Get().
		SubResourcef("/v1/open/rbac/model/systems/%s/auth-token/", systemID).
		WithContext(ctx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return "", err
	}

	if err := resp.IsFailed(); err != nil {
		return "", fmt.Errorf("get token failed: %w", err)
	}

	return resp.Data.AuthToken, nil
}
