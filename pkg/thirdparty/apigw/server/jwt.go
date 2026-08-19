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

	"github.com/golang-jwt/jwt/v5"
)

// app blueking apigw application info.
type app struct {
	AppCode           string `json:"app_code"`
	Verified          bool   `json:"verified"`
	ValidErrorMessage string `json:"valid_error_message"`
}

// Validate app.
func (a *app) Validate() error {
	if a.AppCode == "" {
		return errors.New("app code is required")
	}

	return nil
}

// user blueking apigw user info.
type user struct {
	UserName          string `json:"username"`
	Verified          bool   `json:"verified"`
	ValidErrorMessage string `json:"valid_error_message"`
}

// Validate user.
func (u *user) Validate() error {
	if u.UserName == "" {
		return errors.New("username is required")
	}

	return nil
}

// bkAppStateClaims jwt certification for application states.
type bkAppStateClaims struct {
	App  app  `json:"app"`
	User user `json:"user"`
	jwt.RegisteredClaims
}

// Validate bkAppStateClaims.
func (claims *bkAppStateClaims) Validate() error {
	if err := claims.App.Validate(); err != nil {
		return err
	}

	if err := claims.User.Validate(); err != nil {
		return err
	}

	if !claims.App.Verified {
		return errors.New(claims.App.ValidErrorMessage)
	}

	return nil
}

// bkUserStateClaims jwt certification for user states.
type bkUserStateClaims struct {
	App  app  `json:"app"`
	User user `json:"user"`
	jwt.RegisteredClaims
}

// Validate bkUserStateClaims.
func (claims *bkUserStateClaims) Validate() error {
	if err := claims.App.Validate(); err != nil {
		return err
	}

	if err := claims.User.Validate(); err != nil {
		return err
	}

	if !claims.App.Verified {
		return errors.New(claims.App.ValidErrorMessage)
	}

	if !claims.User.Verified {
		return errors.New(claims.User.ValidErrorMessage)
	}

	return nil
}

// parseToken parse token by jwt token and secret for application states.
func parseToken(token string, pem []byte, claims jwt.Claims) error {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// parse public key.
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pem)
		if err != nil {
			return nil, fmt.Errorf("parse public key error: %v", err)
		}

		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
	}

	tokenClaims, err := jwt.ParseWithClaims(token, claims, keyFunc)
	if err != nil {
		return err
	}

	if tokenClaims == nil {
		return errors.New("can not get token from parse with claims")
	}

	_, ok := tokenClaims.Claims.(jwt.Claims)
	if !ok {
		return errors.New("token claims type error")
	}

	if !tokenClaims.Valid {
		return errors.New("token claims valid failed")
	}

	return nil
}
