/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb provides handlers to operate cmd api.
package cmdb

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
)

// This file only supports requesting and getting responses.

// HeaderSetter ...
type HeaderSetter interface {
	GetAuthHeader() (string, error)
}

// Config the config of cmdb.
type Config struct {
	SupplierAccount string
	HeaderSetter    HeaderSetter
}

// cli client for cmdb.
type cli struct {
	client rest.ClientInterface
	config *Config
}

// newClient initialize a new cmdb client.
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

// getCommonHeader get cmdb common header.
func (c *cli) getCommonHeader(tenantID string) (http.Header, error) {
	header := http.Header{}
	header.Set(restheader.RIDKey, restheader.RIDGenerator())
	header.Set(restheader.TenantIDKey, tenantID)

	authHeader, err := c.config.HeaderSetter.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	header.Set(restheader.BKGWAuthKey, authHeader)

	return header, nil
}

// ListBizHosts ...
func (c *cli) listBizHosts(ctx context.Context, req *ListBizHostsReq) (*ListBizHostsResp, error) {
	resp := new(BaseBroker[*ListBizHostsResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/hosts/app/%d/list_hosts", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list biz hosts failed, err: %v", err)
	}

	return resp.Data, nil
}

// searchBusiness search cmdb business.
func (c *cli) searchBusiness(ctx context.Context, req *SearchBusinessReq) (*SearchBusinessResp, error) {
	resp := new(BaseBroker[*SearchBusinessResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/biz/search/%s", c.config.SupplierAccount).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("search business failed, err: %v", err)
	}

	return resp.Data, nil
}

// searchCloudArea search cloud area.
func (c *cli) searchCloudArea(ctx context.Context, req *SearchCloudAreaReq) (*SearchCloudAreaResp, error) {
	resp := new(BaseBroker[*SearchCloudAreaResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/cloudarea").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("search cloud area failed, err: %v", err)
	}

	return resp.Data, nil
}

// createCloudArea create cloud area.
func (c *cli) createCloudArea(ctx context.Context, req *CreateCloudAreaReq) (*CreateCloudAreaResp, error) {
	resp := new(BaseBroker[*CreateCloudAreaResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/create/cloudarea").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("create cloud area failed, err: %v", err)
	}

	return resp.Data, nil
}

// updateCloudArea update cloud area.
func (c *cli) updateCloudArea(ctx context.Context, req *UpdateCloudAreaReq) error {
	resp := new(BaseBroker[*UpdateCloudAreaResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return err
	}

	err = c.client.Put().
		SubResourcef("/update/cloudarea/%d", req.BKCloudID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("update cloud area failed, err: %v", err)
	}

	return nil
}

// deleteCloudArea delete cloud area.
func (c *cli) deleteCloudArea(ctx context.Context, req *DeleteCloudAreaReq) error {
	resp := new(BaseBroker[*UpdateCloudAreaResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return err
	}

	err = c.client.Delete().
		SubResourcef("/delete/cloudarea/%d", req.BKCloudID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("delete cloud area failed, err: %v", err)
	}

	return nil
}

// updateHostCloudAreaField update host cloud area field.
func (c *cli) updateHostCloudAreaField(ctx context.Context, req *UpdateHostCloudAreaFieldReq) error {
	resp := new(BaseBroker[*UpdateHostCloudAreaFieldResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return err
	}

	err = c.client.Put().
		SubResourcef("/updatemany/hosts/cloudarea_field").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("update host cloud area field failed, err: %v", err)
	}

	return nil
}

// searchBizInstTopo search biz inst topo.
func (c *cli) searchBizInstTopo(ctx context.Context, req *SearchBizInstTopoReq) (*SearchBizInstTopoResp, error) {
	resp := new(BaseBroker[*SearchBizInstTopoResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/find/topoinst/biz/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("search biz inst topo failed, err: %v", err)
	}

	return resp.Data, nil
}

// getBizInternalModule get biz internal module.
func (c *cli) getBizInternalModule(ctx context.Context, req *GetBizInternalModuleReq) (*GetBizInternalModuleResp, error) {
	resp := new(BaseBroker[*GetBizInternalModuleResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Get().
		SubResourcef("/topo/internal/%d/%d", c.config.SupplierAccount, req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("get biz internal module failed, err: %v", err)
	}

	return resp.Data, nil
}

// findTopoNodePaths find topo node paths.
func (c *cli) findTopoNodePaths(ctx context.Context, req *FindTopoNodePathsReq) (*FindTopoNodePathsResp, error) {
	resp := new(BaseBroker[*FindTopoNodePathsResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/cache/find/cache/topo/node_path/biz/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("find topo node paths failed, err: %v", err)
	}

	return resp.Data, nil
}

// findModuleBatch find module batch.
func (c *cli) findModuleBatch(ctx context.Context, req *FindModuleBatchReq) (*FindModuleBatchResp, error) {
	resp := new(BaseBroker[*FindModuleBatchResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/module/bk_biz_id/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("find module batch failed, err: %v", err)
	}

	return resp.Data, nil
}

// searchObjectAttribute search object attribute.
func (c *cli) searchObjectAttribute(ctx context.Context, req *SearchObjectAttributeReq) (
	*SearchObjectAttributeResp, error) {

	resp := new(BaseBroker[*SearchObjectAttributeResp])
	header, err := c.getCommonHeader(req.TenantID)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/find/objectattr").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("search object attribute failed, err: %v", err)
	}

	return resp.Data, nil
}
