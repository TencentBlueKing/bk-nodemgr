/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package v3 ...
package v3

import "errors"

// Validate check request body.
func (x *GetAgentConfReq) Validate() error {
	if x.Token == "" {
		return errors.New("token is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *GetAgentConfReq) AutoConvert() {
}

// Validate check request body.
func (x *GetDataProxyConfReq) Validate() error {
	if x.Token == "" {
		return errors.New("token is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *GetDataProxyConfReq) AutoConvert() {
}

// Validate check request body.
func (x *GetFileProxyConfReq) Validate() error {
	if x.Token == "" {
		return errors.New("token is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *GetFileProxyConfReq) AutoConvert() {
}

// Validate check request body.
func (x *GetCheckListReq) Validate() error {
	if x.Token == "" {
		return errors.New("token is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *GetCheckListReq) AutoConvert() {
}
