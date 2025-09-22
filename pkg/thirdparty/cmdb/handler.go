/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// This document is responsible for processing the conversion of original requests and responses
// from the third-party system and the internal data of the nodeman system.

// IHandler the Handler of cmdb.
// nolint: interfacebloat
type IHandler interface {
	IEnum
	IBiz
	IHost
	IHostIdentifier
	INetworkArea
	IBindHostAgent
	IUnbindHostAgent
	IUpdateHostNetworkAreaField
	IDynamicGroup
	IServiceTemplate
	IWatch
}

// IDynamicGroup this interface is used to edit dynamic group.
type IDynamicGroup interface {
	// SearchDynamicGroup search dynamic group.
	SearchDynamicGroup(ctx contextx.ITenantUserContext, bizID int64, page types.Page) ([]*types.DynamicGroup, error)

	// ExecuteHostDynamicGroup execute dynamic grouping rules to return hosts within the group.
	ExecuteHostDynamicGroup(ctx contextx.ITenantUserContext, bizID int64, groupID string, page types.Page) ([]*types.Host, error)
}

// IBiz this interface is used to edit biz info.
type IBiz interface {
	// SearchBusiness search business.
	SearchBusiness(ctx contextx.ITenantUserContext, page types.Page) ([]*types.Business, error)
}

// IEnum this interface is used to get enum resource.
type IEnum interface {
	// GetCloudVendors get cloud vendors.
	GetCloudVendors() []string

	// GetOSTypes get os types.
	GetOSTypes() []string
}

// IHostIdentifier host identifier.
type IHostIdentifier interface {
	// PushHostIdentifier push host identifier.
	PushHostIdentifier(ctx contextx.ITenantUserContext, hostIDs ...int64) (taskID string, err error)

	// FindHostIdentifierPushResult find host identifier push result.
	FindHostIdentifierPushResult(ctx contextx.ITenantUserContext, taskID string) (successList []int64, failedList []int64,
		pendingList []int64, err error)
}

// IHost this interface is used to edit host info.
type IHost interface {
	// ListBizHosts list biz hosts.
	ListBizHosts(ctx contextx.ITenantUserContext, bizID int64, page types.Page) ([]*types.Host, error)

	// ListHostsWithoutBusiness list hosts without business.
	ListHostsWithoutBusiness(ctx contextx.ITenantUserContext, page types.Page) ([]*types.Host, error)

	// FindHostByServiceTemplate find host by service template.
	FindHostByServiceTemplate(ctx contextx.ITenantUserContext, bizID int64, page types.Page, serviceTemplateIDs ...int64) (
		[]*types.Host, error)

	// CheckBizHostByIP check biz host by ip
	CheckBizHostByIP(ctx contextx.ITenantUserContext, bizID int64, cloudID int64, ip string) (bool, error)

	// ListResourcePoolHosts list resource pool hosts.
	ListResourcePoolHosts(ctx contextx.ITenantUserContext, page types.Page) ([]*types.Host, error)

	// AddHostToResourcePool add host to resource pool
	AddHostToResourcePool(ctx contextx.ITenantUserContext, hosts ...*types.Host) (successHost []*types.Host,
		failedIndexMsg []string, err error)

	// AddHostToBusinessIdle add host to business idle
	AddHostToBusinessIdle(ctx contextx.ITenantUserContext, bizID int64, hosts ...*types.Host) ([]int64, error)
}

// INetworkArea this interface is used to edit network area.
type INetworkArea interface {
	// SearchNetworkArea search network area.
	SearchNetworkArea(ctx contextx.ITenantUserContext, page types.Page) ([]*types.NetworkArea, error)

	// CreateNetworkArea create network area.
	CreateNetworkArea(ctx contextx.ITenantUserContext, networkAreaName string, cloudVendor string) (*types.NetworkArea, error)

	// UpdateNetworkArea update network area.
	UpdateNetworkArea(ctx contextx.ITenantUserContext, id int64, networkAreaName string, cloudVendor string) error

	// DeleteNetworkArea delete network area.
	DeleteNetworkArea(ctx contextx.ITenantUserContext, id int64) error
}

// IBindHostAgent this interface is used to bind host agent.
type IBindHostAgent interface {
	BindHostAgent(ctx contextx.ITenantUserContext, hostInfo ...*types.Host) error
}

// IUnbindHostAgent this interface is used to unbind host agent.
type IUnbindHostAgent interface {
	UnbindHostAgent(ctx contextx.ITenantUserContext, hostInfo ...*types.Host) error
}

// IUpdateHostNetworkAreaField this interface is used to update host network area field.
type IUpdateHostNetworkAreaField interface {
	// UpdateHostNetworkAreaField update host network area field.
	UpdateHostNetworkAreaField(ctx contextx.ITenantUserContext, bizID int64, networkAreaID int64, hostIDs ...int64) error
}

// IServiceTemplate this interface is used to list service template.
type IServiceTemplate interface {
	// ListServiceTemplate list service template.
	ListServiceTemplate(ctx contextx.ITenantUserContext, bizID int64, page types.Page) ([]*types.ServiceTemplate, error)
}

// IWatch this interface is used to watch resource event.
type IWatch interface {
	// WatchHostResourceEvent watch host resource event.
	WatchHostResourceEvent(ctx contextx.ITenantUserContext, cursor string) ([]*types.HostEvent, error)

	// WatchHostRelationResourceEvent watch host relation resource event.
	WatchHostRelationResourceEvent(ctx contextx.ITenantUserContext, cursor string) ([]*types.HostEvent, error)
}

// Handler the Handler of cmdb.
type Handler struct {
	cli    *cli
	logger logger.ILogger

	scheduler scheduler.Scheduler

	cloudVendorKeeper iEnumResourceKeeper
	osTypeKeeper      iEnumResourceKeeper
	cpuArchKeeper     iEnumResourceKeeper
}

const (
	enumResourceSyncInterval = 30 * time.Minute
	enumResourceSyncTimeout  = 30 * time.Second
)

// OptionFn ...
type OptionFn func(*Handler)

// WithLogger this func will set the logger of the Handler.
func WithLogger(logger logger.ILogger) OptionFn {
	return func(s *Handler) {
		s.logger = logger
	}
}

// New initialize a new cmdb Handler.
func New(c *restclient.Capability, conf *Config, opts ...OptionFn) (IHandler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		cli:    cli,
		logger: logger.LoggerDefault{},

		cloudVendorKeeper: newCloudVendorKeeper(cli),
		osTypeKeeper:      newOSTypeKeeper(cli),
		cpuArchKeeper:     newCPUArchKeeper(cli),
	}

	for _, opt := range opts {
		opt(h)
	}

	err = h.initEnumKeepers()
	if err != nil {
		h.logger.Errorf("failed to init enum keepers: %v", err)
		return nil, err
	}

	return h, nil
}

// nolint: gocognit, funlen
func (h *Handler) initEnumKeepers() error {
	h.logger.Infof("initializing enum keepers from cmdb")

	if h.scheduler != nil {
		h.scheduler.Terminate()
	}
	h.scheduler = scheduler.NewScheduler(scheduler.WithLogger(h.logger))
	syncTasks := []*scheduler.Task{
		scheduler.NewTask(
			"sync_cloud_vendor",
			enumResourceSyncInterval,
			enumResourceSyncTimeout,
			func(ctx contextx.IContext) error {
				tenantIDs := tenant.GetAllTenantIDs()
				for _, tenantID := range tenantIDs {
					tenantUserCtx := contextx.NewTenantUserContext(ctx, tenantID, access.GetVirtualUser())

					if err := h.cloudVendorKeeper.update(tenantUserCtx); err != nil {
						return err
					}
				}

				return nil
			},
		),
		scheduler.NewTask(
			"sync_os_type",
			enumResourceSyncInterval,
			enumResourceSyncTimeout,
			func(ctx contextx.IContext) error {
				tenantIDs := tenant.GetAllTenantIDs()
				for _, tenantID := range tenantIDs {
					tenantUserCtx := contextx.NewTenantUserContext(ctx, tenantID, access.GetVirtualUser())
					if err := h.osTypeKeeper.update(tenantUserCtx); err != nil {
						return err
					}
				}

				return nil
			},
		),
		scheduler.NewTask(
			"sync_cpu_arch",
			enumResourceSyncInterval,
			enumResourceSyncTimeout,
			func(ctx contextx.IContext) error {
				tenantIDs := tenant.GetAllTenantIDs()
				for _, tenantID := range tenantIDs {
					tenantUserCtx := contextx.NewTenantUserContext(ctx, tenantID, access.GetVirtualUser())
					if err := h.cpuArchKeeper.update(tenantUserCtx); err != nil {
						return err
					}
				}

				return nil
			},
		),
	}

	for _, task := range syncTasks {
		err := h.scheduler.RegisterTask(task)
		if err != nil {
			h.logger.Errorf("failed to register sync task, task-id(%s): %v", task.ID, err)
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), enumResourceSyncTimeout)
	defer cancel()

	tenantIDs := tenant.GetAllTenantIDs()
	for _, tenantID := range tenantIDs {
		tenantUserCtx := contextx.NewTenantUserContext(ctx, tenantID, access.GetVirtualUser())

		if err := h.cloudVendorKeeper.update(tenantUserCtx); err != nil {
			h.logger.Warnf("failed to sync cloud vendor: %v", err)
		}
		if err := h.osTypeKeeper.update(tenantUserCtx); err != nil {
			h.logger.Warnf("failed to sync os type: %v", err)
		}
		if err := h.cpuArchKeeper.update(tenantUserCtx); err != nil {
			h.logger.Warnf("failed to sync cpu arch: %v", err)
		}
	}

	h.scheduler.Start()
	h.logger.Infof("start schedule enum resource sync tasks")

	return nil
}

// ListBizHosts list biz hosts.
func (h *Handler) ListBizHosts(ctx contextx.ITenantUserContext, bizID int64, page types.Page) ([]*types.Host, error) {
	tenantID := ctx.TenantID()
	loginUsername := ctx.BKUsername()

	executor := pageexecutor.NewPageExecutor[*types.Host](CCPageSizeLimit, 1*time.Hour) // nolint: mnd
	fn := func(ctx context.Context, p types.Page) ([]*types.Host, error) {
		req := &ListBizHostsReq{
			BKBizID: bizID,
			Page: Page{
				Start: p.Offset,
				Limit: p.Limit,
				Sort:  p.Sort,
			},
		}
		tenantUserCtx := contextx.NewTenantUserContext(ctx, tenantID, loginUsername)
		resp, err := h.cli.listBizHosts(tenantUserCtx, req) // nolint: contextcheck
		if err != nil {
			return nil, err
		}

		hosts := make([]*types.Host, len(resp.Info))
		for idx, host := range resp.Info {
			hosts[idx] = h.convHostInfoToTypes(tenantID, host, bizID)
		}

		return hosts, nil
	}

	result, err := executor.Execute(ctx, page, fn)
	if err != nil {
		return nil, fmt.Errorf("execute page executor failed: %v", err)
	}

	return result.Items, nil
}

// SearchBusiness search business.
func (h *Handler) SearchBusiness(ctx contextx.ITenantUserContext, page types.Page) ([]*types.Business, error) {
	req := &SearchBusinessReq{
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
		Fields: nil,
	}

	resp, err := h.cli.searchBusiness(ctx, req)
	if err != nil {
		return nil, err
	}

	bizs := make([]*types.Business, len(resp.Info))
	for idx, business := range resp.Info {
		bizs[idx] = &types.Business{
			TenantID: ctx.TenantID(),
			BizID:    business.BKBizID,
			BizName:  business.BKBizName,
		}
	}

	return bizs, nil
}

// SearchNetworkArea search network area.
func (h *Handler) SearchNetworkArea(ctx contextx.ITenantUserContext, page types.Page) ([]*types.NetworkArea, error) {
	req := &SearchCloudAreaReq{
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.searchCloudArea(ctx, req)
	if err != nil {
		return nil, err
	}

	netAreas := make([]*types.NetworkArea, len(resp.Info))
	for idx, networkarea := range resp.Info {
		netAreas[idx] = h.convCloudAreaToTypes(ctx.TenantID(), networkarea)
	}

	return netAreas, nil
}

// CreateNetworkArea create network area.
func (h *Handler) CreateNetworkArea(ctx contextx.ITenantUserContext, networkAreaName string, cloudVendor string) (
	*types.NetworkArea, error) {

	req := &CreateCloudAreaReq{
		BKCloudName:   networkAreaName,
		BKCloudVendor: h.cloudVendorKeeper.getKey(cloudVendor),
	}

	resp, err := h.cli.createCloudArea(ctx, req)
	if err != nil {
		return nil, err
	}

	netArea := &types.NetworkArea{
		TenantID:    ctx.TenantID(),
		ID:          resp.Created.ID,
		Name:        networkAreaName,
		CloudVendor: cloudVendor,
	}

	return netArea, nil
}

// UpdateNetworkArea update network area.
func (h *Handler) UpdateNetworkArea(
	ctx contextx.ITenantUserContext, id int64, networkAreaName string, cloudVendor string) error {

	req := &UpdateCloudAreaReq{
		BKCloudID:     id,
		BKCloudName:   networkAreaName,
		BKCloudVendor: h.cloudVendorKeeper.getKey(cloudVendor),
	}

	err := h.cli.updateCloudArea(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// DeleteNetworkArea delete network area.
func (h *Handler) DeleteNetworkArea(ctx contextx.ITenantUserContext, id int64) error {
	req := &DeleteCloudAreaReq{
		BKCloudID: id,
	}

	err := h.cli.deleteCloudArea(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// UpdateHostNetworkAreaField update host network area field.
func (h *Handler) UpdateHostNetworkAreaField(
	ctx contextx.ITenantUserContext, bizID int64, networkAreaID int64, hostIDs ...int64) error {

	req := &UpdateHostCloudAreaFieldReq{
		BKBizID:   bizID,
		BKCloudID: networkAreaID,
		BKHostIDs: hostIDs,
	}

	err := h.cli.updateHostCloudAreaField(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// GetCloudVendors get cloud vendors.
func (h *Handler) GetCloudVendors() []string {
	return h.cloudVendorKeeper.values()
}

// GetOSTypes get os types.
func (h *Handler) GetOSTypes() []string {
	return h.osTypeKeeper.values()
}

// BindHostAgent bind host agent.
func (h *Handler) BindHostAgent(ctx contextx.ITenantUserContext, hostInfo ...*types.Host) error {
	if len(hostInfo) == 0 {
		return errors.New("host info list is empty")
	}

	req := &BindHostAgentReq{
		List: make([]*HostAgentIDInfo, 0, len(hostInfo)),
	}

	virtualUser := access.GetVirtualUser()
	h.logger.InfoCtxf(ctx, "use virtual user to bind host agent, virtual-user(%s), req(%v)", virtualUser, req)
	tenantUserCtx := contextx.NewTenantUserContext(ctx, ctx.TenantID(), virtualUser)

	for _, host := range hostInfo {
		req.List = append(req.List, &HostAgentIDInfo{
			BKHostID:  host.HostID,
			BKAgentID: host.Dynamic.AgentID,
		})
	}
	err := h.cli.bindHostAgent(tenantUserCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// UnbindHostAgent bind host agent.
func (h *Handler) UnbindHostAgent(ctx contextx.ITenantUserContext, hostInfo ...*types.Host) error {
	if len(hostInfo) == 0 {
		return errors.New("host info list is empty")
	}

	req := &UnbindHostAgentReq{
		List: make([]*HostAgentIDInfo, 0, len(hostInfo)),
	}

	virtualUser := access.GetVirtualUser()
	h.logger.InfoCtxf(ctx, "use virtual user to un bind host agent, virtual-user(%s), req(%v)", virtualUser, req)
	tenantUserCtx := contextx.NewTenantUserContext(ctx, ctx.TenantID(), virtualUser)

	for _, host := range hostInfo {
		req.List = append(req.List, &HostAgentIDInfo{
			BKHostID:  host.HostID,
			BKAgentID: host.Dynamic.AgentID,
		})
	}
	err := h.cli.unbindHostAgent(tenantUserCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// AddHostToBusinessIdle add host to business idle.
func (h *Handler) AddHostToBusinessIdle(ctx contextx.ITenantUserContext, bizID int64, hosts ...*types.Host) (
	[]int64, error) {

	if len(hosts) == 0 {
		return nil, errors.New("host list is empty")
	}

	req := &AddHostToBusinessIdleReq{BKBizID: bizID}
	for _, host := range hosts {
		req.BKHostList = append(req.BKHostList, h.convCreateHostInfoFromTypes(host))
	}

	virtualUser := access.GetVirtualUser()
	h.logger.InfoCtxf(ctx, "use virtual user to add host to business idle, virtual-user(%s), req(%v)", virtualUser, req)
	tenantUserCtx := contextx.NewTenantUserContext(ctx, ctx.TenantID(), virtualUser)

	resp, err := h.cli.addHostToBusinessIdle(tenantUserCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.BKHostIDs, nil
}

// PushHostIdentifier push host identifier.
// nolint: nonamedreturns
func (h *Handler) PushHostIdentifier(ctx contextx.ITenantUserContext, hostIDs ...int64) (taskID string, err error) {
	req := &PushHostIdentifierReq{
		BKHostIDs: hostIDs,
	}

	virtualUser := access.GetVirtualUser()
	h.logger.InfoCtxf(ctx, "use virtual user to push host identifier, virtual-user(%s), req(%v)", virtualUser, req)
	tenantUserCtx := contextx.NewTenantUserContext(ctx, ctx.TenantID(), virtualUser)

	resp, err := h.cli.pushHostIdentifier(tenantUserCtx, req)
	if err != nil {
		return "", err
	}

	return resp.TaskID, nil
}

// FindHostIdentifierPushResult find host identifier push result.
// nolint: nonamedreturns
func (h *Handler) FindHostIdentifierPushResult(ctx contextx.ITenantUserContext, taskID string) (successList []int64,
	failedList []int64, pendingList []int64, err error) {

	req := &FindHostIdentifierPushResultReq{
		TaskID: taskID,
	}

	virtualUser := access.GetVirtualUser()
	h.logger.InfoCtxf(ctx, "use virtual user to find host identifier push result, virtual-user(%s), req(%v)", virtualUser, req)
	tenantUserCtx := contextx.NewTenantUserContext(ctx, ctx.TenantID(), virtualUser)

	resp, err := h.cli.findHostIdentifierPushResult(tenantUserCtx, req)
	if err != nil {
		return nil, nil, nil, err
	}

	return resp.SuccessList, resp.FailedList, resp.PendingList, nil
}

// ListResourcePoolHosts list resource pool hosts.
func (h *Handler) ListResourcePoolHosts(ctx contextx.ITenantUserContext, page types.Page) ([]*types.Host, error) {
	req := &ListResourcePoolHostsReq{
		Fields: ccHostFields(),
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listResourcePoolHosts(ctx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]*types.Host, len(resp.Info))
	for idx, host := range resp.Info {
		hosts[idx] = h.convHostInfoToTypes(ctx.TenantID(), host, CCResourcePoolBusinessID)
	}

	return hosts, nil
}

// ListHostsWithoutBusiness list hosts without business.
func (h *Handler) ListHostsWithoutBusiness(ctx contextx.ITenantUserContext, page types.Page) ([]*types.Host, error) {
	req := &ListHostsWithoutBusinessReq{
		Fields: ccHostFields(),
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listHostsWithoutBusiness(ctx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]*types.Host, len(resp.Info))
	for idx, host := range resp.Info {
		hosts[idx] = h.convHostInfoToTypes(ctx.TenantID(), host, CCNoBusinessID)
	}

	return hosts, nil
}

// AddHostToResourcePool add host to resource pool.
// nolint: nonamedreturns
func (h *Handler) AddHostToResourcePool(ctx contextx.ITenantUserContext, hosts ...*types.Host) (successHost []*types.Host,
	failedIndexMsg []string, err error) {

	req := &AddHostToResourcePoolReq{
		HostInfo: make([]*CreateHostInfo, 0, len(hosts)),
	}

	virtualUser := access.GetVirtualUser()
	h.logger.InfoCtxf(ctx, "use virtual user to add host to resource pool, virtual-user(%s), req(%v)", virtualUser, req)
	tenantUserCtx := contextx.NewTenantUserContext(ctx, ctx.TenantID(), virtualUser)

	for _, host := range hosts {
		req.HostInfo = append(req.HostInfo, h.convCreateHostInfoFromTypes(host))
	}

	resp, err := h.cli.addHostToResource(tenantUserCtx, req)
	if err != nil {
		return nil, nil, err
	}

	successHost = make([]*types.Host, len(resp.Success))
	for index, host := range resp.Success {
		successHost[index] = hosts[host.Index]
		successHost[index].HostID = host.BKHostID
	}

	failedIndexMsg = make([]string, len(resp.Error))
	for index, errmsg := range resp.Error {
		failedIndexMsg[index] = fmt.Sprintf("host params index: %d, errmsg: %s", errmsg.Index, errmsg.ErrorMessage)
	}

	return successHost, failedIndexMsg, nil
}

// ExecuteHostDynamicGroup execute dynamic grouping rules to return hosts within the group.
func (h *Handler) ExecuteHostDynamicGroup(ctx contextx.ITenantUserContext, bizID int64, groupID string, page types.Page) (
	[]*types.Host, error) {

	req := &ExecuteDynamicGroupReq{
		BKBizID:        bizID,
		ID:             groupID,
		Fields:         ccHostFields(),
		DisableCounter: false,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.executeDynamicGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Host, len(resp.Info))
	for index, host := range resp.Info {
		jsonData, err := json.Marshal(host)
		if err != nil {
			return nil, err
		}

		var hostData HostInfo
		err = json.Unmarshal(jsonData, &hostData)
		if err != nil {
			return nil, err
		}

		result[index] = h.convHostInfoToTypes(ctx.TenantID(), &hostData, bizID)
	}

	return result, nil
}

// SearchDynamicGroup search dynamic group.
func (h *Handler) SearchDynamicGroup(ctx contextx.ITenantUserContext, bizID int64, page types.Page) (
	[]*types.DynamicGroup, error) {

	req := &SearchDynamicGroupReq{
		BKBizID: bizID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.searchDynamicGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.DynamicGroup, len(resp.Info))
	for index, group := range resp.Info {
		result[index] = &types.DynamicGroup{
			ID:    group.ID,
			BizID: group.BKBizID,
			ObjID: group.BKObjID,
			Name:  group.Name,
		}
	}

	return result, nil
}

// ListServiceTemplate list service template.
func (h *Handler) ListServiceTemplate(ctx contextx.ITenantUserContext, bizID int64, page types.Page) ([]*types.ServiceTemplate,
	error) {

	req := &ListServiceTemplateReq{
		BKBizID: bizID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listServiceTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	serviceTemplates := make([]*types.ServiceTemplate, len(resp.Info))
	for index, serviceTemplate := range resp.Info {
		serviceTemplates[index] = &types.ServiceTemplate{
			ID:                  serviceTemplate.ID,
			BizID:               serviceTemplate.BKBizID,
			ServiceTemplateName: serviceTemplate.ServiceTemplateName,
			ServiceCategoryID:   serviceTemplate.ServiceCategoryID,
			HostApplyEnabled:    serviceTemplate.HostApplyEnabled,
		}
	}

	return serviceTemplates, nil
}

// FindHostByServiceTemplate find host by service template.
func (h *Handler) FindHostByServiceTemplate(ctx contextx.ITenantUserContext, bizID int64, page types.Page,
	serviceTemplateIDs ...int64) ([]*types.Host, error) {

	req := &FindHostByServiceTemplateReq{
		BKBizID:              bizID,
		BKServiceTemplateIDs: serviceTemplateIDs,
		Fields:               ccHostFields(),
		Page: Page{
			Start: page.Offset,
			Sort:  page.Sort,
			Limit: page.Limit,
		},
	}

	resp, err := h.cli.findHostByServiceTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Host, len(resp.Info))
	for index, host := range resp.Info {
		result[index] = h.convHostInfoToTypes(ctx.TenantID(), host, bizID)
	}

	return result, nil
}

// convHostInfoToTypes convert host info to types.Host.
func (h *Handler) convHostInfoToTypes(tenantID string, hostInfo *HostInfo, bizID int64) *types.Host {
	if hostInfo == nil {
		return nil
	}

	data := &types.Host{
		HostID:   hostInfo.BKHostID,
		TenantID: tenantID,
		Static: &types.HostStatic{
			BizID:         bizID,
			NetworkAreaID: hostInfo.BKCloudID,
			RegionID:      hostInfo.BKCloudRegion,
			CityID:        hostInfo.IdcCityID,
			HostName:      hostInfo.BKHostName,
			DeptName:      hostInfo.DeptName,
			InnerIP:       hostInfo.BKHostInnerIPV4,
			InnerIPV6:     hostInfo.BKHostInnerIPV6,
			OuterIP:       hostInfo.BKHostOuterIPV4,
			OuterIPV6:     hostInfo.BKHostOuterIPV6,
			Mac:           hostInfo.BKMac,
			OSTypeCCID:    hostInfo.BKOSType,
			OSType:        h.osTypeKeeper.getValue(hostInfo.BKOSType),
			Arch:          h.cpuArchKeeper.getValue(hostInfo.BKCpuArchitecture),
			Addressing:    types.Addressing(hostInfo.BKAddressing),
			SyncedAgentID: hostInfo.BKAgentID,
		},
		Dynamic: types.NewBlankNodeDynamic(),
	}

	return data
}

func (h *Handler) convCreateHostInfoFromTypes(host *types.Host) *CreateHostInfo {
	return &CreateHostInfo{
		BKCloudID:         host.Static.NetworkAreaID,
		BKHostInnerIP:     host.Static.InnerIP,
		BKHostInnerIPV6:   host.Static.InnerIPV6,
		BKHostOuterIP:     host.Static.OuterIP,
		BKHostOuterIPV6:   host.Static.OuterIPV6,
		BKOSType:          h.osTypeKeeper.getKey(host.Static.OSType),
		BKCpuArchitecture: h.cpuArchKeeper.getKey(host.Static.Arch),
		BKAddressing:      string(host.Static.Addressing),
	}
}

func (h *Handler) convCloudAreaToTypes(tenantID string, cloudArea *CloudArea) *types.NetworkArea {
	return &types.NetworkArea{
		TenantID:    tenantID,
		ID:          cloudArea.BKCloudID,
		Name:        cloudArea.BKCloudName,
		CloudVendor: h.cloudVendorKeeper.getValue(cloudArea.BKCloudVendor),
	}
}

// convHostTopoRelationToTypes convert host topo relation to types.HostRel.
func convHostTopoRelationToTypes(hostRel *HostTopoRelation) *types.Host {
	if hostRel == nil {
		return nil
	}

	return &types.Host{
		HostID: hostRel.BKHostID,
		Static: &types.HostStatic{
			BizID: hostRel.BKBizID,
		},
	}
}

// CheckBizHostByIP check biz host by ip.
func (h *Handler) CheckBizHostByIP(ctx contextx.ITenantUserContext, bizID int64, cloudID int64, ip string) (bool, error) {
	req := &ListBizHostsReq{
		BKBizID: bizID,
		Page: Page{
			Start: 0,
			Limit: 1,
		},
		Fields: ccHostFields(),
	}

	req.HostPropertyFilter = new(HostPropertyFilter)
	req.HostPropertyFilter.Condition = "AND"
	req.HostPropertyFilter.Rules = make([]*FieldCondition, 0)
	req.HostPropertyFilter.Rules = append(req.HostPropertyFilter.Rules, &FieldCondition{
		Field:    "bk_host_innerip",
		Operator: "equal",
		Value:    ip,
	})
	req.HostPropertyFilter.Rules = append(req.HostPropertyFilter.Rules, &FieldCondition{
		Field:    "bk_cloud_id",
		Operator: "equal",
		Value:    cloudID,
	})

	resp, err := h.cli.listBizHosts(ctx, req)
	if err != nil {
		return false, err
	}

	if len(resp.Info) == 0 {
		return false, nil
	}

	return true, nil
}

// WatchHostResourceEvent watch host resource event.
func (h *Handler) WatchHostResourceEvent(ctx contextx.ITenantUserContext, cursor string) ([]*types.HostEvent, error) {
	req := &ResourceWatchReq{
		BKCursor:   cursor,
		BKResource: string(types.ResourceTypeHost),
		BKFields:   ccHostFields(),
	}

	resp, err := h.cli.resourceWatch(ctx, req)
	if err != nil {
		return nil, err
	}

	hostEvents := make([]*types.HostEvent, 0, len(resp.BKEvents))
	for _, hostEvent := range resp.BKEvents {
		hostData := new(HostEventInfo)
		if err := conv.MapToStruct(*hostEvent, hostData); err != nil {
			return nil, err
		}

		hostEvents = append(hostEvents, &types.HostEvent{
			Cursor:    hostData.BKCursor,
			Resource:  types.ResourceTypeHost,
			EventType: types.EventType(hostData.BKEventType),
			Detail:    h.convHostInfoToTypes(ctx.TenantID(), hostData.BKDetail, CCNoBusinessID),
		})
	}

	return hostEvents, nil
}

// WatchHostRelationResourceEvent get host relation resource by watch.
func (h *Handler) WatchHostRelationResourceEvent(ctx contextx.ITenantUserContext, cursor string) ([]*types.HostEvent, error) {
	req := &ResourceWatchReq{
		BKCursor:   cursor,
		BKResource: string(types.ResourceTypeHostRelation),
	}

	resp, err := h.cli.resourceWatch(ctx, req)
	if err != nil {
		return nil, err
	}

	hostEvents := make([]*types.HostEvent, 0)
	for _, relationEvent := range resp.BKEvents {
		relationData := new(HostRelationEventInfo)
		if err := conv.MapToStruct(*relationEvent, relationData); err != nil {
			return nil, err
		}

		hostEvents = append(hostEvents, &types.HostEvent{
			Cursor:    relationData.BKCursor,
			Resource:  types.ResourceTypeHostRelation,
			EventType: types.EventType(relationData.BKEventType),
			Detail:    convHostTopoRelationToTypes(relationData.BKDetail),
		})
	}

	return hostEvents, nil
}
