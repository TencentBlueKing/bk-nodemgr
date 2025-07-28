/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package apigw

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// app blueking apigw application info.
type app struct {
	Version  int64  `json:"version"`
	AppCode  string `json:"app_code"`
	Verified bool   `json:"verified"`
}

// Validate app.
func (a *app) Validate() error {
	if !a.Verified {
		return errors.New("app not verified")
	}

	if a.AppCode == "" {
		return errors.New("app code is required")
	}

	return nil
}

// user blueking apigw user info.
type user struct {
	UserName string `json:"username"`
	Verified bool   `json:"verified"`
}

// Validate user.
func (u *user) Validate() error {
	if !u.Verified {
		return errors.New("user not verified")
	}

	if u.UserName == "" {
		return errors.New("username is required")
	}

	return nil
}

// bkClaims blueking apigw api gateway jwt struct.
type bkClaims struct {
	App  app  `json:"app"`
	User user `json:"user"`
	jwt.RegisteredClaims
}

// Validate bkClaims.
func (c *bkClaims) Validate() error {
	if err := c.App.Validate(); err != nil {
		return err
	}

	if err := c.User.Validate(); err != nil {
		return err
	}

	return nil
}

// parseToken parse token by jwt token and secret.
func parseToken(token string, pem []byte) (*bkClaims, error) {
	// parse public key.
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pem)
	if err != nil {
		return nil, err
	}

	tokenClaims, err := jwt.ParseWithClaims(token, &bkClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if tokenClaims == nil {
		return nil, errors.New("can not get token from parse with claims")
	}

	claims, ok := tokenClaims.Claims.(*bkClaims)
	if !ok {
		return nil, errors.New("token claims type error")
	}

	if !tokenClaims.Valid {
		return nil, errors.New("token claims valid failed")
	}

	return claims, nil
}
