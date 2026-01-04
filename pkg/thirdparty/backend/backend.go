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

// ===============================================================================
// Topo Related Interfaces
// ===============================================================================

func (c *cli) listBusiness(ctx contextx.IContext, req *protoBackend.TopoBusinessListReq) (*protoBackend.TopoBusinessListResp, error) {
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

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list business failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
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

func (c *cli) selectHostID(ctx contextx.IContext, req *protoBackend.TopoHostSelectHostIDReq,
) (*protoBackend.TopoHostSelectHostIDResp, error) {

	resp := new(protoBackend.TopoHostSelectHostIDResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/host/scenario/select_host_id").
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
		return nil, fmt.Errorf("select host id failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("select host id failed,get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) selectInnerIP(ctx contextx.IContext, req *protoBackend.TopoHostSelectInnerIPReq,
) (*protoBackend.TopoHostSelectInnerIPResp, error) {

	resp := new(protoBackend.TopoHostSelectInnerIPResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/host/scenario/select_inner_ip").
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
		return nil, fmt.Errorf("select inner ip failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("select inner ip failed,get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) selectInnerIPV6(ctx contextx.IContext, req *protoBackend.TopoHostSelectInnerIPV6Req,
) (*protoBackend.TopoHostSelectInnerIPV6Resp, error) {

	resp := new(protoBackend.TopoHostSelectInnerIPV6Resp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/host/scenario/select_inner_ipv6").
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
		return nil, fmt.Errorf("select inner ipv6 failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("select inner ipv6 failed,get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) selectNetWorkareaIDAndInnerIP(ctx contextx.IContext, req *protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPReq,
) (*protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPResp, error) {

	resp := new(protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/host/scenario/select_networkarea_id_and_inner_ip").
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
		return nil, fmt.Errorf("select networkarea id and inner ip failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("select networkarea id and inner ip failed,get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) selectNetWorkareaIDAndInnerIPV6(ctx contextx.IContext, req *protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPV6Req,
) (*protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPV6Resp, error) {

	resp := new(protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPV6Resp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/topo/host/scenario/select_networkarea_id_and_inner_ipv6").
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
		return nil, fmt.Errorf("select networkarea id and inner ipv6 failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("select networkarea id and inner ipv6 failed,get empty data. code(%d), message(%s), request-id(%s)",
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

func (c *cli) listNetworkArea(ctx contextx.IContext, req *protoBackend.TopoNetworkAreaListReq) (*protoBackend.TopoNetworkAreaListResp, error) {
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

	if resp.GetData() == nil {
		return nil, fmt.Errorf("list networkarea failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
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
			fmt.Errorf("list plugin workflows failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

// ===============================================================================
// Node Workflow Related Interfaces
// ===============================================================================

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

func (c *cli) getNodeWorkflowOperationInstanceLog(
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

func (c *cli) retryNodeWorkflowOperation(ctx contextx.IContext, req *protoBackend.NodeWorkflowOperationRetryReq,
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

func (c *cli) terminateNodeWorkflowOperation(ctx contextx.IContext, req *protoBackend.NodeWorkflowOperationTerminateReq,
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

func (c *cli) getNodeWorkflowManualInfo(
	ctx contextx.IContext, req *protoBackend.NodeWorkflowOperationManualInfoGetReq,
) (*protoBackend.NodeWorkflowOperationManualInfoGetResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationManualInfoGetResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/workflow/operation/manual/info/get").
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
		return nil, fmt.Errorf("get workflow operation manual info failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("get workflow operation manual info failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listNodeWorkflowOperationInstanceStatusDistribution(
	ctx contextx.IContext, req *protoBackend.NodeWorkflowOperationInstanceStatusDistributionListReq,
) (*protoBackend.NodeWorkflowOperationInstanceStatusDistributionListResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationInstanceStatusDistributionListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/node/workflow/operation/instance/status_distribution/list").
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
		return nil, fmt.Errorf("list operation instance status distribution failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list operation instance status distribution failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

// ===============================================================================
// Node Agent Related Interfaces
// ===============================================================================

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

func (c *cli) checkInstallAgent(ctx contextx.IContext, req *protoBackend.NodeAgentInstallCheckReq,
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
		return nil, fmt.Errorf("check install agent failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("check install agent failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

// ===============================================================================
// Node Proxy Related Interfaces
// ===============================================================================

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

// ===============================================================================
// Package Release Related Interfaces
// ===============================================================================

func (c *cli) listReleaseAgent(ctx contextx.IContext, req *protoBackend.PackageReleaseAgentListReq,
) (*protoBackend.PackageReleaseAgentListResp, error) {

	resp := new(protoBackend.PackageReleaseAgentListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/agent/list").
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
		return nil, fmt.Errorf("list agent release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list agent release failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctReleaseAgent(ctx contextx.IContext, req *protoBackend.PackageReleaseAgentDistinctReq) (
	*protoBackend.PackageReleaseAgentDistinctResp, error) {

	resp := new(protoBackend.PackageReleaseAgentDistinctResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/agent/distinct").
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
		return nil, fmt.Errorf("distinct agent release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("distinct agent release failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) setReleaseAgentLabelsMany(ctx contextx.IContext, req *protoBackend.PackageReleaseAgentSetLabelsManyReq) error {
	resp := new(protoBackend.PackageReleaseAgentSetLabelsManyResp)

	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/agent/set_labels_many").
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
		return fmt.Errorf("set many agent release labels failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) enableReleaseAgent(ctx contextx.IContext, req *protoBackend.PackageReleaseAgentEnableReq) error {
	resp := new(protoBackend.PackageReleaseAgentEnableResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/agent/enable").
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
		return fmt.Errorf("enable agent release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) disableReleaseAgent(ctx contextx.IContext, req *protoBackend.PackageReleaseAgentDisableReq) error {
	resp := new(protoBackend.PackageReleaseAgentDisableResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/agent/disable").
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
		return fmt.Errorf("disable agent release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setAsDefaultReleaseAgent(ctx contextx.IContext, req *protoBackend.PackageReleaseAgentSetAsDefaultReq) error {
	resp := new(protoBackend.PackageReleaseAgentSetAsDefaultResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/agent/set_as_default").
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
		return fmt.Errorf("set agent release as default failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) cancelAsDefaultReleaseAgent(ctx contextx.IContext, req *protoBackend.PackageReleaseAgentCancelAsDefaultReq,
) error {

	resp := new(protoBackend.PackageReleaseAgentCancelAsDefaultResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/agent/cancel_as_default").
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
		return fmt.Errorf("cancel agent release as default failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) deleteReleaseAgent(ctx contextx.IContext, req *protoBackend.PackageReleaseAgentDeleteReq) error {
	resp := new(protoBackend.PackageReleaseAgentDeleteResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/agent/delete").
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
		return fmt.Errorf("delete agent release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) listReleaseProxy(ctx contextx.IContext, req *protoBackend.PackageReleaseProxyListReq,
) (*protoBackend.PackageReleaseProxyListResp, error) {

	resp := new(protoBackend.PackageReleaseProxyListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/proxy/list").
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

func (c *cli) distinctReleaseProxy(ctx contextx.IContext, req *protoBackend.PackageReleaseProxyDistinctReq) (
	*protoBackend.PackageReleaseProxyDistinctResp, error) {

	resp := new(protoBackend.PackageReleaseProxyDistinctResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/proxy/distinct").
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
		return nil, fmt.Errorf("distinct release proxy failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("distinct release proxy failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) setReleaseProxyLabelsMany(ctx contextx.IContext, req *protoBackend.PackageReleaseProxySetLabelsManyReq) error {
	resp := new(protoBackend.PackageReleaseProxySetLabelsManyResp)

	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/proxy/set_labels_many").
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
		return fmt.Errorf("set many proxy release labels failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) enableReleaseProxy(ctx contextx.IContext, req *protoBackend.PackageReleaseProxyEnableReq) error {
	resp := new(protoBackend.PackageReleaseProxyEnableResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/proxy/enable").
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
		return fmt.Errorf("enable proxy release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) disableReleaseProxy(ctx contextx.IContext, req *protoBackend.PackageReleaseProxyDisableReq) error {
	resp := new(protoBackend.PackageReleaseProxyDisableResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/proxy/disable").
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
		return fmt.Errorf("disable proxy release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setAsDefaultReleaseProxy(ctx contextx.IContext, req *protoBackend.PackageReleaseProxySetAsDefaultReq) error {
	resp := new(protoBackend.PackageReleaseProxySetAsDefaultResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/proxy/set_as_default").
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
		return fmt.Errorf("set proxy release as default failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) cancelAsDefaultReleaseProxy(ctx contextx.IContext, req *protoBackend.PackageReleaseProxyCancelAsDefaultReq,
) error {

	resp := new(protoBackend.PackageReleaseProxyCancelAsDefaultResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/proxy/cancel_as_default").
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
		return fmt.Errorf("cancel proxy release as default failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) deleteReleaseProxy(ctx contextx.IContext, req *protoBackend.PackageReleaseProxyDeleteReq) error {
	resp := new(protoBackend.PackageReleaseProxyDeleteResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/proxy/delete").
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
		return fmt.Errorf("delete proxy release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) listReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginListReq,
) (*protoBackend.PackageReleasePluginListResp, error) {

	resp := new(protoBackend.PackageReleasePluginListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/plugin/list").
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
		return nil, fmt.Errorf("list release proxy failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) enableReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginEnableReq) error {
	resp := new(protoBackend.PackageReleasePluginEnableResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/plugin/enable").
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

func (c *cli) disableReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginDisableReq) error {
	resp := new(protoBackend.PackageReleasePluginDisableResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/plugin/disable").
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

func (c *cli) setAsDefaultReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginSetAsDefaultReq) error {
	resp := new(protoBackend.PackageReleasePluginSetAsDefaultResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/plugin/set_as_default").
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

func (c *cli) cancelAsDefaultReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginCancelAsDefaultReq) error {
	resp := new(protoBackend.PackageReleasePluginCancelAsDefaultResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/plugin/cancel_as_default").
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

func (c *cli) deleteReleasePlugin(ctx contextx.IContext, req *protoBackend.PackageReleasePluginDeleteReq) error {
	resp := new(protoBackend.PackageReleasePluginDeleteResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/plugin/delete").
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

func (c *cli) listReleaseCert(ctx contextx.IContext, req *protoBackend.PackageReleaseCertListReq,
) (*protoBackend.PackageReleaseCertListResp, error) {

	resp := new(protoBackend.PackageReleaseCertListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/cert/list").
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
		return nil, fmt.Errorf("list cert release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list cert release failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) deleteReleaseCert(ctx contextx.IContext, req *protoBackend.PackageReleaseCertDeleteReq) error {
	resp := new(protoBackend.PackageReleaseCertDeleteResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/cert/delete").
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
		return fmt.Errorf("delete cert release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) listReleaseBinTool(ctx contextx.IContext, req *protoBackend.PackageReleaseBinToolListReq,
) (*protoBackend.PackageReleaseBinToolListResp, error) {

	resp := new(protoBackend.PackageReleaseBinToolListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/bintool/list").
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
		return nil, fmt.Errorf("list bintool release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list bintool release failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) deleteReleaseBinTool(ctx contextx.IContext, req *protoBackend.PackageReleaseBinToolDeleteReq) error {
	resp := new(protoBackend.PackageReleaseBinToolDeleteResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/bintool/delete").
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
		return fmt.Errorf("delete bintool release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) listReleasePluginBinTool(ctx contextx.IContext, req *protoBackend.PackageReleasePluginBinToolListReq,
) (*protoBackend.PackageReleasePluginBinToolListResp, error) {

	resp := new(protoBackend.PackageReleasePluginBinToolListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/plugin_bintool/list").
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
		return nil, fmt.Errorf("list plugin bintool release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list plugin bintool release failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) deleteReleasePluginBinTool(ctx contextx.IContext, req *protoBackend.PackageReleasePluginBinToolDeleteReq) error {
	resp := new(protoBackend.PackageReleasePluginBinToolDeleteResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/package/release/plugin_bintool/delete").
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
		return fmt.Errorf("delete plugin bintool release failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

// ===============================================================================
// Policy Related Interfaces
// ===============================================================================

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

// ===============================================================================
// Package Event Related Interfaces
// ===============================================================================

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

// ===============================================================================
// Plugin Related Interfaces
// ===============================================================================

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

// ===============================================================================
// Process Related Interfaces
// ===============================================================================

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

func (c *cli) getProcessDistributionByHostID(ctx contextx.IContext, req *protoBackend.GetProcessDistributionByHostIDReq) (
	*protoBackend.GetProcessDistributionByHostIDResp, error) {

	resp := new(protoBackend.GetProcessDistributionByHostIDResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/process/get_distribution_by_host_id").
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
		return nil, fmt.Errorf("get process distribution by host id failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("get process distribution by host id failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getProcessDistributionByPluginName(ctx contextx.IContext, req *protoBackend.GetProcessDistributionByPluginNameReq) (
	*protoBackend.GetProcessDistributionByPluginNameResp, error) {

	resp := new(protoBackend.GetProcessDistributionByPluginNameResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/process/get_distribution_by_plugin_name").
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
		return nil, fmt.Errorf("get process distribution by plugin name failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("get process distribution by plugin name failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

// distinctProcess distinct process by conditions.
func (c *cli) distinctProcess(ctx contextx.IContext, req *protoBackend.ProcessDistinctReq) (
	*protoBackend.ProcessDistinctResp, error) {

	resp := new(protoBackend.ProcessDistinctResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/process/distinct").
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
		return nil, fmt.Errorf("distinct process failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil, fmt.Errorf("distinct process failed, get empty data. code(%d), message(%s), request-id(%s)",
			resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

// ===============================================================================
// Plugin Workflow Related Interfaces
// ===============================================================================

func (c *cli) listPluginWorkflows(ctx contextx.IContext, req *protoBackend.PluginWorkflowListReq) (
	*protoBackend.PluginWorkflowListResp, error) {

	resp := new(protoBackend.PluginWorkflowListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/workflow/list").
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
		return nil, fmt.Errorf("list plugin workflows failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list plugin workflows failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) distinctPluginWorkflows(ctx contextx.IContext, req *protoBackend.PluginWorkflowDistinctReq) (
	*protoBackend.PluginWorkflowDistinctResp, error) {

	resp := new(protoBackend.PluginWorkflowDistinctResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/workflow/distinct").
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
		return nil, fmt.Errorf("distinct plugin workflows failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("distinct plugin workflows failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listPluginWorkflowOperation(ctx contextx.IContext, req *protoBackend.PluginWorkflowOperationListReq) (
	*protoBackend.PluginWorkflowOperationListResp, error) {

	resp := new(protoBackend.PluginWorkflowOperationListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/workflow/operation/list").
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
		return nil, fmt.Errorf("list plugin workflow operation failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list plugin workflow operation failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listPluginWorkflowOperationInstance(ctx contextx.IContext, req *protoBackend.PluginWorkflowOperationInstanceListReq) (
	*protoBackend.PluginWorkflowOperationInstanceListResp, error) {

	resp := new(protoBackend.PluginWorkflowOperationInstanceListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/workflow/operation/instance/list").
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
		return nil, fmt.Errorf("list plugin workflow operation instance failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list plugin workflow operation instance failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getPluginWorkflowOperationInstanceLog(ctx contextx.IContext, req *protoBackend.PluginWorkflowOperationInstanceLogGetReq) (
	*protoBackend.PluginWorkflowOperationInstanceLogGetResp, error) {

	resp := new(protoBackend.PluginWorkflowOperationInstanceLogGetResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/workflow/operation/instance/log/get").
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
		return nil, fmt.Errorf("get plugin workflow operation instance log failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("get plugin workflow operation instance log failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listPluginWorkflowOperationInstanceStatusDistribution(ctx contextx.IContext,
	req *protoBackend.PluginWorkflowOperationInstanceStatusDistributionListReq) (
	*protoBackend.PluginWorkflowOperationInstanceStatusDistributionListResp, error) {

	resp := new(protoBackend.PluginWorkflowOperationInstanceStatusDistributionListResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/workflow/operation/instance/status_distribution/list").
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
		return nil, fmt.Errorf("list plugin workflow operation instance status distribution failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("list plugin workflow operation instance status distribution failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) retryPluginWorkflowOperation(ctx contextx.IContext, req *protoBackend.PluginWorkflowOperationRetryReq) error {
	resp := new(protoBackend.PluginWorkflowOperationRetryResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/workflow/operation/retry").
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
		return fmt.Errorf("retry plugin workflow operation failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) terminatePluginWorkflowOperation(ctx contextx.IContext, req *protoBackend.PluginWorkflowOperationTerminateReq) error {
	resp := new(protoBackend.PluginWorkflowOperationTerminateResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/workflow/operation/terminate").
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
		return fmt.Errorf("terminate plugin workflow operation failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

func (c *cli) setPluginMemo(ctx contextx.IContext, req *protoBackend.PluginSetMemoReq) error {
	resp := new(protoBackend.PluginSetMemoResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/plugin/set_memo").
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
		return fmt.Errorf("set plugin memo failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	return nil
}

// ===============================================================================
// Encryption Related Interfaces
// ===============================================================================

func (c *cli) getRSAPublicKey(ctx contextx.IContext, req *protoBackend.GetRSAPublicKeyReq) (*protoBackend.GetRSAPublicKeyResp, error) {
	resp := new(protoBackend.GetRSAPublicKeyResp)
	header := c.getHeader(ctx)

	err := c.client.Post().
		SubResourcef("/cipher/rsa/get_public_key").
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
		return nil, fmt.Errorf("get RSA public key failed. code(%d), message(%s), error(%v), request-id(%s)",
			code, resp.GetMessage(), resp.GetError(), resp.GetRequestId())
	}

	if resp.GetData() == nil {
		return nil,
			fmt.Errorf("get RSA public key failed, get empty data. code(%d), message(%s), request-id(%s)",
				resp.GetCode(), resp.GetMessage(), resp.GetRequestId())
	}

	return resp, nil
}
