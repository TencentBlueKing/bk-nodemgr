/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bklogin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
)

// cli client for bkoa.
type cli struct {
	client rest.ClientInterface
}

// newClient initialize a new bkoa client.
func newClient(c *client.Capability) (*cli, error) {
	restCli, err := rest.NewClient(c, "/")
	if err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
	}, nil
}

// getCommonHeader get cmdb common header.
func (c *cli) getCommonHeader() (http.Header, error) {
	header := http.Header{}
	header.Set(restheader.BKRIDKey, restheader.BKRIDGenerator())

	return header, nil
}

// getUserInfo get user info.
func (c *cli) getUserInfo(ctx context.Context, req *GetUserInfoReq) (*GetUserInfoResp, error) {
	resp := new(BaseBroker[*GetUserInfoResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Get().
		SubResourcef("/user/get_info").
		WithContext(ctx).
		WithHeaders(header).
		WithParam("bk_ticket", req.BKTicket).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return resp.Data, nil
}
