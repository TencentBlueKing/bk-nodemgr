/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package client provides an api gateway client.
package client

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// AppConfig defines the api gateway related app info.
type AppConfig struct {
	// endpoints is a seed list of host:port addresses of api gateway nodes.
	endpoints []string
	// appCode is the BlueKing app code of nodeman to request api gateway.
	appCode string
	// appSecret is the BlueKing app secret of nodeman to request api gateway.
	appSecret string
}

// NewAppConfig creates a new AppConfig.
func NewAppConfig(endpoints []string, appCode string, appSecret string) AppConfig {
	return AppConfig{
		endpoints: endpoints,
		appCode:   appCode,
		appSecret: appSecret,
	}
}

// GetAppCode returns the app code from the config.
func (conf *AppConfig) GetAppCode() string {
	return conf.appCode
}

// Validate api gateway runtime.
func (conf *AppConfig) Validate() error {
	if len(conf.endpoints) == 0 {
		return errors.New("failed to validated app config: api gateway endpoints is not set")
	}

	if len(conf.appCode) == 0 {
		return errors.New("failed to validated app config: api gateway app code is not set")
	}

	if len(conf.appSecret) == 0 {
		return errors.New("failed to validated app config: api gateway app secret is not set")
	}

	return nil
}

// UserConfig defines the api gateway related runtime.
type UserConfig struct {
	// AppConfig is the api gateway related app info.
	AppConfig
	// AuthMode is the BlueKing api authentication mode.
	AuthMode AuthMode
	// BKUsername is the BlueKing BKUsername of nodeman to request api gateway.
	BKUsername string
	// AccessToken is the BlueKing access token of nodeman to request api gateway.
	AccessToken string
}

// AuthMode is the mode to do api auth verification.
type AuthMode string

const (
	// AuthModeAt at auth mode, use user accessToken for user authentication.
	AuthModeAt AuthMode = "at"
	// AuthModeUn un auth mode, use username for user authentication.
	AuthModeUn AuthMode = "un"
)

// Validate api gateway runtime.
func (conf *UserConfig) Validate() error {
	if err := conf.AppConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validated user config: %w", err)
	}

	switch conf.AuthMode {
	case AuthModeAt:
		if len(conf.AccessToken) == 0 {
			return errors.New("failed to validated user config: api gateway access token is not set")
		}
	case AuthModeUn:
		if len(conf.BKUsername) == 0 {
			return errors.New("failed to validated user config: api gateway user is not set")
		}
	default:
		return fmt.Errorf("failed to validated user config, env(%s): api gateway not support", conf.AuthMode)
	}

	return nil
}

// GetAuthHeader get api gateway auth header.
// # 调用目标 API 开启: 应用认证+用户认证
// # Call the target API to enable: application authentication+user authentication required.
// X-Bkapi-Authorization: {"bk_app_code": "x", "bk_app_secret": "y", "bk_ticket": "z"}
//
// # 调用目标 API 开启: 应用认证
// # Call the target API to enable: application authentication required.
// X-Bkapi-Authorization: {"bk_app_code": "x", "bk_app_secret": "y", "bk_username": "username"}
//
// # 调用目标 API 开启: 用户认证
// # Call the target API to enable: user authentication required.
// X-Bkapi-Authorization: {"bk_ticket": "z"}
//
// # 使用 access_token
// # Use access_token
// X-Bkapi-Authorization: {"access_token": "z"}.
func (conf *UserConfig) GetAuthHeader() string {
	var auth string
	switch conf.AuthMode {
	case AuthModeAt:
		auth = fmt.Sprintf("{\"access_token\":\"%s\"}", conf.AccessToken)
	case AuthModeUn:
		auth = fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\"}",
			conf.appCode, conf.appSecret, conf.BKUsername)
	default:
		// default use un mode.
		auth = fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\"}",
			conf.appCode, conf.appSecret, conf.BKUsername)
	}

	return auth
}

func sentinelFields() []string {
	return []string{
		"bk_app_secret",
		"access_token",
	}
}

func maskValue(value string) string {
	if len(value) > 6 {
		return value[:3] + "***" + value[len(value)-3:]
	}

	return strings.Repeat("*", len(value))
}

// AuthHeaderMasker mask auth header.
func AuthHeaderMasker(value string) string {
	result := value

	// handle every field that needs to be desensitized.
	for _, field := range sentinelFields() {
		// build a regular expression that matches the field value
		pattern := `"` + field + `":\s*"([^"]*)"`
		re := regexp.MustCompile(pattern)

		// if a match is found, replace it
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			value := re.FindStringSubmatch(match)[1]
			maskedValue := maskValue(value)

			return `"` + field + `": "` + maskedValue + `"`
		})
	}

	return result
}
