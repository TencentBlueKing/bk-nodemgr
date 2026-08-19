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

package criteria

import "fmt"

// AdminUser define the default system administrator user.
type AdminUser string

const (
	// AdminUserUnixRoot this defines the default administrator user of unix.
	AdminUserUnixRoot AdminUser = "root"

	// AdminUserWindowsAdministrator this defines the default administrator user of windows.
	AdminUserWindowsAdministrator AdminUser = "Administrator"
)

// DefaultAdminUser returns the default system administrator user of the os type.
func DefaultAdminUser(osType OSType) (AdminUser, error) {
	if err := osType.Validate(); err != nil {
		return "", err
	}

	switch osType {
	case OSWindows:
		return AdminUserWindowsAdministrator, nil
	case OSAix, OSAix6, OSAix7, OSAndroid, OSDarwin, OSDragonfly, OSFreebsd, OSHurd,
		OSIllumos, OSIos, OSLinux, OSNetbsd, OSOpenbsd, OSSolaris, OSZos:
		return AdminUserUnixRoot, nil
	default:
		return "", fmt.Errorf("unsupported os type for default admin user: %s", osType)
	}
}

// Validate checks if the admin user is valid.
func (user AdminUser) Validate() error {
	switch user {
	case AdminUserUnixRoot, AdminUserWindowsAdministrator:
		return nil
	default:
		return fmt.Errorf("invalid admin user: %s", user)
	}
}

// String converts the AdminUser to its string representation.
func (user AdminUser) String() string {
	return string(user)
}
