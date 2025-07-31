/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package gse provides handlers to operate gse API.
// nolint:dupl
package gse

import (
	"context"
	"fmt"
	"net/http"

	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
)

// Config the config of gse.
type Config struct {
	APIGWClientConfig apigwclient.Config
}

// Validate the config.
func (conf *Config) Validate() error {
	if err := conf.APIGWClientConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate gse config, err: %v", err)
	}

	return nil
}

// cli client for gse.
type cli struct {
	client restclient.IClient
	config *Config
}

// newClient initialize a new gse client.
func newClient(c *restclient.Capability, conf *Config) (*cli, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("failed to new gse client: %v", err)
	}

	restCli, err := apigwclient.NewClient(c, "/api/v2", conf.APIGWClientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to new gse client: %v", err)
	}

	return &cli{
		client: restCli,
		config: conf,
	}, nil
}

// getCommonHeader get gse common header.
func (c *cli) getCommonHeader() (http.Header, error) {
	header := http.Header{}

	// TODO: 接入租户信息

	return header, nil
}

// listAgentInfo list agent info.
func (c *cli) listAgentInfo(ctx context.Context, req *ListAgentInfoReq) (ListAgentInfoResp, error) {
	resp := new(BaseBroker[ListAgentInfoResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/cluster/list_agent_info").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list agent info failed, err: %v", err)
	}

	return resp.Data, nil
}

// listAgentState list agent state.
func (c *cli) listAgentState(ctx context.Context, req *ListAgentStateReq) (ListAgentStateResp, error) {
	resp := new(BaseBroker[ListAgentStateResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/cluster/list_agent_state").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list agent info failed, err: %v", err)
	}

	return resp.Data, nil
}

// asyncExecuteScript async execute script.
func (c *cli) asyncExecuteScript(ctx context.Context, req *AsyncExecuteScriptReq) (*AsyncExecuteScriptResp, error) {
	resp := new(BaseBroker[*AsyncExecuteScriptResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/async_execute_script").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("async execute script failed, err: %v", err)
	}

	return resp.Data, nil
}

// getExecuteScriptResult get execute script result.
func (c *cli) getExecuteScriptResult(ctx context.Context, req *GetExecuteScriptResultReq) (
	*GetExecuteScriptResultResp, error) {

	resp := new(BaseBroker[*GetExecuteScriptResultResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/get_execute_script_result").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("get execute script result failed, err: %v", err)
	}

	return resp.Data, nil
}

// asyncTerminateExecuteScript async terminate execute script.
func (c *cli) asyncTerminateExecuteScript(ctx context.Context, req *AsyncTerminateExecuteScriptReq) (
	*AsyncTerminateExecuteScriptResp, error) {

	resp := new(BaseBroker[*AsyncTerminateExecuteScriptResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/async_terminate_execute_script").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("async terminate execute script failed, err: %v", err)
	}

	return resp.Data, nil
}

// asyncTransferFile async transfer file.
func (c *cli) asyncTransferFile(ctx context.Context, req *AsyncTransferFileReq) (*AsyncTransferFileResp, error) {
	resp := new(BaseBroker[*AsyncTransferFileResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/async_transfer_file").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("async transfer file failed, err: %v", err)
	}

	return resp.Data, nil
}

// getTransferFileResult get transfer file result.
func (c *cli) getTransferFileResult(ctx context.Context, req *GetTransferFileResultReq) (
	*GetTransferFileResultResp, error) {

	resp := new(BaseBroker[*GetTransferFileResultResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/get_transfer_file_result").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("get transfer file result failed, err: %v", err)
	}

	return resp.Data, nil
}

// asyncTerminateTransferFile async terminate transfer file.
func (c *cli) asyncTerminateTransferFile(ctx context.Context, req *AsyncTerminateTransferFileReq) (
	*AsyncTerminateTransferFileResp, error) {

	resp := new(BaseBroker[*AsyncTerminateTransferFileResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/async_terminate_transfer_file").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("async transfer file failed, err: %v", err)
	}

	return resp.Data, nil
}

func (c *cli) operateAgent(ctx context.Context, req *OperateAgentReq) (*OperateAgentResp, error) {
	resp := new(BaseBroker[*OperateAgentResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/cluster/operate_agent").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("operate agent failed, err: %v", err)
	}

	return resp.Data, nil
}
