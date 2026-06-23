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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
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
	IUpdateHostOpsFields
	IObjectAttribute
	IDynamicGroup
	IServiceTemplate
	IWatch
	IScope
}

// IDynamicGroup this interface is used to edit dynamic group.
type IDynamicGroup interface {
	// SearchDynamicGroup search dynamic group.
	SearchDynamicGroup(nCtx contextx.IContext, bizID int64, page types.Page) ([]*types.DynamicGroup, error)
}

// IBiz this interface is used to edit biz info.
type IBiz interface {
	// SearchBusiness search business.
	SearchBusiness(nCtx contextx.IContext, page types.Page) ([]*types.Business, error)

	// SearchBizInstTopo search business instance topology.
	SearchBizInstTopo(nCtx contextx.IContext, bizID int64) ([]*types.TopoNodeInfo, error)
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
	PushHostIdentifier(nCtx contextx.IContext, hostIDs ...int64) (taskID string, err error)

	// FindHostIdentifierPushResult find host identifier push result.
	FindHostIdentifierPushResult(nCtx contextx.IContext, taskID string) (successList []int64, failedList []int64,
		pendingList []int64, err error)
}

// INetworkArea this interface is used to edit network area.
type INetworkArea interface {
	// SearchNetworkArea search network area.
	SearchNetworkArea(nCtx contextx.IContext, page types.Page) ([]*types.NetworkArea, error)

	// CreateNetworkArea create network area.
	CreateNetworkArea(nCtx contextx.IContext, networkAreaName string, cloudVendor string) (*types.NetworkArea, error)

	// UpdateNetworkArea update network area.
	UpdateNetworkArea(nCtx contextx.IContext, id int64, networkAreaName string, cloudVendor string) error

	// DeleteNetworkArea delete network area.
	DeleteNetworkArea(nCtx contextx.IContext, id int64) error
}

// IBindHostAgent this interface is used to bind host agent.
type IBindHostAgent interface {
	BindHostAgent(nCtx contextx.IContext, hostInfo ...*types.Host) error
}

// IUnbindHostAgent this interface is used to unbind host agent.
type IUnbindHostAgent interface {
	UnbindHostAgent(nCtx contextx.IContext, hostInfo ...*types.Host) error
}

// IUpdateHostNetworkAreaField this interface is used to update host network area field.
type IUpdateHostNetworkAreaField interface {
	// UpdateHostNetworkAreaField update host network area field.
	UpdateHostNetworkAreaField(nCtx contextx.IContext, bizID int64, networkAreaID int64, hostIDs ...int64) error
}

// IUpdateHostOpsFields updates host out-of-band (ops) fields in CMDB.
type IUpdateHostOpsFields interface {
	// UpdateHostOpsFields batch-updates ops fields for one or more hosts in CMDB.
	UpdateHostOpsFields(nCtx contextx.IContext, hosts ...*types.Host) error
}

// IServiceTemplate this interface is used to list service template.
type IServiceTemplate interface {
	// ListServiceTemplate list service template.
	ListServiceTemplate(nCtx contextx.IContext, bizID int64, page types.Page) ([]*types.ServiceTemplate, error)
}

// IWatch this interface is used to watch resource event.
type IWatch interface {
	// WatchHostResourceEvent watch host resource event.
	WatchHostResourceEvent(nCtx contextx.IContext, cursor string) ([]*types.HostEvent, error)

	// WatchHostRelationResourceEvent watch host relation resource event.
	WatchHostRelationResourceEvent(nCtx contextx.IContext, cursor string) ([]*types.HostEvent, error)
}

// IFinder this interface is used to find host.
type IFinder interface {
}

// Handler the Handler of cmdb.
type Handler struct {
	cli *cli

	scheduler scheduler.Scheduler

	cloudVendorKeeper iEnumResourceKeeper
	osTypeKeeper      iEnumResourceKeeper
	cpuArchKeeper     iEnumResourceKeeper

	// combined handler group splited by tenant.
	combinedHandlerGroupMu sync.RWMutex
	combinedHandlerGroup   map[string]*combinedHandler
}

const (
	enumResourceSyncInterval = 30 * time.Minute
	enumResourceSyncTimeout  = 30 * time.Second
)

// OptionFn ...
type OptionFn func(*Handler)

// New initialize a new cmdb Handler.
func New(c *restclient.Capability, conf *Config, opts ...OptionFn) (IHandler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		cli: cli,

		cloudVendorKeeper: newCloudVendorKeeper(cli),
		osTypeKeeper:      newOSTypeKeeper(cli),
		cpuArchKeeper:     newCPUArchKeeper(cli),

		combinedHandlerGroup: make(map[string]*combinedHandler),
	}

	for _, opt := range opts {
		opt(h)
	}

	err = h.initEnumKeepers()
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to init enum keepers")

		return nil, err
	}

	return h, nil
}

// nolint: gocognit, funlen
func (h *Handler) initEnumKeepers() error {
	logger.G.Sys().Info("initializing enum keepers from cmdb")

	if h.scheduler != nil {
		h.scheduler.Terminate()
	}
	h.scheduler = scheduler.NewScheduler()
	syncTasks := []*scheduler.Task{
		scheduler.NewTask(
			"sync_cloud_vendor",
			enumResourceSyncInterval,
			enumResourceSyncTimeout,
			func(nCtx contextx.IContext) error {
				tenantIDs := tenant.GetAllTenantIDs()
				for _, tenantID := range tenantIDs {
					newCtx := contextx.From(nCtx, contextx.WithTenantID(tenantID), contextx.WithBKUsername(h.cli.config.VirtualUser))
					if err := h.cloudVendorKeeper.update(newCtx); err != nil {
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
			func(nCtx contextx.IContext) error {
				tenantIDs := tenant.GetAllTenantIDs()
				for _, tenantID := range tenantIDs {
					newCtx := contextx.From(nCtx, contextx.WithTenantID(tenantID), contextx.WithBKUsername(h.cli.config.VirtualUser))
					if err := h.osTypeKeeper.update(newCtx); err != nil {
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
			func(nCtx contextx.IContext) error {
				tenantIDs := tenant.GetAllTenantIDs()
				for _, tenantID := range tenantIDs {
					newCtx := contextx.From(nCtx, contextx.WithTenantID(tenantID), contextx.WithBKUsername(h.cli.config.VirtualUser))
					if err := h.cpuArchKeeper.update(newCtx); err != nil {
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
			logger.G.Sys().WithErr(err).With("task-id", task.ID).Error("failed to register sync task")

			return err
		}
	}

	nCtx, cancel := context.WithTimeout(context.Background(), enumResourceSyncTimeout)
	defer cancel()

	tenantIDs := tenant.GetAllTenantIDs()
	for _, tenantID := range tenantIDs {
		newCtx := contextx.New(nCtx, contextx.WithTenantID(tenantID), contextx.WithBKUsername(h.cli.config.VirtualUser))
		if err := h.cloudVendorKeeper.update(newCtx); err != nil {
			logger.G.Sys().WithErr(err).Warn("failed to sync cloud vendor")
		}
		if err := h.osTypeKeeper.update(newCtx); err != nil {
			logger.G.Sys().WithErr(err).Warn("failed to sync os type")
		}
		if err := h.cpuArchKeeper.update(newCtx); err != nil {
			logger.G.Sys().WithErr(err).Warn("failed to sync cpu arch")
		}
	}

	h.scheduler.Start()
	logger.G.Sys().Info("started schedule enum resource sync tasks")

	return nil
}

// SearchBusiness search business.
func (h *Handler) SearchBusiness(nCtx contextx.IContext, page types.Page) ([]*types.Business, error) {
	req := &SearchBusinessReq{
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
		Fields: nil,
	}

	resp, err := h.cli.searchBusiness(nCtx, req)
	if err != nil {
		return nil, err
	}

	bizs := make([]*types.Business, len(resp.Info))
	for idx, business := range resp.Info {
		bizs[idx] = &types.Business{
			TenantID: nCtx.TenantID(),
			BizID:    business.BKBizID,
			BizName:  business.BKBizName,
		}
	}

	return bizs, nil
}

// SearchBizInstTopo search business instance topology.
func (h *Handler) SearchBizInstTopo(nCtx contextx.IContext, bizID int64) ([]*types.TopoNodeInfo, error) {
	req := &SearchBizInstTopoReq{BKBizID: bizID}
	resp, err := h.cli.searchBizInstTopo(nCtx, req)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, nil
	}

	result := make([]*types.TopoNodeInfo, len(*resp))
	for idx, topo := range *resp {
		result[idx] = convBizInstTopoToTypes(topo)
	}

	return result, nil
}

func convBizInstTopoToTypes(topo *BizInstTopo) *types.TopoNodeInfo {
	if topo == nil {
		return nil
	}

	children := make([]*types.TopoNodeInfo, 0, len(topo.Children))
	for _, child := range topo.Children {
		if child == nil {
			continue
		}
		children = append(children, convBizInstTopoToTypes(child))
	}

	return &types.TopoNodeInfo{
		InstID:   topo.BKInstID,
		InstName: topo.BKInstName,
		ObjID:    topo.BKObjID,
		ObjName:  topo.BKObjName,
		Children: children,
	}
}

// SearchNetworkArea search network area.
func (h *Handler) SearchNetworkArea(nCtx contextx.IContext, page types.Page) ([]*types.NetworkArea, error) {
	req := &SearchCloudAreaReq{
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.searchCloudArea(nCtx, req)
	if err != nil {
		return nil, err
	}

	netAreas := make([]*types.NetworkArea, len(resp.Info))
	for idx, networkarea := range resp.Info {
		netAreas[idx] = h.convCloudAreaToTypes(nCtx.TenantID(), networkarea)
	}

	return netAreas, nil
}

// CreateNetworkArea create network area.
func (h *Handler) CreateNetworkArea(nCtx contextx.IContext, networkAreaName string, cloudVendor string) (
	*types.NetworkArea, error) {

	req := &CreateCloudAreaReq{
		BKCloudName:   networkAreaName,
		BKCloudVendor: h.cloudVendorKeeper.getKey(cloudVendor),
	}

	resp, err := h.cli.createCloudArea(nCtx, req)
	if err != nil {
		return nil, err
	}

	netArea := &types.NetworkArea{
		TenantID:    nCtx.TenantID(),
		ID:          resp.Created.ID,
		Name:        networkAreaName,
		CloudVendor: cloudVendor,
	}

	return netArea, nil
}

// UpdateNetworkArea update network area.
func (h *Handler) UpdateNetworkArea(
	nCtx contextx.IContext, id int64, networkAreaName string, cloudVendor string) error {

	req := &UpdateCloudAreaReq{
		BKCloudID:     id,
		BKCloudName:   networkAreaName,
		BKCloudVendor: h.cloudVendorKeeper.getKey(cloudVendor),
	}

	err := h.cli.updateCloudArea(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// DeleteNetworkArea delete network area.
func (h *Handler) DeleteNetworkArea(nCtx contextx.IContext, id int64) error {
	req := &DeleteCloudAreaReq{
		BKCloudID: id,
	}

	err := h.cli.deleteCloudArea(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// UpdateHostNetworkAreaField update host network area field.
func (h *Handler) UpdateHostNetworkAreaField(
	nCtx contextx.IContext, bizID int64, networkAreaID int64, hostIDs ...int64) error {

	req := &UpdateHostCloudAreaFieldReq{
		BKBizID:   bizID,
		BKCloudID: networkAreaID,
		BKHostIDs: hostIDs,
	}

	err := h.cli.updateHostCloudAreaField(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// UpdateHostOpsFields batch-updates ops (out-of-band) fields for the given hosts in CMDB.
func (h *Handler) UpdateHostOpsFields(nCtx contextx.IContext, hosts ...*types.Host) error {
	if len(hosts) == 0 {
		return nil
	}

	updates := make([]*UpdateHostProperties, 0, len(hosts))
	for _, host := range hosts {
		if host == nil || host.Dynamic == nil {
			continue
		}

		item := &UpdateHostProperties{BKHostID: host.HostID}
		item.Properties.OpsConsoleHostID = &host.Dynamic.OpsConsoleHostID
		item.Properties.OpsOutBandType = &host.Dynamic.OpsOutBandType
		item.Properties.OpsOutBandProtocol = &host.Dynamic.OpsOutBandProtocol
		item.Properties.OpsBMCIP = &host.Dynamic.OpsBMCIP
		item.Properties.OpsBMCPort = &host.Dynamic.OpsBMCPort
		updates = append(updates, item)
	}

	if len(updates) == 0 {
		return nil
	}

	req := &BatchUpdateHostReq{Update: updates}

	return h.cli.batchUpdateHost(nCtx, req)
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
func (h *Handler) BindHostAgent(nCtx contextx.IContext, hostInfo ...*types.Host) error {
	if len(hostInfo) == 0 {
		return errors.New("host info list is empty")
	}

	reqList := make([]*HostAgentIDInfo, len(hostInfo))
	for idx := range hostInfo {
		reqList[idx] = &HostAgentIDInfo{
			BKHostID:  hostInfo[idx].HostID,
			BKAgentID: hostInfo[idx].Dynamic.AgentID,
		}
	}

	_, _, err := h.getCombinedHandler(nCtx).bindHostAgentCombinedHandler.Call(nCtx, reqList...)

	return err
}

// UnbindHostAgent unbind host agent.
func (h *Handler) UnbindHostAgent(nCtx contextx.IContext, hostInfo ...*types.Host) error {
	if len(hostInfo) == 0 {
		return errors.New("host info list is empty")
	}

	reqList := make([]*HostAgentIDInfo, len(hostInfo))
	for idx := range hostInfo {
		reqList[idx] = &HostAgentIDInfo{
			BKHostID:  hostInfo[idx].HostID,
			BKAgentID: hostInfo[idx].Dynamic.AgentID,
		}
	}

	_, _, err := h.getCombinedHandler(nCtx).unbindHostAgentCombinedHandler.Call(nCtx, reqList...)

	return err
}

// AddHostToBusinessIdle add host to business idle.
func (h *Handler) AddHostToBusinessIdle(nCtx contextx.IContext, bizID int64, hosts ...*types.Host) (
	[]int64, error) {

	if len(hosts) == 0 {
		return nil, errors.New("host list is empty")
	}

	reqList := make([]*CreateHostInfo, len(hosts))
	for idx := range hosts {
		reqList[idx] = h.convCreateHostInfoFromTypes(hosts[idx])
	}

	bizIDStr, _ := conv.ToString(bizID)
	resp, beginIndex, err := h.getCombinedHandler(nCtx).addHostToBusinessIdleCombinedHandler.CallWithAggregationKey(nCtx, bizIDStr, reqList...)
	if err != nil {
		return nil, err
	}

	endIndex := beginIndex + len(reqList)
	if beginIndex < 0 || endIndex > len(resp.BKHostIDs) {
		return nil, fmt.Errorf("host ids length is not match, begin(%d), len(%d), total(%d)", beginIndex, len(reqList), len(resp.BKHostIDs))
	}

	// Note: the returned slice shares the underlying array with resp.BKHostIDs. Callers must drop their references
	// right after short-lived usage and must NOT cache it in long-running goroutines or structs to avoid memory leaks.
	return resp.BKHostIDs[beginIndex:endIndex], nil
}

// PushHostIdentifier push host identifier.
// nolint: nonamedreturns
func (h *Handler) PushHostIdentifier(nCtx contextx.IContext, hostIDs ...int64) (taskID string, err error) {
	resp, _, err := h.getCombinedHandler(nCtx).pushHostIdentifierCombinedHandler.Call(nCtx, hostIDs...)
	if err != nil {
		return "", err
	}

	return resp.TaskID, nil
}

// FindHostIdentifierPushResult find host identifier push result.
// nolint: nonamedreturns
func (h *Handler) FindHostIdentifierPushResult(nCtx contextx.IContext, taskID string) (successList []int64,
	failedList []int64, pendingList []int64, err error) {

	resp, _, err := h.getCombinedHandler(nCtx).findHostIdentifierPushResultCombinedHandler.CallWithAggregationKey(nCtx, taskID, struct{}{})
	if err != nil {
		return nil, nil, nil, err
	}

	return resp.SuccessList, resp.FailedList, resp.PendingList, nil
}

// AddHostToResourcePool add host to resource pool.
// nolint: nonamedreturns
func (h *Handler) AddHostToResourcePool(nCtx contextx.IContext, hosts ...*types.Host) (successHost []*types.Host,
	failedIndexMsg []string, err error) {

	reqList := make([]*CreateHostInfo, len(hosts))

	for idx := range hosts {
		reqList[idx] = h.convCreateHostInfoFromTypes(hosts[idx])
	}

	resp, _, err := h.getCombinedHandler(nCtx).addHostToResourcePoolCombinedHandler.Call(nCtx, reqList...)
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

// SearchDynamicGroup search dynamic group.
func (h *Handler) SearchDynamicGroup(nCtx contextx.IContext, bizID int64, page types.Page) (
	[]*types.DynamicGroup, error) {

	req := &SearchDynamicGroupReq{
		BKBizID: bizID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.searchDynamicGroup(nCtx, req)
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
func (h *Handler) ListServiceTemplate(nCtx contextx.IContext, bizID int64, page types.Page) ([]*types.ServiceTemplate,
	error) {

	req := &ListServiceTemplateReq{
		BKBizID: bizID,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listServiceTemplate(nCtx, req)
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

// ipSeparator is the separator of inner and outer ip.
const ipSeparator = ","

// convHostInfoToTypes convert host info to types.Host.
func (h *Handler) convHostInfoToTypes(tenantID string, hostInfo *HostInfo, bizID int64) *types.Host {
	if hostInfo == nil {
		return nil
	}

	data := &types.Host{
		HostID:   hostInfo.BKHostID,
		TenantID: tenantID,
		Static: &types.HostStatic{
			BizID:                    bizID,
			NetworkAreaID:            hostInfo.BKCloudID,
			RegionID:                 hostInfo.BKCloudRegion,
			CityID:                   hostInfo.IdcCityID,
			HostName:                 hostInfo.BKHostName,
			DeptName:                 hostInfo.DeptName,
			InnerIPList:              strings.Split(hostInfo.BKHostInnerIPV4, ipSeparator),
			InnerIPV6List:            strings.Split(hostInfo.BKHostInnerIPV6, ipSeparator),
			OuterIPList:              strings.Split(hostInfo.BKHostOuterIPV4, ipSeparator),
			OuterIPV6List:            strings.Split(hostInfo.BKHostOuterIPV6, ipSeparator),
			Operator:                 hostInfo.Operator,
			Mac:                      hostInfo.BKMac,
			CPUNum:                   hostInfo.BKCpu,
			MemCap:                   hostInfo.BKMem,
			OSTypeCCID:               hostInfo.BKOSType,
			OSType:                   h.osTypeKeeper.getValue(hostInfo.BKOSType),
			Arch:                     h.cpuArchKeeper.getValue(hostInfo.BKCpuArchitecture),
			Addressing:               types.Addressing(hostInfo.BKAddressing),
			SyncedAgentID:            hostInfo.BKAgentID,
			SyncedOpsConsoleHostID:   hostInfo.OpsConsoleHostID,
			SyncedOpsOutBandType:     hostInfo.OpsOutBandType,
			SyncedOpsOutBandProtocol: hostInfo.OpsOutBandProtocol,
			SyncedOpsBMCIP:           hostInfo.OpsBMCIP,
			SyncedOpsBMCPort:         hostInfo.OpsBMCPort,
		},
		Dynamic: types.NewBlankNodeDynamic(),
	}

	return data
}

func (h *Handler) convCreateHostInfoFromTypes(host *types.Host) *CreateHostInfo {
	return &CreateHostInfo{
		BKCloudID:         host.Static.NetworkAreaID,
		BKHostInnerIP:     strings.Join(host.Static.InnerIPList, ipSeparator),
		BKHostInnerIPV6:   strings.Join(host.Static.InnerIPV6List, ipSeparator),
		BKHostOuterIP:     strings.Join(host.Static.OuterIPList, ipSeparator),
		BKHostOuterIPV6:   strings.Join(host.Static.OuterIPV6List, ipSeparator),
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
func convHostTopoRelationToTypes(tenantID string, hostRel *HostTopoRelation) *types.Host {
	if hostRel == nil {
		return nil
	}

	return &types.Host{
		TenantID: tenantID,
		HostID:   hostRel.BKHostID,
		Static: &types.HostStatic{
			BizID:    hostRel.BKBizID,
			SetID:    hostRel.BKSetID,
			ModuleID: hostRel.BKModuleID,
		},
	}
}

// WatchHostResourceEvent watch host resource event.
func (h *Handler) WatchHostResourceEvent(nCtx contextx.IContext, cursor string) ([]*types.HostEvent, error) {
	req := &ResourceWatchReq{
		BKCursor:   cursor,
		BKResource: string(types.ResourceTypeHost),
		BKFields:   ccHostFields(),
	}

	resp, err := h.cli.resourceWatch(nCtx, req)
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
			Detail:    h.convHostInfoToTypes(nCtx.TenantID(), hostData.BKDetail, CCNoBusinessID),
		})
	}

	return hostEvents, nil
}

// WatchHostRelationResourceEvent get host relation resource by watch.
func (h *Handler) WatchHostRelationResourceEvent(nCtx contextx.IContext, cursor string) ([]*types.HostEvent, error) {
	req := &ResourceWatchReq{
		BKCursor:   cursor,
		BKResource: string(types.ResourceTypeHostRelation),
	}

	resp, err := h.cli.resourceWatch(nCtx, req)
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
			Detail:    convHostTopoRelationToTypes(nCtx.TenantID(), relationData.BKDetail),
		})
	}

	return hostEvents, nil
}

// IHost this interface is used to edit host info.
type IHost interface {
	// ListBizHosts list biz hosts.
	ListBizHosts(nCtx contextx.IContext, bizID int64, page types.Page) ([]*types.Host, error)

	// ListBizHostTopoRelations lists biz host topo relations.
	ListBizHostTopoRelations(nCtx contextx.IContext, bizID int64, page types.Page) ([]*types.Host, error)

	// ListHostsWithoutBusiness list hosts without business.
	ListHostsWithoutBusiness(nCtx contextx.IContext, page types.Page) ([]*types.Host, error)

	// CheckBizHostByIP check biz host by ip
	CheckBizHostByIP(nCtx contextx.IContext, bizID int64, cloudID int64, ip string) (bool, error)

	// ListResourcePoolHosts list resource pool hosts.
	ListResourcePoolHosts(nCtx contextx.IContext, page types.Page) ([]*types.Host, error)

	// AddHostToResourcePool add host to resource pool
	AddHostToResourcePool(nCtx contextx.IContext, hosts ...*types.Host) (successHost []*types.Host,
		failedIndexMsg []string, err error)

	// AddHostToBusinessIdle add host to business idle
	AddHostToBusinessIdle(nCtx contextx.IContext, bizID int64, hosts ...*types.Host) ([]int64, error)

	// FindHostWithCondition find host with condition.
	FindHostWithCondition(nCtx contextx.IContext, page types.Page, cond *types.HostStaticExactCondition) ([]*types.Host, error)

	// FindHostByServiceTemplate find host by service template.
	FindHostByServiceTemplate(nCtx contextx.IContext, bizID int64, page types.Page, serviceTemplateIDs []int64, moduleIDs []int64) (
		[]*types.Host, error)

	// FindHostBySetTemplate find host by set template.
	FindHostBySetTemplate(nCtx contextx.IContext, bizID int64, page types.Page, setTemplateIDs []int64, setIDs []int64) ([]*types.Host, error)
}

// ListBizHosts list biz hosts.
// Deprecated: use FindHostWithCondition instead.
func (h *Handler) ListBizHosts(nCtx contextx.IContext, bizID int64, page types.Page) ([]*types.Host, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to list biz hosts, nCtx is nil")
	}

	if bizID == CCInvalidID {
		return nil, fmt.Errorf("failed to list biz hosts, bizID is invalid")
	}

	tenantID := nCtx.TenantID()
	bkUsername := nCtx.BKUsername()

	executor := pageexecutor.NewPageExecutor[*types.Host](CCPageSizeLimit, 1*time.Hour) // nolint: mnd
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		req := &ListBizHostsReq{
			BKBizID: bizID,
			Page: Page{
				Start: p.Offset,
				Limit: p.Limit,
				Sort:  p.Sort,
			},
		}

		newCtx := contextx.New(nCtx, contextx.WithTenantID(tenantID), contextx.WithBKUsername(bkUsername))
		resp, err := h.cli.listBizHosts(newCtx, req)
		if err != nil {
			return nil, err
		}

		hosts := make([]*types.Host, len(resp.Info))
		for idx, host := range resp.Info {
			hosts[idx] = h.convHostInfoToTypes(tenantID, host, bizID)
		}

		return hosts, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, fmt.Errorf("execute page executor failed: %v", err)
	}

	return result.Items, nil
}

// ListBizHostTopoRelations lists biz host topo relations.
// Deprecated: use FindHostWithCondition instead.
func (h *Handler) ListBizHostTopoRelations(nCtx contextx.IContext, bizID int64, page types.Page) ([]*types.Host, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to list biz host topo relations, nCtx is nil")
	}

	if bizID == CCInvalidID {
		return nil, fmt.Errorf("failed to list biz host topo relations, bizID is invalid")
	}

	tenantID := nCtx.TenantID()
	bkUsername := nCtx.BKUsername()

	executor := pageexecutor.NewPageExecutor[*types.Host](CCPageSizeLimit, 1*time.Hour) // nolint: mnd
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		req := &FindHostTopoRelationReq{
			BKBizID: bizID,
			Page: Page{
				Start: p.Offset,
				Limit: p.Limit,
				Sort:  p.Sort,
			},
		}

		newCtx := contextx.New(nCtx, contextx.WithTenantID(tenantID), contextx.WithBKUsername(bkUsername))
		resp, err := h.cli.findHostTopoRelation(newCtx, req)
		if err != nil {
			return nil, err
		}

		hosts := make([]*types.Host, len(resp.Data))
		for idx, relation := range resp.Data {
			hosts[idx] = convHostTopoRelationToTypes(tenantID, relation)
		}

		return hosts, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, fmt.Errorf("execute page executor failed: %w", err)
	}

	return result.Items, nil
}

// ListHostsWithoutBusiness list hosts without business.
// Deprecated: use FindHostWithCondition instead.
func (h *Handler) ListHostsWithoutBusiness(nCtx contextx.IContext, page types.Page) ([]*types.Host, error) {
	req := &ListHostsWithoutBusinessReq{
		Fields: ccHostFields(),
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listHostsWithoutBusiness(nCtx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]*types.Host, len(resp.Info))
	for idx, host := range resp.Info {
		hosts[idx] = h.convHostInfoToTypes(nCtx.TenantID(), host, CCNoBusinessID)
	}

	return hosts, nil
}

// FindHostWithCondition find host with condition.
func (h *Handler) FindHostWithCondition(nCtx contextx.IContext, page types.Page, cond *types.HostStaticExactCondition) (
	[]*types.Host, error) {

	if nCtx == nil {
		return nil, fmt.Errorf("failed to find host with condition, nCtx is nil")
	}

	bizIDs := convHostStaticExactConditionToBizIDs(cond)
	filter := convHostStaticExactConditionToFilter(cond)
	executor := pageexecutor.NewPageExecutor[*types.Host](CCPageSizeLimit, 1*time.Hour) // nolint: mnd

	hosts := make([]*types.Host, 0)
	if len(bizIDs) > 0 {
		for idx := range bizIDs {
			bizID := bizIDs[idx]

			fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
				return h.listHostWithBiz(nCtx, p, bizID, filter)
			}

			result, err := executor.Execute(nCtx, page, fn)
			if err != nil {
				return nil, fmt.Errorf("execute page executor failed: %v", err)
			}

			hosts = append(hosts, result.Items...)
		}
	} else {
		fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
			return h.listHostWithoutBiz(nCtx, p, filter)
		}

		result, err := executor.Execute(nCtx, page, fn)
		if err != nil {
			return nil, fmt.Errorf("execute page executor failed: %v", err)
		}

		hosts = append(hosts, result.Items...)
	}

	return hosts, nil
}

func (h *Handler) listHostWithBiz(nCtx contextx.IContext, p types.Page, bizID int64, filter *HostPropertyFilter) ([]*types.Host, error) {
	req := &ListBizHostsReq{
		Page: Page{
			Start: p.Offset,
			Limit: p.Limit,
			Sort:  p.Sort,
		},
		BKBizID:            bizID,
		Fields:             ccHostFields(),
		HostPropertyFilter: filter,
	}

	// if filter is not nil and has no rules, set it to nil to avoid unnecessary filtering
	if req.HostPropertyFilter != nil && len(req.HostPropertyFilter.Rules) == 0 {
		req.HostPropertyFilter = nil
	}

	resp, err := h.cli.listBizHosts(nCtx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]*types.Host, len(resp.Info))
	hostIDs := make([]int64, len(resp.Info))
	for idx, host := range resp.Info {
		hosts[idx] = h.convHostInfoToTypes(nCtx.TenantID(), host, bizID)
		hostIDs[idx] = host.BKHostID
	}

	findHostBizRelationsReq := &FindHostBizRelationsReq{
		BKHostID: hostIDs,
	}
	findHostBizRelationsResp, err := h.cli.findHostBizRelations(nCtx, findHostBizRelationsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list host without biz: %w", err)
	}

	hostRel, err := conv.SliceToMap[int64, *HostTopoRelation](*findHostBizRelationsResp, func(rel *HostTopoRelation) int64 {
		return rel.BKHostID
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert host biz relations to map: %w", err)
	}

	for _, host := range hosts {
		if _, ok := hostRel[host.HostID]; !ok {
			continue
		}

		host.Static.SetID = hostRel[host.HostID].BKSetID
		host.Static.ModuleID = hostRel[host.HostID].BKModuleID
	}

	return hosts, nil
}

func (h *Handler) listHostWithoutBiz(nCtx contextx.IContext, p types.Page, filter *HostPropertyFilter) ([]*types.Host, error) {
	listHostsWithoutBusinessReq := &ListHostsWithoutBusinessReq{
		Page: Page{
			Start: p.Offset,
			Limit: p.Limit,
			Sort:  p.Sort,
		},
		Fields:             ccHostFields(),
		HostPropertyFilter: filter,
	}

	// if filter is not nil and has no rules, set it to nil to avoid unnecessary filtering
	if listHostsWithoutBusinessReq.HostPropertyFilter != nil && len(listHostsWithoutBusinessReq.HostPropertyFilter.Rules) == 0 {
		listHostsWithoutBusinessReq.HostPropertyFilter = nil
	}

	listHostsWithoutBusinessResp, err := h.cli.listHostsWithoutBusiness(nCtx, listHostsWithoutBusinessReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list host without biz: %w", err)
	}

	hosts := make([]*types.Host, len(listHostsWithoutBusinessResp.Info))
	for idx, host := range listHostsWithoutBusinessResp.Info {
		hosts[idx] = h.convHostInfoToTypes(nCtx.TenantID(), host, 0)
	}

	hostIDs := conv.SliceToSlice[*types.Host, int64](hosts, func(host *types.Host) int64 {
		return host.HostID
	})

	findHostBizRelationsReq := &FindHostBizRelationsReq{
		BKHostID: hostIDs,
	}
	findHostBizRelationsResp, err := h.cli.findHostBizRelations(nCtx, findHostBizRelationsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list host without biz: %w", err)
	}

	hostBizRel, err := conv.SliceToMap[int64, *HostTopoRelation](*findHostBizRelationsResp, func(rel *HostTopoRelation) int64 {
		return rel.BKHostID
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list host without biz: %w", err)
	}

	for _, host := range hosts {
		if _, ok := hostBizRel[host.HostID]; !ok {
			continue
		}

		host.Static.BizID = hostBizRel[host.HostID].BKBizID
		host.Static.SetID = hostBizRel[host.HostID].BKSetID
		host.Static.ModuleID = hostBizRel[host.HostID].BKModuleID
	}

	return hosts, nil
}

func convHostStaticExactConditionToBizIDs(cond *types.HostStaticExactCondition) []int64 {
	if cond == nil {
		return []int64{}
	}

	if cond.StaticExactInclude == nil {
		return []int64{}
	}

	return cond.StaticExactInclude.BizID
}

func convHostStaticExactConditionToFilter(cond *types.HostStaticExactCondition) *HostPropertyFilter {
	if cond == nil {
		return &HostPropertyFilter{}
	}

	filter := &HostPropertyFilter{
		Condition: hostPropertyFilterConditionAnd,
		Rules:     make([]*FieldCondition, 0),
	}

	if cond.StaticExactInclude != nil {
		if len(cond.StaticExactInclude.HostID) > 0 {
			filter.Rules = append(filter.Rules, &FieldCondition{
				Field:    ccFieldBKHostID,
				Operator: ruleOperatorIn,
				Value:    cond.StaticExactInclude.HostID,
			})
		}

		if len(cond.StaticExactInclude.NetworkAreaID) > 0 {
			filter.Rules = append(filter.Rules, &FieldCondition{
				Field:    ccFieldBKCloudID,
				Operator: ruleOperatorIn,
				Value:    cond.StaticExactInclude.NetworkAreaID,
			})
		}

		if len(cond.StaticExactInclude.InnerIP) > 0 {
			filter.Rules = append(filter.Rules, &FieldCondition{
				Field:    ccFieldBKInnerIP,
				Operator: ruleOperatorIn,
				Value:    cond.StaticExactInclude.InnerIP,
			})
		}

		if len(cond.StaticExactInclude.Addressing) > 0 {
			filter.Rules = append(filter.Rules, &FieldCondition{
				Field:    ccFieldBKAddressing,
				Operator: ruleOperatorIn,
				Value:    cond.StaticExactInclude.Addressing,
			})
		}
	}

	if cond.StaticExactExclude != nil {
		if len(cond.StaticExactExclude.HostID) > 0 {
			filter.Rules = append(filter.Rules, &FieldCondition{
				Field:    ccFieldBKHostID,
				Operator: ruleOperatorIn,
				Value:    cond.StaticExactExclude.HostID,
			})
		}

		if len(cond.StaticExactExclude.NetworkAreaID) > 0 {
			filter.Rules = append(filter.Rules, &FieldCondition{
				Field:    ccFieldBKCloudID,
				Operator: ruleOperatorIn,
				Value:    cond.StaticExactExclude.NetworkAreaID,
			})
		}

		if len(cond.StaticExactExclude.InnerIP) > 0 {
			filter.Rules = append(filter.Rules, &FieldCondition{
				Field:    ccFieldBKInnerIP,
				Operator: ruleOperatorIn,
				Value:    cond.StaticExactExclude.InnerIP,
			})
		}

		if len(cond.StaticExactExclude.Addressing) > 0 {
			filter.Rules = append(filter.Rules, &FieldCondition{
				Field:    ccFieldBKAddressing,
				Operator: ruleOperatorIn,
				Value:    cond.StaticExactExclude.Addressing,
			})
		}
	}

	return filter
}

// CheckBizHostByIP check biz host by ip.
func (h *Handler) CheckBizHostByIP(nCtx contextx.IContext, bizID int64, cloudID int64, ip string) (bool, error) {
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
		Field:    ccFieldBKInnerIP,
		Operator: ruleOperatorEqual,
		Value:    ip,
	})
	req.HostPropertyFilter.Rules = append(req.HostPropertyFilter.Rules, &FieldCondition{
		Field:    ccFieldBKCloudID,
		Operator: ruleOperatorEqual,
		Value:    cloudID,
	})

	resp, err := h.cli.listBizHosts(nCtx, req)
	if err != nil {
		return false, err
	}

	if len(resp.Info) == 0 {
		return false, nil
	}

	return true, nil
}

// ListResourcePoolHosts list resource pool hosts.
func (h *Handler) ListResourcePoolHosts(nCtx contextx.IContext, page types.Page) ([]*types.Host, error) {
	req := &ListResourcePoolHostsReq{
		Fields: ccHostFields(),
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listResourcePoolHosts(nCtx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]*types.Host, len(resp.Info))
	for idx, host := range resp.Info {
		hosts[idx] = h.convHostInfoToTypes(nCtx.TenantID(), host, CCResourcePoolBusinessID)
	}

	return hosts, nil
}

// FindHostByServiceTemplate find host by service template.
func (h *Handler) FindHostByServiceTemplate(nCtx contextx.IContext, bizID int64, page types.Page, serviceTemplateIDs []int64, modleIDs []int64) (
	[]*types.Host, error) {

	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		req := &FindHostByServiceTemplateReq{
			BKBizID:              bizID,
			BKServiceTemplateIDs: serviceTemplateIDs,
			BKModuleIDs:          modleIDs,
			Fields:               ccHostFields(),
			Page: Page{
				Start: p.Offset,
				Sort:  p.Sort,
				Limit: p.Limit,
			},
		}

		resp, err := h.cli.findHostByServiceTemplate(nCtx, req)
		if err != nil {
			return nil, err
		}

		host := make([]*types.Host, len(resp.Info))
		for idx, info := range resp.Info {
			host[idx] = h.convHostInfoToTypes(nCtx.TenantID(), info, bizID)
		}

		return host, nil
	}

	pExecutor := pageexecutor.NewPageExecutor[*types.Host](CCPageSizeLimit, ccQueryTimeout)
	result, err := pExecutor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// FindHostBySetTemplate find host by set template.
func (h *Handler) FindHostBySetTemplate(nCtx contextx.IContext, bizID int64, page types.Page, setTemplateIDs []int64, setIDs []int64) (
	[]*types.Host, error) {

	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		req := &FindHostBySetTemplateReq{
			BKBizID:          bizID,
			BKSetTemplateIDs: setTemplateIDs,
			BKSetIDs:         setIDs,
			Fields:           ccHostFields(),
			Page: Page{
				Start: p.Offset,
				Sort:  p.Sort,
				Limit: p.Limit,
			},
		}

		resp, err := h.cli.findHostBySetTemplate(nCtx, req)
		if err != nil {
			return nil, err
		}

		host := make([]*types.Host, len(resp.Info))
		for idx, info := range resp.Info {
			host[idx] = h.convHostInfoToTypes(nCtx.TenantID(), info, bizID)
		}

		return host, nil
	}

	pExecutor := pageexecutor.NewPageExecutor[*types.Host](CCPageSizeLimit, ccQueryTimeout)
	result, err := pExecutor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// FindHostByDynamicGroup find host by dynamic group.
func (h *Handler) FindHostByDynamicGroup(nCtx contextx.IContext, bizID int64, dynamicGroupIDs []string, page types.Page) ([]*types.Host, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to get host by dynamic group, nCtx is nil")
	}

	if bizID == CCInvalidID {
		return nil, fmt.Errorf("failed to get host by dynamic group, bizID is invalid")
	}

	pExecutor := pageexecutor.NewPageExecutor[*types.Host](CCPageSizeLimit, ccQueryTimeout)
	hosts := make([]*types.Host, 0)
	for idx := range dynamicGroupIDs {
		dynamicGroupID := dynamicGroupIDs[idx]
		fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
			return h.executeHostDynamicGroup(nCtx, bizID, dynamicGroupID, p)
		}

		result, err := pExecutor.Execute(nCtx, page, fn)
		if err != nil {
			return nil, err
		}

		hosts = append(hosts, result.Items...)
	}

	return hosts, nil
}

// executeHostDynamicGroup execute dynamic grouping rules to return hosts within the group.
func (h *Handler) executeHostDynamicGroup(nCtx contextx.IContext, bizID int64, groupID string, page types.Page) (
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

	resp, err := h.cli.executeDynamicGroup(nCtx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Host, len(resp.Info))
	for index := range resp.Info {
		hostData := HostInfo{}
		if err := conv.MapToStruct(resp.Info[index], &hostData); err != nil {
			return nil, fmt.Errorf("failed to convert host info: %w", err)
		}

		result[index] = h.convHostInfoToTypes(nCtx.TenantID(), &hostData, bizID)
	}

	return result, nil
}

// FindSetByDynamicGroup find set by dynamic group.
func (h *Handler) FindSetByDynamicGroup(nCtx contextx.IContext, bizID int64, dynamicGroupIDs []string, page types.Page) ([]*SetInfo, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to get set by dynamic group, nCtx is nil")
	}

	if bizID == CCInvalidID {
		return nil, fmt.Errorf("failed to get set by dynamic group, bizID is invalid")
	}

	pExecutor := pageexecutor.NewPageExecutor[*SetInfo](CCPageSizeLimit, ccQueryTimeout)
	sets := make([]*SetInfo, 0)
	for idx := range dynamicGroupIDs {
		dynamicGroupID := dynamicGroupIDs[idx]
		fn := func(nCtx contextx.IContext, p types.Page) ([]*SetInfo, error) {
			return h.executeSetDynamicGroup(nCtx, bizID, dynamicGroupID, p)
		}

		result, err := pExecutor.Execute(nCtx, page, fn)
		if err != nil {
			return nil, err
		}

		sets = append(sets, result.Items...)
	}

	return sets, nil
}

// executeSetDynamicGroup execute dynamic grouping rules to return sets within the group.
func (h *Handler) executeSetDynamicGroup(nCtx contextx.IContext, bizID int64, groupID string, page types.Page) (
	[]*SetInfo, error) {

	// Fields needed: bk_set_id, set_template_id (following Python: fields=["bk_set_id", "set_template_id"])
	fields := []ccField{
		ccFieldBKSetID,
		ccFieldSetTemplateID,
		ccFieldBKSetName,
		ccFieldBKBizID,
	}

	req := &ExecuteDynamicGroupReq{
		BKBizID:        bizID,
		ID:             groupID,
		Fields:         fields,
		DisableCounter: false,
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.executeDynamicGroup(nCtx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*SetInfo, 0, len(resp.Info))
	for index := range resp.Info {
		setData := SetInfo{}
		if err := conv.MapToStruct(resp.Info[index], &setData); err != nil {
			return nil, fmt.Errorf("failed to convert set info: %w", err)
		}

		result = append(result, &setData)
	}

	return result, nil
}

const (
	// TopoNodeObjIDBiz topo node object id for biz.
	TopoNodeObjIDBiz = "biz"
	// TopoNodeObjIDHost topo node object id for host.
	TopoNodeObjIDHost = "host"
	// TopoNodeObjIDModule topo node object id for module.
	TopoNodeObjIDModule = "module"
)

// FindHostByTopo find host by topo.
// nolint: gocognit
func (h *Handler) FindHostByTopo(nCtx contextx.IContext, bizID int64, topoNodes ...*types.ScopeTopoNode) ([]*types.Host, error) {
	// this has a special logic, so we need to split it for three parts:
	// 1. find obj id is biz
	// 2. find obj id is host
	// 3. find obj id is other
	bizTopoNodes := make([]*types.ScopeTopoNode, 0)
	hostTopoNodes := make([]*types.ScopeTopoNode, 0)
	otherTopoNodes := make([]*types.ScopeTopoNode, 0)
	for idx := range topoNodes {
		switch topoNodes[idx].TopoObjID {
		case TopoNodeObjIDBiz:
			bizTopoNodes = append(bizTopoNodes, topoNodes[idx])
		case TopoNodeObjIDHost:
			hostTopoNodes = append(hostTopoNodes, topoNodes[idx])
		default:
			otherTopoNodes = append(otherTopoNodes, topoNodes[idx])
		}
	}

	finalhost := make([]*types.Host, 0)
	pExecutor := pageexecutor.NewPageExecutor[*types.Host](CCPageSizeLimit, ccQueryTimeout)

	if len(bizTopoNodes) > 0 {
		for idx := range bizTopoNodes {
			instID := topoNodes[idx].TopoInstID
			if bizID != instID {
				return nil, fmt.Errorf("faied to find host by topo node biz: biz-id(%d), inst-id(%d): invalid topo node",
					instID, topoNodes[idx].TopoInstID)
			}

			fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
				return h.findHostByTopoNodeBiz(nCtx, instID, p)
			}

			result, err := pExecutor.Execute(nCtx, types.UnlimitedPage(), fn)
			if err != nil {
				return nil, fmt.Errorf("failed to find host by topo node biz: %w", err)
			}

			finalhost = append(finalhost, result.Items...)
		}
	}

	if len(hostTopoNodes) > 0 {
		hostIDs := conv.SliceToSlice[*types.ScopeTopoNode, int64](hostTopoNodes, func(node *types.ScopeTopoNode) int64 {
			return node.TopoInstID
		})
		cond := &types.HostStaticExactCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				HostID: hostIDs,
			},
		}

		hosts, err := h.FindHostWithCondition(nCtx, types.UnlimitedPage(), cond)
		if err != nil {
			return nil, fmt.Errorf("failed to find host by topo node host: %w", err)
		}

		finalhost = append(finalhost, hosts...)
	}

	if len(otherTopoNodes) > 0 {
		for idx := range otherTopoNodes {
			objID := otherTopoNodes[idx].TopoObjID
			instID := otherTopoNodes[idx].TopoInstID
			fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
				return h.findHostByTopoNodeBelowBiz(nCtx, p, bizID, objID, instID)
			}

			result, err := pExecutor.Execute(nCtx, types.UnlimitedPage(), fn)
			if err != nil {
				return nil, fmt.Errorf("failed to find host by topo node other: %w", err)
			}

			finalhost = append(finalhost, result.Items...)
		}
	}

	return finalhost, nil
}

// findHostByTopoNodeBelowBiz find host by biz topo node.
// notice: this topo node is below biz and above host.
func (h *Handler) findHostByTopoNodeBelowBiz(nCtx contextx.IContext, p types.Page, bizID int64, objID string, instID int64) ([]*types.Host, error) {
	req := &FindHostByTopoReq{
		BKBizID:  bizID,
		BKObjID:  objID,
		BKInstID: instID,
		Fields:   ccHostFields(),
		Page: Page{
			Start: p.Offset,
			Limit: p.Limit,
			Sort:  p.Sort,
		},
	}

	resp, err := h.cli.findHostByTopo(nCtx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to find host by topo node other: %w", err)
	}

	hosts := make([]*types.Host, len(resp.Info))
	for idx := range resp.Info {
		hosts[idx] = h.convHostInfoToTypes(nCtx.TenantID(), resp.Info[idx], bizID)
	}

	return hosts, nil
}

func (h *Handler) findHostByTopoNodeBiz(nCtx contextx.IContext, bizID int64, p types.Page) ([]*types.Host, error) {
	req := &ListBizHostsReq{
		BKBizID: bizID,
		Fields:  ccHostFields(),
		Page: Page{
			Start: p.Offset,
			Limit: p.Limit,
			Sort:  p.Sort,
		},
	}

	resp, err := h.cli.listBizHosts(nCtx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]*types.Host, len(resp.Info))
	for idx := range resp.Info {
		hosts[idx] = h.convHostInfoToTypes(nCtx.TenantID(), resp.Info[idx], bizID)
	}

	return hosts, nil
}
