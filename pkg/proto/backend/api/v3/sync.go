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

import "errors"

// Validate check body.
func (x *SyncCmdbHostReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *SyncCmdbHostReq) AutoConvert() {
}

// Validate check body.
func (req *SyncCmdbNetworkAreaReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (req *SyncCmdbNetworkAreaReq) AutoConvert() {
}

// Validate check body.
func (x *SyncAgentStateReq) Validate() error {
	if len(x.HostIds) == 0 {
		return errors.New("host_ids is required")
	}
	return nil
}

// AutoConvert auto convert.
func (x *SyncAgentStateReq) AutoConvert() {
}

// Validate check body.
func (x *SyncAllAgentStateReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *SyncAllAgentStateReq) AutoConvert() {
}
