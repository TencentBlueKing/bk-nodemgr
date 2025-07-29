/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package iegtjj provides handlers to operate on iegtjj api.
package iegtjj

import (
	"context"
	"net/http"

	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
)

// cli client for iegtjj.
type cli struct {
	client restclient.IClient
}

// newClient initialize a new iegtjj restclient.
func newClient(c *restclient.Capability) (*cli, error) {
	restCli, err := restclient.NewClient(c, "/")
	if err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
	}, nil
}

// getCommonHeader get a common header.
func (c *cli) getCommonHeader() (http.Header, error) {
	header := http.Header{}

	return header, nil
}

func (c *cli) getDevicePassword(ctx context.Context, req *GetDevicePasswordReq) (*GetDevicePasswordResp, error) {
	resp := new(BaseBroker[*GetDevicePasswordResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/pwd/getDevicePassword").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
