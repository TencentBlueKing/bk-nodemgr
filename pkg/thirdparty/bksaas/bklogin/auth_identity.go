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

package bklogin

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

const (
	// CookieKeyBKTicket is the key of cookie.
	CookieKeyBKTicket = "bk_ticket"

	// CookieKeyBKToken is the key of cookie.
	CookieKeyBKToken = "bk_token"
)

var _ restserver.IAuthIdentity = &AuthIdentity{}

// AuthIdentity verify the ticket.
type AuthIdentity struct {
	handler *Handler
}

// Verify the ticket.
func (identity *AuthIdentity) Verify(r restserver.IRequest) error {
	token, err := r.GetCookie(identity.handler.conf.AuthType)
	if err != nil {
		return fmt.Errorf("failed to verify authentication by get cookie: %w", err)
	}

	tenantID, bkUserName, loginName, err := identity.handler.Verify(contextx.New(r.GContext()), token)
	if err != nil {
		return fmt.Errorf("failed to verify authentication: %w", err)
	}

	r.Data().SetTenantID(tenantID)
	r.Data().SetBKUsername(bkUserName)
	r.Data().SetLoginName(loginName)

	return nil
}
