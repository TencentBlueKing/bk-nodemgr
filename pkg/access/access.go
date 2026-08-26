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

// Package access ...
package access

import (
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

const (
	defaultVirtualUser = "bk-nodemgr"
)

// nolint: gochecknoglobals // virtual user is configured once during service initialization.
var virtualUser = struct {
	once sync.Once
	user string
}{
	user: defaultVirtualUser,
}

// GetVirtualUserBKUsername gets the tenant-scoped bk username for the system virtual user.
func GetVirtualUserBKUsername(ctx contextx.IContext) (string, error) {
	return GetBKUsernameByLoginName(ctx, GetVirtualUser())
}

// GetVirtualUser gets the system user.
// System user is used to execute the system.
func GetVirtualUser() string {
	virtualUser.once.Do(func() {
		virtualUser.user = defaultVirtualUser
	})

	return virtualUser.user
}

// SetVirtualUser sets the system user.
// System user is used to execute the system.
func SetVirtualUser(user string) {
	virtualUser.once.Do(func() {
		virtualUser.user = user
	})
}
