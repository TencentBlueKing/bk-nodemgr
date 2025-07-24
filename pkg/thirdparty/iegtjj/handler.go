/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package iegtjj

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

// IHandler define the interface of handler.
type IHandler interface {
	// LoadPassword load password by username and ip.
	// notice: this username is the user of the wework's rtx not the device.
	LoadPassword(ctx context.Context, rtx string, networkAreaID int64, ip string, loginUser string) (string, error)
}

// Config the config of iegtjj.
type Config struct {
	HeaderSetter HeaderSetter
	Key          string
	SecretKey    string
}

// Validate validate config.
func (conf Config) Validate() error {
	if conf.Key == "" {
		return errors.New("iegtjj config key is empty")
	}

	if conf.SecretKey == "" {
		return errors.New("iegtjj config secret_key is empty")
	}

	return nil
}

// Handler get vault and cli.
type Handler struct {
	cli       *cli
	logger    logger.Logger
	conf      *Config
	decryptor *DevicePasswdDecryptor
}

// OptionFn ...
type OptionFn func(*Handler)

// WithLogger this func will set the logger of the Handler.
func WithLogger(logger logger.Logger) OptionFn {
	return func(s *Handler) {
		s.logger = logger
	}
}

// New initialize a new iegtjj Handler.
func New(c *client.Capability, conf *Config, opts ...OptionFn) (*Handler, error) {
	cli, err := newClient(c, conf.HeaderSetter)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		cli:       cli,
		conf:      conf,
		decryptor: NewDecryptor(conf.SecretKey),
		logger:    logger.LoggerDefault{},
	}

	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

// LoadPassword load password.
func (h *Handler) LoadPassword(
	ctx context.Context,
	rtx string,
	networkAreaID int64,
	ip string,
	loginUser string,
) (string, error) {

	if networkAreaID != 0 {
		return "", fmt.Errorf("iegtjj unsupported this network_area_id. network_area_id(%d)", networkAreaID)
	}

	switch loginUser {
	case "root", "administrator", "Administrator":
	default:
		return "", fmt.Errorf("iegtjj unsupported this login user. login-user(%s)", loginUser)
	}

	req := &GetDevicePasswordReq{
		Key:      h.conf.Key,
		Username: rtx,
		IPList:   []string{ip},
	}

	resp, err := h.cli.getDevicePassword(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get device password: %w", err)
	}

	if len(resp.IPList) > 1 || len(resp.IPList) == 0 {
		return "", fmt.Errorf("iplist length is not 1, ip(%v), resp(%v)", ip, resp)
	}

	if resp.IPList[ip].Password == "" {
		return "", fmt.Errorf("failed to get device password, message(%v)", resp.IPList[ip].Message)
	}

	passwd, err := h.decryptor.Decrypt(resp.IPList[ip].Password)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt device password: %w", err)
	}

	return passwd, nil
}
