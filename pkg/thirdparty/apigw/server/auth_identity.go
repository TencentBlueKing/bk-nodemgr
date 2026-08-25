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

package server

import (
	"errors"
	"fmt"
	"regexp"

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

// TenantIDRegexp tenant id regexp.
const TenantIDRegexp = `^[a-z][a-z0-9-]{1,30}[a-z0-9]$`

// Verify the jwt from apigateway.
func (identity *BKGWJWTAuthIdentityAppState) Verify(r restserver.IRequest) error {
	if r == nil {
		return errors.New("failed to verify user authentication, rest context is nil")
	}

	tenantID := r.GetRequestHeader(apigwheader.BKGWTenantIDKey)

	mustCompile := regexp.MustCompile(TenantIDRegexp)
	if !mustCompile.MatchString(tenantID) {
		return errors.New("failed to set tenant id: invalid tenant id")
	}

	r.Data().SetTenantID(tenantID)

	jwtStr := r.GetRequestHeader(apigwheader.BKGWJWTTokenKey)
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

	// notice: apigw only can return the bk_username,
	// if need the login name, need to use exchange login name from user manager by bk_username.
	r.Data().SetBKUsername(claims.User.UserName)

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
func (identity *BKGWJWTAuthIdentityUserState) Verify(rCtx restserver.IContext) error {
	if rCtx == nil {
		return errors.New("failed to verify user authentication, rest context is nil")
	}

	tenantID := rCtx.GetRequestHeader(apigwheader.BKGWTenantIDKey)

	mustCompile := regexp.MustCompile(TenantIDRegexp)
	if !mustCompile.MatchString(tenantID) {
		return errors.New("failed to set tenant id: invalid tenant id")
	}

	rCtx.Data().SetTenantID(tenantID)

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

	// notice: apigw only can return the bk_username,
	// if need the login name, need to use exchange login name from user manager by bk_username.
	rCtx.Data().SetBKUsername(claims.User.UserName)

	return nil
}
