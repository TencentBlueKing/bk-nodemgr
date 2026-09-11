/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
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
	restCli, err := restclient.NewClient(c, "/api/v3")
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

func (c *cli) uploadOriginProxy(
	nCtx contextx.IContext, tenantID string, req *protoFile.UploadOriginProxyReq, fileName string, file io.Reader) (
	*protoFile.UploadOriginProxyResp_Data, error) {

	resp := new(protoFile.UploadOriginProxyResp)
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
		SubResourcef("/upload/origin/proxy").
		WithContext(nCtx).
		WithHeaders(header).
		BodyReader(body).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to upload origin proxy. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to upload origin proxy, get empty data. code(%d), message(%s), request-id(%s)",
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
		if err := writer.WriteField("filename", fileName); err != nil {
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

func (c *cli) downloadRemoteFile(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadRemoteFileReq) (*restserver.StreamResponse, error) {
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	result := c.client.Post().
		SubResourcef("/download/remote_file").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do()
	reader, err := result.RawStream()
	if err != nil {
		return nil, fmt.Errorf("failed to download remote file: %w", err)
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

func (c *cli) downloadReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadPluginReq) (
	*restserver.StreamResponse, error) {

	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	result := c.client.Post().
		SubResourcef("/download/plugin").
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

func (c *cli) downloadReleaseCert(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadCertReq) (
	*restserver.StreamResponse, error) {

	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	result := c.client.Post().
		SubResourcef("/download/cert").
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

func (c *cli) downloadReleaseBinTool(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadBinToolReq) (
	*restserver.StreamResponse, error) {

	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	result := c.client.Post().
		SubResourcef("/download/bintool").
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

func (c *cli) downloadReleasePluginBinTool(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadPluginBinToolReq) (
	*restserver.StreamResponse, error) {

	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	result := c.client.Post().
		SubResourcef("/download/plugin_bintool").
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

func (c *cli) downloadInstaller(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadInstallerReq) (
	*restserver.StreamResponse, error) {

	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	result := c.client.Post().
		SubResourcef("/download/installer").
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

func (c *cli) uploadOriginPluginV2(
	nCtx contextx.IContext, tenantID string, req *protoFile.UploadOriginPluginV2Req, fileName string, file io.Reader) (
	*protoFile.UploadOriginPluginV2Resp_Data, error) {

	resp := new(protoFile.UploadOriginPluginV2Resp)
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
		SubResourcef("/upload/origin/v2/plugin").
		WithContext(nCtx).
		WithHeaders(header).
		BodyReader(body).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to upload origin plugin v2. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to upload origin plugin v2, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) uploadOriginExternalPluginV2(
	nCtx contextx.IContext, tenantID string, req *protoFile.UploadOriginExternalPluginV2Req, fileName string, file io.Reader) (
	*protoFile.UploadOriginExternalPluginV2Resp_Data, error) {

	resp := new(protoFile.UploadOriginExternalPluginV2Resp)
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
		SubResourcef("/upload/origin/v2/external_plugin").
		WithContext(nCtx).
		WithHeaders(header).
		BodyReader(body).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to upload origin external plugin v2. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to upload origin external plugin v2, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) uploadOriginPluginV3(
	nCtx contextx.IContext, tenantID string, req *protoFile.UploadOriginPluginV3Req, fileName string, file io.Reader) (
	*protoFile.UploadOriginPluginV3Resp_Data, error) {

	resp := new(protoFile.UploadOriginPluginV3Resp)
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
		SubResourcef("/upload/origin/v3/plugin").
		WithContext(nCtx).
		WithHeaders(header).
		BodyReader(body).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to upload origin plugin v3. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to upload origin plugin v3, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) publishReleasePluginV2(nCtx contextx.IContext, tenantID string, req *protoFile.PublishReleasePluginV2Req) (
	*protoFile.PublishReleasePluginV2Resp_Data, error) {

	resp := new(protoFile.PublishReleasePluginV2Resp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/publish/release/v2/plugin").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to publish release plugin v2. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp.GetData(), nil
}

func (c *cli) publishReleaseExternalPluginV2(nCtx contextx.IContext, tenantID string, req *protoFile.PublishReleaseExternalPluginV2Req) (
	*protoFile.PublishReleaseExternalPluginV2Resp_Data, error) {

	resp := new(protoFile.PublishReleaseExternalPluginV2Resp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/publish/release/v2/external_plugin").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to publish release external plugin v2. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp.GetData(), nil
}

func (c *cli) publishReleasePluginV3(nCtx contextx.IContext, tenantID string, req *protoFile.PublishReleasePluginV3Req) (
	*protoFile.PublishReleasePluginV3Resp_Data, error) {

	resp := new(protoFile.PublishReleasePluginV3Resp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/publish/release/v3/plugin").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to publish release plugin v3. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp.GetData(), nil
}

func (c *cli) uploadOriginPluginBinTool(
	nCtx contextx.IContext, tenantID string, req *protoFile.UploadOriginPluginBinToolReq, fileName string, file io.Reader) (
	*protoFile.UploadOriginPluginBinToolResp_Data, error) {

	resp := new(protoFile.UploadOriginPluginBinToolResp)
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
		SubResourcef("/upload/origin/plugin_bintool").
		WithContext(nCtx).
		WithHeaders(header).
		BodyReader(body).
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to upload origin plugin bin tool. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to upload origin plugin bin tool, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) publishReleasePluginBinTool(nCtx contextx.IContext, tenantID string, req *protoFile.PublishReleasePluginBinToolReq) (
	*protoFile.PublishReleasePluginBinToolResp_Data, error) {

	resp := new(protoFile.PublishReleasePluginBinToolResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/publish/release/plugin_bintool").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to publish release plugin bin tool. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp.GetData(), nil
}

func (c *cli) infoReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadAgentReq) (
	*protoFile.FileInfoResp_Data, error) {

	resp := new(protoFile.FileInfoResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/info/agent").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to get release agent info. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to get release agent info, empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) infoReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadProxyReq) (
	*protoFile.FileInfoResp_Data, error) {

	resp := new(protoFile.FileInfoResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/info/proxy").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to get release proxy info. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to get release proxy info, empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) infoReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadPluginReq) (
	*protoFile.FileInfoResp_Data, error) {

	resp := new(protoFile.FileInfoResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/info/plugin").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to get release plugin info. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to get release plugin info, empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) infoReleaseCert(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadCertReq) (
	*protoFile.FileInfoResp_Data, error) {

	resp := new(protoFile.FileInfoResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/info/cert").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to get release cert info. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to get release cert info, empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) infoReleaseBinTool(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadBinToolReq) (
	*protoFile.FileInfoResp_Data, error) {

	resp := new(protoFile.FileInfoResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/info/bintool").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to get release bintool info. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to get release bintool info, empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) infoReleasePluginBinTool(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadPluginBinToolReq) (
	*protoFile.FileInfoResp_Data, error) {

	resp := new(protoFile.FileInfoResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/info/plugin_bintool").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to get release plugin bintool info. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to get release plugin bintool info, empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) infoInstaller(nCtx contextx.IContext, tenantID string, req *protoFile.DownloadInstallerReq) (
	*protoFile.FileInfoResp_Data, error) {

	resp := new(protoFile.FileInfoResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/info/installer").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to get installer info. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("failed to get installer info, empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) exportPrepareOriginPluginPackage(nCtx contextx.IContext, tenantID string, req *protoFile.ExportPrepareOriginPluginPackageReq) (
	*protoFile.ExportPrepareOriginPluginPackageResp_Data, error) {

	resp := new(protoFile.ExportPrepareOriginPluginPackageResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/export/prepare/origin_plugin_package").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to prepare origin plugin package export. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, errors.New("failed to prepare origin plugin package export, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) exportGetOriginPluginPackageDownloadAddress(
	nCtx contextx.IContext, tenantID string, req *protoFile.ExportGetOriginPluginPackageDownloadAddressReq) (
	*protoFile.ExportGetOriginPluginPackageDownloadAddressResp_Data, error) {

	resp := new(protoFile.ExportGetOriginPluginPackageDownloadAddressResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/export/get_address/origin_plugin_package").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to get origin plugin package export download address. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, errors.New("failed to get origin plugin package export download address, get empty data")
	}

	return resp.GetData(), nil
}

// ===============================================================================
// ReleaseAgent Related Interface
// ===============================================================================

func (c *cli) listReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseAgentListReq) (
	*protoFile.ReleaseAgentListResp_Data, error) {

	resp := new(protoFile.ReleaseAgentListResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/agent/list").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query agent release list. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query agent release list, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) countReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseAgentCountReq) (
	*protoFile.ReleaseAgentCountResp_Data, error) {

	resp := new(protoFile.ReleaseAgentCountResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/agent/count").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query agent release count. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query agent release count, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) getReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseAgentGetReq) (
	*protoFile.ReleaseAgentGetResp_Data, error) {

	resp := new(protoFile.ReleaseAgentGetResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/agent/get").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query agent release get. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query agent release get, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) distinctReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseAgentDistinctReq) (
	*protoFile.ReleaseAgentDistinctResp_Data, error) {

	resp := new(protoFile.ReleaseAgentDistinctResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/agent/distinct").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query agent release distinct. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query agent release distinct, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) enableReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseAgentEnableReq) error {
	resp := new(protoFile.ReleaseAgentEnableResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/agent/enable").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to enable release agent. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) disableReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseAgentDisableReq) error {
	resp := new(protoFile.ReleaseAgentDisableResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/agent/disable").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to disable release agent. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setAsDefaultReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseAgentSetAsDefaultReq) error {
	resp := new(protoFile.ReleaseAgentSetAsDefaultResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/agent/set_as_default").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to set release agent as default. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) cancelAsDefaultReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseAgentCancelAsDefaultReq) error {
	resp := new(protoFile.ReleaseAgentCancelAsDefaultResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/agent/cancel_as_default").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to cancel release agent default setting. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) deleteReleaseAgent(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseAgentDeleteReq) error {
	resp := new(protoFile.ReleaseAgentDeleteResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/agent/delete").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to delete release agent. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setReleaseAgentLabelsMany(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseAgentSetLabelsManyReq) error {
	resp := new(protoFile.ReleaseAgentSetLabelsManyResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/agent/set_labels_many").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to set release agent labels. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

// ===============================================================================
// ReleaseProxy Related Interface
// ===============================================================================

func (c *cli) listReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseProxyListReq) (
	*protoFile.ReleaseProxyListResp_Data, error) {

	resp := new(protoFile.ReleaseProxyListResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/proxy/list").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query proxy release list. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query proxy release list, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) countReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseProxyCountReq) (
	*protoFile.ReleaseProxyCountResp_Data, error) {

	resp := new(protoFile.ReleaseProxyCountResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/proxy/count").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query proxy release count. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query proxy release count, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) getReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseProxyGetReq) (
	*protoFile.ReleaseProxyGetResp_Data, error) {

	resp := new(protoFile.ReleaseProxyGetResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/proxy/get").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query proxy release get. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query proxy release get, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) distinctReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseProxyDistinctReq) (
	*protoFile.ReleaseProxyDistinctResp_Data, error) {

	resp := new(protoFile.ReleaseProxyDistinctResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/proxy/distinct").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query proxy release distinct. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query proxy release distinct, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) enableReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseProxyEnableReq) error {
	resp := new(protoFile.ReleaseProxyEnableResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/proxy/enable").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to enable release proxy. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) disableReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseProxyDisableReq) error {
	resp := new(protoFile.ReleaseProxyDisableResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/proxy/disable").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to disable release proxy. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setAsDefaultReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseProxySetAsDefaultReq) error {
	resp := new(protoFile.ReleaseProxySetAsDefaultResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/proxy/set_as_default").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to set release proxy as default. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) cancelAsDefaultReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseProxyCancelAsDefaultReq) error {
	resp := new(protoFile.ReleaseProxyCancelAsDefaultResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/proxy/cancel_as_default").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to cancel release proxy default setting. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) deleteReleaseProxy(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseProxyDeleteReq) error {
	resp := new(protoFile.ReleaseProxyDeleteResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/proxy/delete").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to delete release proxy. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setReleaseProxyLabelsMany(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseProxySetLabelsManyReq) error {
	resp := new(protoFile.ReleaseProxySetLabelsManyResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/proxy/set_labels_many").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to set release proxy labels. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

// ===============================================================================
// ReleasePlugin Related Interface
// ===============================================================================

func (c *cli) listReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginListReq) (
	*protoFile.ReleasePluginListResp_Data, error) {

	resp := new(protoFile.ReleasePluginListResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/plugin/list").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query plugin release list. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query plugin release list, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) countReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginCountReq) (
	*protoFile.ReleasePluginCountResp_Data, error) {

	resp := new(protoFile.ReleasePluginCountResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/plugin/count").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query plugin release count. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query plugin release count, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) getReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginGetReq) (
	*protoFile.ReleasePluginGetResp_Data, error) {

	resp := new(protoFile.ReleasePluginGetResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/plugin/get").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query plugin release get. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query plugin release get, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) distinctReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginDistinctReq) (
	*protoFile.ReleasePluginDistinctResp_Data, error) {

	resp := new(protoFile.ReleasePluginDistinctResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/plugin/distinct").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query plugin release distinct. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query plugin release distinct, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) existReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginExistReq) (
	*protoFile.ReleasePluginExistResp_Data, error) {

	resp := new(protoFile.ReleasePluginExistResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/plugin/exist").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query plugin release exist. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query plugin release exist, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) getReleasePluginDefaultVersion(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginGetDefaultVersionReq) (
	*protoFile.ReleasePluginGetDefaultVersionResp_Data, error) {

	resp := new(protoFile.ReleasePluginGetDefaultVersionResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/plugin/default_version").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query plugin release default_version. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query plugin release default_version, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) distinctNameReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginDistinctNameReq) (
	*protoFile.ReleasePluginDistinctNameResp_Data, error) {

	resp := new(protoFile.ReleasePluginDistinctNameResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/plugin/distinct_name").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query plugin release distinct_name. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query plugin release distinct_name, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) enableReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginEnableReq) error {
	resp := new(protoFile.ReleasePluginEnableResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/plugin/enable").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to enable release plugin. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) disableReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginDisableReq) error {
	resp := new(protoFile.ReleasePluginDisableResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/plugin/disable").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to disable release plugin. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setAsDefaultReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginSetAsDefaultReq) error {
	resp := new(protoFile.ReleasePluginSetAsDefaultResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/plugin/set_as_default").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to set release plugin as default. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) cancelAsDefaultReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginCancelAsDefaultReq) error {
	resp := new(protoFile.ReleasePluginCancelAsDefaultResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/plugin/cancel_as_default").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to cancel release plugin default setting. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) deleteReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginDeleteReq) error {
	resp := new(protoFile.ReleasePluginDeleteResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/plugin/delete").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to delete release plugin. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setHiddenReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginSetHiddenReq) error {
	resp := new(protoFile.ReleasePluginSetHiddenResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/plugin/set_hidden").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to set release plugin hidden state. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) cancelHiddenReleasePlugin(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginCancelHiddenReq) error {
	resp := new(protoFile.ReleasePluginCancelHiddenResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/plugin/cancel_hidden").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to cancel release plugin hidden state. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

// ===============================================================================
// ReleaseCert Related Interface
// ===============================================================================

func (c *cli) listReleaseCert(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseCertListReq) (
	*protoFile.ReleaseCertListResp_Data, error) {

	resp := new(protoFile.ReleaseCertListResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/cert/list").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query cert release list. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query cert release list, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) deleteReleaseCert(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseCertDeleteReq) error {
	resp := new(protoFile.ReleaseCertDeleteResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/cert/delete").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to delete release cert. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

// ===============================================================================
// ReleaseBinTool Related Interface
// ===============================================================================

func (c *cli) listReleaseBinTool(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseBinToolListReq) (
	*protoFile.ReleaseBinToolListResp_Data, error) {

	resp := new(protoFile.ReleaseBinToolListResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/bintool/list").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query bintool release list. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query bintool release list, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) deleteReleaseBinTool(nCtx contextx.IContext, tenantID string, req *protoFile.ReleaseBinToolDeleteReq) error {
	resp := new(protoFile.ReleaseBinToolDeleteResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/bintool/delete").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to delete release bin tool. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

// ===============================================================================
// ReleasePluginBinTool Related Interface
// ===============================================================================

func (c *cli) listReleasePluginBinTool(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginBinToolListReq) (
	*protoFile.ReleasePluginBinToolListResp_Data, error) {

	resp := new(protoFile.ReleasePluginBinToolListResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/plugin_bintool/list").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query plugin_bintool release list. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query plugin_bintool release list, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) distinctNameReleasePluginBinTool(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginBinToolDistinctNameReq) (
	*protoFile.ReleasePluginBinToolDistinctNameResp_Data, error) {

	resp := new(protoFile.ReleasePluginBinToolDistinctNameResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}

	if err = c.client.Post().
		SubResourcef("/release/plugin_bintool/distinct_name").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp); err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to query plugin_bintool release distinct_name. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to query plugin_bintool release distinct_name, get empty data")
	}

	return resp.GetData(), nil
}

func (c *cli) deleteReleasePluginBinTool(nCtx contextx.IContext, tenantID string, req *protoFile.ReleasePluginBinToolDeleteReq) error {
	resp := new(protoFile.ReleasePluginBinToolDeleteResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/release/plugin_bintool/delete").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return fmt.Errorf("failed to do post request: %w", err)
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("failed to delete release plugin bin tool. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

// ===============================================================================
// PackageEvent Related Interface
// ===============================================================================

func (c *cli) listPackageEvent(nCtx contextx.IContext, tenantID string, req *protoFile.PackageEventListReq) (
	*protoFile.PackageEventListResp_Data, error) {

	resp := new(protoFile.PackageEventListResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}
	err = c.client.Post().
		SubResourcef("/event/list").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}
	if resp.GetCode() != CodeOK {
		return nil, fmt.Errorf("failed to list package event. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}
	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to list package event, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp.GetData(), nil
}

func (c *cli) distinctPackageEvent(nCtx contextx.IContext, tenantID string, req *protoFile.PackageEventDistinctReq) (
	*protoFile.PackageEventDistinctResp_Data, error) {

	resp := new(protoFile.PackageEventDistinctResp)
	header, err := c.getCommonHeader(nCtx, tenantID)
	if err != nil {
		return nil, err
	}
	err = c.client.Post().
		SubResourcef("/event/distinct").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to do post request: %w", err)
	}
	if resp.GetCode() != CodeOK {
		return nil, fmt.Errorf("failed to distinct package event. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}
	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to distinct package event, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp.GetData(), nil
}
