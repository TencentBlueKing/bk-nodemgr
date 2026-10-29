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

package client

import (
	"fmt"

	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

var _ restclient.IClient = &Client{}

// Client apigw client.
type Client struct {
	restClient restclient.IClient
	config     VirtualUserConfig
}

// NewClient new apigw client.
// this application will use the special user config to request.
func NewClient(capability *restclient.Capability, baseURL string, config VirtualUserConfig, opts ...restclient.Opt) (*Client, error) {
	restOpts := []restclient.Opt{
		restclient.WithCustomHeaderMasker(apigwheader.BKGWAuthKey, AuthHeaderMasker),
	}
	restOpts = append(restOpts, opts...)

	restClient, err := restclient.NewClient(capability, baseURL, restOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to new apigw client: %v", err)
	}

	return &Client{
		restClient: restClient,
		config:     config,
	}, nil
}

// Post implement rest.IClient.
func (c *Client) Post() *restclient.Request {
	return c.restClient.Post()
}

// Put implement rest.IClient.
func (c *Client) Put() *restclient.Request {
	return c.restClient.Put()
}

// Get implement rest.IClient.
func (c *Client) Get() *restclient.Request {
	return c.restClient.Get()
}

// Delete implement rest.IClient.
func (c *Client) Delete() *restclient.Request {
	return c.restClient.Delete()
}

// Patch implement rest.IClient.
func (c *Client) Patch() *restclient.Request {
	return c.restClient.Patch()
}

// Head implement rest.IClient.
func (c *Client) Head() *restclient.Request {
	return c.restClient.Head()
}
