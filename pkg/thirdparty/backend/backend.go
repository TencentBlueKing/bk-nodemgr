/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package backend provides handlers to operate nodeman backend api.
// nolint:dupl
package backend

import (
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

// Config the config of backend.
type Config struct {
	APIGWAppConfig apigwclient.AppConfig
}

// Validate the config.
func (conf *Config) Validate() error {
	if err := conf.APIGWAppConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate backend client config: %v", err)
	}

	return nil
}

// cli client for backend.
type cli struct {
	client restclient.IClient
	config Config
}

// newClient initialize a new backend client.
func newClient(c *restclient.Capability, conf Config) (*cli, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("failed to new backend client: %v", err)
	}

	restCli, err := restclient.NewClient(c, "/api/v3", restclient.WithSensitiveHeader(apigwheader.BKGWAuthKey))
	if err != nil {
		return nil, fmt.Errorf("failed to new backend client: %v", err)
	}

	return &cli{
		client: restCli,
		config: conf,
	}, nil
}

// getHeader get backend common header.
// nolint: unparam
func (c *cli) getHeader(ctx contextx.ITenantUserContext) (http.Header, error) {
	header := http.Header{}
	header.Set(restheader.BKTenantIDKey, ctx.TenantID())
	header.Set(apigwheader.BKGWRIDKey, identifier.GenRequestID())

	// backend apigw open the user auth.
	userConfig := apigwclient.UserConfig{
		AppConfig: c.config.APIGWAppConfig,
		User:      ctx.LoginName(),
	}
	header.Set(apigwheader.BKGWAuthKey, userConfig.GetAuthHeader())

	return header, nil
}

func (c *cli) listBusiness(ctx contextx.ITenantUserContext, req *protoBackend.TopoBusinessListReq,
) (*protoBackend.TopoBusinessListResp_Data, error) {

	resp := new(protoBackend.TopoBusinessListResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/business/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list business failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("list business failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) listHost(ctx contextx.ITenantUserContext, req *protoBackend.TopoHostListReq,
) (*protoBackend.TopoHostListResp, error) {

	resp := new(protoBackend.TopoHostListResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/host/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list host failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list host failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctHost(ctx contextx.ITenantUserContext, req *protoBackend.TopoHostDistinctReq,
) (*protoBackend.TopoHostDistinctResp, error) {

	resp := new(protoBackend.TopoHostDistinctResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/host/distinct").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("distinct host failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("distinct host failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) createNetworkArea(ctx contextx.ITenantUserContext, req *protoBackend.TopoNetworkAreaCreateReq,
) (*protoBackend.TopoNetworkAreaCreateResp, error) {

	resp := new(protoBackend.TopoNetworkAreaCreateResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/networkarea/create").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("create networkarea failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("create networkarea failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) updateNetworkArea(ctx contextx.ITenantUserContext, req *protoBackend.TopoNetworkAreaUpdateReq,
) (*protoBackend.TopoNetworkAreaUpdateResp, error) {

	resp := new(protoBackend.TopoNetworkAreaUpdateResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/networkarea/update").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("update networkarea failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("update networkarea failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNetworkArea(ctx contextx.ITenantUserContext, req *protoBackend.TopoNetworkAreaListReq,
) (*protoBackend.TopoNetworkAreaListResp_Data, error) {

	resp := new(protoBackend.TopoNetworkAreaListResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/networkarea/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list networkarea failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	data := resp.GetData()
	if resp.GetData() == nil {
		return nil, fmt.Errorf("list networkarea failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) getNetworkArea(ctx contextx.ITenantUserContext, req *protoBackend.TopoNetworkAreaGetReq,
) (*protoBackend.TopoNetworkAreaGetResp, error) {

	resp := new(protoBackend.TopoNetworkAreaGetResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/networkarea/get").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get networkarea failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("get networkarea failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) deleteNetworkArea(ctx contextx.ITenantUserContext, req *protoBackend.TopoNetworkAreaDeleteReq,
) (*protoBackend.TopoNetworkAreaDeleteResp, error) {

	resp := new(protoBackend.TopoNetworkAreaDeleteResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/networkarea/delete").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("delete networkarea failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("delete networkarea failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) createNetworkUnit(ctx contextx.ITenantUserContext, req *protoBackend.TopoNetworkUnitCreateReq,
) (*protoBackend.TopoNetworkUnitCreateResp, error) {

	resp := new(protoBackend.TopoNetworkUnitCreateResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/networkunit/create").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("create networkunit failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("create networkunit failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) updateNetworkUnit(ctx contextx.ITenantUserContext, req *protoBackend.TopoNetworkUnitUpdateReq,
) (*protoBackend.TopoNetworkUnitUpdateResp, error) {

	resp := new(protoBackend.TopoNetworkUnitUpdateResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/networkunit/update").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("update networkunit failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("update networkunit failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getNetworkUnit(ctx contextx.ITenantUserContext, req *protoBackend.TopoNetworkUnitGetReq,
) (*protoBackend.TopoNetworkUnitGetResp, error) {

	resp := new(protoBackend.TopoNetworkUnitGetResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/networkunit/get").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get networkunit failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("get networkunit failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNetworkUnit(ctx contextx.ITenantUserContext, req *protoBackend.TopoNetworkUnitListReq,
) (*protoBackend.TopoNetworkUnitListResp, error) {

	resp := new(protoBackend.TopoNetworkUnitListResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/networkunit/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list networkunit failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list networkunit failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) deleteNetworkUnit(ctx contextx.ITenantUserContext, req *protoBackend.TopoNetworkUnitDeleteReq,
) (*protoBackend.TopoNetworkUnitDeleteResp, error) {

	resp := new(protoBackend.TopoNetworkUnitDeleteResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/networkunit/delete").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("delete networkunit failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("delete networkunit failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listTopoEvent(ctx contextx.ITenantUserContext, req *protoBackend.TopoEventListReq,
) (*protoBackend.TopoEventListResp, error) {

	resp := new(protoBackend.TopoEventListResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/event/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list topoevent failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list topoevent failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctTopoEvent(ctx contextx.ITenantUserContext, req *protoBackend.TopoEventDistinctReq,
) (*protoBackend.TopoEventDistinctResp, error) {

	resp := new(protoBackend.TopoEventDistinctResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/event/distinct").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("distinct topoevent failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("distinct topoevent failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listAccessPoint(ctx contextx.ITenantUserContext, req *protoBackend.TopoAccessPointListReq,
) (*protoBackend.TopoAccessPointListResp, error) {

	resp := new(protoBackend.TopoAccessPointListResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/accesspoint/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list accesspoint failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list accesspoint failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getConstant(ctx contextx.ITenantUserContext, req *protoBackend.TopoConstantGetReq,
) (*protoBackend.TopoConstantGetResp, error) {

	resp := new(protoBackend.TopoConstantGetResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/topo/constant/get").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get constant failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("get constant failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNodeWorkflow(ctx contextx.ITenantUserContext, req *protoBackend.NodeWorkflowListReq,
) (*protoBackend.NodeWorkflowListResp, error) {

	resp := new(protoBackend.NodeWorkflowListResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/node/workflow/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list workflow failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list workflow failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctNodeWorkflow(ctx contextx.ITenantUserContext, req *protoBackend.NodeWorkflowDistinctReq,
) (*protoBackend.NodeWorkflowDistinctResp, error) {

	resp := new(protoBackend.NodeWorkflowDistinctResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/node/workflow/distinct").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("distinct workflow failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("distinct workflow failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNodeWorkflowOperation(ctx contextx.ITenantUserContext, req *protoBackend.NodeWorkflowOperationListReq,
) (*protoBackend.NodeWorkflowOperationListResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationListResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/node/workflow/operation/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list workflow operation failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list workflow operation failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNodeWorkflowOperationInstance(
	ctx contextx.ITenantUserContext,
	req *protoBackend.NodeWorkflowOperationInstanceListReq,
) (*protoBackend.NodeWorkflowOperationInstanceListResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationInstanceListResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/node/workflow/operation/instance/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list workflow operation instance failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list workflow operation instance failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getOperationInstanceLog(
	ctx contextx.ITenantUserContext, req *protoBackend.NodeWorkflowOperationInstanceLogGetReq,
) (*protoBackend.NodeWorkflowOperationInstanceLogGetResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationInstanceLogGetResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/node/workflow/operation/instance/log/get").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get workflow operation instance logs failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("get workflow operation instance logs failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNodeWorkflowOpInstanceStatus(
	ctx contextx.ITenantUserContext,
	req *protoBackend.NodeWorkflowOperationInstanceListStatusReq,
) (*protoBackend.NodeWorkflowOperationInstanceListStatusResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationInstanceListStatusResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/node/workflow/operation/instance/status/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get workflow operation instance logs failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("get workflow operation instance logs failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) installNodeAgent(ctx contextx.ITenantUserContext, req *protoBackend.NodeAgentInstallReq,
) (*protoBackend.NodeAgentInstallResp, error) {

	resp := new(protoBackend.NodeAgentInstallResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/node/agent/install").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("install node agent failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("install node agent failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) retryOperation(ctx contextx.ITenantUserContext, req *protoBackend.NodeWorkflowOperationRetryReq,
) (*protoBackend.NodeWorkflowOperationRetryResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationRetryResp)
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/node/workflow/operation/retry").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("retry operation failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("retry operation failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listRelease(ctx contextx.ITenantUserContext, req *protoBackend.PackageReleaseListReq,
) (*protoBackend.PackageReleaseListResp, error) {

	resp := new(protoBackend.PackageReleaseListResp)
	err := c.client.Post().
		SubResourcef("/package/release/list").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list release failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list release failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctRelease(ctx contextx.ITenantUserContext, req *protoBackend.PackageReleaseDistinctReq) (
	*protoBackend.PackageReleaseDistinctResp, error) {

	resp := new(protoBackend.PackageReleaseDistinctResp)
	err := c.client.Post().
		SubResourcef("/package/release/distinct").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("distinct release failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("distinct release failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) setReleaseLabels(ctx contextx.ITenantUserContext, req *protoBackend.PackageReleaseSetLabelsReq) error {
	resp := new(protoBackend.PackageReleaseSetLabelsResp)
	err := c.client.Post().
		SubResourcef("/package/release/set_labels").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("set release labels failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) enableRelease(ctx contextx.ITenantUserContext, req *protoBackend.PackageReleaseEnableReq) error {
	resp := new(protoBackend.PackageReleaseEnableResp)
	err := c.client.Post().
		SubResourcef("/package/release/enable").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("enable release failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) disableRelease(ctx contextx.ITenantUserContext, req *protoBackend.PackageReleaseDisableReq) error {
	resp := new(protoBackend.PackageReleaseDisableResp)
	err := c.client.Post().
		SubResourcef("/package/release/disable").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("disable release failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setAsDefaultRelease(ctx contextx.ITenantUserContext, req *protoBackend.PackageReleaseSetAsDefaultReq) error {
	resp := new(protoBackend.PackageReleaseSetAsDefaultResp)
	err := c.client.Post().
		SubResourcef("/package/release/set_as_default").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("set release as default failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) cancelAsDefaultRelease(ctx contextx.ITenantUserContext, req *protoBackend.PackageReleaseCancelAsDefaultReq,
) error {

	resp := new(protoBackend.PackageReleaseCancelAsDefaultResp)
	err := c.client.Post().
		SubResourcef("/package/release/cancel_as_default").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("cancel release as default failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) deleteRelease(ctx contextx.ITenantUserContext, req *protoBackend.PackageReleaseDeleteReq) error {
	resp := new(protoBackend.PackageReleaseDeleteResp)
	err := c.client.Post().
		SubResourcef("/package/release/delete").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("delete release failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) listConfigPolicy(ctx contextx.ITenantUserContext, req *protoBackend.ConfigPolicyListReq) (
	*protoBackend.ConfigPolicyListResp, error) {

	resp := new(protoBackend.ConfigPolicyListResp)
	err := c.client.Post().
		SubResourcef("/policy/config/list").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list config policy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getConfigPolicy(ctx contextx.ITenantUserContext, req *protoBackend.ConfigPolicyGetReq) (
	*protoBackend.ConfigPolicyGetResp, error) {

	resp := new(protoBackend.ConfigPolicyGetResp)
	err := c.client.Post().
		SubResourcef("/policy/config/get").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get config policy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) createConfigPolicy(ctx contextx.ITenantUserContext, req *protoBackend.ConfigPolicyCreateReq) (
	*protoBackend.ConfigPolicyCreateResp, error) {

	resp := new(protoBackend.ConfigPolicyCreateResp)
	err := c.client.Post().
		SubResourcef("/policy/config/create").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("create config policy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) updateConfigPolicy(ctx contextx.ITenantUserContext, req *protoBackend.ConfigPolicyUpdateReq) (
	*protoBackend.ConfigPolicyUpdateResp, error) {

	resp := new(protoBackend.ConfigPolicyUpdateResp)
	err := c.client.Post().
		SubResourcef("/policy/config/update").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("update config policy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) enableConfigPolicy(ctx contextx.ITenantUserContext, req *protoBackend.ConfigPolicyEnableReq) (
	*protoBackend.ConfigPolicyEnableResp, error) {

	resp := new(protoBackend.ConfigPolicyEnableResp)
	err := c.client.Post().
		SubResourcef("/policy/config/enable").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("enable config policy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) disableConfigPolicy(ctx contextx.ITenantUserContext, req *protoBackend.ConfigPolicyDisableReq) (
	*protoBackend.ConfigPolicyDisableResp, error) {

	resp := new(protoBackend.ConfigPolicyDisableResp)
	err := c.client.Post().
		SubResourcef("/policy/config/disable").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("disable config policy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) deleteConfigPolicy(ctx contextx.ITenantUserContext, req *protoBackend.ConfigPolicyDeleteReq) (
	*protoBackend.ConfigPolicyDeleteResp, error) {

	resp := new(protoBackend.ConfigPolicyDeleteResp)
	err := c.client.Post().
		SubResourcef("/policy/config/delete").
		WithContext(ctx).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("delete config policy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}
