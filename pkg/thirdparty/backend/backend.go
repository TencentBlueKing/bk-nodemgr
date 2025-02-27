/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
)

// This file only supports requesting and getting responses.

// HeaderSetter ...
type HeaderSetter interface {
	GetAuthHeader() (string, error)
}

// Config the config of cmdb.
type Config struct {
	TenantID     string
	HeaderSetter HeaderSetter
}

// cli client for cmdb.
type cli struct {
	client rest.ClientInterface
	config *Config
}

// newClient initialize a new cmdb client.
func newClient(c *client.Capability, conf *Config) (*cli, error) {
	restCli, err := rest.NewClient(c, "/api/v3")
	if err != nil {
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
	header.Set(restheader.RIDKey, restheader.RIDGenerator())
	header.Set(restheader.TenantIDKey, tenantID)

	authHeader, err := c.config.HeaderSetter.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	header.Set(restheader.BKGWAuthKey, authHeader)

	return header, nil
}
