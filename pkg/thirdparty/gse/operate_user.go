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

package gse

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// OperateUser define the user that the GSE channel uses to operate plugin
// processes and files on the host.
type OperateUser string

const (
	// OperateUserUnixRoot this defines the operate user of unix.
	OperateUserUnixRoot OperateUser = "root"

	// OperateUserWindowsSystem this defines the operate user of windows. Operating as
	// a named user on windows requires its password, so the system user is used instead.
	OperateUserWindowsSystem OperateUser = "system"
)

// DefaultOperateUser returns the default user that the GSE channel uses to operate plugin
// processes and files on the host of the os type.
// The host login user is only the SSH authentication identity at install time and is never
// used for GSE channel operations.
func DefaultOperateUser(osType criteria.OSType) (OperateUser, error) {
	if err := osType.Validate(); err != nil {
		return "", err
	}

	switch osType {
	case criteria.OSWindows:
		return OperateUserWindowsSystem, nil
	case criteria.OSAix, criteria.OSAix6, criteria.OSAix7, criteria.OSAndroid, criteria.OSDarwin,
		criteria.OSDragonfly, criteria.OSFreebsd, criteria.OSHurd, criteria.OSIllumos, criteria.OSIos,
		criteria.OSLinux, criteria.OSNetbsd, criteria.OSOpenbsd, criteria.OSSolaris, criteria.OSZos:
		return OperateUserUnixRoot, nil
	default:
		return "", fmt.Errorf("unsupported os type for default operate user: %s", osType)
	}
}

// Validate checks if the operate user is valid.
func (user OperateUser) Validate() error {
	switch user {
	case OperateUserUnixRoot, OperateUserWindowsSystem:
		return nil
	default:
		return fmt.Errorf("invalid operate user: %s", user)
	}
}

// String converts the OperateUser to its string representation.
func (user OperateUser) String() string {
	return string(user)
}
