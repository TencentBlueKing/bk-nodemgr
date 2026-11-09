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

// nolint: revive
package utils

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// EnsureHostLoginUser ensures host login user is set for plugin action runtime.
// The login user is only the SSH authentication identity at install time;
// GSE channel process/file operations must use the host operate user instead,
// see EnsureHostOperateUser.
func EnsureHostLoginUser(std *PluginActionStandarder, host *types.Host) error {
	if host.Dynamic.LoginUser != "" {
		return nil
	}

	defaultUser, err := criteria.DefaultAdminUser(host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get default admin user, host-id(%d), os-type(%s): %w",
			host.HostID, host.Dynamic.NodeOsType, err)
	}

	host.Dynamic.LoginUser = string(defaultUser)
	std.InstanceData().Log().
		Zh("主机(%d)登录用户为空, 已根据操作系统类型(%s)回退为默认系统管理员账户(%s)",
			host.HostID, host.Dynamic.NodeOsType, host.Dynamic.LoginUser).
		En("host(%d) login user is empty, fallback to default system administrator(%s) by os type(%s).",
			host.HostID, host.Dynamic.LoginUser, host.Dynamic.NodeOsType).
		Info()

	return nil
}

// EnsureHostOperateUser ensures host operate user is set for plugin action runtime.
// The operate user is the identity that the GSE channel uses to operate plugin processes
// and files on the host (root on unix, system on windows); it is written with a default
// value at host sync time, and falls back to the default here for legacy hosts.
func EnsureHostOperateUser(std *PluginActionStandarder, host *types.Host) error {
	if host.Dynamic.OperateUser != "" {
		return nil
	}

	defaultUser, err := gse.DefaultOperateUser(host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get default operate user, host-id(%d), os-type(%s): %w",
			host.HostID, host.Dynamic.NodeOsType, err)
	}

	host.Dynamic.OperateUser = defaultUser.String()
	std.InstanceData().Log().
		Zh("主机(%d)操作用户为空, 已根据操作系统类型(%s)回退为默认操作用户(%s)",
			host.HostID, host.Dynamic.NodeOsType, host.Dynamic.OperateUser).
		En("host(%d) operate user is empty, fallback to default operate user(%s) by os type(%s).",
			host.HostID, host.Dynamic.OperateUser, host.Dynamic.NodeOsType).
		Info()

	return nil
}
