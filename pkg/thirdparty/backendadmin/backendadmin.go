/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package backendadmin provides a client for backend admin APIs.
package backendadmin

import (
	"fmt"
	"net/http"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Config holds the configuration for the backend admin client.
type Config struct {
	RestJWTSecret          string
	RestJWTTokenExpiration time.Duration
}

type cli struct {
	client                           client.IClient
	config                           *Config
	jwtGenerator                     restheader.IBKNodeMgrAuthorizationGenerator
	getNetworkUnitSegmentRulesFn     func(contextx.IContext) (types.NetworkUnitSegmentRuleConfig, error)
	upsertNetworkUnitSegmentRulesFn  func(contextx.IContext, types.NetworkUnitSegmentRuleConfig) error
	syncUnassignedAgentNetworkUnitFn func(contextx.IContext, []int64) (*types.NodeAgentAssignUnitResult, error)
}

func newClient(c *client.Capability, conf *Config) (*cli, error) {
	if conf == nil {
		return nil, fmt.Errorf("backendadmin config is nil")
	}

	restCli, err := client.NewClient(c, "/admin")
	if err != nil {
		return nil, err
	}

	return &cli{
		client:       restCli,
		config:       conf,
		jwtGenerator: restheader.NewNodeMgrAuthorizationManager(conf.RestJWTSecret, restheader.WithJwtTokenExpiration(conf.RestJWTTokenExpiration)),
	}, nil
}

func (c *cli) getCommonHeader(nCtx contextx.IContext) (http.Header, error) {
	header := http.Header{}
	header.Set(restheader.BKTenantIDKey, nCtx.TenantID())
	header.Set(restheader.BKNodemgrRequestIDKey, identifier.GenRequestID())

	authorization, err := c.jwtGenerator.Generate(nCtx.LoginName(), nCtx.BKUsername())
	if err != nil {
		return nil, err
	}
	header.Set(restheader.BKNodemgrAuthorization, authorization)

	return header, nil
}
