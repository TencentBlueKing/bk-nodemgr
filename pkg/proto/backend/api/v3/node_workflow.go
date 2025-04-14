/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"errors"
)

// Validate check body.
func (x *NodeAgentInstallReq) Validate() error {
	if x.InnerIp == "" && x.InnerIpv6 == "" {
		return errors.New("inner_ip and inner_ipv6 can not be empty at the same time")
	}

	if x.GetBizId() < 0 {
		return errors.New("biz_id must be equal or greater than 0")
	}

	if x.OsType == "" {
		return errors.New("os_type can not be empty")
	}

	//if err := types.Addressing(x.GetAddressing()).Validate(); err != nil {
	//	return err
	//}

	if x.GetNetworkUnitId() < 0 {
		return errors.New("network_unit_id must be greater than or equal to 0")
	}

	if x.TargetVersion == "" {
		return errors.New("target_version can not be empty")
	}

	if x.GetTargetGeneration() <= 0 || x.GetTargetGeneration() > 2 {
		return errors.New("target_generation must be 1 or 2")
	}

	if x.LoginIp == "" {
		return errors.New("login_ip can not be empty")
	}

	if x.GetLoginPort() < 0 {
		return errors.New("login_port must be greater than 0")
	}

	if len(x.GetLoginUser()) == 0 {
		return errors.New("login_user can not be empty")
	}

	switch x.GetLoginMode() {
	case NodeAgentInstallReq_LOGIN_MODE_PASSWORD:
		if len(x.LoginPassword) == 0 {
			return errors.New("login_password can not be empty")
		}
	case NodeAgentInstallReq_LOGIN_MODE_KEY:
		if len(x.LoginKeyFile) == 0 {
			return errors.New("login_key_file can not be empty")
		}
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentInstallReq) AutoConvert() {
	if x.NetworkUnitId == nil {
		x.NetworkUnitId = new(int64)
		*x.NetworkUnitId = -1
	}

	if x.BizId == nil {
		x.BizId = new(int64)
		*x.BizId = -1
	}

	if x.LoginPort == nil {
		x.LoginPort = new(int64)
		*x.LoginPort = -1
	}

	if x.TargetGeneration == nil {
		x.TargetGeneration = new(int64)
		*x.TargetGeneration = -1
	}
}
