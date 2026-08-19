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

package contextx

import "context"

// IContext this is the context of the nodemgr.
type IContext interface {
	context.Context
	IContextValues
}

// IContextValues defines the context values.
type IContextValues interface {
	// Values get all values.
	Values() map[string]any

	// TenantID get tenant-id from values.
	TenantID() string
	// CheckTenantID validate tenant-id.
	CheckTenantID() error

	// BKUsername get bk-username from values.
	BKUsername() string
	// CheckBKUsername validate bk-username.
	CheckBKUsername() error

	// LoginName get login-name from values.
	LoginName() string
	// CheckLoginName validate login-name.
	CheckLoginName() error

	// MessageID get message-id from values.
	MessageID() string
	// CheckMessageID validate message-id.
	CheckMessageID() error
}

// IContextInfo describes the context values.
type IContextInfo interface {
	IContextValues

	// GetValue get value by key.
	GetValue(key any) (any, bool)

	// clone a new context-info.
	// this is inner interface, do not use it outside.
	clone() IContextInfo
}
