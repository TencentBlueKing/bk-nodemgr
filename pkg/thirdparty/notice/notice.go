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

// Package notice provides handlers to operate notice api.
package notice

import (
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

// cli client for notice.
type cli struct {
	client restclient.IClient
	config *Config
}

// newClient initialize a new notice client.
func newClient(c *restclient.Capability, conf *Config) (*cli, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	// Use /apigw/v1 as baseURL (consistent with bk-hcm)
	restCli, err := apigwclient.NewClient(c, "/apigw/v1", conf.APIGWUserConfig)
	if err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
		config: conf,
	}, nil
}

// getHeader get notice common header.
func (c *cli) getHeader(nCtx contextx.IContext) http.Header {
	header := http.Header{}
	header.Set(apigwheader.BKGWTenantIDKey, nCtx.TenantID())
	header.Set(apigwheader.BKGWAuthKey, c.config.APIGWUserConfig.GetAuthHeader(nCtx))

	return header
}

// getCurrentAnnouncements get current announcements from notice center.
func (c *cli) getCurrentAnnouncements(nCtx contextx.IContext,
	params *getCurrentAnnouncementsParams) ([]*announcementData, error) {

	resp := new(BaseBroker[[]*announcementData])
	header := c.getHeader(nCtx)

	request := c.client.Get().
		SubResourcef("/announcement/get_current_announcements").
		WithContext(nCtx).
		WithHeaders(header)

	// Add query parameters
	if params != nil {
		for key, value := range params.toQueryParams() {
			request = request.WithParam(key, value)
		}
	}

	err := request.Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("get current announcements failed: %w", err)
	}

	return resp.Data, nil
}

// registerApplication registers an application with notice center.
// No request body needed - authentication is via header.
func (c *cli) registerApplication(nCtx contextx.IContext) (*registerApplicationResp, error) {
	resp := new(BaseBroker[*registerApplicationResp])
	header := c.getHeader(nCtx)

	err := c.client.Post().
		SubResourcef("/register").
		WithContext(nCtx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("register application failed: %w", err)
	}

	return resp.Data, nil
}
