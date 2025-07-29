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
	"context"
	"errors"
	"fmt"
	"net/url"

	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

// IHandler defines the handler interface
type IHandler interface {
	// Verify verify the bk_ticket.
	Verify(ctx context.Context, bkTicket string) (string, error)

	// GetAuthIdentity get the auth identity.
	GetAuthIdentity() *AuthIdentity
}

// Handler the Handler of cmdb.
type Handler struct {
	cli    *cli
	logger logger.Logger
	conf   *Config
}

// Config the config of bkoa.
type Config struct {
	LoginURL string
}

// Validate validates the config.
func (conf *Config) Validate() error {
	_, err := url.Parse(conf.LoginURL)
	if err != nil {
		return fmt.Errorf("failed to validate bklogin config: %w", err)
	}

	return nil
}

// OptionFn ...
type OptionFn func(*Handler)

// WithLogger this func will set the logger of the Handler.
func WithLogger(logger logger.Logger) OptionFn {
	return func(s *Handler) {
		s.logger = logger
	}
}

// New initialize a new cmdb Handler.
func New(c *restclient.Capability, conf *Config, opts ...OptionFn) (IHandler, error) {
	cli, err := newClient(c)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		cli:    cli,
		logger: logger.LoggerDefault{},
		conf:   conf,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

// Verify the bk_ticket is valid or not, and return the bk_username.
func (h *Handler) Verify(ctx context.Context, bkTicket string) (string, error) {
	if ctx == nil {
		return "", errors.New("failed to verify bk_ticket: invalid context")
	}

	if bkTicket == "" {
		return "", errors.New("failed to verify bk_ticket: invalid param")
	}

	resp, err := h.cli.getUserInfo(ctx, &GetUserInfoReq{BKTicket: bkTicket})
	if err != nil {
		return "", fmt.Errorf("failed to verify bk_ticket: %w", err)
	}

	return resp.Username, nil
}

// GetAuthIdentity ...
func (h *Handler) GetAuthIdentity() *AuthIdentity {
	return &AuthIdentity{
		handler: h,
	}
}
