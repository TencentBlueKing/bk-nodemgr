/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bklogin

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
)

// IHandler defines the handler interface.
type IHandler interface {
	// GetLoginURL get the login url.
	GetLoginURL() string

	// Verify verify the bk_ticket or bk_token.
	Verify(nCtx contextx.IContext, token string) (string, string, string, error)

	// GetAuthIdentity get the auth identity.
	GetAuthIdentity() *AuthIdentity

	// GetAuthType returns the auth type (cookie key name: bk_token or bk_ticket).
	GetAuthType() string
}


// Handler the Handler of cmdb.
type Handler struct {
	cli *cli

	conf *Config
}

// GetAuthType returns the auth type.
func (h *Handler) GetAuthType() string {
	return h.conf.AuthType
}

// Config the config of bkoa.
type Config struct {
	LoginURL string
	AuthType string
}

// Validate validates the config.
func (conf *Config) Validate() error {
	_, err := url.Parse(conf.LoginURL)
	if err != nil {
		return fmt.Errorf("failed to validate bklogin config: %w", err)
	}

	switch conf.AuthType {
	case CookieKeyBKTicket, CookieKeyBKToken:
	default:
		return fmt.Errorf("failed to validate bklogin config: unsupported auth type: %s", conf.AuthType)
	}

	return nil
}

// OptionFn ...
type OptionFn func(*Handler)

// New initialize a new cmdb Handler.
func New(c *restclient.Capability, conf *Config, opts ...OptionFn) (IHandler, error) {
	cli, err := newClient(c)
	if err != nil {
		return nil, err
	}

	err = conf.Validate()
	if err != nil {
		return nil, err
	}

	handler := &Handler{
		cli:  cli,
		conf: conf,
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler, nil
}

// Verify the bk_ticket or bk_token is valid or not, and return the bk_username.
// nolint: nonamedreturns
func (h *Handler) Verify(nCtx contextx.IContext, token string) (tenantID string, bkUsername string, loginName string, err error) {
	if nCtx == nil {
		return "", "", "", errors.New("failed to verify token: invalid context")
	}

	switch h.conf.AuthType {
	case CookieKeyBKTicket:
		resp, err := h.cli.getUserInfoByBKTicket(nCtx, &GetUserInfoByBKTicketReq{BKTicket: token})
		if err != nil {
			return "", "", "", fmt.Errorf("failed to verify bk_ticket: %w", err)
		}

		return "default", resp.Username, resp.Username, nil
	case CookieKeyBKToken:
		resp, err := h.cli.getUserInfoByBKToken(nCtx, &GetUserInfoByBKTokenReq{BKToken: token})
		if err != nil {
			return "", "", "", fmt.Errorf("failed to verify bk_token: %w", err)
		}

		return "default", resp.Username, resp.Username, nil
	default:
		return "", "", "", fmt.Errorf("failed to verify token: unsupported auth type: %s", h.conf.AuthType)
	}
}

// GetAuthIdentity ...
func (h *Handler) GetAuthIdentity() *AuthIdentity {
	return &AuthIdentity{
		handler: h,
	}
}

// GetLoginURL ...
func (h *Handler) GetLoginURL() string {
	return h.conf.LoginURL
}
