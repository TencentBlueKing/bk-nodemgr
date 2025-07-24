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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
)

// HeaderSetter get auth header.
type HeaderSetter interface {
	GetAuthHeader() (string, error)
}

// cli client for iegtjj.
type cli struct {
	client       rest.ClientInterface
	headerSetter HeaderSetter
}

// newClient initialize a new iegtjj client.
func newClient(c *client.Capability, headerSetter HeaderSetter) (*cli, error) {
	restCli, err := rest.NewClient(c, "/")
	if err != nil {
		return nil, err
	}

	return &cli{
		client:       restCli,
		headerSetter: headerSetter,
	}, nil
}

// getCommonHeader get a common header.
func (c *cli) getCommonHeader() (http.Header, error) {
	header := http.Header{}
	header.Set(restheader.BKRIDKey, restheader.BKRIDGenerator())

	authHeader, err := c.headerSetter.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	header.Set(restheader.BKGWAuthKey, authHeader)

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
