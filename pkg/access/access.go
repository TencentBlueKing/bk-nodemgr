/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package access ...
package access

import (
	"sync"
)

const (
	defaultVirtualUser = "bk-nodemgr"
)

var virtualUser = struct {
	sync.Once
	user string
}{
	user: defaultVirtualUser,
}

// GetVirtualUser gets the system user.
// system user is used to execute the system
func GetVirtualUser() string {
	virtualUser.Once.Do(func() {
		virtualUser.user = defaultVirtualUser
	})

	return virtualUser.user
}

// SetVirtualUser sets the system user.
// system user is used to execute the system
func SetVirtualUser(user string) {
	virtualUser.Once.Do(func() {
		virtualUser.user = user
	})
}
