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
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
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
	restCli, err := restclient.NewClient(c, "/", restclient.WithHeaderMasker(HeaderKeyAuth))
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
	if tenant.GetMode() == tenant.ModeMultiple {
		header.Set(restheader.BKTenantIDKey, tenant.SystemTenantID)
	}

	header.Set(HeaderKeyAuth, auth{username: c.config.Username, password: c.config.Password}.GetHeader())

	return header
}

func (c *cli) effectiveProjectID() string {
	if tenant.GetMode() == tenant.ModeMultiple {
		return fmt.Sprintf("%s.%s", tenant.SystemTenantID, c.config.ProjectID)
	}

	return c.config.ProjectID
}

// DownloadFile download file from bkrepo.
func (c *cli) DownloadFile(nCtx contextx.IContext, req *DownloadFileReq) (*DownloadFileResp, error) {
	resp := new(DownloadFileResp)
	header := c.getCommonHeader()

	var err error
	resp.Data, err = c.client.Get().
		SubResourcef("generic/%s/%s/%s", c.effectiveProjectID(), c.config.RepoName, req.Path).
		WithParam("download", "true").
		WithContext(nCtx).
		WithHeaders(header).
		Do().RawData()
	if err != nil {
		return nil, fmt.Errorf("download file failed: %w", err)
	}

	return resp, nil
}

// UploadFile upload file to bkrepo.
func (c *cli) UploadFile(nCtx contextx.IContext, req *UploadFileReq) (*UploadFileResp, error) {
	resp := new(BaseBroker[*UploadFileResp])
	header := c.getCommonHeader()
	header = req.Info.BindHeader(header)

	err := c.client.Put().
		SubResourcef("generic/%s/%s/%s", c.effectiveProjectID(), c.config.RepoName, req.Path).
		WithContext(nCtx).
		WithHeaders(header).
		BodyReader(req.Reader).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("upload file failed: %w", err)
	}

	return resp.Data, nil
}

// QueryNodeInfo query node info.
func (c *cli) QueryNodeInfo(nCtx contextx.IContext, req *QueryNodeInfoReq) (*QueryNodeInfoResp, error) {
	resp := new(BaseBroker[*QueryNodeInfoResp])
	header := c.getCommonHeader()

	err := c.client.Get().
		SubResourcef("/repository/api/node/detail/%s/%s/%s", c.effectiveProjectID(), c.config.RepoName, req.Path).
		WithContext(nCtx).
		WithHeaders(header).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("query node info failed: %w", err)
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("query node info failed: %w", err)
	}

	return resp.Data, nil
}

// ListNode list nodes request.
func (c *cli) ListNode(nCtx contextx.IContext, req *ListNodeReq) (*ListNodeResp, error) {
	resp := new(BaseBroker[*ListNodeResp])
	header := c.getCommonHeader()

	err := c.client.Get().
		SubResourcef("/repository/api/node/page/%s/%s/%s", c.effectiveProjectID(), c.config.RepoName, req.Path).
		WithContext(nCtx).
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
		return nil, fmt.Errorf("list node failed: %w", err)
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list node failed: %w", err)
	}

	return resp.Data, nil
}

// MkDir create directory request.
func (c *cli) MkDir(nCtx contextx.IContext, req *MkdirReq) error {
	resp := new(BaseBroker[*MkdirResp])
	header := c.getCommonHeader()

	err := c.client.Post().
		SubResourcef("/repository/api/node/mkdir/%s/%s/%s", c.effectiveProjectID(), c.config.RepoName, req.Path).
		WithContext(nCtx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	return nil
}
