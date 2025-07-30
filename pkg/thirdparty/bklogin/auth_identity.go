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

	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

const (
	// CookieKeyBKTicket is the key of cookie.
	CookieKeyBKTicket = "bk_ticket"
)

var _ restserver.AuthIdentity = &AuthIdentity{}

// AuthIdentity verify the ticket.
type AuthIdentity struct {
	handler *Handler
}

// Verify the ticket.
func (identity *AuthIdentity) Verify(rCtx *restserver.Context) error {
	bkTicket, err := rCtx.GetCookie(CookieKeyBKTicket)
	if err != nil {
		return fmt.Errorf("failed to verify authentication: %w", err)
	}

	loginUsername, err := identity.handler.Verify(rCtx, bkTicket)
	if err != nil {
		return fmt.Errorf("failed to verify authentication: %w", err)
	}

	// TODO: 等到多租户版本上线，LoginName 需要绑定新的 headerKey
	rCtx.BKUsername = loginUsername
	rCtx.LoginName = loginUsername

	return nil
}
