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
	"context"
	"fmt"
	"net/http"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
)

// This file only supports requesting and getting responses.

// HeaderSetter ...
type HeaderSetter interface {
	GetAuthHeader() (string, error)
}

// Config the config of backend.
type Config struct {
	HeaderSetter HeaderSetter
}

// cli client for backend.
type cli struct {
	client rest.ClientInterface
	config *Config
}

// newClient initialize a new backend client.
func newClient(c *client.Capability, conf *Config) (*cli, error) {
	restCli, err := rest.NewClient(c, "/api/v3")
	if err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
		config: conf,
	}, nil
}

// getCommonHeader get backend common header.
func (c *cli) getCommonHeader(tenantID string) (http.Header, error) {
	header := http.Header{}
	header.Set(restheader.BKRIDKey, restheader.BKRIDGenerator())
	header.Set(restheader.BKTenantIDKey, tenantID)

	authHeader, err := c.config.HeaderSetter.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	header.Set(restheader.BKGWAuthKey, authHeader)

	return header, nil
}

func (c *cli) listBusiness(ctx context.Context, tenantID string, req *protoBackend.TopoBusinessListReq) (
	*protoBackend.TopoBusinessListResp_Data, error) {

	resp := new(protoBackend.TopoBusinessListResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) listHost(ctx context.Context, tenantID string, req *protoBackend.TopoHostListReq) (
	*protoBackend.TopoHostListResp, error) {

	resp := new(protoBackend.TopoHostListResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) distinctHost(ctx context.Context, tenantID string, req *protoBackend.TopoHostDistinctReq) (
	*protoBackend.TopoHostDistinctResp, error) {

	resp := new(protoBackend.TopoHostDistinctResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) createNetworkArea(ctx context.Context, tenantID string, req *protoBackend.TopoNetworkAreaCreateReq) (
	*protoBackend.TopoNetworkAreaCreateResp, error) {

	resp := new(protoBackend.TopoNetworkAreaCreateResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) updateNetworkArea(ctx context.Context, tenantID string, req *protoBackend.TopoNetworkAreaUpdateReq) (
	*protoBackend.TopoNetworkAreaUpdateResp, error) {

	resp := new(protoBackend.TopoNetworkAreaUpdateResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) listNetworkArea(ctx context.Context, tenantID string, req *protoBackend.TopoNetworkAreaListReq) (
	*protoBackend.TopoNetworkAreaListResp_Data, error) {

	resp := new(protoBackend.TopoNetworkAreaListResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) getNetworkArea(ctx context.Context, tenantID string, req *protoBackend.TopoNetworkAreaGetReq) (
	*protoBackend.TopoNetworkAreaGetResp, error) {

	resp := new(protoBackend.TopoNetworkAreaGetResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) deleteNetworkArea(ctx context.Context, tenantID string, req *protoBackend.TopoNetworkAreaDeleteReq) (
	*protoBackend.TopoNetworkAreaDeleteResp, error) {

	resp := new(protoBackend.TopoNetworkAreaDeleteResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) createNetworkUnit(ctx context.Context, tenantID string, req *protoBackend.TopoNetworkUnitCreateReq) (
	*protoBackend.TopoNetworkUnitCreateResp, error) {

	resp := new(protoBackend.TopoNetworkUnitCreateResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) updateNetworkUnit(ctx context.Context, tenantID string, req *protoBackend.TopoNetworkUnitUpdateReq) (
	*protoBackend.TopoNetworkUnitUpdateResp, error) {

	resp := new(protoBackend.TopoNetworkUnitUpdateResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) getNetworkUnit(ctx context.Context, tenantID string, req *protoBackend.TopoNetworkUnitGetReq) (
	*protoBackend.TopoNetworkUnitGetResp, error) {

	resp := new(protoBackend.TopoNetworkUnitGetResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) listNetworkUnit(ctx context.Context, tenantID string, req *protoBackend.TopoNetworkUnitListReq) (
	*protoBackend.TopoNetworkUnitListResp, error) {

	resp := new(protoBackend.TopoNetworkUnitListResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) deleteNetworkUnit(ctx context.Context, tenantID string, req *protoBackend.TopoNetworkUnitDeleteReq) (
	*protoBackend.TopoNetworkUnitDeleteResp, error) {

	resp := new(protoBackend.TopoNetworkUnitDeleteResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) listTopoEvent(ctx context.Context, tenantID string, req *protoBackend.TopoEventListReq) (
	*protoBackend.TopoEventListResp, error) {

	resp := new(protoBackend.TopoEventListResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) distinctTopoEvent(ctx context.Context, tenantID string, req *protoBackend.TopoEventDistinctReq) (
	*protoBackend.TopoEventDistinctResp, error) {

	resp := new(protoBackend.TopoEventDistinctResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) listAccessPoint(ctx context.Context, tenantID string, req *protoBackend.TopoAccessPointListReq) (
	*protoBackend.TopoAccessPointListResp, error) {

	resp := new(protoBackend.TopoAccessPointListResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) getConstant(ctx context.Context, tenantID string, req *protoBackend.TopoConstantGetReq) (
	*protoBackend.TopoConstantGetResp, error) {

	resp := new(protoBackend.TopoConstantGetResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) listNodeWorkflow(ctx context.Context, tenantID string, req *protoBackend.NodeWorkflowListReq) (
	*protoBackend.NodeWorkflowListResp, error) {

	resp := new(protoBackend.NodeWorkflowListResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) distinctNodeWorkflow(ctx context.Context, tenantID string, req *protoBackend.NodeWorkflowDistinctReq) (
	*protoBackend.NodeWorkflowDistinctResp, error) {

	resp := new(protoBackend.NodeWorkflowDistinctResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) listNodeWorkflowOperation(ctx context.Context,
	tenantID string, req *protoBackend.NodeWorkflowOperationListReq) (
	*protoBackend.NodeWorkflowOperationListResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationListResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) listNodeWorkflowOperationInstance(ctx context.Context,
	tenantID string, req *protoBackend.NodeWorkflowOperationInstanceListReq) (
	*protoBackend.NodeWorkflowOperationInstanceListResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationInstanceListResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) getOperationInstanceLog(ctx context.Context,
	tenantID string, req *protoBackend.NodeWorkflowOperationInstanceLogGetReq) (
	*protoBackend.NodeWorkflowOperationInstanceLogGetResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationInstanceLogGetResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) listNodeWorkflowOpInstanceStatus(ctx context.Context, tenantID string,
	req *protoBackend.NodeWorkflowOperationInstanceListStatusReq) (
	*protoBackend.NodeWorkflowOperationInstanceListStatusResp, error) {

	resp := new(protoBackend.NodeWorkflowOperationInstanceListStatusResp)
	header, err := c.getCommonHeader(tenantID)
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

func (c *cli) installNodeAgent(ctx context.Context, tenantID string, req *protoBackend.NodeAgentInstallReq) (
	*protoBackend.NodeAgentInstallResp, error) {

	resp := new(protoBackend.NodeAgentInstallResp)
	header, err := c.getCommonHeader(tenantID)
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
