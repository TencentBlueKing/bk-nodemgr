/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package monitor provides handlers to operate monitor API.
package monitor

import (
	"fmt"
	"net/http"
	"strconv"

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

	restCli, err := apigwclient.NewClient(c, "/", conf.VirtualUserConfig)
	if err != nil {
		return nil, err
	}

	return &cli{client: restCli, config: conf}, nil
}

func (c *cli) getHeader(nCtx contextx.IContext) (http.Header, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	header := http.Header{}
	header.Set(apigwheader.BKGWTenantIDKey, nCtx.TenantID())
	header.Set(apigwheader.BKGWAuthKey, c.config.VirtualUserConfig.GetAuthHeader(nCtx))

	return header, nil
}

func (c *cli) getOrCreateAgentEventDataID(
	nCtx contextx.IContext,
	bkBizID int64,
) (*GetOrCreateAgentEventDataIDResp, error) {

	resp := new(BaseBroker[*GetOrCreateAgentEventDataIDResp])
	header, err := c.getHeader(nCtx)
	if err != nil {
		return nil, err
	}

	err = c.client.Get().
		SubResourcef("/app/metadata/get_or_create_agent_event_data_id/").
		WithContext(nCtx).
		WithHeaders(header).
		WithParam("bk_biz_id", strconv.FormatInt(bkBizID, 10)).
		EnableLogBody().
		EnableLogResponse().
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("get or create agent event data-id failed: %w", err)
	}

	return resp.Data, nil
}
