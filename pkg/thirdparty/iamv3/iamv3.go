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

package iamv3

import (
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
)

const (
	// HeaderKeyIAMVersion is the key for IAM version header.
	HeaderKeyIAMVersion = "X-Bk-IAM-Version"
	// HeaderValueIAMVersion is the value for IAM version header.
	HeaderValueIAMVersion = "1"
)

// cli client for IAM v3.
type cli struct {
	client restclient.IClient
	config *Config
}

// newClient initialize a new IAM v3 client.
func newClient(c *restclient.Capability, conf *Config) (*cli, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	// Use /api as baseURL, v1/v2 will be handled by specific API methods
	restCli, err := apigwclient.NewClient(c, "/api", conf.APIGWUserConfig)
	if err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
		config: conf,
	}, nil
}

// getHeader returns common HTTP headers for IAM API requests.
func (c *cli) getHeader(ctx contextx.IContext) http.Header {
	header := http.Header{}
	// Use non-canonical header name as required by IAM API
	header[HeaderKeyIAMVersion] = []string{HeaderValueIAMVersion}
	header.Set(restheader.BKTenantIDKey, ctx.TenantID())

	return header
}

// v2PolicyQuery performs V2 policy query API call.
func (c *cli) v2PolicyQuery(ctx contextx.IContext, req *PolicyQueryInput) (*expression.ExprCell, error) {
	resp := new(BaseBroker[*expression.ExprCell])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v2/policy/systems/%s/query/", req.System).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("v2 policy query failed: %w", err)
	}

	return resp.Data, nil
}

// v2PolicyQueryByActions performs V2 policy query by actions API call.
func (c *cli) v2PolicyQueryByActions(ctx contextx.IContext, req *PolicyQueryByActionsInput) (
	[]map[string]interface{}, error) {

	resp := new(BaseBroker[[]map[string]interface{}])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v2/policy/systems/%s/query_by_actions/", req.System).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("v2 policy query by actions failed: %w", err)
	}

	return resp.Data, nil
}

// policyQuery performs V1 policy query API call.
// nolint:unused // Reserved for future use or V1 API compatibility.
func (c *cli) policyQuery(ctx contextx.IContext, req *PolicyQueryInput) (map[string]interface{}, error) {
	resp := new(BaseBroker[map[string]interface{}])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v1/policy/query").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("policy query failed: %w", err)
	}

	return resp.Data, nil
}

// policyQueryByActions performs V1 policy query by actions API call.
// nolint:unused // Reserved for future use or V1 API compatibility.
func (c *cli) policyQueryByActions(ctx contextx.IContext, req *PolicyQueryByActionsInput) (
	[]map[string]interface{}, error) {

	resp := new(BaseBroker[[]map[string]interface{}])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v1/policy/query_by_actions").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("policy query by actions failed: %w", err)
	}

	return resp.Data, nil
}

// policyAuth performs V1 policy auth API call (with resources).
// nolint:unused // Reserved for future use.
func (c *cli) policyAuth(ctx contextx.IContext, req *PolicyAuthInput) (map[string]interface{}, error) {
	resp := new(BaseBroker[map[string]interface{}])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v1/policy/auth").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("policy auth failed: %w", err)
	}

	return resp.Data, nil
}

// v2PolicyAuth performs V2 policy auth API call (with resources).
// nolint:unused // Reserved for future use.
func (c *cli) v2PolicyAuth(ctx contextx.IContext, req *PolicyAuthInput) (map[string]interface{}, error) {
	resp := new(BaseBroker[map[string]interface{}])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v2/policy/systems/%s/auth/", req.System).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("v2 policy auth failed: %w", err)
	}

	return resp.Data, nil
}

// policyAuthByResources performs V1 policy auth by resources API call.
// nolint:unused // Reserved for future use.
func (c *cli) policyAuthByResources(ctx contextx.IContext, req *PolicyAuthByResourcesInput) (
	[]map[string]interface{}, error) {

	resp := new(BaseBroker[[]map[string]interface{}])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v1/policy/auth_by_resources").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("policy auth by resources failed: %w", err)
	}

	return resp.Data, nil
}

// policyAuthByActions performs V1 policy auth by actions API call.
// nolint:unused // Reserved for future use.
func (c *cli) policyAuthByActions(ctx contextx.IContext, req *PolicyAuthByActionsInput) (
	[]map[string]interface{}, error) {

	resp := new(BaseBroker[[]map[string]interface{}])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v1/policy/auth_by_actions").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("policy auth by actions failed: %w", err)
	}

	return resp.Data, nil
}

// getToken retrieves the system token from IAM.
func (c *cli) getToken(ctx contextx.IContext, system string) (string, error) {
	resp := new(BaseBroker[TokenResponse])
	header := c.getHeader(ctx)

	err := c.client.Get().
		SubResourcef("/v1/model/systems/%s/token", system).
		WithContext(ctx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return "", err
	}

	if err := resp.IsFailed(); err != nil {
		return "", fmt.Errorf("get token failed: %w", err)
	}

	return resp.Data.Token, nil
}

// getApplyURL retrieves the permission apply URL from IAM.
func (c *cli) getApplyURL(ctx contextx.IContext, app *Application) (string, error) {
	resp := new(BaseBroker[ApplyURLResponse])
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/v1/open/application/").
		WithContext(ctx).
		WithHeaders(header).
		Body(app).
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

// PolicyGetInput is the input for policy get API.
type PolicyGetInput struct {
	System   string `json:"system"`
	PolicyID int64  `json:"policy_id"`
}

// PolicyListInput is the input for policy list API.
type PolicyListInput struct {
	System string `json:"system"`
	Page   int    `json:"page"`
	Limit  int    `json:"page_size"`
}

// policyGet retrieves a specific policy by ID.
// nolint:unused // Reserved for future use.
func (c *cli) policyGet(ctx contextx.IContext, system string, policyID int64) (map[string]interface{}, error) {
	resp := new(BaseBroker[map[string]interface{}])
	header := c.getHeader(ctx)

	err := c.client.Get().
		SubResourcef("/v1/systems/%s/policies/%d", system, policyID).
		WithContext(ctx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("policy get failed: %w", err)
	}

	return resp.Data, nil
}

// PolicyListResponse is the response for policy list API.
type PolicyListResponse struct {
	Count   int64                    `json:"count"`
	Results []map[string]interface{} `json:"results"`
}

// policyList retrieves a list of policies.
// nolint:unused // Reserved for future use.
func (c *cli) policyList(ctx contextx.IContext, system string, page, pageSize int) (*PolicyListResponse, error) {
	resp := new(BaseBroker[PolicyListResponse])
	header := c.getHeader(ctx)

	err := c.client.Get().
		SubResourcef("/v1/systems/%s/policies/?page=%d&page_size=%d", system, page, pageSize).
		WithContext(ctx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("policy list failed: %w", err)
	}

	return &resp.Data, nil
}

// policySubjects retrieves subjects for a specific policy.
// nolint:unused // Reserved for future use.
func (c *cli) policySubjects(ctx contextx.IContext, system string, policyIDs string) (
	[]map[string]interface{}, error) {

	resp := new(BaseBroker[[]map[string]interface{}])
	header := c.getHeader(ctx)

	err := c.client.Get().
		SubResourcef("/v1/systems/%s/policies/-/subjects/?ids=%s", system, policyIDs).
		WithContext(ctx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("policy subjects failed: %w", err)
	}

	return resp.Data, nil
}
