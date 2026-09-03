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

package bklogin

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler defines the handler interface.
type IHandler interface {
	// GetLoginURL get the login url.
	GetLoginURL() string

	// GetAppCode get the BlueKing app code, used to build the bklogin redirect url.
	GetAppCode() string

	// Verify verify the bk_ticket or bk_token.
	Verify(nCtx contextx.IContext, token string) (string, string, string, error)

	// GetAuthIdentity get the auth identity.
	GetAuthIdentity() *AuthIdentity

	// GetAuthType returns the auth type (cookie key name: bk_token or bk_ticket).
	GetAuthType() string

	// GetWebUserInfo returns the unified web user info for the web login scenario.
	GetWebUserInfo(nCtx contextx.IContext, token string) (*types.WebUserInfo, error)
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
	// AppCode is the BlueKing app code, used to build the bklogin redirect url.
	AppCode           string
	AuthType          string
	VirtualUserConfig apigwclient.VirtualUserConfig
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

	if tenant.GetMode() == tenant.ModeMultiple && conf.AuthType == CookieKeyBKToken {
		if err := conf.VirtualUserConfig.Validate(); err != nil {
			return fmt.Errorf("failed to validate bklogin api-gateway config: %w", err)
		}
	}

	return nil
}

// OptionFn ...
type OptionFn func(*Handler)

// New initialize a new cmdb Handler.
func New(c *restclient.Capability, conf *Config, opts ...OptionFn) (IHandler, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	cli, err := newClient(c, conf)
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

		return tenant.SingleModeTenantID, resp.Username, resp.Username, nil
	case CookieKeyBKToken:
		tenantMode := tenant.GetMode()
		switch tenantMode {
		case tenant.ModeSingle:
			req := &GetUserInfoByBKTokenSingleTenantModeReq{BKToken: token}
			resp, err := h.cli.getUserInfoByBKTokenSingleTenantMode(nCtx, req)
			if err != nil {
				return "", "", "", fmt.Errorf("failed to verify bk_token: %w", err)
			}

			return tenant.SingleModeTenantID, resp.Username, resp.Username, nil
		case tenant.ModeMultiple:
			req := &GetUserInfoByBKTokenMultipleTenantModeReq{BKToken: token}
			resp, err := h.cli.getUserInfoByBKTokenMultipleTenantMode(nCtx, req)
			if err != nil {
				return "", "", "", fmt.Errorf("failed to verify bk_token: %w", err)
			}

			return resp.TenantID, resp.BKUsername, resp.LoginName, nil
		default:
			return "", "", "", fmt.Errorf("failed to verify token: unsupported tenant mode: %s", tenantMode)
		}
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

// GetAppCode returns the BlueKing app code.
func (h *Handler) GetAppCode() string {
	return h.conf.AppCode
}

// GetWebUserInfo returns the unified web user info for the web login scenario.
// It abstracts the bk_ticket/bk_token branch and guarantees username is non-empty on success.
func (h *Handler) GetWebUserInfo(nCtx contextx.IContext, token string) (*types.WebUserInfo, error) {
	if nCtx == nil {
		return nil, errors.New("failed to get web user info: invalid context")
	}

	info := new(types.WebUserInfo)
	switch h.conf.AuthType {
	case CookieKeyBKTicket:
		resp, err := h.cli.getUserInfoByBKTicket(nCtx, &GetUserInfoByBKTicketReq{BKTicket: token})
		if err != nil {
			return nil, fmt.Errorf("failed to get web user info by bk_ticket: %w", err)
		}
		info.BKUsername = resp.Username
		info.LoginName = resp.Username
		tenantMode := tenant.GetMode()
		switch tenantMode {
		case tenant.ModeSingle:
			info.TenantID = tenant.SingleModeTenantID
		case tenant.ModeMultiple:
			info.TenantID = tenant.SystemTenantID
		default:
			return nil, fmt.Errorf("failed to get web user info: unsupported tenant mode: %s", tenantMode)
		}
	case CookieKeyBKToken:
		tenantMode := tenant.GetMode()
		switch tenantMode {
		case tenant.ModeSingle:
			req := &GetUserInfoByBKTokenSingleTenantModeReq{BKToken: token}
			resp, err := h.cli.getUserInfoByBKTokenSingleTenantMode(nCtx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to get web user info by bk_token: %w", err)
			}
			info.BKUsername = resp.Username
			info.LoginName = resp.Username
			info.TenantID = tenant.SingleModeTenantID
		case tenant.ModeMultiple:
			req := &GetUserInfoByBKTokenMultipleTenantModeReq{BKToken: token}
			resp, err := h.cli.getUserInfoByBKTokenMultipleTenantMode(nCtx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to get web user info by bk_token: %w", err)
			}
			info.BKUsername = resp.BKUsername
			info.LoginName = resp.LoginName
			info.TimeZone = resp.TimeZone
			info.TenantID = resp.TenantID
		default:
			return nil, fmt.Errorf("failed to get web user info: unsupported tenant mode: %s", tenantMode)
		}

	default:
		return nil, fmt.Errorf("failed to get web user info: unsupported auth type: %s", h.conf.AuthType)
	}

	if info.LoginName == "" {
		return nil, errors.New("failed to get web user info: username is empty")
	}

	return info, nil
}
