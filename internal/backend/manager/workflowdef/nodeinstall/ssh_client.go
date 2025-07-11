/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func buildSSHClient(ctx context.Context,
	logger logger.Logger,
	crypter crypter.Crypter,
	info *types.DeploymentInfo,
) (*sshx.Client, error) {

	sshConf := &sshx.Config{
		Network: sshx.NetworkTCP,
		IP:      info.LoginInfo.IP,
		Port:    int(info.LoginInfo.Port),
		User:    info.LoginInfo.User,
		Logger:  logger,
	}

	switch info.LoginInfo.Mode {
	case types.LoginModePassword:
		passwd, err := crypter.Decrypt(info.LoginInfo.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password, err: %w", err)
		}

		sshConf.AuthMethod = sshx.AuthMethodPassword
		sshConf.Password = string(passwd)

	case types.LoginModeKeyFile:
		privateKey, err := crypter.Decrypt(info.LoginInfo.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt private key, err: %w", err)
		}

		sshConf.AuthMethod = sshx.AuthMethodPrivateKey
		sshConf.PrivateKey = privateKey
	case types.LoginModeNone:
		sshConf.AuthMethod = sshx.AuthMethodNone
	default:
		return nil, fmt.Errorf("unsupported login mode, mode(%s)", info.LoginInfo.Mode)
	}

	client, err := sshx.NewClient(ctx, sshConf, sshx.DefaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host, host(%s), err: %w",
			fmt.Sprintf("%s:%d", info.LoginInfo.IP, info.LoginInfo.Port), err)
	}

	return client, nil
}
