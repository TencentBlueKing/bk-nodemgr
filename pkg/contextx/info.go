/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package contextx

import "errors"

var _ IContextInfo = &Info{}

// Info defines the context info of the nodemgr.
type Info struct {
	values map[string]any

	tenantID   string
	bkUsername string
	loginName  string
	messageID  string
}

const (
	attributeKeyTenantID   = "tenant_id"
	attributeKeyBKUsername = "bk_username"
	attributeKeyLoginName  = "login_name"
	attributeKeyMessageID  = "message_id"
)

// GetValue get value from values.
func (info *Info) GetValue(key any) (any, bool) {
	strKey, ok := key.(string)
	if ok {
		if val, exists := info.values[strKey]; exists {
			return val, true
		}
	}

	return nil, false
}

// Values get all values.
func (info *Info) Values() map[string]any {
	return info.values
}

// TenantID get tenant-id from values.
func (info *Info) TenantID() string {
	return info.tenantID
}

// CheckTenantID check tenant-id.
func (info *Info) CheckTenantID() error {
	if info.tenantID == "" {
		return errors.New("tenant-id not found")
	}

	return nil
}

// BKUsername get bk-username from values.
func (info *Info) BKUsername() string {
	return info.bkUsername
}

// CheckBKUsername valcheckidate bk-username.
func (info *Info) CheckBKUsername() error {
	if info.bkUsername == "" {
		return errors.New("bk-username not found")
	}

	return nil
}

// LoginName get login-name from values.
func (info *Info) LoginName() string {
	return info.loginName
}

// CheckLoginName valcheckidate login-name.
func (info *Info) CheckLoginName() error {
	if info.loginName == "" {
		return errors.New("login-name not found")
	}

	return nil
}

// MessageID get message-id from values.
func (info *Info) MessageID() string {
	return info.messageID
}

// CheckMessageID check message-id.
func (info *Info) CheckMessageID() error {
	if info.messageID == "" {
		return errors.New("message-id not found")
	}

	return nil
}

// clone a context.
func (info *Info) clone() IContextInfo {
	valuesCopy := make(map[string]any, len(info.values))
	for k, v := range info.values {
		valuesCopy[k] = v
	}

	return &Info{
		values:     valuesCopy,
		tenantID:   info.tenantID,
		bkUsername: info.bkUsername,
		loginName:  info.loginName,
		messageID:  info.messageID,
	}
}
