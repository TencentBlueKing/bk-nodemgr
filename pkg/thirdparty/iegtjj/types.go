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

package iegtjj

// BaseBroker describe the common part of broker data.
type BaseBroker[T any] struct {
	Result[T] `json:"Result"`
}

// Result define the iegtjj api response common part.
type Result[T any] struct {
	RespCommon
	Data T `json:"ResponseItems"`
}

// RespCommon describe the common part of response data.
type RespCommon struct {
	RequestID string `json:"RequestId"`
	HasError  bool   `json:"HasError"`
	Message   string `json:"message"`
}

// GetDevicePasswordReq describe the request of LoadPassword.
type GetDevicePasswordReq struct {
	Key      string   `json:"Key"`
	Username string   `json:"Username"`
	IPList   []string `json:"IpList"`
}

// GetDevicePasswordResp describe the response of LoadPassword.
type GetDevicePasswordResp struct {
	IPList map[string]IPListItem `json:"IpList"`
}

// IPListItem describe the ip item.
type IPListItem struct {
	Code     int    `json:"Code"`
	Message  string `json:"Message"`
	Password string `json:"Password"`
}
