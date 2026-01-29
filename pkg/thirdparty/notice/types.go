/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package notice

import (
	"fmt"

	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
)

// Config the config of notice.
type Config struct {
	APIGWUserConfig apigwclient.UserConfig
}

// Validate configures the config.
func (conf *Config) Validate() error {
	if err := conf.APIGWUserConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate notice client config: %v", err)
	}

	return nil
}

const (
	// codeOK define the success code.
	codeOK = 0
)

// BaseBroker describe the base broker.
type BaseBroker[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// IsFailed check the response is ok.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Code != codeOK {
		return fmt.Errorf("code(%d), msg(%s)", resp.Code, resp.Message)
	}

	return nil
}

// getCurrentAnnouncementsParams describe the parameters for getting current announcements.
// This is an internal type, not exposed in the public API.
type getCurrentAnnouncementsParams struct {
	Platform string
	Username string
}

// toQueryParams converts params to URL query parameters.
func (p *getCurrentAnnouncementsParams) toQueryParams() map[string]string {
	params := make(map[string]string)
	if p.Platform != "" {
		params["platform"] = p.Platform
	}
	if p.Username != "" {
		params["username"] = p.Username
	}

	return params
}

// announcementContent describe the content of an announcement.
type announcementContent struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// announcementData describe the announcement data from API response.
type announcementData struct {
	ID           int64               `json:"id"`
	Content      announcementContent `json:"content"`
	AnnounceType string              `json:"announce_type"`
	StartTime    string              `json:"start_time"`
	EndTime      string              `json:"end_time"`
}

// registerApplicationResp describe the response data after registering an application.
type registerApplicationResp struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// appRegistration represents the result of registering an application (internal use).
type appRegistration struct {
	ID   int64
	Code string
	Name string
}
