/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package notice provides handlers to operate notice api.
package notice

import (
	"fmt"
	"net/http"

	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
)

// This file only supports requesting and getting responses.

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
// nolint: unparam
func (c *cli) getHeader() (http.Header, error) {
	header := http.Header{}

	// TODO: integrate tenant info

	return header, nil
}
