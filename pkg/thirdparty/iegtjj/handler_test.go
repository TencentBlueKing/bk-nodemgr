/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package iegtjj

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

// This file only supports requesting and getting responses.

const (
	languageHeaderKey   = "X-Bkcmdb-Language"
	languageHeaderValue = "en"
)

// Config the config of cmdb.
type Config struct {
	HeaderSetter HeaderSetter
}

// Validate configures the config.
func (conf *Config) Validate() error {
	if conf.HeaderSetter == nil {
		return errors.New("header setter is nil")
	}

	return nil
}

// cli restclient for cmdb.
type cli struct {
	client restclient.IClient
	config *Config
}

// newClient initialize a new cmdb restclient.
func newClient(c *restclient.Capability, conf *Config) (*cli, error) {
	restCli, err := restclient.NewClient(c, "/api/v3")
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

// getCommonHeader get cmdb common header.
func (c *cli) getCommonHeader(tenantID string) (http.Header, error) {
	header := http.Header{}
	header.Set(apigwheader.BKRIDKey, restheader.RIDGenerator())
	header.Set(restheader.BKTenantIDKey, tenantID)
	header.Set(languageHeaderKey, languageHeaderValue)

	authHeader := c.config.HeaderSetter.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	header.Set(apigwheader.BKGWAuthKey, authHeader)

	return header, nil
}

type getPasswdReq struct {
}

type getPasswdResp struct {
}

// getPasswd ...
func (c *cli) getPasswd(ctx context.Context, req *getPasswdReq) (*getPasswdResp, error) {
	resp := new(BaseBroker[*getPasswdResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/hosts/app/%d/list_hosts", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list biz hosts failed, err: %v", err)
	}

	return resp.Data, nil
}
