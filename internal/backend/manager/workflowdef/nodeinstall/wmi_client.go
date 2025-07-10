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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"
)

func buildWMIClient(_ context.Context,
	logger logger.Logger,
	crypter crypter.Crypter,
	info *types.DeploymentInfo,
) (*wmix.Client, error) {
	wmiConf := &wmix.Config{
		IP:      info.LoginIP,
		User:    info.LoginUser,
		Timeout: wmix.DefaultTimeout,
		Logger:  logger,
	}

	switch info.LoginMode {
	case types.LoginModePassword:
		passwd, err := crypter.Decrypt(info.LoginPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password, err: %w", err)
		}

		wmiConf.AuthMethod = wmix.AuthMethodPassword
		wmiConf.Password = string(passwd)

	case types.LoginModeKeyFile:
		// todo implement
	case types.LoginModeNone:
		wmiConf.AuthMethod = wmix.AuthMethodNone
	default:
		return nil, fmt.Errorf("unsupported login mode, mode(%s)", info.LoginMode)
	}

	client, err := wmix.NewClient(wmiConf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host, host(%s), err: %w",
			fmt.Sprintf("%s:%d", info.LoginIP, info.LoginPort), err)
	}

	return client, nil
}
