/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bkrepo

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

type auth struct {
	username string
	password string
}

// GetHeader get auth header.
func (setter auth) GetHeader() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(setter.username+":"+setter.password))
}

// Config the config of gse.
type Config struct {
	RepoName  string
	ProjectID string

	Username string
	Password string
}

// cli client for gse.
type cli struct {
	client restclient.IClient
	config *Config
}

// newClient initialize a new gse client.
func newClient(c *restclient.Capability, conf *Config) (*cli, error) {
	restCli, err := restclient.NewClient(c, "/", restclient.WithSensitiveHeader(HeaderKeyAuth))
	if err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
		config: conf,
	}, nil
}

// getCommonHeader get gse common header.
func (c *cli) getCommonHeader() http.Header {
	header := http.Header{}
	header.Set(apigwheader.BKGWRIDKey, identifier.GenRequestID())

	header.Set(HeaderKeyAuth, auth{username: c.config.Username, password: c.config.Password}.GetHeader())

	return header
}

// DownloadFile download file from bkrepo.
func (c *cli) DownloadFile(ctx context.Context, req *DownloadFileReq) (*DownloadFileResp, error) {
	resp := new(DownloadFileResp)
	header := c.getCommonHeader()

	subPath, err := url.JoinPath("generic", c.config.ProjectID, c.config.RepoName, req.Path)
	if err != nil {
		return nil, err
	}

	resp.Data, err = c.client.Get().
		SubResourcef(subPath).
		WithParam("download", "true").
		WithContext(ctx).
		WithHeaders(header).
		Do().RawData()
	if err != nil {
		return nil, fmt.Errorf("download file failed, err: %w", err)
	}

	return resp, nil
}

// UploadFile upload file to bkrepo.
func (c *cli) UploadFile(ctx context.Context, req *UploadFileReq) (*UploadFileResp, error) {
	resp := new(BaseBroker[*UploadFileResp])
	header := c.getCommonHeader()
	header = req.Info.BindHeader(header)

	subPath, err := url.JoinPath("generic", c.config.ProjectID, c.config.RepoName, req.Path)
	if err != nil {
		return nil, fmt.Errorf("upload file failed, err: %w", err)
	}

	err = c.client.Put().
		SubResourcef(subPath).
		WithContext(ctx).
		WithHeaders(header).
		BodyReader(req.Reader).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("upload file failed, err: %w", err)
	}

	return resp.Data, nil
}

// QueryNodeInfo query node info.
func (c *cli) QueryNodeInfo(ctx context.Context, req *QueryNodeInfoReq) (*QueryNodeInfoResp, error) {
	resp := new(BaseBroker[*QueryNodeInfoResp])
	header := c.getCommonHeader()

	subPath, err := url.JoinPath("/repository/api/node/detail", c.config.ProjectID, c.config.RepoName, req.Path)
	if err != nil {
		return nil, fmt.Errorf("query node info failed, err: %w", err)
	}

	err = c.client.Get().
		SubResourcef(subPath).
		WithContext(ctx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("query node info failed, err: %w", err)
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("query node info failed, err: %w", err)
	}

	return resp.Data, nil
}

// ListNode list nodes request.
func (c *cli) ListNode(ctx context.Context, req *ListNodeReq) (*ListNodeResp, error) {
	resp := new(BaseBroker[*ListNodeResp])
	header := c.getCommonHeader()

	subPath, err := url.JoinPath("/repository/api/node/page", c.config.ProjectID, c.config.RepoName, req.Path)
	if err != nil {
		return nil, fmt.Errorf("list node failed, err: %w", err)
	}

	err = c.client.Get().
		SubResourcef(subPath).
		WithContext(ctx).
		WithHeaders(header).
		WithParams(map[string]string{
			"pageSize":        strconv.Itoa(req.PageSize),
			"pageNum":         strconv.Itoa(req.PageNum),
			"includeFolder":   "true",
			"includeMetadata": "false",
			"sortProperty":    "id",
			"direction":       "DESC",
		}).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("list node failed, err: %w", err)
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list node failed, err: %w", err)
	}

	return resp.Data, nil
}

// MkDir create directory request.
func (c *cli) MkDir(ctx context.Context, req *MkdirReq) error {
	resp := new(BaseBroker[*MkdirResp])
	header := c.getCommonHeader()

	subPath, err := url.JoinPath("/repository/api/node/mkdir", c.config.ProjectID, c.config.RepoName, req.Path)
	if err != nil {
		return fmt.Errorf("mkdir failed, err: %w", err)
	}

	err = c.client.Post().
		SubResourcef(subPath).
		WithContext(ctx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("mkdir failed, err: %w", err)
	}

	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("mkdir failed, err: %w", err)
	}

	return nil
}
