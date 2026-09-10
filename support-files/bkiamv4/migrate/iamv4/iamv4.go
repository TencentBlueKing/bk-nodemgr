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

package iamv4

import (
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

type cli struct {
	client restclient.IClient
	config *Config
}

func newClient(c *restclient.Capability, conf *Config) (*cli, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	client, err := apigwclient.NewClient(c, conf.BaseURL+"/api/v1/open", apigwclient.VirtualUserConfig{AppConfig: conf.AppConfig},
		restclient.WithCustomHeaderMasker(apigwheader.BKGWAuthKey, func(string) string { return "[REDACTED]" }))
	if err != nil {
		return nil, fmt.Errorf("create APIGW client: %w", err)
	}

	return &cli{client: client, config: conf}, nil
}

func (c *cli) getHeader(ctx contextx.IContext) http.Header {
	header := http.Header{}
	header.Set(apigwheader.BKGWAuthKey, c.config.GetAuthHeader())
	header.Set(apigwheader.BKGWTenantIDKey, ctx.TenantID())

	return header
}
