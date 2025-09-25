/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package server

import (
	"errors"
	"fmt"

	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

// BKGWJWTAuthIdentityAppState verify the jwt from apigateway.
type BKGWJWTAuthIdentityAppState struct {
	pem []byte
}

// NewBKGWJWTAuthIdentity ...
func NewBKGWJWTAuthIdentity(pem []byte) *BKGWJWTAuthIdentityAppState {
	return &BKGWJWTAuthIdentityAppState{
		pem: pem,
	}
}

// Verify the jwt from apigateway.
func (identity *BKGWJWTAuthIdentityAppState) Verify(rCtx restserver.IRequest) error {
	if rCtx == nil {
		return errors.New("failed to verify user authentication, rest context is nil")
	}
	jwtStr := rCtx.GetRequestHeader(apigwheader.BKGWJWTTokenKey)
	if jwtStr == "" {
		return errors.New("failed to verify user authentication, jwt token is empty")
	}

	claims := new(bkAppStateClaims)
	err := parseToken(jwtStr, identity.pem, claims)
	if err != nil {
		return fmt.Errorf("failed to verify user authentication: %w", err)
	}

	if claims.Validate() != nil {
		return fmt.Errorf("failed to verify user authentication: %w", err)
	}

	rCtx.Data().SetLoginName(claims.User.UserName)
	// TODO: 等待多租户版本上线后，需要修改 rCtx.BKUsername1 的赋值
	rCtx.Data().SetBKUsername(claims.User.UserName)

	return nil
}

// BKGWJWTAuthIdentityUserState verify the jwt from apigateway for user state.
type BKGWJWTAuthIdentityUserState struct {
	pem []byte
}

// NewBKGWJWTAuthIdentityUserState ...
func NewBKGWJWTAuthIdentityUserState(pem []byte) *BKGWJWTAuthIdentityUserState {
	return &BKGWJWTAuthIdentityUserState{
		pem: pem,
	}
}

// Verify the jwt from apigateway.
func (identity *BKGWJWTAuthIdentityUserState) Verify(rCtx *restserver.Context) error {
	if rCtx == nil {
		return errors.New("failed to verify user authentication, rest context is nil")
	}
	jwtStr := rCtx.GetRequestHeader(apigwheader.BKGWJWTTokenKey)
	if jwtStr == "" {
		return errors.New("failed to verify user authentication, jwt token is empty")
	}

	claims := new(bkUserStateClaims)
	err := parseToken(jwtStr, identity.pem, claims)
	if err != nil {
		return fmt.Errorf("failed to verify user authentication: %w", err)
	}

	if claims.Validate() != nil {
		return fmt.Errorf("failed to verify user authentication: %w", err)
	}

	rCtx.Data().SetLoginName(claims.User.UserName)
	// TODO: 等待多租户版本上线后，需要修改 rCtx.BKUsername1 的赋值
	rCtx.Data().SetBKUsername(claims.User.UserName)

	return nil
}
