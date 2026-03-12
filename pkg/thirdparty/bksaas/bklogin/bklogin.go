/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package bklogin provides handlers to operate bklogin API.
package bklogin

import (
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	bksaasheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bksaas/header"
)

// cli client for bkoa.
type cli struct {
	client restclient.IClient
}

// newClient initialize a new bkoa client.
func newClient(c *restclient.Capability) (*cli, error) {
	restCli, err := restclient.NewClient(c, "/")
	if err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
	}, nil
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

// getUserInfoByBKToken get user info by bk_token.
func (c *cli) getUserInfoByBKToken(
	nCtx contextx.IContext, req *GetUserInfoByBKTokenReq) (*GetUserInfoByBKTokenResp, error) {

	resp := new(BKTokenBroker[*GetUserInfoByBKTokenResp])
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
