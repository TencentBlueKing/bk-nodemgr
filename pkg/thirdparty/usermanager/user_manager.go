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

// Package usermanager provides handlers to operate user manager api.
package usermanager

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
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
	restCli, err := apigwclient.NewClient(c, "/api/v3", conf.APIGWUserConfig)
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
	header.Set(apigwheader.BKGWTenantIDKey, nCtx.TenantID())

	return header, nil
}

// listTenant list tenant.
func (c *cli) listTenant(nCtx contextx.IContext) (listTenantResp, error) {
	resp := new(BaseBroker[listTenantResp])
	header, err := c.getHeader(nCtx)
	if err != nil {
		return nil, err
	}

	// notice: this api is special, it only need the app config.
	header.Set(apigwheader.BKGWAuthKey, c.config.APIGWUserConfig.AppConfig.GetAuthHeader())

	err = c.client.Get().
		SubResourcef("/open/tenants/").
		WithContext(nCtx).
		WithHeaders(header).
		EnableLogBody().
		EnableLogResponse().
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

func (c *cli) batchLookupVirtualUser(nCtx contextx.IContext, loginName string) (batchLookupVirtualUserResp, error) {
	resp := new(BaseBroker[batchLookupVirtualUserResp])
	header, err := c.getHeader(nCtx)
	if err != nil {
		return nil, err
	}

	// notice: this api is special, it only need the app config.
	header.Set(apigwheader.BKGWAuthKey, c.config.APIGWUserConfig.AppConfig.GetAuthHeader())

	err = c.client.Get().
		SubResourcef("/open/tenant/virtual-users/-/lookup/").
		WithContext(nCtx).
		WithHeaders(header).
		WithParams(map[string]string{
			"lookups":      loginName,
			"lookup_field": lookupFieldLoginName,
		}).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("failed to batch lookup virtual user: %w", err)
	}
	if resp.Data == nil {
		return nil, errors.New("failed to batch lookup virtual user: missing virtual user data")
	}

	return *resp.Data, nil
}
