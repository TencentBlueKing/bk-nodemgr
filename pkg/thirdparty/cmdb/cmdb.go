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
// nolint:dupl
package cmdb

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

// This file only supports requesting and getting responses.

const (
	// HeaderKeyLanguage the language header key.
	HeaderKeyLanguage = "X-Bkcmdb-Language"

	// HeaderValueLanguage the language header value.
	HeaderValueLanguage = "en"
)

// Config the config of cmdb.
type Config struct {
	SupplierAccount string
	APIGWAppConfig  apigwclient.AppConfig
}

// Validate configures the config.
func (conf *Config) Validate() error {
	if conf.SupplierAccount == "" {
		return errors.New("failed to validate cmdb client config: supplier account is empty")
	}

	if err := conf.APIGWAppConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate cmdb client config: %v", err)
	}

	return nil
}

// cli client for cmdb.
type cli struct {
	client restclient.IClient
	config *Config
}

// newClient initialize a new cmdb client.
func newClient(c *restclient.Capability, conf *Config) (*cli, error) {
	restCli, err := restclient.NewClient(c, "/api/v3", restclient.WithSensitiveHeader(apigwheader.BKGWAuthKey))
	if err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
		config: conf,
	}, nil
}

// getHeader get cmdb common header.
// nolint: unparam
func (c *cli) getHeader(ctx contextx.ITenantUserContext) (http.Header, error) {
	header := http.Header{}
	header.Set(restheader.BKTenantIDKey, ctx.TenantID())
	header.Set(HeaderKeyLanguage, HeaderValueLanguage)
	header.Set(apigwheader.BKGWRIDKey, identifier.GenRequestID())

	// backend apigw open the user auth.
	userConfig := apigwclient.UserConfig{
		AppConfig: c.config.APIGWAppConfig,
		User:      ctx.LoginName(),
	}
	header.Set(apigwheader.BKGWAuthKey, userConfig.GetAuthHeader())

	return header, nil
}

// ListBizHosts ...
func (c *cli) listBizHosts(ctx contextx.ITenantUserContext, req *ListBizHostsReq) (*ListBizHostsResp, error) {
	resp := new(BaseBroker[*ListBizHostsResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/hosts/app/%d/list_hosts", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
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
func (c *cli) searchBusiness(ctx contextx.ITenantUserContext, req *SearchBusinessReq) (*SearchBusinessResp, error) {
	resp := new(BaseBroker[*SearchBusinessResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/biz/search/%s", c.config.SupplierAccount).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
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
func (c *cli) searchCloudArea(ctx contextx.ITenantUserContext, req *SearchCloudAreaReq) (*SearchCloudAreaResp, error) {
	resp := new(BaseBroker[*SearchCloudAreaResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/cloudarea").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
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
func (c *cli) createCloudArea(ctx contextx.ITenantUserContext, req *CreateCloudAreaReq) (*CreateCloudAreaResp, error) {
	resp := new(BaseBroker[*CreateCloudAreaResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/create/cloudarea").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
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
func (c *cli) updateCloudArea(ctx contextx.ITenantUserContext, req *UpdateCloudAreaReq) error {
	resp := new(BaseBroker[*UpdateCloudAreaResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return err
	}

	err = c.client.Put().
		SubResourcef("/update/cloudarea/%d", req.BKCloudID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
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
func (c *cli) deleteCloudArea(ctx contextx.ITenantUserContext, req *DeleteCloudAreaReq) error {
	resp := new(BaseBroker[*UpdateCloudAreaResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return err
	}

	err = c.client.Delete().
		SubResourcef("/delete/cloudarea/%d", req.BKCloudID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
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
func (c *cli) updateHostCloudAreaField(ctx contextx.ITenantUserContext, req *UpdateHostCloudAreaFieldReq) error {
	resp := new(BaseBroker[*UpdateHostCloudAreaFieldResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return err
	}

	err = c.client.Put().
		SubResourcef("/updatemany/hosts/cloudarea_field").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
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
// nolint: unused
func (c *cli) searchBizInstTopo(ctx contextx.ITenantUserContext, req *SearchBizInstTopoReq) (*SearchBizInstTopoResp, error) {
	resp := new(BaseBroker[*SearchBizInstTopoResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/find/topoinst/biz/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
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
// nolint: unused
func (c *cli) getBizInternalModule(ctx contextx.ITenantUserContext, req *GetBizInternalModuleReq) (
	*GetBizInternalModuleResp, error) {

	resp := new(BaseBroker[*GetBizInternalModuleResp])
	header, err := c.getHeader(ctx)
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
// nolint: unused
func (c *cli) findTopoNodePaths(ctx contextx.ITenantUserContext, req *FindTopoNodePathsReq) (*FindTopoNodePathsResp, error) {
	resp := new(BaseBroker[*FindTopoNodePathsResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/cache/find/cache/topo/node_path/biz/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
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
// nolint: unused
func (c *cli) findModuleBatch(ctx contextx.ITenantUserContext, req *FindModuleBatchReq) (*FindModuleBatchResp, error) {
	resp := new(BaseBroker[*FindModuleBatchResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/module/bk_biz_id/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
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
func (c *cli) searchObjectAttribute(ctx contextx.ITenantUserContext, req *SearchObjectAttributeReq) (
	*SearchObjectAttributeResp, error) {

	resp := new(BaseBroker[*SearchObjectAttributeResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/find/objectattr").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("search object attribute failed, err: %v", err)
	}

	return resp.Data, nil
}

// bindHostAgent bind host agent.
func (c *cli) bindHostAgent(ctx contextx.ITenantUserContext, req *BindHostAgentReq) error {
	resp := new(BaseBroker[*BindHostAgentResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/host/bind/agent").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("bind host agent failed, err: %v", err)
	}

	return nil
}

// unbindHostAgent bind host agent.
func (c *cli) unbindHostAgent(ctx contextx.ITenantUserContext, req *UnbindHostAgentReq) error {
	resp := new(BaseBroker[*UnbindHostAgentResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return err
	}

	err = c.client.Post().
		SubResourcef("/host/unbind/agent").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("unbind host agent failed, err: %v", err)
	}

	return nil
}

// addHostToBusiessIdle add host to business idle.
func (c *cli) addHostToBusinessIdle(ctx contextx.ITenantUserContext, req *AddHostToBusinessIdleReq) (*AddHostToBusinessIdleResp,
	error) {

	resp := new(BaseBroker[*AddHostToBusinessIdleResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/hosts/add/business_idle").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("add host to business idle failed, err: %v", err)
	}

	return resp.Data, nil
}

// pushHostIdentifier push host identifier.
func (c *cli) pushHostIdentifier(ctx contextx.ITenantUserContext, req *PushHostIdentifierReq) (*PushHostIdentifierResp, error) {
	resp := new(BaseBroker[*PushHostIdentifierResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/event/push/host_identifier").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("push host identifier failed, err: %v", err)
	}

	return resp.Data, nil
}

// findHostIdentifierPushResult find host identifier push result.
func (c *cli) findHostIdentifierPushResult(ctx contextx.ITenantUserContext, req *FindHostIdentifierPushResultReq) (
	*FindHostIdentifierPushResultResp, error) {

	resp := new(BaseBroker[*FindHostIdentifierPushResultResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/event/find/host_identifier_push_result").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("find host identifier push result failed, err: %v", err)
	}

	return resp.Data, nil
}

// addHostToResource add host to resource.
func (c *cli) addHostToResource(ctx contextx.ITenantUserContext, req *AddHostToResourcePoolReq) (*AddHostToResourcePoolResp,
	error) {

	resp := new(BaseBroker[*AddHostToResourcePoolResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/hosts/add/resource").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("add host to resource failed, err: %v", err)
	}

	return resp.Data, nil
}

func (c *cli) listResourcePoolHosts(ctx contextx.ITenantUserContext, req *ListResourcePoolHostsReq) (*ListResourcePoolHostsResp,
	error) {

	resp := new(BaseBroker[*ListResourcePoolHostsResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/hosts/list_resource_pool_hosts").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list resource pool hosts failed, err: %v", err)
	}

	return resp.Data, nil
}

// createDynamicGroup create dynamic group.
// nolint: unused
func (c *cli) createDynamicGroup(ctx contextx.ITenantUserContext, req *CreateDynamicGroupReq) (*CreateDynamicGroupResp, error) {
	resp := new(BaseBroker[*CreateDynamicGroupResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/dynamicgroup").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("create dynamic group failed, err: %v", err)
	}

	return resp.Data, nil
}

// executeDynamicGroup execute host dynamic group.
func (c *cli) executeDynamicGroup(ctx contextx.ITenantUserContext, req *ExecuteDynamicGroupReq) (*ExecuteDynamicGroupResp, error) {
	resp := new(BaseBroker[*ExecuteDynamicGroupResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/dynamicgroup/data/%d/%s", req.BKBizID, req.ID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("execute host dynamic group failed, err: %v", err)
	}

	return resp.Data, nil
}

// searchDynamicGroup search dynamic group.
func (c *cli) searchDynamicGroup(ctx contextx.ITenantUserContext, req *SearchDynamicGroupReq) (*SearchDynamicGroupResp, error) {
	resp := new(BaseBroker[*SearchDynamicGroupResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/dynamicgroup/search/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("search dynamic group failed, err: %v", err)
	}

	return resp.Data, nil
}

// deleteDynamicGroup delete dynamic group.
// nolint: unused
func (c *cli) deleteDynamicGroup(ctx contextx.ITenantUserContext, req *DeleteDynamicGroupReq) error {
	resp := new(BaseBroker[*DeleteDynamicGroupResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return err
	}

	err = c.client.Delete().
		SubResourcef("/dynamicgroup/%d/%s", req.BKBizID, req.ID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("delete dynamic group failed, err: %v", err)
	}

	return nil
}

// getDynamicGroup get dynamic group.
// nolint: unused
func (c *cli) getDynamicGroup(ctx contextx.ITenantUserContext, req *GetDynamicGroupReq) (*GetDynamicGroupResp, error) {
	resp := new(BaseBroker[*GetDynamicGroupResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Get().
		SubResourcef("/dynamicgroup/%d/%s", req.BKBizID, req.ID).
		WithContext(ctx).
		WithHeaders(header).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("get dynamic group failed, err: %v", err)
	}

	return resp.Data, nil
}

// updateDynamicGroup update dynamic group.
// nolint: unused
func (c *cli) updateDynamicGroup(ctx contextx.ITenantUserContext, req *UpdateDynamicGroupReq) error {
	resp := new(BaseBroker[*UpdateDynamicGroupResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return err
	}

	err = c.client.Put().
		SubResourcef("/dynamicgroup/%d/%s", req.BKBizID, req.ID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("update dynamic group failed, err: %v", err)
	}

	return nil
}

// listHostsWithoutBusiness list hosts without business.
func (c *cli) listHostsWithoutBusiness(ctx contextx.ITenantUserContext, req *ListHostsWithoutBusinessReq) (
	*ListHostsWithoutBusinessResp, error) {

	resp := new(BaseBroker[*ListHostsWithoutBusinessResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/hosts/list_hosts_without_app").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list hosts without business failed, err: %v", err)
	}

	return resp.Data, nil
}

// getMainlineObjectTopo get mainline object topo.
// nolint: unused
func (c *cli) getMainlineObjectTopo(ctx contextx.ITenantUserContext, req *GetMainlineObjectTopoReq) (
	*GetMainlineObjectTopoResp, error) {

	resp := new(BaseBroker[*GetMainlineObjectTopoResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/find/topomodelmainline").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("get mainline object failed, err: %v", err)
	}

	return resp.Data, nil
}

// listServiceTemplate list service template.
func (c *cli) listServiceTemplate(ctx contextx.ITenantUserContext, req *ListServiceTemplateReq) (
	*ListServiceTemplateResp, error) {

	resp := new(BaseBroker[*ListServiceTemplateResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/proc/service_template").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list service template failed, err: %v", err)
	}

	return resp.Data, nil
}

// listServiceInstance list service instance.
// nolint: unused
func (c *cli) listServiceInstance(ctx contextx.ITenantUserContext, req *ListServiceInstanceReq) (
	*ListServiceInstanceResp, error) {

	resp := new(BaseBroker[*ListServiceInstanceResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/proc/service_instance").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list service instance failed, err: %v", err)
	}

	return resp.Data, nil
}

// listProcessInstance list process instance.
// nolint: unused
func (c *cli) listProcessInstance(ctx contextx.ITenantUserContext, req *ListProcessInstanceReq) (
	*ListProcessInstanceResp, error) {

	resp := new(BaseBroker[*ListProcessInstanceResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/proc/process_instance").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list process instance failed, err: %v", err)
	}

	return resp.Data, nil
}

// listProcTemplate list proc template.
// nolint: unused
func (c *cli) listProcTemplate(ctx contextx.ITenantUserContext, req *ListProcTemplateReq) (
	*ListProcTemplateResp, error) {

	resp := new(BaseBroker[*ListProcTemplateResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/proc/proc_template").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list proc template failed, err: %v", err)
	}

	return resp.Data, nil
}

// findSetBatch find set batch.
// nolint: unused
func (c *cli) findSetBatch(ctx contextx.ITenantUserContext, req *FindSetBatchReq) (*FindSetBatchResp, error) {
	resp := new(BaseBroker[*FindSetBatchResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/set/bk_biz_id/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("find set batch failed, err: %v", err)
	}

	return resp.Data, nil
}

// searchSet search set.
// nolint: unused
func (c *cli) searchSet(ctx contextx.ITenantUserContext, req *SearchSetReq) (*SearchSetResp, error) {
	resp := new(BaseBroker[*SearchSetResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/set/search/%s/%d", c.config.SupplierAccount, req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("search set failed, err: %v", err)
	}

	return resp.Data, nil
}

// searchModule search module.
// nolint: unused
func (c *cli) searchModule(ctx contextx.ITenantUserContext, req *SearchModuleReq) (*SearchModuleResp, error) {
	resp := new(BaseBroker[*SearchModuleResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/module/search/%s/%d/%d", c.config.SupplierAccount, req.BKBizID, req.BKSetID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("search module failed, err: %v", err)
	}

	return resp.Data, nil
}

// findHostTopoRelation find host topo relation.
// nolint: unused
func (c *cli) findHostTopoRelation(ctx contextx.ITenantUserContext, req *FindHostTopoRelationReq) (
	*FindHostTopoRelationResp, error) {

	resp := new(BaseBroker[*FindHostTopoRelationResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/host/topo/relation/read").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("find host topo relation failed, err: %v", err)
	}

	return resp.Data, nil
}

// findHostBizRelations find host biz relations.
// nolint: unused
func (c *cli) findHostBizRelations(ctx contextx.ITenantUserContext, req *FindHostBizRelationsReq) (
	*FindHostBizRelationsResp, error) {

	resp := new(BaseBroker[*FindHostBizRelationsResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/hosts/modules/read").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("find host biz relations failed, err: %v", err)
	}

	return resp.Data, nil
}

// findHostByServiceTemplate find host by service template.
func (c *cli) findHostByServiceTemplate(ctx contextx.ITenantUserContext, req *FindHostByServiceTemplateReq) (
	*FindHostByServiceTemplateResp, error) {

	resp := new(BaseBroker[*FindHostByServiceTemplateResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("findmany/hosts/by_service_templates/biz/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("find host by service template failed, err: %v", err)
	}

	return resp.Data, nil
}

// findHostBySetTemplate find host by set template.
// nolint: unused
func (c *cli) findHostBySetTemplate(ctx contextx.ITenantUserContext, req *FindHostBySetTemplateReq) (
	*FindHostBySetTemplateResp, error) {

	resp := new(BaseBroker[*FindHostBySetTemplateResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/hosts/by_set_templates/biz/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("find host by set template failed, err: %v", err)
	}

	return resp.Data, nil
}

// findHostByTopo find host by topo.
// nolint: unused
func (c *cli) findHostByTopo(ctx contextx.ITenantUserContext, req *FindHostByTopoReq) (
	*FindHostByTopoResp, error) {

	resp := new(BaseBroker[*FindHostByTopoResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/hosts/by_topo/biz/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("find host by topo failed, err: %v", err)
	}

	return resp.Data, nil
}

// findHostRelationsWithTopo find host relations with topo.
// nolint: unused
func (c *cli) findHostRelationsWithTopo(ctx contextx.ITenantUserContext, req *FindHostRelationsWithTopoReq) (
	*FindHostRelationsWithTopoResp, error) {

	resp := new(BaseBroker[*FindHostRelationsWithTopoResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/hosts/relation/with_topo").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("find host relations with topo failed, err: %v", err)
	}

	return resp.Data, nil
}

// listServiceInstanceDetail list service instance detail.
// nolint: unused
func (c *cli) listServiceInstanceDetail(ctx contextx.ITenantUserContext, req *ListServiceInstanceDetailReq) (
	*ListServiceInstanceDetailResp, error) {

	resp := new(BaseBroker[*ListServiceInstanceDetailResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/proc/service_instance/details").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list service instance detail failed, err: %v", err)
	}

	return resp.Data, nil
}

// listBizHostsTopo list biz hosts topo.
// nolint: unused
func (c *cli) listBizHostsTopo(ctx contextx.ITenantUserContext, req *ListBizHostsTopoReq) (
	*ListBizHostsTopoResp, error) {

	resp := new(BaseBroker[*ListBizHostsTopoResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/hosts/app/%d/list_hosts_topo", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list hosts topo failed, err: %v", err)
	}

	return resp.Data, nil
}

// listServiceInstanceByHost list service instance by host.
// nolint: unused
func (c *cli) listServiceInstanceByHost(ctx contextx.ITenantUserContext, req *ListServiceInstanceByHostReq) (
	*ListServiceInstanceByHostResp, error) {

	resp := new(BaseBroker[*ListServiceInstanceByHostResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/proc/service_instance/with_host").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list service instance by host failed, err: %v", err)
	}

	return resp.Data, nil
}

// listServiceInstanceBySetTemplate list service instance by host.
// nolint: unused
func (c *cli) listServiceInstanceBySetTemplate(ctx contextx.ITenantUserContext, req *ListServiceInstanceBySetTemplateReq) (
	*ListServiceInstanceBySetTemplateResp, error) {

	resp := new(BaseBroker[*ListServiceInstanceBySetTemplateResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/proc/service/set_template/list_service_instance/biz/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list service instance by set template failed, err: %v", err)
	}

	return resp.Data, nil
}

// listSetTemplate list set template.
// nolint: unused
func (c *cli) listSetTemplate(ctx contextx.ITenantUserContext, req *ListSetTemplateReq) (
	*ListSetTemplateResp, error) {

	resp := new(BaseBroker[*ListSetTemplateResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("findmany/topo/set_template/bk_biz_id/%d", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list set template failed, err: %v", err)
	}

	return resp.Data, nil
}

// batchUpdateHost batch update host.
// nolint: unused
func (c *cli) batchUpdateHost(ctx contextx.ITenantUserContext, req *BatchUpdateHostReq) error {
	resp := new(BaseBroker[*BatchUpdateHostResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return err
	}

	err = c.client.Put().
		SubResourcef("/hosts/property/batch").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return err
	}

	if err := resp.IsFailed(); err != nil {
		return fmt.Errorf("batch update host failed, err: %v", err)
	}

	return nil
}

// findHostServiceTemplate find host service template.
// nolint: unused
func (c *cli) findHostServiceTemplate(ctx contextx.ITenantUserContext, req *FindHostServiceTemplateReq) (
	*FindHostServiceTemplateResp, error) {

	resp := new(BaseBroker[*FindHostServiceTemplateResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/hosts/service_template").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("find host service template failed, err: %v", err)
	}

	return resp.Data, nil
}

// resourceWatch resource watch.
// cc resource_watch interface using a short-long chain design
// if there are any event changes within 20 seconds, the events will be pushed back directly.
func (c *cli) resourceWatch(ctx contextx.ITenantUserContext, req *ResourceWatchReq) (*ResourceWatchResp, error) {
	resp := new(BaseBroker[*ResourceWatchResp])
	header, err := c.getHeader(ctx)
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/event/watch/resource/%s", req.BKResource).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		EnableLogBody().
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("resource watch failed, err: %v", err)
	}

	return resp.Data, nil
}
