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
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
)

// Config the config of gse.
type Config struct {
	APIGWUserConfig apigwclient.UserConfig
}

// Validate the config.
func (conf *Config) Validate() error {
	if err := conf.APIGWUserConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate gse config: %v", err)
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

	restCli, err := apigwclient.NewClient(c, "/api/v2", conf.APIGWUserConfig)
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
func (c *cli) listAgentInfo(nCtx contextx.IContext, req *ListAgentInfoReq) (ListAgentInfoResp, error) {
	resp := new(BaseBroker[ListAgentInfoResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/cluster/list_agent_info").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list agent info failed: %v", err)
	}

	return resp.Data, nil
}

// listAgentState list agent state.
func (c *cli) listAgentState(nCtx contextx.IContext, req *ListAgentStateReq) (ListAgentStateResp, error) {
	resp := new(BaseBroker[ListAgentStateResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/cluster/list_agent_state").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list agent info failed: %v", err)
	}

	return resp.Data, nil
}

// asyncExecuteScript async execute script.
func (c *cli) asyncExecuteScript(nCtx contextx.IContext, req *AsyncExecuteScriptReq) (*AsyncExecuteScriptResp, error) {
	resp := new(BaseBroker[*AsyncExecuteScriptResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/async_execute_script").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("async execute script failed: %v", err)
	}

	return resp.Data, nil
}

// getExecuteScriptResult get execute script result.
func (c *cli) getExecuteScriptResult(nCtx contextx.IContext, req *GetExecuteScriptResultReq) (
	*GetExecuteScriptResultResp, error) {

	resp := new(BaseBroker[*GetExecuteScriptResultResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/get_execute_script_result").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("get execute script result failed: %v", err)
	}

	return resp.Data, nil
}

// asyncTerminateExecuteScript async terminate execute script.
func (c *cli) asyncTerminateExecuteScript(nCtx contextx.IContext, req *AsyncTerminateExecuteScriptReq) (
	*AsyncTerminateExecuteScriptResp, error) {

	resp := new(BaseBroker[*AsyncTerminateExecuteScriptResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/async_terminate_execute_script").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("async terminate execute script failed: %v", err)
	}

	return resp.Data, nil
}

// asyncTransferFile async transfer file.
func (c *cli) asyncTransferFile(nCtx contextx.IContext, req *AsyncTransferFileReq) (*AsyncTransferFileResp, error) {
	resp := new(BaseBroker[*AsyncTransferFileResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/async_transfer_file").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("async transfer file failed: %v", err)
	}

	return resp.Data, nil
}

// getTransferFileResult get transfer file result.
func (c *cli) getTransferFileResult(nCtx contextx.IContext, req *GetTransferFileResultReq) (
	*GetTransferFileResultResp, error) {

	resp := new(BaseBroker[*GetTransferFileResultResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/get_transfer_file_result").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("get transfer file result failed: %v", err)
	}

	return resp.Data, nil
}

// asyncTerminateTransferFile async terminate transfer file.
func (c *cli) asyncTerminateTransferFile(nCtx contextx.IContext, req *AsyncTerminateTransferFileReq) (
	*AsyncTerminateTransferFileResp, error) {

	resp := new(BaseBroker[*AsyncTerminateTransferFileResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/task/extensions/async_terminate_transfer_file").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("async transfer file failed: %v", err)
	}

	return resp.Data, nil
}

func (c *cli) operateAgent(nCtx contextx.IContext, req *OperateAgentReq) (*OperateAgentResp, error) {
	resp := new(BaseBroker[*OperateAgentResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/cluster/operate_agent").
		WithContext(nCtx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("operate agent failed: %v", err)
	}

	return resp.Data, nil
}
