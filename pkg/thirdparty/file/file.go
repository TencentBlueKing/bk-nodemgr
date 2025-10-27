/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package file provides handlers to operate nodeman file api.
// nolint:dupl
package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
)

// CodeOK defines the success code.
const CodeOK = 0

// This file only supports requesting and getting responses.

// Config the config of backend.
type Config struct {
	RestJwtSecret          string
	RestJwtTokenExpiration time.Duration
}

// cli client for backend.
type cli struct {
	client       restclient.IClient
	config       *Config
	jwtGenerator restheader.IBKNodeMgrAuthorizationGenerator
}

// newClient initialize a new backend client.
func newClient(c *restclient.Capability, conf *Config) (*cli, error) {
	restCli, err := restclient.NewClient(c, "/")
	if err != nil {
		return nil, err
	}

	jwtGenerator := restheader.NewNodeMgrAuthorizationManager(conf.RestJwtSecret, restheader.WithJwtTokenExpiration(conf.RestJwtTokenExpiration))

	return &cli{
		client:       restCli,
		config:       conf,
		jwtGenerator: jwtGenerator,
	}, nil
}

// getCommonHeader get backend common header.
// nolint: unparam
func (c *cli) getCommonHeader(nCtx contextx.IContext, tenantID string) (http.Header, error) {
	header := http.Header{}
	header.Set(restheader.BKTenantIDKey, tenantID)
	header.Set(restheader.BKNodemgrRequestIDKey, identifier.GenRequestID())

	authorization, err := c.jwtGenerator.Generate(nCtx.LoginName(), nCtx.BKUsername())
	if err != nil {
		return nil, err
	}

	header.Set(restheader.BKNodemgrAuthorization, authorization)

	return header, nil
}

func (c *cli) uploadOriginAgent(
	nCtx contextx.IContext, tenantID string, req *protoFile.UploadOriginAgentReq, fileName string, file io.Reader) (
	*protoFile.UploadOriginAgentResp_Data, error) {

	resp := new(protoFile.UploadOriginAgentResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	body, err := c.generateUploadFileBody(req, header, fileName, file)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload file body: %w", err)
	}
	defer func() {
		_ = body.Close()
	}()

	err = c.client.Post().
		SubResourcef("/upload/origin/agent").
		WithContext(nCtx).
		WithHeaders(header).
		BodyReader(body).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to upload origin agent. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to upload origin agent, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) uploadOriginServer(
	nCtx contextx.IContext, tenantID string, req *protoFile.UploadOriginServerReq, fileName string, file io.Reader) (
	*protoFile.UploadOriginServerResp_Data, error) {

	resp := new(protoFile.UploadOriginServerResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	body, err := c.generateUploadFileBody(req, header, fileName, file)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload file body: %w", err)
	}
	defer func() {
		_ = body.Close()
	}()

	err = c.client.Post().
		SubResourcef("/upload/origin/server").
		WithContext(nCtx).
		WithHeaders(header).
		BodyReader(body).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to upload origin server. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to upload origin server, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) uploadOriginCert(
	nCtx contextx.IContext, tenantID string, req *protoFile.UploadOriginCertReq, fileName string, file io.Reader) (
	*protoFile.UploadOriginCertResp_Data, error) {

	resp := new(protoFile.UploadOriginCertResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	body, err := c.generateUploadFileBody(req, header, fileName, file)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload file body: %w", err)
	}
	defer func() {
		_ = body.Close()
	}()

	err = c.client.Post().
		SubResourcef("/upload/origin/cert").
		WithContext(nCtx).
		WithHeaders(header).
		BodyReader(body).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to upload origin cert. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to upload origin cert, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) uploadOriginBinTool(
	nCtx contextx.IContext, tenantID string, req *protoFile.UploadOriginBinToolReq, fileName string, file io.Reader) (
	*protoFile.UploadOriginBinToolResp_Data, error) {

	resp := new(protoFile.UploadOriginBinToolResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	body, err := c.generateUploadFileBody(req, header, fileName, file)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload file body: %w", err)
	}
	defer func() {
		_ = body.Close()
	}()

	err = c.client.Post().
		SubResourcef("/upload/origin/bintool").
		WithContext(nCtx).
		WithHeaders(header).
		BodyReader(body).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to upload origin bintool. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to upload origin bintool, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) generateUploadFileBody(
	metadata interface{}, header http.Header, fileName string, file io.Reader) (io.ReadCloser, error) {

	metadataStr, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// upload file pipe
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	header.Set("Content-Type", writer.FormDataContentType())

	gp := gopool.NewPool()
	gp.Go(func() error {
		defer func() {
			_ = writer.Close()
			_ = pw.Close()
		}()

		if err := writer.WriteField("metadata", string(metadataStr)); err != nil {
			return err
		}
		fileField, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			return fmt.Errorf("failed to create form file: %w", err)
		}

		_, _ = io.Copy(fileField, file)

		return nil
	})

	return pr, nil
}

func (c *cli) publishReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.PublishReleaseAgentReq) (
	*protoFile.PublishReleaseAgentResp_Data, error) {

	resp := new(protoFile.PublishReleaseAgentResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/publish/release/agent").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to publish release agent. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp.GetData(), nil
}

func (c *cli) downloadReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadAgentReq) (
	*restserver.StreamResponse, error) {

	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	result := c.client.Post().
		SubResourcef("/download/agent").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do()
	reader, err := result.RawStream()
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	return &restserver.StreamResponse{
		Data:       reader,
		StatusCode: result.StatusCode,
		Headers:    result.Header,
	}, nil
}

func (c *cli) downloadReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadProxyReq) (
	*restserver.StreamResponse, error) {

	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	result := c.client.Post().
		SubResourcef("/download/proxy").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do()
	reader, err := result.RawStream()
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	return &restserver.StreamResponse{
		Data:       reader,
		StatusCode: result.StatusCode,
		Headers:    result.Header,
	}, nil
}

func (c *cli) publishReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.PublishReleaseProxyReq) (
	*protoFile.PublishReleaseProxyResp_Data, error) {

	resp := new(protoFile.PublishReleaseProxyResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/publish/release/proxy").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to publish release proxy. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp.GetData(), nil
}

func (c *cli) publishReleaseCert(nCtx contextx.IContext, tenantID string, req *protoFile.PublishReleaseCertReq) (
	*protoFile.PublishReleaseCertResp_Data, error) {

	resp := new(protoFile.PublishReleaseCertResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/publish/release/cert").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to publish release cert. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp.GetData(), nil
}

func (c *cli) publishReleaseBinTool(nCtx contextx.IContext, tenantID string, req *protoFile.PublishReleaseBinToolReq) (
	*protoFile.PublishReleaseBinToolResp_Data, error) {

	resp := new(protoFile.PublishReleaseBinToolResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/publish/release/bintool").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to publish release bintool. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp.GetData(), nil
}

func (c *cli) launchTransferNode(nCtx contextx.IContext, tenantID string, req *protoFile.TransferLaunchNodeReq) (
	*protoFile.TransferLaunchNodeResp, error) {

	resp := new(protoFile.TransferLaunchNodeResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/transfer/launch/node").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to transfer node launch. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, errors.New("failed to transfer node launch. data is nil")
	}

	return resp, nil
}

func (c *cli) launchTransferPlugin(nCtx contextx.IContext, tenantID string, req *protoFile.TransferLaunchPluginReq) (
	*protoFile.TransferLaunchPluginResp, error) {

	resp := new(protoFile.TransferLaunchPluginResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/transfer/launch/plugin").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to transfer plugin launch. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, errors.New("failed to transfer plugin launch. data is nil")
	}

	return resp, nil
}

func (c *cli) launchTransferInstaller(nCtx contextx.IContext, tenantID string, req *protoFile.TransferLaunchInstallerReq) (
	*protoFile.TransferLaunchInstallerResp, error) {

	resp := new(protoFile.TransferLaunchInstallerResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/transfer/launch/installer").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to transfer installer launch. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, errors.New("failed to transfer installer launch. data is nil")
	}

	return resp, nil
}

func (c *cli) queryTransfer(nCtx contextx.IContext, tenantID string, req *protoFile.TransferQueryReq) (
	*protoFile.TransferQueryResp, error) {

	resp := new(protoFile.TransferQueryResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/transfer/query").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query transfer. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, errors.New("failed to query transfer. data is nil")
	}

	return resp, nil
}
