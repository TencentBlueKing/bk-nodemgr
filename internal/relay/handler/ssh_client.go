/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package handler ...
package handler

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func generateSSHClient(
	ctx context.Context,
	ip string, port int, user string,
	loginMode types.LoginMode, hostCredit []byte, passwordVault string,
	logger logger.Logger,
) (*sshx.Client, error) {

	sshConf := &sshx.Config{
		Network: sshx.NetworkTCP,
		IP:      ip,
		Port:    port,
		User:    user,
		Logger:  logger,
	}

	switch loginMode {
	case types.LoginModePassword:
		sshConf.AuthMethod = sshx.AuthMethodPassword
		sshConf.Password = string(hostCredit)

	case types.LoginModeKeyFile:
		sshConf.AuthMethod = sshx.AuthMethodPrivateKey
		sshConf.PrivateKey = hostCredit

	case types.LoginModePasswordVault:
		sshConf.AuthMethod = sshx.AuthMethodPassword
		sshConf.Password = passwordVault
	default:
		return nil, fmt.Errorf("unsupported login mode, mode(%s)", loginMode)
	}

	client, err := sshx.NewClient(ctx, sshConf, sshx.DefaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host, host(%s), err: %w",
			fmt.Sprintf("%s:%d", ip, port), err)
	}

	return client, nil
}
