/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package apigw provides an APIGatewayClient client.
package client

import (
	"errors"
	"fmt"
)

// Config defines the api gateway related runtime.
type Config struct {
	// Endpoints is a seed list of host:port addresses of api gateway nodes.
	Endpoints []string
	// AppCode is the BlueKing app code of nodeman to request api gateway.
	AppCode string
	// AppSecret is the BlueKing app secret of nodeman to request api gateway.
	AppSecret string
	// User is the BlueKing User of nodeman to request api gateway.
	User string
	// AuthMode is the BlueKing api authentication mode.
	AuthMode AuthMode
	// BkTicket is the BlueKing access ticket of nodeman to request api gateway.
	BkTicket string
	// BkToken is the BlueKing User token of nodeman to request api gateway.
	BkToken string
	// AccessToken is the BlueKing access token of nodeman to request api gateway.
	AccessToken string
}

// AuthMode is the mode to do api auth verification,
// which supports 'oa', 'ee' and 'at' mode for now.
// 1. oa use bk_ticket auth.
// 2. ee use bk_token auth.
// 3. at use access_token auth.
type AuthMode string

const (
	// AuthModeEe ee auth mode, use bkToken for user authentication.
	AuthModeEe AuthMode = "ee"
	// AuthModeOa oa auth mode, use bkTicket for user authentication.
	AuthModeOa AuthMode = "oa"
	// AuthModeAt at auth mode, use accessToken for user authentication.
	AuthModeAt AuthMode = "at"
	// AuthModeUn un auth mode, use username for user authentication.
	AuthModeUn AuthMode = "un"
)

// Validate api gateway runtime.
func (c *Config) Validate() error {
	if len(c.Endpoints) == 0 {
		return errors.New("api gateway endpoints is not set")
	}
	if len(c.AppCode) == 0 {
		return errors.New("api gateway app code is not set")
	}
	if len(c.AppSecret) == 0 {
		return errors.New("api gateway app secret is not set")
	}

	switch c.AuthMode {
	case AuthModeEe:
		if len(c.BkToken) == 0 {
			return errors.New("api gateway bk token is not set")
		}
	case AuthModeOa:
		if len(c.BkToken) == 0 {
			return errors.New("api gateway bk token is not set")
		}
	case AuthModeAt:
		if len(c.AccessToken) == 0 {
			return errors.New("api gateway access token is not set")
		}
	case AuthModeUn:
		if len(c.User) == 0 {
			return errors.New("api gateway user is not set")
		}
	default:
		return fmt.Errorf("api gateway not support, env(%s)", c.AuthMode)
	}

	return nil
}

// GetAuthHeader get api gateway auth header.
// # 调用目标 API 开启: 应用认证+用户认证
// # Call the target API to enable: application authentication+user authentication required.
// X-Bkapi-Authorization: {"bk_app_code": "x", "bk_app_secret": "y", "bk_token": "z"}
// # 调用目标 API 开启: 应用认证
// # Call the target API to enable: application authentication required.
// X-Bkapi-Authorization: {"bk_app_code": "x", "bk_app_secret": "y"}
// # 调用目标 API 开启: 用户认证
// # Call the target API to enable: user authentication required.
// X-Bkapi-Authorization: {"bk_token": "z"}
// # 使用 access_token
// # Use access_token
// X-Bkapi-Authorization: {"access_token": "z"}
// # 调用目标 API 开启: 免用户认证
// # Call the target API to enable: no user authentication required.
// X-Bkapi-Authorization: {"bk_app_code": "x", "bk_app_secret": "y", "bk_username": "z"}
// when not set auth mode this func will return empty string and error.
func (c *Config) GetAuthHeader() string {
	auth := ""
	switch c.AuthMode {
	case AuthModeOa:
		auth = fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_ticket\":\"%s\"}",
			c.AppCode, c.AppSecret, c.BkTicket)
	case AuthModeEe:
		auth = fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_token\":\"%s\"}",
			c.AppCode, c.AppSecret, c.BkToken)
	case AuthModeAt:
		auth = fmt.Sprintf("{\"access_token\":\"%s\"}", c.AccessToken)
	case AuthModeUn:
		auth = fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\"}",
			c.AppCode, c.AppSecret, c.User)
	default:
		// default use un mode.
		auth = fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\"}",
			c.AppCode, c.AppSecret, c.User)
	}

	return auth
}
