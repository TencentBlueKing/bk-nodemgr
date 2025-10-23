/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package header provide the header of rest server.
package header

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// BKTenantIDKey is tenant id header key.
	BKTenantIDKey = "X-Bk-Tenant-Id"

	// BKNodemgrRequestIDKey is request id header key.
	BKNodemgrRequestIDKey = "X-Bknodemgr-Request-Id"

	// BKNodemgrAuthorization is authorization header key.
	BKNodemgrAuthorization = "X-Bknodemgr-Authorization"
)

// BKTenantIDGetter get tenant id value.
func BKTenantIDGetter(req *http.Request) string {
	id := req.Header.Get(BKTenantIDKey)

	return id
}

// BKNodemgrRequestIDGetter get request id value.
func BKNodemgrRequestIDGetter(req *http.Request) string {
	id := req.Header.Get(BKNodemgrRequestIDKey)

	return id
}

// BKNodeMgrAuthorizationGetter get authorization value.
func BKNodeMgrAuthorizationGetter(req *http.Request) string {
	authorization := req.Header.Get(BKNodemgrAuthorization)

	return authorization
}

// BKNodeMgrAuthorization is nodemgr authorization.
type BKNodeMgrAuthorization struct {
	LoginName  string `json:"login_name"`
	BkUserName string `json:"bk_username"`
}

// BKNodeMgrClaims is nodemgr claims.
type BKNodeMgrClaims struct {
	BKNodeMgrAuthorization `json:",inline"`
	jwt.RegisteredClaims
}

const (
	// BKNodeMgrJWTIssuer is nodemgr jwt issuer.
	BKNodeMgrJWTIssuer = "bk-nodemgr"

	// BKNodeMgrJWTExpirationTime is nodemgr jwt expiration time.
	BKNodeMgrJWTExpirationTime = 24 * time.Hour
)

// IBKNodeMgrAuthorizationParser is nodemgr authorization parser.
type IBKNodeMgrAuthorizationParser interface {
	Parse(authorizationStr string) (*BKNodeMgrAuthorization, error)
}

// IBKNodeMgrAuthorizationGenerator is nodemgr authorization generator.
type IBKNodeMgrAuthorizationGenerator interface {
	Generate(loginName, bkUserName string) (string, error)
}

var (
	_ IBKNodeMgrAuthorizationParser    = &NodeMgrAuthorizationManager{}
	_ IBKNodeMgrAuthorizationGenerator = &NodeMgrAuthorizationManager{}
)

// NodeMgrAuthorizationManager is nodemgr authorization manager.
type NodeMgrAuthorizationManager struct {
	jwtSecret []byte
}

// NewNodeMgrAuthorizationManager create nodemgr authorization manager.
func NewNodeMgrAuthorizationManager(jwtSecretStr string) *NodeMgrAuthorizationManager {
	return &NodeMgrAuthorizationManager{
		jwtSecret: []byte(jwtSecretStr),
	}
}

// Generate generate nodemgr authorization.
func (manager *NodeMgrAuthorizationManager) Generate(loginName, bkUserName string) (string, error) {
	claims := BKNodeMgrClaims{
		BKNodeMgrAuthorization: BKNodeMgrAuthorization{
			LoginName:  loginName,
			BkUserName: bkUserName,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(BKNodeMgrJWTExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    BKNodeMgrJWTIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(manager.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate bk nodemgr authorization: %w", err)
	}

	return tokenString, nil
}

// Parse parse bk-nodemgr authorization.
func (manager *NodeMgrAuthorizationManager) Parse(authorizationStr string) (*BKNodeMgrAuthorization, error) {
	token, err := jwt.ParseWithClaims(authorizationStr, &BKNodeMgrClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %w", token.Header["alg"])
		}

		return manager.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse bk nodemgr authorization: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("failed to parse bk nodemgr authorization: invalid token")
	}

	claims, ok := token.Claims.(*BKNodeMgrClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse bk nodemgr authorization: invalid claims")
	}

	return &BKNodeMgrAuthorization{
		LoginName:  claims.LoginName,
		BkUserName: claims.BkUserName,
	}, nil
}
