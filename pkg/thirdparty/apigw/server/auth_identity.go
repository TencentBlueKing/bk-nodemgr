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
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
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

	if err := claims.Validate(); err != nil {
		return fmt.Errorf("failed to verify user authentication: %w", err)
	}

	ctx := context.Background()
	if req := r.GetRequest(); req != nil {
		ctx = req.Context()
	}

	loginName := claims.User.UserName
	nCtx := contextx.New(ctx,
		contextx.WithTenantID(tenantID),
		contextx.WithLoginName(loginName),
	)
	bkUsername, err := access.GetBKUsernameByLoginName(nCtx, loginName)
	if err != nil {
		return fmt.Errorf("failed to verify user authentication: failed to get bk username by login name: %w", err)
	}

	r.Data().SetLoginName(loginName)
	r.Data().SetBKUsername(bkUsername)

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

	if err := claims.Validate(); err != nil {
		return fmt.Errorf("failed to verify user authentication: %w", err)
	}

	loginName := claims.User.UserName
	nCtx := contextx.New(rCtx,
		contextx.WithTenantID(tenantID),
		contextx.WithLoginName(loginName),
	)
	bkUsername, err := access.GetBKUsernameByLoginName(nCtx, loginName)
	if err != nil {
		return fmt.Errorf("failed to verify user authentication: failed to get bk username by login name: %w", err)
	}

	rCtx.Data().SetLoginName(loginName)
	rCtx.Data().SetBKUsername(bkUsername)

	return nil
}
