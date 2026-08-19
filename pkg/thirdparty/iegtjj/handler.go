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

package iegtjj

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
)

// IHandler define the interface of handler.
type IHandler interface {
	// LoadPassword load password by username and ip.
	// notice: this username is the user of the wework's rtx not the device.
	LoadPassword(nCtx contextx.IContext, rtx string, networkAreaID int64, ip string, loginUser string) (string, error)
}

// Config the config of iegtjj.
type Config struct {
	Key       string
	SecretKey string
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
	cli *cli

	conf      *Config
	decryptor *DevicePasswdDecryptor
}

// OptionFn ...
type OptionFn func(*Handler)

// New initialize a new iegtjj Handler.
func New(c *restclient.Capability, conf *Config, opts ...OptionFn) (*Handler, error) {
	cli, err := newClient(c)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		cli:       cli,
		conf:      conf,
		decryptor: NewDecryptor(conf.SecretKey),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

// LoadPassword load password.
func (h *Handler) LoadPassword(
	nCtx contextx.IContext,
	rtx string,
	networkAreaID int64,
	ip string,
	loginUser string,
) (string, error) {

	if networkAreaID != 0 {
		return "", fmt.Errorf("iegtjj unsupported this networkarea_id. networkarea_id(%d)", networkAreaID)
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

	resp, err := h.cli.getDevicePassword(nCtx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get device password: %w", err)
	}

	if len(resp.IPList) != 1 {
		return "", fmt.Errorf("unexpected iplist length from iegtjj: expected 1, got %d, ip(%s)", len(resp.IPList), ip)
	}

	item, ok := resp.IPList[ip]
	if !ok {
		for k, v := range resp.IPList {
			return "", fmt.Errorf("iegtjj returned mismatched ip key: requested %s, got %s, message: %s", ip, k, v.Message)
		}
	}
	if item.Code != 0 {
		return "", fmt.Errorf("iegtjj device password error for ip(%s): %s (code: %d)", ip, item.Message, item.Code)
	}

	if item.Password == "" {
		return "", fmt.Errorf("iegtjj returned empty password for ip(%s), message: %s", ip, item.Message)
	}

	passwd, err := h.decryptor.Decrypt(item.Password)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt device password: %w", err)
	}

	return passwd, nil
}
