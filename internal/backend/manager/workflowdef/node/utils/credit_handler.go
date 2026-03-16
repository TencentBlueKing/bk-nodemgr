/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package utils

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"
)

// NewCreditHandler creates a new CreditHandler.
func NewCreditHandler(storageHostCredit credit.IStorageHostCredit, passwordVault creditvault.IHostPasswordVault) *CreditHandler {
	return &CreditHandler{
		storageHostCredit: storageHostCredit,
		passwordVault:     passwordVault,
	}
}

// CreditHandler defines the credit handler.
type CreditHandler struct {
	storageHostCredit credit.IStorageHostCredit
	passwordVault     creditvault.IHostPasswordVault
}

// GetSSHCredit gets the ssh credit.
func (c *CreditHandler) GetSSHCredit(std *NodeActionStandarder) (sshx.AuthMethod, string, error) {
	switch std.DeployInfo().Host.Dynamic.LoginMode {
	case types.LoginModePassword:
		passwd, err := c.storageHostCredit.LoadHostCredit(
			std.Context(),
			std.DeployInfo().Host.Dynamic.LoginCreditID)
		if err != nil {
			return sshx.AuthMethodNone, "", fmt.Errorf("failed to load password from host credit storage: %w", err)
		}

		return sshx.AuthMethodPassword, string(passwd), nil

	case types.LoginModeKeyFile:
		privateKey, err := c.storageHostCredit.LoadHostCredit(
			std.Context(),
			std.DeployInfo().Host.Dynamic.LoginCreditID)
		if err != nil {
			return sshx.AuthMethodNone, "", fmt.Errorf("failed to load private key from host credit storage: %w", err)
		}

		return sshx.AuthMethodPrivateKey, string(privateKey), nil

	case types.LoginModePasswordVault:
		logger.G.Biz(std.Context()).With(
			"operator", std.Operator(),
			"network_area_id", std.DeployInfo().Host.Static.NetworkAreaID,
			"login_ip", std.DeployInfo().Host.Dynamic.LoginIP,
			"login_user", std.DeployInfo().Host.Dynamic.LoginUser,
		).Info("loading password from password vault")

		passwd, err := c.passwordVault.LoadPassword(
			std.Context(),
			std.Operator(),
			std.DeployInfo().Host.Static.NetworkAreaID,
			std.DeployInfo().Host.Dynamic.LoginIP,
			std.DeployInfo().Host.Dynamic.LoginUser)
		if err != nil {
			return sshx.AuthMethodNone, "", fmt.Errorf("failed to load password from password vault: %w", err)
		}

		if len(passwd) == 0 {
			return sshx.AuthMethodNone, "", fmt.Errorf("password vault returned empty password. ip(%s), user(%s)",
				std.DeployInfo().Host.Dynamic.LoginIP, std.DeployInfo().Host.Dynamic.LoginUser)
		}

		logger.G.Biz(std.Context()).With(
			"login_ip", std.DeployInfo().Host.Dynamic.LoginIP,
			"password_length", len(passwd),
		).Info("password vault loaded successfully")

		return sshx.AuthMethodPassword, passwd, nil

	default:
		return sshx.AuthMethodNone, "", fmt.Errorf("unsupported login mode. mode(%s)", std.DeployInfo().Host.Dynamic.LoginMode)
	}
}

// GetWMICredit gets the wmic credit.
func (c *CreditHandler) GetWMICredit(std *NodeActionStandarder) (wmix.AuthMethod, string, error) {
	switch std.DeployInfo().Host.Dynamic.LoginMode {
	case types.LoginModePassword:
		passwd, err := c.storageHostCredit.LoadHostCredit(
			std.Context(),
			std.DeployInfo().Host.Dynamic.LoginCreditID)
		if err != nil {
			return wmix.AuthMethodNone, "", fmt.Errorf("failed to load password from host credit storage: %w", err)
		}

		return wmix.AuthMethodPassword, string(passwd), nil

	case types.LoginModePasswordVault:
		logger.G.Biz(std.Context()).With(
			"operator", std.Operator(),
			"network_area_id", std.DeployInfo().Host.Static.NetworkAreaID,
			"login_ip", std.DeployInfo().Host.Dynamic.LoginIP,
			"login_user", std.DeployInfo().Host.Dynamic.LoginUser,
		).Info("loading password from password vault (wmi)")

		passwd, err := c.passwordVault.LoadPassword(
			std.Context(),
			std.Operator(),
			std.DeployInfo().Host.Static.NetworkAreaID,
			std.DeployInfo().Host.Dynamic.LoginIP,
			std.DeployInfo().Host.Dynamic.LoginUser)
		if err != nil {
			return wmix.AuthMethodNone, "", fmt.Errorf("failed to load password from password vault: %w", err)
		}

		if len(passwd) == 0 {
			return wmix.AuthMethodNone, "", fmt.Errorf("password vault returned empty password. ip(%s), user(%s)",
				std.DeployInfo().Host.Dynamic.LoginIP, std.DeployInfo().Host.Dynamic.LoginUser)
		}

		logger.G.Biz(std.Context()).With(
			"login_ip", std.DeployInfo().Host.Dynamic.LoginIP,
			"password_length", len(passwd),
		).Info("password vault loaded successfully (wmi)")

		return wmix.AuthMethodPassword, passwd, nil

	default:
		return wmix.AuthMethodNone, "", fmt.Errorf("unsupported login mode. mode(%s)", std.DeployInfo().Host.Dynamic.LoginMode)
	}
}
