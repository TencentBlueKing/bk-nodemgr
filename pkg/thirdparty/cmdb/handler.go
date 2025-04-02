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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// This document is responsible for processing the conversion of original requests and responses
// from the third-party system and the internal data of the nodeman system.

// IHandler the Handler of cmdb.
type IHandler interface {
	// ListBizHosts list biz hosts.
	ListBizHosts(ctx context.Context, bizID int64, page types.Page) ([]*types.Host, error)

	// SearchBusiness search business.
	SearchBusiness(ctx context.Context, page types.Page) ([]*types.Business, error)

	// SearchNetworkArea search network area.
	SearchNetworkArea(ctx context.Context, page types.Page) ([]*types.NetworkArea, error)

	// CreateNetworkArea create network area.
	CreateNetworkArea(ctx context.Context, networkAreaName string, cloudVendor string) (*types.NetworkArea, error)

	// UpdateNetworkArea update network area.
	UpdateNetworkArea(ctx context.Context, id int64, networkAreaName string, cloudVendor string) error

	// DeleteNetworkArea delete network area.
	DeleteNetworkArea(ctx context.Context, id int64) error

	// UpdateHostNetworkAreaField update host network area field.
	UpdateHostNetworkAreaField(ctx context.Context, bizID int64, networkAreaID int64, hostIDs ...int64) error

	// GetCloudVendors get cloud vendors.
	GetCloudVendors(ctx context.Context) ([]string, error)

	// BindHostAgent bind host agent
	BindHostAgent(ctx context.Context, hostInfo ...*types.Host) error

	// UnbindHostAgent bind host agent
	UnbindHostAgent(ctx context.Context, hostInfo ...*types.Host) error

	// AddHostToBusinessIdle add host to business idle
	AddHostToBusinessIdle(ctx context.Context, bizID int64, hosts ...*types.Host) ([]int64, error)

	// PushHostIdentifier push host identifier.
	PushHostIdentifier(ctx context.Context, hostIDs ...int64) (taskID string, err error)

	// FindHostIdentifierPushResult find host identifier push result.
	FindHostIdentifierPushResult(ctx context.Context, taskID string) (successList []int64, failedList []int64,
		pendingList []int64, err error)

	// ListResourcePoolHosts list resource pool hosts.
	ListResourcePoolHosts(ctx context.Context, page types.Page) ([]*types.Host, error)

	// ListHostsWithoutBusiness list hosts without business.
	ListHostsWithoutBusiness(ctx context.Context, page types.Page) ([]*types.Host, error)

	// AddHostToResourcePool add host to resource pool
	AddHostToResourcePool(ctx context.Context, hosts ...*types.Host) (successHost []*types.Host,
		failedIndexMsg []string, err error)

	// SearchDynamicGroup search dynamic group.
	SearchDynamicGroup(ctx context.Context, bizID int64, page types.Page) ([]*types.DynamicGroup, error)

	// ExecuteHostDynamicGroup execute dynamic grouping rules to return hosts within the group.
	ExecuteHostDynamicGroup(ctx context.Context, bizID int64, groupID string, page types.Page) ([]*types.Host, error)

	// ListServiceTemplate list service template.
	ListServiceTemplate(ctx context.Context, bizID int64, page types.Page) ([]*types.ServiceTemplate, error)

	// FindHostByServiceTemplate find host by service template.
	FindHostByServiceTemplate(ctx context.Context, bizID int64, page types.Page, serviceTemplateIDs ...int64) (
		[]*types.Host, error)

	// NewWatcher new watcher.
	NewWatcher(tenantID string) (IWatcher, error)

	// CheckBizHostByIP check biz host by ip
	CheckBizHostByIP(ctx context.Context, bizID int64, cloudID int64, ip string) (bool, error)
}

// Handler the Handler of cmdb.
type Handler struct {
	cli    *cli
	logger logger.Logger

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
func WithLogger(logger logger.Logger) OptionFn {
	return func(s *Handler) {
		s.logger = logger
	}
}

// New initialize a new cmdb Handler.
func New(c *client.Capability, conf *Config, opts ...OptionFn) (IHandler, error) {
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
	h.initEnumKeepers()

	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

func (h *Handler) initEnumKeepers() {
	h.scheduler = scheduler.NewScheduler()

	h.scheduler.RegisterTask(&scheduler.Task{
		ID:       "sync_cloud_vendor",
		Interval: enumResourceSyncInterval,
		Timeout:  enumResourceSyncTimeout,
		Fn: func(ctx context.Context) error {
			return h.cloudVendorKeeper.update(ctx)
		},
	})

	h.scheduler.RegisterTask(&scheduler.Task{
		ID:       "sync_os_type",
		Interval: enumResourceSyncInterval,
		Timeout:  enumResourceSyncTimeout,
		Fn: func(ctx context.Context) error {
			return h.osTypeKeeper.update(ctx)
		},
	})

	h.scheduler.RegisterTask(&scheduler.Task{
		ID:       "sync_cpu_arch",
		Interval: enumResourceSyncInterval,
		Timeout:  enumResourceSyncTimeout,
		Fn: func(ctx context.Context) error {
			return h.cpuArchKeeper.update(ctx)
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), enumResourceSyncTimeout)
	defer cancel()

	if err := h.cloudVendorKeeper.update(ctx); err != nil {
		h.logger.Warnf("failed to sync cloud vendor, err: %v", err)
	}
	if err := h.osTypeKeeper.update(ctx); err != nil {
		h.logger.Warnf("failed to sync os type, err: %v", err)
	}
	if err := h.cpuArchKeeper.update(ctx); err != nil {
		h.logger.Warnf("failed to sync cpu arch, err: %v", err)
	}
}

// ListBizHosts list biz hosts.
func (h *Handler) ListBizHosts(ctx context.Context, bizID int64, page types.Page) ([]*types.Host, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListBizHostsReq{
		TenantID: tenantID,
		BKBizID:  bizID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listBizHosts(ctx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]*types.Host, len(resp.Info))
	for idx, host := range resp.Info {
		hosts[idx] = h.convHostInfoToTypes(tenantID, host)
	}

	return hosts, nil
}

// SearchBusiness search business.
func (h *Handler) SearchBusiness(ctx context.Context, page types.Page) ([]*types.Business, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchBusinessReq{
		TenantID: tenantID,
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
			TenantID: tenantID,
			BizID:    business.BKBizID,
			BizName:  business.BKBizName,
		}
	}

	return bizs, nil
}

// SearchNetworkArea search network area.
func (h *Handler) SearchNetworkArea(ctx context.Context, page types.Page) ([]*types.NetworkArea, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchCloudAreaReq{
		TenantID: tenantID,
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
		netAreas[idx] = h.convCloudAreaToTypes(tenantID, networkarea)
	}

	return netAreas, nil
}

// CreateNetworkArea create network area.
func (h *Handler) CreateNetworkArea(ctx context.Context, networkAreaName string, cloudVendor string) (
	*types.NetworkArea, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &CreateCloudAreaReq{
		TenantID:      tenantID,
		BKCloudName:   networkAreaName,
		BKCloudVendor: h.cloudVendorKeeper.getKey(cloudVendor),
	}

	resp, err := h.cli.createCloudArea(ctx, req)
	if err != nil {
		return nil, err
	}

	netArea := &types.NetworkArea{
		TenantID:    tenantID,
		ID:          resp.Created.ID,
		Name:        networkAreaName,
		CloudVendor: cloudVendor,
	}

	return netArea, nil
}

// UpdateNetworkArea update network area.
func (h *Handler) UpdateNetworkArea(
	ctx context.Context, id int64, networkAreaName string, cloudVendor string) error {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &UpdateCloudAreaReq{
		TenantID:      tenantID,
		BKCloudID:     id,
		BKCloudName:   networkAreaName,
		BKCloudVendor: h.cloudVendorKeeper.getKey(cloudVendor),
	}

	err = h.cli.updateCloudArea(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// DeleteNetworkArea delete network area.
func (h *Handler) DeleteNetworkArea(ctx context.Context, id int64) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &DeleteCloudAreaReq{
		TenantID:  tenantID,
		BKCloudID: id,
	}

	err = h.cli.deleteCloudArea(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// UpdateHostNetworkAreaField update host network area field.
func (h *Handler) UpdateHostNetworkAreaField(
	ctx context.Context, bizID int64, networkAreaID int64, hostIDs ...int64) error {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &UpdateHostCloudAreaFieldReq{
		TenantID:  tenantID,
		BKBizID:   bizID,
		BKCloudID: networkAreaID,
		BKHostIDs: hostIDs,
	}

	err = h.cli.updateHostCloudAreaField(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// GetCloudVendors get cloud vendors.
func (h *Handler) GetCloudVendors(_ context.Context) ([]string, error) {
	return h.cloudVendorKeeper.values(), nil
}

// BindHostAgent bind host agent.
func (h *Handler) BindHostAgent(ctx context.Context, hostInfo ...*types.Host) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(hostInfo) == 0 {
		return errors.New("host info list is empty")
	}

	req := &BindHostAgentReq{TenantID: tenantID}
	for _, host := range hostInfo {
		req.List = append(req.List, &HostAgentIDInfo{
			BKHostID:  host.HostID,
			BKAgentID: host.Dynamic.AgentID,
		})
	}
	err = h.cli.bindHostAgent(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// UnbindHostAgent bind host agent.
func (h *Handler) UnbindHostAgent(ctx context.Context, hostInfo ...*types.Host) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(hostInfo) == 0 {
		return errors.New("host info list is empty")
	}

	req := &UnbindHostAgentReq{TenantID: tenantID}
	for _, host := range hostInfo {
		req.List = append(req.List, &HostAgentIDInfo{
			BKHostID:  host.HostID,
			BKAgentID: host.Dynamic.AgentID,
		})
	}
	err = h.cli.unbindHostAgent(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// AddHostToBusinessIdle add host to business idle.
func (h *Handler) AddHostToBusinessIdle(ctx context.Context, bizID int64, hosts ...*types.Host) (
	[]int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	if len(hosts) == 0 {
		return nil, errors.New("host list is empty")
	}

	req := &AddHostToBusinessIdleReq{TenantID: tenantID, BKBizID: bizID}
	for _, host := range hosts {
		req.BKHostList = append(req.BKHostList, h.convCreateHostInfoFromTypes(host))
	}
	resp, err := h.cli.addHostToBusinessIdle(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.BKHostIDs, nil
}

// PushHostIdentifier push host identifier.
// nolint: nonamedreturns
func (h *Handler) PushHostIdentifier(ctx context.Context, hostIDs ...int64) (taskID string, err error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	req := &PushHostIdentifierReq{
		TenantID:  tenantID,
		BKHostIDs: hostIDs,
	}
	resp, err := h.cli.pushHostIdentifier(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.TaskID, nil
}

// FindHostIdentifierPushResult find host identifier push result.
// nolint: nonamedreturns
func (h *Handler) FindHostIdentifierPushResult(ctx context.Context, taskID string) (successList []int64,
	failedList []int64, pendingList []int64, err error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	req := &FindHostIdentifierPushResultReq{
		TenantID: tenantID,

		TaskID: taskID,
	}
	resp, err := h.cli.findHostIdentifierPushResult(ctx, req)
	if err != nil {
		return nil, nil, nil, err
	}

	return resp.SuccessList, resp.FailedList, resp.PendingList, nil
}

// ListResourcePoolHosts list resource pool hosts.
func (h *Handler) ListResourcePoolHosts(ctx context.Context, page types.Page) ([]*types.Host, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListResourcePoolHostsReq{
		TenantID: tenantID,
		Fields:   ccHostFields(),
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
		hosts[idx] = h.convHostInfoToTypes(tenantID, host)
	}

	return hosts, nil
}

// ListHostsWithoutBusiness list hosts without business.
func (h *Handler) ListHostsWithoutBusiness(ctx context.Context, page types.Page) ([]*types.Host, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListHostsWithoutBusinessReq{
		TenantID: tenantID,
		Fields:   ccHostFields(),
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
		hosts[idx] = h.convHostInfoToTypes(tenantID, host)
	}

	return hosts, nil
}

// AddHostToResourcePool add host to resource pool.
// nolint: nonamedreturns
func (h *Handler) AddHostToResourcePool(ctx context.Context, hosts ...*types.Host) (successHost []*types.Host,
	failedIndexMsg []string, err error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, nil, err
	}

	req := &AddHostToResourcePoolReq{TenantID: tenantID}
	for _, host := range hosts {
		req.HostInfo = append(req.HostInfo, h.convCreateHostInfoFromTypes(host))
	}

	resp, err := h.cli.addHostToResource(ctx, req)
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
func (h *Handler) ExecuteHostDynamicGroup(ctx context.Context, bizID int64, groupID string, page types.Page) (
	[]*types.Host, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ExecuteDynamicGroupReq{
		TenantID:       tenantID,
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

	var resp *ExecuteDynamicGroupResp
	resp, err = h.cli.executeDynamicGroup(ctx, req)
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

		result[index] = h.convHostInfoToTypes(tenantID, &hostData)
	}

	return result, nil
}

// SearchDynamicGroup search dynamic group.
func (h *Handler) SearchDynamicGroup(ctx context.Context, bizID int64, page types.Page) (
	[]*types.DynamicGroup, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchDynamicGroupReq{
		TenantID: tenantID,
		BKBizID:  bizID,
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
func (h *Handler) ListServiceTemplate(ctx context.Context, bizID int64, page types.Page) ([]*types.ServiceTemplate,
	error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &ListServiceTemplateReq{
		TenantID: tenantID,

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
func (h *Handler) FindHostByServiceTemplate(ctx context.Context, bizID int64, page types.Page,
	serviceTemplateIDs ...int64) ([]*types.Host, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &FindHostByServiceTemplateReq{
		TenantID: tenantID,

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
		result[index] = h.convHostInfoToTypes(tenantID, host)
	}

	return result, nil
}

// NewWatcher new watcher.
func (h *Handler) NewWatcher(tenantID string) (IWatcher, error) {
	watcher := NewWatcher(tenantID, h)

	return watcher, nil
}

// convHostInfoToTypes convert host info to types.Host.
func (h *Handler) convHostInfoToTypes(tenantID string, hostInfo *HostInfo) *types.Host {
	data := &types.Host{
		HostID:   hostInfo.BKHostID,
		TenantID: tenantID,
		Static: &types.HostStatic{
			BizID:         hostInfo.BKBizID,
			NetworkAreaID: hostInfo.BKCloudID,
			HostName:      hostInfo.BKHostName,
			DeptName:      hostInfo.DeptName,
			InnerIP:       hostInfo.BKHostInnerIPV4,
			InnerIPV6:     hostInfo.BKHostInnerIPV6,
			OuterIP:       hostInfo.BKHostOuterIPV4,
			OuterIPV6:     hostInfo.BKHostOuterIPV6,
			Mac:           hostInfo.BKMac,
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
func convHostTopoRelationToTypes(hostRel *HostTopoRelation) *types.HostRel {
	data := &types.HostRel{
		HostID:   hostRel.BKHostID,
		BizID:    hostRel.BKBizID,
		ModuleID: hostRel.BKModuleID,
		SetID:    hostRel.BKSetID,
	}

	return data
}

// CheckBizHostByIP check biz host by ip.
func (h *Handler) CheckBizHostByIP(ctx context.Context, bizID int64, cloudID int64, ip string) (bool, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return false, err
	}

	req := &ListBizHostsReq{
		TenantID: tenantID,
		BKBizID:  bizID,
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
