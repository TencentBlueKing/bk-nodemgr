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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
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

	restCli, err := restclient.NewClient(c, "/api/v3",
		restclient.WithCustomHeaderMasker(apigwheader.BKGWAuthKey, apigwclient.AuthHeaderMasker))
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
func (c *cli) getHeader(ctx contextx.IContext) http.Header {
	header := http.Header{}
	header.Set(restheader.BKTenantIDKey, ctx.TenantID())
	header.Set(apigwheader.BKGWRIDKey, identifier.GenRequestID())

	// backend apigw open the user auth.
	userConfig := apigwclient.UserConfig{
		AppConfig:  c.config.APIGWAppConfig,
		BKUsername: ctx.BKUsername(),
	}
	header.Set(apigwheader.BKGWAuthKey, userConfig.GetAuthHeader())

	return header
}

func (c *cli) listBusiness(ctx contextx.IContext, req *protoBackend.TopoBusinessListReq,
) (*protoBackend.TopoBusinessListResp_Data, error) {

	resp := new(protoBackend.TopoBusinessListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/business/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list business failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("list business failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) listHost(ctx contextx.IContext, req *protoBackend.TopoHostListReq,
) (*protoBackend.TopoHostListResp, error) {

	resp := new(protoBackend.TopoHostListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/host/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list host failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list host failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getHostDistributionByNodeRole(ctx contextx.IContext, req *protoBackend.TopoGetHostDistributionByNodeRoleReq,
) (*protoBackend.TopoGetHostDistributionByNodeRoleResp, error) {

	resp := new(protoBackend.TopoGetHostDistributionByNodeRoleResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/host/get_host_distribution_by_node_role").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to get host distribution by node role. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to get host distribution by node role, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getHostDistributionByNetworkAreaID(ctx contextx.IContext, req *protoBackend.TopoGetHostDistributionByNetworkAreaIDReq,
) (*protoBackend.TopoGetHostDistributionByNetworkAreaIDResp, error) {

	resp := new(protoBackend.TopoGetHostDistributionByNetworkAreaIDResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/host/get_host_distribution_by_networkarea_id").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("failed to get host distribution by network area id. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("failed to get host distribution by network area id, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctHost(ctx contextx.IContext, req *protoBackend.TopoHostDistinctReq,
) (*protoBackend.TopoHostDistinctResp, error) {

	resp := new(protoBackend.TopoHostDistinctResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/host/distinct").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("distinct host failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("distinct host failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) createNetworkArea(ctx contextx.IContext, req *protoBackend.TopoNetworkAreaCreateReq,
) (*protoBackend.TopoNetworkAreaCreateResp, error) {

	resp := new(protoBackend.TopoNetworkAreaCreateResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/networkarea/create").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("create networkarea failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("create networkarea failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) updateNetworkArea(ctx contextx.IContext, req *protoBackend.TopoNetworkAreaUpdateReq,
) (*protoBackend.TopoNetworkAreaUpdateResp, error) {

	resp := new(protoBackend.TopoNetworkAreaUpdateResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/networkarea/update").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("update networkarea failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("update networkarea failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNetworkArea(ctx contextx.IContext, req *protoBackend.TopoNetworkAreaListReq,
) (*protoBackend.TopoNetworkAreaListResp_Data, error) {

	resp := new(protoBackend.TopoNetworkAreaListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/networkarea/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list networkarea failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	data := resp.GetData()
	if data == nil {
		return nil, fmt.Errorf("list networkarea failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return data, nil
}

func (c *cli) getNetworkArea(ctx contextx.IContext, req *protoBackend.TopoNetworkAreaGetReq,
) (*protoBackend.TopoNetworkAreaGetResp, error) {

	resp := new(protoBackend.TopoNetworkAreaGetResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/networkarea/get").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get networkarea failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("get networkarea failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) deleteNetworkArea(ctx contextx.IContext, req *protoBackend.TopoNetworkAreaDeleteReq,
) (*protoBackend.TopoNetworkAreaDeleteResp, error) {

	resp := new(protoBackend.TopoNetworkAreaDeleteResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/networkarea/delete").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("delete networkarea failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("delete networkarea failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) createNetworkUnit(ctx contextx.IContext, req *protoBackend.TopoNetworkUnitCreateReq,
) (*protoBackend.TopoNetworkUnitCreateResp, error) {

	resp := new(protoBackend.TopoNetworkUnitCreateResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/networkunit/create").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("create networkunit failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("create networkunit failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) updateNetworkUnit(ctx contextx.IContext, req *protoBackend.TopoNetworkUnitUpdateReq,
) (*protoBackend.TopoNetworkUnitUpdateResp, error) {

	resp := new(protoBackend.TopoNetworkUnitUpdateResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/networkunit/update").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("update networkunit failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("update networkunit failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getNetworkUnit(ctx contextx.IContext, req *protoBackend.TopoNetworkUnitGetReq,
) (*protoBackend.TopoNetworkUnitGetResp, error) {

	resp := new(protoBackend.TopoNetworkUnitGetResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/networkunit/get").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get networkunit failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("get networkunit failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNetworkUnit(ctx contextx.IContext, req *protoBackend.TopoNetworkUnitListReq,
) (*protoBackend.TopoNetworkUnitListResp, error) {

	resp := new(protoBackend.TopoNetworkUnitListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/networkunit/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list networkunit failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list networkunit failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) deleteNetworkUnit(ctx contextx.IContext, req *protoBackend.TopoNetworkUnitDeleteReq,
) (*protoBackend.TopoNetworkUnitDeleteResp, error) {

	resp := new(protoBackend.TopoNetworkUnitDeleteResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/networkunit/delete").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("delete networkunit failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("delete networkunit failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listTopoEvent(ctx contextx.IContext, req *protoBackend.TopoEventListReq,
) (*protoBackend.TopoEventListResp, error) {

	resp := new(protoBackend.TopoEventListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/event/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list topoevent failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list topoevent failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctTopoEvent(ctx contextx.IContext, req *protoBackend.TopoEventDistinctReq,
) (*protoBackend.TopoEventDistinctResp, error) {

	resp := new(protoBackend.TopoEventDistinctResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/event/distinct").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("distinct topoevent failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("distinct topoevent failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listAccessPoint(ctx contextx.IContext, req *protoBackend.TopoAccessPointListReq,
) (*protoBackend.TopoAccessPointListResp, error) {

	resp := new(protoBackend.TopoAccessPointListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/accesspoint/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list accesspoint failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list accesspoint failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getConstant(ctx contextx.IContext, req *protoBackend.TopoConstantGetReq,
) (*protoBackend.TopoConstantGetResp, error) {

	resp := new(protoBackend.TopoConstantGetResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/constant/get").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get constant failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("get constant failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNodeWorkflow(ctx contextx.IContext, req *protoBackend.NodeWorkflowListReq,
) (*protoBackend.NodeWorkflowListResp, error) {

	resp := new(protoBackend.NodeWorkflowListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/workflow/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list workflow failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list workflow failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctNodeWorkflow(ctx contextx.IContext, req *protoBackend.NodeWorkflowDistinctReq,
) (*protoBackend.NodeWorkflowDistinctResp, error) {

	resp := new(protoBackend.NodeWorkflowDistinctResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/workflow/distinct").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("distinct workflow failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("distinct workflow failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNodeWorkflowOperation(ctx contextx.IContext, req *protoBackend.NodeWorkflowOperationListReq,
) (*protoBackend.NodeWorkflowOperationListResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/workflow/operation/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list workflow operation failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list workflow operation failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNodeWorkflowOperationInstance(
	ctx contextx.IContext,
	req *protoBackend.NodeWorkflowOperationInstanceListReq,
) (*protoBackend.NodeWorkflowOperationInstanceListResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationInstanceListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/workflow/operation/instance/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list workflow operation instance failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list workflow operation instance failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getOperationInstanceLog(
	ctx contextx.IContext, req *protoBackend.NodeWorkflowOperationInstanceLogGetReq,
) (*protoBackend.NodeWorkflowOperationInstanceLogGetResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationInstanceLogGetResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/workflow/operation/instance/log/get").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get workflow operation instance logs failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("get workflow operation instance logs failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNodeWorkflowOpInstanceStatus(
	ctx contextx.IContext,
	req *protoBackend.NodeWorkflowOperationInstanceListStatusReq,
) (*protoBackend.NodeWorkflowOperationInstanceListStatusResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationInstanceListStatusResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/workflow/operation/instance/status/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get workflow operation instance logs failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("get workflow operation instance logs failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) installNodeAgent(ctx contextx.IContext, req *protoBackend.NodeAgentInstallReq,
) (*protoBackend.NodeAgentInstallResp, error) {

	resp := new(protoBackend.NodeAgentInstallResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/agent/install").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("install node agent failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("install node agent failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) upgradeNodeAgent(ctx contextx.IContext, req *protoBackend.NodeAgentUpgradeReq) (
	*protoBackend.NodeAgentUpgradeResp, error) {

	resp := new(protoBackend.NodeAgentUpgradeResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/agent/upgrade").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("upgrade node agent failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("upgrade node agent failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) restartNodeAgent(ctx contextx.IContext, req *protoBackend.NodeAgentRestartReq) (
	*protoBackend.NodeAgentRestartResp, error) {

	resp := new(protoBackend.NodeAgentRestartResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/agent/restart").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("restart node agent failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("restart node agent failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) reconfigNodeAgent(ctx contextx.IContext, req *protoBackend.NodeAgentReconfigReq) (
	*protoBackend.NodeAgentReconfigResp, error) {

	resp := new(protoBackend.NodeAgentReconfigResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/agent/reconfig").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("reconfig node agent failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("reconfig node agent failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) uninstallNodeAgent(ctx contextx.IContext, req *protoBackend.NodeAgentUninstallReq) (
	*protoBackend.NodeAgentUninstallResp, error) {

	resp := new(protoBackend.NodeAgentUninstallResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/agent/uninstall").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("uninstall node agent failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("uninstall node agent failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) checkAgentInstall(ctx contextx.IContext, req *protoBackend.NodeAgentInstallCheckReq,
) (*protoBackend.NodeAgentInstallCheckResp, error) {

	resp := new(protoBackend.NodeAgentInstallCheckResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/agent/install_check").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("check agent install node failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("check agent install node failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) retryOperation(ctx contextx.IContext, req *protoBackend.NodeWorkflowOperationRetryReq,
) (*protoBackend.NodeWorkflowOperationRetryResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationRetryResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/workflow/operation/retry").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("retry operation failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) terminateOperation(ctx contextx.IContext, req *protoBackend.NodeWorkflowOperationTerminateReq,
) (*protoBackend.NodeWorkflowOperationTerminateResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationTerminateResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/workflow/operation/terminate").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("terminate operation instance failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return resp, nil
}
func (c *cli) installNodeProxy(ctx contextx.IContext, req *protoBackend.NodeProxyInstallReq) (
	*protoBackend.NodeProxyInstallResp, error) {

	resp := new(protoBackend.NodeProxyInstallResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/proxy/install").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("install node proxy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("install node proxy failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) upgradeNodeProxy(ctx contextx.IContext, req *protoBackend.NodeProxyUpgradeReq) (
	*protoBackend.NodeProxyUpgradeResp, error) {

	resp := new(protoBackend.NodeProxyUpgradeResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/proxy/upgrade").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("upgrade node proxy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("upgrade node proxy failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) restartNodeProxy(ctx contextx.IContext, req *protoBackend.NodeProxyRestartReq) (
	*protoBackend.NodeProxyRestartResp, error) {

	resp := new(protoBackend.NodeProxyRestartResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/proxy/restart").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("restart node proxy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("restart node proxy failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) reconfigNodeProxy(ctx contextx.IContext, req *protoBackend.NodeProxyReconfigReq) (
	*protoBackend.NodeProxyReconfigResp, error) {

	resp := new(protoBackend.NodeProxyReconfigResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/proxy/reconfig").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("reconfig node proxy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("reconfig node proxy failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) updateNodeProxy(ctx contextx.IContext, req *protoBackend.NodeProxyUpdateReq) (
	*protoBackend.NodeProxyUpdateResp, error) {

	resp := new(protoBackend.NodeProxyUpdateResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/proxy/update").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("update node proxy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("update node proxy failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) uninstallNodeProxy(ctx contextx.IContext, req *protoBackend.NodeProxyUninstallReq) (
	*protoBackend.NodeProxyUninstallResp, error) {

	resp := new(protoBackend.NodeProxyUninstallResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/proxy/uninstall").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("uninstall node proxy failed. code(%d), message(%s), request-id(%s)",
			code, resp.GetMessage(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("uninstall node proxy failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listRelease(ctx contextx.IContext, req *protoBackend.PackageReleaseListReq,
) (*protoBackend.PackageReleaseListResp, error) {

	resp := new(protoBackend.PackageReleaseListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list release failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listReleaseAgent(ctx contextx.IContext, req *protoBackend.PackageReleaseAgentListReq,
) (*protoBackend.PackageReleaseAgentListResp, error) {

	resp := new(protoBackend.PackageReleaseAgentListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release_agent/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list release agent failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list release agent failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listReleaseProxy(ctx contextx.IContext, req *protoBackend.PackageReleaseProxyListReq,
) (*protoBackend.PackageReleaseProxyListResp, error) {

	resp := new(protoBackend.PackageReleaseProxyListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release_proxy/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list release proxy failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list release proxy failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctRelease(ctx contextx.IContext, req *protoBackend.PackageReleaseDistinctReq) (
	*protoBackend.PackageReleaseDistinctResp, error) {

	resp := new(protoBackend.PackageReleaseDistinctResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/distinct").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("distinct release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("distinct release failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) setReleaseLabels(ctx contextx.IContext, req *protoBackend.PackageReleaseSetLabelsReq) error {
	resp := new(protoBackend.PackageReleaseSetLabelsResp)

	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/set_labels").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("set release labels failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setReleaseLabelsMany(ctx contextx.IContext, req *protoBackend.PackageReleaseSetLabelsManyReq) error {
	resp := new(protoBackend.PackageReleaseSetLabelsManyResp)

	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/set_labels_many").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("set many release labels failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) enableRelease(ctx contextx.IContext, req *protoBackend.PackageReleaseEnableReq) error {
	resp := new(protoBackend.PackageReleaseEnableResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/enable").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("enable release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) disableRelease(ctx contextx.IContext, req *protoBackend.PackageReleaseDisableReq) error {
	resp := new(protoBackend.PackageReleaseDisableResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/disable").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("disable release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setAsDefaultRelease(ctx contextx.IContext, req *protoBackend.PackageReleaseSetAsDefaultReq) error {
	resp := new(protoBackend.PackageReleaseSetAsDefaultResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/set_as_default").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("set release as default failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) cancelAsDefaultRelease(ctx contextx.IContext, req *protoBackend.PackageReleaseCancelAsDefaultReq,
) error {

	resp := new(protoBackend.PackageReleaseCancelAsDefaultResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/cancel_as_default").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("cancel release as default failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) deleteRelease(ctx contextx.IContext, req *protoBackend.PackageReleaseDeleteReq) error {
	resp := new(protoBackend.PackageReleaseDeleteResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/delete").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("delete release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) listReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginListReq,
) (*protoBackend.PackageReleasePluginListResp, error) {

	resp := new(protoBackend.PackageReleasePluginListResp)
	header := c.getHeader(ctx)

	result := c.client.Post().
		SubResourcef("/package/release_plugin/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do()
	logger.G.Biz(ctx).With("body", result.MaskResponseBody(), "url", result.FullURL).Info("get response data")

	if err := result.Into(resp); err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list release proxy failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list release proxy failed, get empty data. code(%d), message(%s), request-id(%s)",
		    resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) enableReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginEnableReq) error {
	resp := new(protoBackend.PackageReleasePluginEnableResp)
	header := c.getHeader(ctx)

	result := c.client.Post().
		SubResourcef("/package/release_plugin/enable").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do()
	logger.G.Biz(ctx).With("body", result.MaskResponseBody(), "url", result.FullURL).Info("get response data")

	if err := result.Into(resp); err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("enable release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) disableReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginDisableReq) error {
	resp := new(protoBackend.PackageReleasePluginDisableResp)
	header := c.getHeader(ctx)

	result := c.client.Post().
		SubResourcef("/package/release_plugin/disable").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do()
	logger.G.Biz(ctx).With("body", result.MaskResponseBody(), "url", result.FullURL).Info("get response data")

	if err := result.Into(resp); err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("disable release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setAsDefaultReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginSetAsDefaultReq) error {
	resp := new(protoBackend.PackageReleasePluginSetAsDefaultResp)
	header := c.getHeader(ctx)

	result := c.client.Post().
		SubResourcef("/package/release_plugin/set_as_default").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do()
	logger.G.Biz(ctx).With("body", result.MaskResponseBody(), "url", result.FullURL).Info("get response data")

	if err := result.Into(resp); err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("set release as default failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) cancelAsDefaultReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginCancelAsDefaultReq) error {
	resp := new(protoBackend.PackageReleasePluginCancelAsDefaultResp)
	header := c.getHeader(ctx)

	result := c.client.Post().
		SubResourcef("/package/release_plugin/cancel_as_default").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do()
	logger.G.Biz(ctx).With("body", result.MaskResponseBody(), "url", result.FullURL).Info("get response data")

	if err := result.Into(resp); err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("cancel release as default failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) deleteReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginDeleteReq) error {
	resp := new(protoBackend.PackageReleasePluginDeleteResp)
	header := c.getHeader(ctx)

	result := c.client.Post().
		SubResourcef("/package/release_plugin/delete").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do()
	logger.G.Biz(ctx).With("body", result.MaskResponseBody(), "url", result.FullURL).Info("get response data")

	if err := result.Into(resp); err != nil {
		return err
	}

	if code := resp.GetCode(); code != CodeOK {
		return fmt.Errorf("delete release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) listConfigPolicy(ctx contextx.IContext, req *protoBackend.ConfigPolicyListReq) (
	*protoBackend.ConfigPolicyListResp, error) {

	resp := new(protoBackend.ConfigPolicyListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/policy/config/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list config policy failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getConfigPolicy(ctx contextx.IContext, req *protoBackend.ConfigPolicyGetReq) (
	*protoBackend.ConfigPolicyGetResp, error) {

	resp := new(protoBackend.ConfigPolicyGetResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/policy/config/get").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get config policy failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) createConfigPolicy(ctx contextx.IContext, req *protoBackend.ConfigPolicyCreateReq) (
	*protoBackend.ConfigPolicyCreateResp, error) {

	resp := new(protoBackend.ConfigPolicyCreateResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/policy/config/create").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("create config policy failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) updateConfigPolicy(ctx contextx.IContext, req *protoBackend.ConfigPolicyUpdateReq) (
	*protoBackend.ConfigPolicyUpdateResp, error) {

	resp := new(protoBackend.ConfigPolicyUpdateResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/policy/config/update").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("update config policy failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) enableConfigPolicy(ctx contextx.IContext, req *protoBackend.ConfigPolicyEnableReq) (
	*protoBackend.ConfigPolicyEnableResp, error) {

	resp := new(protoBackend.ConfigPolicyEnableResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/policy/config/enable").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("enable config policy failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) disableConfigPolicy(ctx contextx.IContext, req *protoBackend.ConfigPolicyDisableReq) (
	*protoBackend.ConfigPolicyDisableResp, error) {

	resp := new(protoBackend.ConfigPolicyDisableResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/policy/config/disable").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("disable config policy failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) deleteConfigPolicy(ctx contextx.IContext, req *protoBackend.ConfigPolicyDeleteReq) (
	*protoBackend.ConfigPolicyDeleteResp, error) {

	resp := new(protoBackend.ConfigPolicyDeleteResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/policy/config/delete").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("delete config policy failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listPackageEvent(ctx contextx.IContext, req *protoBackend.PackageEventListReq,
) (*protoBackend.PackageEventListResp, error) {

	resp := new(protoBackend.PackageEventListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/event/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list package event failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list package event failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctPackageEvent(ctx contextx.IContext, req *protoBackend.PackageEventDistinctReq,
) (*protoBackend.PackageEventDistinctResp, error) {

	resp := new(protoBackend.PackageEventDistinctResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/event/distinct").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("distinct package event failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("distinct package event failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listConfigPolicyEvent(ctx contextx.IContext, req *protoBackend.ConfigPolicyEventListReq,
) (*protoBackend.ConfigPolicyEventListResp, error) {

	resp := new(protoBackend.ConfigPolicyEventListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/policy/config/event/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list config policy event failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list config policy event failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctConfigPolicyEvent(ctx contextx.IContext, req *protoBackend.ConfigPolicyEventDistinctReq,
) (*protoBackend.ConfigPolicyEventDistinctResp, error) {

	resp := new(protoBackend.ConfigPolicyEventDistinctResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/policy/config/event/distinct").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("distinct config policy event failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("distinct config policy event failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) installPlugin(ctx contextx.IContext, req *protoBackend.PluginInstallReq) (
	*protoBackend.PluginInstallResp, error) {

	resp := new(protoBackend.PluginInstallResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/install").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("install plugin failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("install plugin failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listPlugins(ctx contextx.IContext, req *protoBackend.PluginListReq) (
	*protoBackend.PluginListResp, error) {

	resp := new(protoBackend.PluginListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list plugins failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list plugins failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) applyPluginSubConfig(ctx contextx.IContext, req *protoBackend.PluginApplySubConfigReq) (
	*protoBackend.PluginApplySubConfigResp, error) {

	resp := new(protoBackend.PluginApplySubConfigResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/apply_subconfig").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("apply plugin sub config failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("apply plugin sub config failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listProcesses(ctx contextx.IContext, req *protoBackend.ProcessListReq) (
	*protoBackend.ProcessListResp, error) {

	resp := new(protoBackend.ProcessListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/process/list").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("list processes failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list processes failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getGraphNode(ctx contextx.IContext, req *protoBackend.TopoGraphNodeGetReq) (
	*protoBackend.TopoGraphNodeGetResp, error) {

	resp := new(protoBackend.TopoGraphNodeGetResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/graph_node/get").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		EnableLogResponse().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if code := resp.GetCode(); code != CodeOK {
		return nil, fmt.Errorf("get graph node failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("get graph node failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}
