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

// Package bklogin provides handlers to operate bklogin API.
package bklogin

import (
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
	bksaasheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bksaas/header"
)

// cli client for bkoa.
type cli struct {
	client restclient.IClient
	config *Config
}

// newClient initialize a new bkoa client.
func newClient(capability *restclient.Capability, conf *Config) (*cli, error) {
	restCli, err := restclient.NewClient(capability, "/", restclient.WithURLQueryMasker(CookieKeyBKTicket, CookieKeyBKToken))
	if err != nil {
		return nil, err
	}

	client := &cli{
		client: restCli,
		config: conf,
	}
	if tenant.GetMode() != tenant.ModeMultiple || conf.AuthType != CookieKeyBKToken {
		return client, nil
	}

	client.client, err = apigwclient.NewClient(
		capability,
		"/login/api/v3",
		conf.VirtualUserConfig,
		restclient.WithURLQueryMasker(CookieKeyBKTicket, CookieKeyBKToken),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// getCommonHeader get cmdb common header.
// nolint: unparam
func (c *cli) getCommonHeader() (http.Header, error) {
	header := http.Header{}
	header.Set(bksaasheader.KeyBKRequestID, identifier.GenRequestID())

	return header, nil
}

// getUserInfoByBKTicket get user info by bk_ticket.
func (c *cli) getUserInfoByBKTicket(
	nCtx contextx.IContext, req *GetUserInfoByBKTicketReq) (*GetUserInfoByBKTicketResp, error) {

	resp := new(BKTicketBroker[*GetUserInfoByBKTicketResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Get().
		// add '/' after subpath to prevent redirection during requests.
		SubResourcef("/user/get_info/").
		WithContext(nCtx).
		WithHeaders(header).
		WithParam(CookieKeyBKTicket, req.BKTicket).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("failed to get user info by bk_ticket: %w", err)
	}

	return resp.Data, nil
}

func (c *cli) getUserInfoByBKTokenSingleTenantMode(
	nCtx contextx.IContext,
	req *GetUserInfoByBKTokenSingleTenantModeReq,
) (*GetUserInfoByBKTokenSingleTenantModeResp, error) {

	resp := new(BKTokenSingleTenantModeBroker[*GetUserInfoByBKTokenSingleTenantModeResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Get().
		// add '/' after subpath to prevent redirection during requests.
		SubResourcef("/accounts/get_user/").
		WithContext(nCtx).
		WithHeaders(header).
		WithParam(CookieKeyBKToken, req.BKToken).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("failed to get user info by bk_token: %w", err)
	}

	return resp.Data, nil
}

func (c *cli) getUserInfoByBKTokenMultipleTenantMode(
	nCtx contextx.IContext,
	req *GetUserInfoByBKTokenMultipleTenantModeReq,
) (*GetUserInfoByBKTokenMultipleTenantModeResp, error) {

	// This is a special logic: the API gateway is required to take a tenant ID, but this API does not differentiate tenants.
	apigwHeader := http.Header{}
	apigwHeader.Set(apigwheader.BKGWTenantIDKey, tenant.SystemTenantID)
	apigwHeader.Set(apigwheader.BKGWAuthKey, c.config.VirtualUserConfig.GetAuthHeader(nCtx))

	resp := new(BKTokenMultipleTenantModeBroker)
	err := c.client.Get().
		SubResourcef("/open/bk-tokens/userinfo/").
		WithContext(nCtx).
		WithHeaders(apigwHeader).
		WithParam(CookieKeyBKToken, req.BKToken).
		Body(req).EnableLogBody().EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("failed to get user info by bk_token: %w", err)
	}

	return resp.Data, nil
}
