/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bklogin

import (
	"fmt"
)

// RespCommon describe the common part of response data.
type RespCommon struct {
	Ret     int64  `json:"ret"`
	Message string `json:"msg"`
}

// BaseBroker describe the base broker.
type BaseBroker[T any] struct {
	RespCommon
	Data T `json:"data"`
}

// CodeOK define the success code.
const CodeOK = 0

// IsFailed check the response is ok.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Ret != CodeOK {
		return fmt.Errorf("resp ret no equal 0, ret(%d)", resp.Ret)
	}

	return nil
}

// GetUserInfoByBKTicketReq describe the get user info bk_ticket request.
type GetUserInfoByBKTicketReq struct {
	BKTicket string `json:"-"`
}

// GetUserInfoByBKTicketResp describe the get user info by bk_ticket response.
type GetUserInfoByBKTicketResp struct {
	Username string `json:"username"`
}

// GetUserInfoByBKTokenReq describe the get user info by bk_token request.
type GetUserInfoByBKTokenReq struct {
	BKToken string `json:"-"`
}

// GetUserInfoByBKTokenResp describe the get user info by bk_token response.
type GetUserInfoByBKTokenResp struct {
	Username string `json:"username"`
}
