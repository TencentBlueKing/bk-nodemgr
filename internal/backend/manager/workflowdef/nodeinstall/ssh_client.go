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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func generateSSHClient(
	ctx context.Context,
	operator string,
	logger logger.Logger,
	storageHostCredit credit.IStorageHostCredit,
	passwordVault creditvault.IHostPasswordVault,
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
		passwd, err := storageHostCredit.LoadHostCredit(
			ctx,
			info.Host.Static.NetworkAreaID,
			info.LoginInfo.IP,
			info.LoginInfo.User,
			types.LoginModePassword)
		if err != nil {
			return nil, fmt.Errorf("failed to load password from storageHostCredit storage: %w", err)
		}

		sshConf.AuthMethod = sshx.AuthMethodPassword
		sshConf.Password = string(passwd)

	case types.LoginModeKeyFile:
		privateKey, err := storageHostCredit.LoadHostCredit(
			ctx,
			info.Host.Static.NetworkAreaID,
			info.LoginInfo.IP,
			info.LoginInfo.User,
			types.LoginModeKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key from storageHostCredit storage: %w", err)
		}

		sshConf.AuthMethod = sshx.AuthMethodPrivateKey
		sshConf.PrivateKey = privateKey
	case types.LoginModePasswordVault:
		passwd, err := passwordVault.LoadPassword(
			ctx,
			operator,
			info.Host.Static.NetworkAreaID,
			info.LoginInfo.IP,
			info.LoginInfo.User)
		if err != nil {
			return nil, fmt.Errorf("failed to load password from password vault: %w", err)
		}

		sshConf.AuthMethod = sshx.AuthMethodPassword
		sshConf.Password = passwd
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
