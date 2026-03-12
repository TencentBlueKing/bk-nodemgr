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

// BKTicketBroker is the response envelope for bk_ticket-based APIs (e.g. /user/get_info/).
// Success is indicated by ret=0; error message is in msg.
type BKTicketBroker[T any] struct {
	Ret     int64  `json:"ret"`
	Message string `json:"msg"`
	Data    T      `json:"data"`
}

const bkTicketSuccess = 0

// IsFailed check the response is ok.
func (resp *BKTicketBroker[T]) IsFailed() error {
	if resp.Ret != bkTicketSuccess {
		return fmt.Errorf("bklogin returned failure: ret=%d, message=%s", resp.Ret, resp.Message)
	}

	return nil
}

// BKTokenBroker is the response envelope for bk_token-based APIs (e.g. /accounts/get_user/).
// Success is indicated by result=true and code=0; error detail is in code and message.
type BKTokenBroker[T any] struct {
	Result  bool   `json:"result"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

const bkTokenSuccess = "00"

// IsFailed check the response is ok.
func (resp *BKTokenBroker[T]) IsFailed() error {
	if resp.Result && resp.Code == bkTokenSuccess {
		return nil
	}

	return fmt.Errorf("bklogin returned failure: code=%s, message=%s", resp.Code, resp.Message)
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

// WebUserInfo is the unified web user info for the web login scenario.
type WebUserInfo struct {
	Username string
}
