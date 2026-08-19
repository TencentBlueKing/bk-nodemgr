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

// Package handler ...
package handler

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"
)

func generateWMIClient(
	ip string, port int, user string,
	credential string,
	loginMode types.LoginMode,
) (*wmix.Client, error) {

	wmiConf := &wmix.Config{
		IP:      ip,
		User:    user,
		Timeout: wmix.DefaultTimeout,
	}

	switch loginMode {
	case types.LoginModePassword, types.LoginModePasswordVault:
		wmiConf.AuthMethod = wmix.AuthMethodPassword
		wmiConf.Password = credential

	default:
		return nil, fmt.Errorf("unsupported login mode, mode(%s)", loginMode)
	}

	client, err := wmix.NewClient(wmiConf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host, host(%s): %w",
			fmt.Sprintf("%s:%d", ip, port), err)
	}

	return client, nil
}
