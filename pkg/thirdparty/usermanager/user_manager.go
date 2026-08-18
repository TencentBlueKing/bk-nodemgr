/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package usermanager provides handlers to operate user manager api.
package usermanager

import (
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

// cli client for usermanager.
type cli struct {
	client restclient.IClient
	config *Config
}

// newClient initialize a new user manager client.
func newClient(c *restclient.Capability, conf *Config) (*cli, error) {
	restCli, err := restclient.NewClient(c, "/api/v3",
		restclient.WithCustomHeaderMasker(apigwheader.BKGWAuthKey, apigwclient.AuthHeaderMasker))
	if err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
		config: conf,
	}, nil
}

// getHeader get user manager common header.
// nolint: unparam
func (c *cli) getHeader(nCtx contextx.IContext) (http.Header, error) {
	header := http.Header{}
	header.Set(restheader.BKTenantIDKey, nCtx.TenantID())
	header.Set(apigwheader.BKGWRIDKey, identifier.GenRequestID())
	header.Set(apigwheader.BKGWAuthKey, c.config.APIGWUserConfig.GetAuthHeader())

	return header, nil
}

// listTenant list tenant.
func (c *cli) listTenant(nCtx contextx.IContext) (listTenantResp, error) {
	resp := new(BaseBroker[listTenantResp])
	header, err := c.getHeader(nCtx)
	if err != nil {
		return nil, err
	}

	err = c.client.Get().
		SubResourcef("/open/tenants/").
		WithContext(nCtx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("failed to list tenant: %v", err)
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("failed to list tenant: missing tenant data")
	}

	return *resp.Data, nil
}
