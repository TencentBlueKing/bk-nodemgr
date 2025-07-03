/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// Handler is interface for nodeman backend handler.
// nolint: interfacebloat
type Handler interface {
	IHandlerNetworkArea
	IHandlerNetworkUnit
	IHandlerNodeAgent
	IHandlerNodeWorkflow
	IHandlerRelease

	// ListBusiness list business within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return business list with page and the total count with filter.
	ListBusiness(ctx context.Context, page types.Page, condition *types.BusinessCondition) (
		[]*types.Business, int64, error)

	// ListHost list host within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return host list with page and the total count with filter.
	ListHost(ctx context.Context, page types.Page, condition *types.HostCondition) ([]*types.Host, int64, error)

	// DistinctHost distinct host by condition.
	// @param ctx context, contains tenant-id.
	// @param condition the filter conditions.
	// @return the host distinct result.
	DistinctHost(
		ctx context.Context, condition *types.HostCondition) (
		*types.HostDistinctResult, error)

	// CountHost count host within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param condition the filter conditions.
	// @return the host count with filter.
	CountHost(ctx context.Context, condition *types.HostCondition) (int64, error)

	// ListTopoEvent list topo events by page and conditions.
	// @param ctx context, contains tenant-id.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the topo-event list with page and the total count with filter.
	ListTopoEvent(ctx context.Context, page types.Page, condition *types.TopoEventCondition) (
		[]*types.TopoEvent, int64, error)

	// CountTopoEvent count topo events by condition.
	// @param ctx context, contains tenant-id.
	// @param condition the filter conditions.
	// @return the topo-event count with filter.
	CountTopoEvent(ctx context.Context, condition *types.TopoEventCondition) (int64, error)

	// DistinctTopoEvent distinct topo-event by condition.
	// @param ctx context, contains tenant-id.
	// @param condition the filter conditions.
	// @return the topo-event distinct result.
	DistinctTopoEvent(
		ctx context.Context, condition *types.TopoEventCondition) (
		*types.TopoEventDistinctResult, error)

	// ListAccessPoint list access points by page and conditions.
	// @param ctx context, contains tenant-id.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the access-point list with page and the total count with filter.
	ListAccessPoint(ctx context.Context, page types.Page, condition *types.AccessPointCondition) (
		[]*types.AccessPoint, int64, error)

	// GetConstant get constant by fields.
	// @param ctx context, contains tenant-id.
	// @param fields describes the fields to get.
	// @return the constant result.
	GetConstant(ctx context.Context, fields types.TopoConstantFields) (*types.TopoConstant, error)
}

// IHandlerNetworkArea defines the network area handler.
type IHandlerNetworkArea interface {
	// CreateNetworkArea create network area within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param networkArea the network area to create.
	// @return the network-area id.
	CreateNetworkArea(ctx context.Context, networkArea *types.NetworkArea) (int64, error)

	// UpdateNetworkArea update network area within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param networkArea the network area to update.
	// @return the error.
	UpdateNetworkArea(ctx context.Context, networkArea *types.NetworkArea) error

	// ListNetworkArea list network area within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the network-area list with page and the total count with filter.
	ListNetworkArea(ctx context.Context, page types.Page, condition *types.NetworkAreaCondition) (
		[]*types.NetworkArea, int64, error)

	// DeleteNetworkArea delete network area within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param networkAreaID the network area id.
	// @return the error.
	DeleteNetworkArea(ctx context.Context, networkAreaID int64) error

	// GetNetworkArea get specific network area.
	// @param ctx context, contains tenant-id.
	// @param networkAreaID the network area id.
	// @return the network-area.
	GetNetworkArea(ctx context.Context, networkAreaID int64) (*types.NetworkArea, error)
}

// IHandlerNetworkUnit defines the network unit handler.
type IHandlerNetworkUnit interface {
	// CreateNetworkUnit create network unit within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param networkUnit the network unit to create.
	// @param accessPoints the access points of the network unit.
	// @return the network-unit id.
	CreateNetworkUnit(ctx context.Context, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (
		int64, error)

	// UpdateNetworkUnit update network unit within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param networkUnit the network unit to update.
	// @param accessPoints the access points of the network unit.
	// @return the network-unit id.
	UpdateNetworkUnit(ctx context.Context, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) error

	// ListNetworkUnit list network unit within specified tenant in context.
	// @param ctx content, contains tenant-id.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the network-unit list with page and the total count with filter.
	ListNetworkUnit(ctx context.Context, page types.Page, condition *types.NetworkUnitCondition) (
		[]*types.NetworkUnit, int64, error)

	// GetNetworkUnit get specific network unit.
	// @param ctx context, contains tenant-id.
	// @param networkUnitID the network unit id.
	// @return the network-unit and its accesspoints.
	GetNetworkUnit(ctx context.Context, networkUnitID int64) (*types.NetworkUnit, map[int64]*types.AccessPoint, error)

	// DeleteNetworkUnit delete network unit within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param networkUnitID the network unit id.
	// @return the network-unit id.
	DeleteNetworkUnit(ctx context.Context, networkUnitID int64) error
}

// IHandlerNodeAgent defines the node agent handler.
type IHandlerNodeAgent interface {
	// InstallAgent node agent.
	// @param ctx context, contains tenant-id.
	// @param hosts the install param.
	// @return the installing workflow-ids and error.
	InstallAgent(ctx context.Context, hostsParam []*types.NodeAgentInstallParam) (string, error)

	// UpgradeAgent node agent.
	// @param ctx context, contains tenant-id.
	// @param hosts the upgrade param.
	// @return the upgrading workflow-ids and error.
	OperationRetry(ctx context.Context, retryParam *types.NodeOperationRetryParam) ([]string, error)
}

// IHandlerNodeWorkflow defines the node workflow handler.
type IHandlerNodeWorkflow interface {
	// ListNodeWorkflow list node workflow within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the node-workflow list with page and the total count with filter.
	ListNodeWorkflow(ctx context.Context, page types.Page, condition *types.NodeWorkflowCondition) (
		[]*types.NodeWorkflow, int64, error)

	// CountNodeWorkflow count node workflow by conditions.
	//	@param ctx context, contains tenant-id.
	//	@param condition the filter conditions.
	//	@return the node-workflow count with filter.
	CountNodeWorkflow(ctx context.Context, condition *types.NodeWorkflowCondition) (int64, error)

	// DistinctNodeWorkflow distinct node workflow by conditions.
	// @param ctx context, contains tenant-id.
	// @param request the node workflow distinct request.
	// @param conditions the filter conditions.
	// @return the node-workflow distinct result.
	DistinctNodeWorkflow(ctx context.Context, request types.NodeWorkflowDistinctRequest,
		condition *types.NodeWorkflowCondition) (*types.NodeWorkflowDistinctResult, error)

	// ListOperation list network area operation.
	// @param ctx context, contains tenant-id.
	// @param page describes the page info when listing.
	// @param workflowID the workflow id.
	// @return the operation list with page and the total count with filter.
	ListNodeWorkflowOperation(ctx context.Context, page types.Page, condition *types.NodeWorkflowOperationCondition) (
		[]*operation.Operation, int64, error)

	// CountOperation count network area operation.
	// @param ctx context, contains tenant-id.
	// @param workflowID the workflow id.
	// @return the operation count with filter.
	CountNodeWorkflowOperation(ctx context.Context, condition *types.NodeWorkflowOperationCondition) (int64, error)

	// ListOperationInstance list network area operation instance.
	// @param ctx context, contains tenant-id.
	// @param operationID the operation id.
	// @return the operation instance list with page and the total count with filter.
	ListNodeWorkflowOperationInstance(ctx context.Context, operationID ...string) (
		[]*operation.InstanceBriefData, int64, error)

	// CountOperationInstance count network area operation instance.
	// @param ctx context, contains tenant-id.
	// @param operationID the operation id.
	// @return the operation instance count with filter.
	CountNodeWorkflowOperationInstance(ctx context.Context, operationID ...string) (int64, error)

	// DistinctNodeWorkflow distinct node workflow by conditions.
	// @param ctx context, contains tenant-id.
	// @param request the node workflow distinct request.
	// @param condition the filter conditions.
	// @return the node-workflow distinct result.
	GetNodeWorkflowOperationInstanceLog(ctx context.Context, instanceID string) (*operation.InstanceData, error)

	// ListOperationInstanceStatus list network area operation instance status.
	// @param ctx context, contains tenant-id.
	// @param triggerID the trigger id.
	// @return the operation instance status list.
	ListNodeWorkflowOperationInstanceStatus(ctx context.Context,
		condition *types.NodeWorkflowOperInstanceStatusCondition) ([]*operation.InstanceStatus, error)
}

// IHandlerRelease defines the backend handler for release.
type IHandlerRelease interface {
	// ListRelease lists release by page and conditions.
	ListRelease(ctx context.Context, page types.Page, condition *types.ReleaseCondition) (
		[]*types.Release, int64, error)

	// CountRelease counts release by conditions.
	CountRelease(ctx context.Context, condition *types.ReleaseCondition) (int64, error)

	// SetReleaseLabels sets release labels.
	SetReleaseLabels(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string,
		labels []string) error

	// EnableRelease enables release active by generation, release type, platform and version.
	EnableRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error

	// DisableRelease disables release disactive by generation, release type, platform and version.
	DisableRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error

	// SetAsDefaultRelease sets the release as default.
	SetAsDefaultRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error

	// CancelAsDefaultRelease cancels the release as default.
	CancelAsDefaultRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error

	// 	DeleteRelease deletes release by generation, release type, platform and version.
	DeleteRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error
}

type handler struct {
	cli *cli
}

// New initialize a new nodeman backend handler.
func New(c *client.Capability, conf *Config) (Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &handler{cli: cli}, nil
}

// ListBusiness list business within specified tenant in context.
func (h *handler) ListBusiness(ctx context.Context, page types.Page, condition *types.BusinessCondition) (
	[]*types.Business, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	req := &protoBackend.TopoBusinessListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listBusiness(ctx, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	items := resp.GetItems()
	data := make([]*types.Business, len(items))
	for idx, item := range items {
		data[idx] = &types.Business{
			TenantID: item.GetTenantId(),
			BizID:    item.GetBkBizId(),
			BizName:  item.GetBkBizName(),
		}
	}

	return data, resp.GetTotal(), nil
}

// nolint funlen
// ListHost list host within specified tenant in context.
func (h *handler) ListHost(ctx context.Context, page types.Page, condition *types.HostCondition) (
	[]*types.Host, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	req := &protoBackend.TopoHostListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listHost(ctx, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	total, hosts := resp.ConvertHostsToTypes()

	return hosts, total, nil
}

// DistinctHost distinct host within specified tenant in context.
func (h *handler) DistinctHost(
	ctx context.Context, condition *types.HostCondition) (
	*types.HostDistinctResult, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &protoBackend.TopoHostDistinctReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctHost(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// CountHost count host within specified tenant in context.
func (h *handler) CountHost(ctx context.Context, condition *types.HostCondition) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	req := &protoBackend.TopoHostListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listHost(ctx, tenantID, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// CreateNetworkArea creates a new networkarea.
func (h *handler) CreateNetworkArea(ctx context.Context, networkArea *types.NetworkArea) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	req := &protoBackend.TopoNetworkAreaCreateReq{
		BkNetworkareaName: networkArea.Name,
		CloudVendor:       networkArea.CloudVendor,
	}

	resp, err := h.cli.createNetworkArea(ctx, tenantID, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetBkNetworkareaId(), nil
}

// UpdateNetworkArea updates an existing networkarea.
func (h *handler) UpdateNetworkArea(ctx context.Context, networkArea *types.NetworkArea) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &protoBackend.TopoNetworkAreaUpdateReq{
		BkNetworkareaId:   networkArea.ID,
		BkNetworkareaName: networkArea.Name,
		CloudVendor:       networkArea.CloudVendor,
	}

	_, err = h.cli.updateNetworkArea(ctx, tenantID, req)
	if err != nil {
		return err
	}

	return nil
}

// ListNetworkArea list network area within specified tenant in context.
func (h *handler) ListNetworkArea(ctx context.Context, page types.Page, condition *types.NetworkAreaCondition) (
	[]*types.NetworkArea, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	req := &protoBackend.TopoNetworkAreaListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listNetworkArea(ctx, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	items := resp.GetItems()
	data := make([]*types.NetworkArea, len(items))
	for idx, item := range items {
		data[idx] = &types.NetworkArea{
			TenantID:    item.GetTenantId(),
			ID:          item.GetBkNetworkareaId(),
			Name:        item.GetBkNetworkareaName(),
			CloudVendor: item.GetCloudVendor(),
		}
	}

	return data, resp.GetTotal(), nil
}

// GetNetworkArea gets an existing networkarea.
func (h *handler) GetNetworkArea(ctx context.Context, networkAreaID int64) (*types.NetworkArea, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &protoBackend.TopoNetworkAreaGetReq{
		BkNetworkareaId: networkAreaID,
	}

	resp, err := h.cli.getNetworkArea(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}

	return &types.NetworkArea{
		TenantID:    resp.GetData().GetTenantId(),
		ID:          resp.GetData().GetBkNetworkareaId(),
		Name:        resp.GetData().GetBkNetworkareaName(),
		CloudVendor: resp.GetData().GetCloudVendor(),
	}, nil
}

// DeleteNetworkArea deletes an existing networkarea.
func (h *handler) DeleteNetworkArea(ctx context.Context, networkAreaID int64) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &protoBackend.TopoNetworkAreaDeleteReq{
		BkNetworkareaId: networkAreaID,
	}

	_, err = h.cli.deleteNetworkArea(ctx, tenantID, req)
	if err != nil {
		return err
	}

	return nil
}

// CreateNetworkUnit creates a new networkunit.
func (h *handler) CreateNetworkUnit(
	ctx context.Context, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return -1, err
	}

	req := &protoBackend.TopoNetworkUnitCreateReq{
		BkNetworkunitName: networkUnit.Name,
		BkNetworkareaId:   networkUnit.NetworkAreaID,
	}
	req.ConvertAccesspointsFromTypes(accessPoints)
	req.ConvertLinksFromTypes(networkUnit.Links)

	resp, err := h.cli.createNetworkUnit(ctx, tenantID, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetBkNetworkunitId(), nil
}

// UpdateNetworkUnit updates an existing networkunit.
func (h *handler) UpdateNetworkUnit(
	ctx context.Context, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) error {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &protoBackend.TopoNetworkUnitUpdateReq{
		BkNetworkunitId:   networkUnit.ID,
		BkNetworkunitName: networkUnit.Name,
		BkNetworkareaId:   networkUnit.NetworkAreaID,
	}
	req.ConvertAccesspointsFromTypes(accessPoints)
	req.ConvertLinksFromTypes(networkUnit.Links)

	_, err = h.cli.updateNetworkUnit(ctx, tenantID, req)
	if err != nil {
		return err
	}

	return nil
}

// GetNetworkUnit gets an existing networkunit.
func (h *handler) GetNetworkUnit(ctx context.Context, networkUnitID int64) (
	*types.NetworkUnit, map[int64]*types.AccessPoint, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, nil, err
	}

	req := &protoBackend.TopoNetworkUnitGetReq{
		BkNetworkunitId: networkUnitID,
	}

	resp, err := h.cli.getNetworkUnit(ctx, tenantID, req)
	if err != nil {
		return nil, nil, err
	}

	networkUnit, accessPoints := resp.ConvertNetworkUnitToTypes()

	return networkUnit, accessPoints, nil
}

// ListNetworkUnit list network unit within specified tenant in context.
func (h *handler) ListNetworkUnit(ctx context.Context, page types.Page, condition *types.NetworkUnitCondition) (
	[]*types.NetworkUnit, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	req := &protoBackend.TopoNetworkUnitListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listNetworkUnit(ctx, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	total, networkUnits := resp.ConvertNetworkUnitsToTypes()

	return networkUnits, total, nil
}

// DeleteNetworkUnit deletes network unit within specified tenant in context.
func (h *handler) DeleteNetworkUnit(ctx context.Context, networkUnitID int64) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	req := &protoBackend.TopoNetworkUnitDeleteReq{
		BkNetworkunitId: networkUnitID,
	}

	_, err = h.cli.deleteNetworkUnit(ctx, tenantID, req)
	if err != nil {
		return err
	}

	return nil
}

// ListTopoEvent list topo event within specified tenant in context.
func (h *handler) ListTopoEvent(ctx context.Context, page types.Page, condition *types.TopoEventCondition) (
	[]*types.TopoEvent, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	req := &protoBackend.TopoEventListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listTopoEvent(ctx, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	total, events := resp.ConvertTopoEventsToTypes()

	return events, total, nil
}

// CountTopoEvent count the number of topo events by conditions.
func (h *handler) CountTopoEvent(ctx context.Context, condition *types.TopoEventCondition) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	req := &protoBackend.TopoEventListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listTopoEvent(ctx, tenantID, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctTopoEvent distinct the number of topo events by conditions.
func (h *handler) DistinctTopoEvent(
	ctx context.Context, condition *types.TopoEventCondition) (
	*types.TopoEventDistinctResult, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &protoBackend.TopoEventDistinctReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctTopoEvent(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// ListAccessPoint list access point within specified tenant in context.
func (h *handler) ListAccessPoint(ctx context.Context, page types.Page, condition *types.AccessPointCondition) (
	[]*types.AccessPoint, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	req := &protoBackend.TopoAccessPointListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listAccessPoint(ctx, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	total, events := resp.ConvertAccessPointsToTypes()

	return events, total, nil
}

// GetConstant get constant by fields.
func (h *handler) GetConstant(ctx context.Context, fields types.TopoConstantFields) (*types.TopoConstant, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &protoBackend.TopoConstantGetReq{}
	if err := req.ConvertFieldsFromTypes(fields); err != nil {
		return nil, err
	}

	resp, err := h.cli.getConstant(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertConstantToTypes(), nil
}

// ListNodeWorkflow list node workflow within specified tenant in context.
func (h *handler) ListNodeWorkflow(ctx context.Context, page types.Page, condition *types.NodeWorkflowCondition) (
	[]*types.NodeWorkflow, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	req := &protoBackend.NodeWorkflowListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listNodeWorkflow(ctx, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	result, num := resp.ConvertNodeWorkflowsToTypes()

	return result, num, nil
}

// CountHost count host within specified tenant in context.
func (h *handler) CountNodeWorkflow(ctx context.Context, condition *types.NodeWorkflowCondition) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	req := &protoBackend.NodeWorkflowListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listNodeWorkflow(ctx, tenantID, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctNodeWorkflow distinct node workflow by conditions.
func (h *handler) DistinctNodeWorkflow(ctx context.Context, _ types.NodeWorkflowDistinctRequest,
	conditions *types.NodeWorkflowCondition) (*types.NodeWorkflowDistinctResult, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &protoBackend.NodeWorkflowDistinctReq{}
	if err := req.ConvertConditionsFromTypes(conditions); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctNodeWorkflow(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertWorkflowDistinctToTypes(), nil
}

// ListOperation list workflow  operation.
func (h *handler) ListNodeWorkflowOperation(ctx context.Context,
	page types.Page, condition *types.NodeWorkflowOperationCondition) (
	[]*operation.Operation, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	req := &protoBackend.NodeWorkflowOperationListReq{
		Page:      convertPage(page),
		OnlyCount: false,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listNodeWorkflowOperation(ctx, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	operations, total := resp.ConvertWorkflowOperationToTypes()

	return operations, total, nil
}

// CountOperation count workflow  operation.
func (h *handler) CountNodeWorkflowOperation(ctx context.Context,
	condition *types.NodeWorkflowOperationCondition) (int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	req := &protoBackend.NodeWorkflowOperationListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}
	resp, err := h.cli.listNodeWorkflowOperation(ctx, tenantID, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotalCount(), nil
}

// ListNodeWorkflowOperationInstance list workflow operation instance.
func (h *handler) ListNodeWorkflowOperationInstance(
	ctx context.Context, operationID ...string) ([]*operation.InstanceBriefData, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	req := &protoBackend.NodeWorkflowOperationInstanceListReq{
		OnlyCount:   false,
		OperationId: []string(operationID),
	}

	resp, err := h.cli.listNodeWorkflowOperationInstance(ctx, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	operations, total := resp.ConvertWorkflowOperationInstanceToTypes()

	return operations, total, nil
}

// CountOperationInstance count workflow operation instance.
func (h *handler) CountNodeWorkflowOperationInstance(ctx context.Context, operationID ...string) (int64, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return 0, err
	}

	req := &protoBackend.NodeWorkflowOperationInstanceListReq{
		OnlyCount:   true,
		OperationId: []string(operationID),
	}

	resp, err := h.cli.listNodeWorkflowOperationInstance(ctx, tenantID, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// GetOperationInstanceLog get workflow operation instance log.
func (h *handler) GetNodeWorkflowOperationInstanceLog(ctx context.Context, instanceID string) (
	*operation.InstanceData, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &protoBackend.NodeWorkflowOperationInstanceLogGetReq{
		OperInstId: instanceID,
	}

	resp, err := h.cli.getOperationInstanceLog(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}

	operations := resp.ConvertWorkflowOperationInstanceLogToTypes()

	return operations, nil
}

// ListNodeWorkflowOpInstanceStatus list workflow operation instance status.
func (h *handler) ListNodeWorkflowOperationInstanceStatus(ctx context.Context,
	conditions *types.NodeWorkflowOperInstanceStatusCondition) ([]*operation.InstanceStatus, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &protoBackend.NodeWorkflowOperationInstanceListStatusReq{
		Page: &protoBackend.Page{},
	}

	if err := req.ConvertConditionsFromTypes(conditions); err != nil {
		return nil, err
	}

	resp, err := h.cli.listNodeWorkflowOpInstanceStatus(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}

	instanceStatus := resp.ConvertWorkflowOperationInstanceStatusToTypes()

	return instanceStatus, nil
}

func (h *handler) OperationRetry(ctx context.Context, retryParam *types.NodeOperationRetryParam) ([]string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &protoBackend.NodeWorkflowOperationRetryReq{}

	req.ConvertOperationRetryParamFromTypes(*retryParam)

	resp, err := h.cli.retryOperation(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}

	result := resp.ConvertResultToComm()

	return result, nil
}

// ListRelease lists release by page and conditions.
func (h *handler) ListRelease(ctx context.Context, page types.Page, condition *types.ReleaseCondition) (
	[]*types.Release, int64, error) {

	req := &protoBackend.PackageReleaseListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listRelease(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	total, releases := resp.ConvertReleasesToTypes()

	return releases, total, nil
}

// CountRelease counts release by conditions.
func (h *handler) CountRelease(ctx context.Context, condition *types.ReleaseCondition) (int64, error) {
	req := &protoBackend.PackageReleaseListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listRelease(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// SetReleaseLabels sets release labels.
func (h *handler) SetReleaseLabels(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string,
	labels []string) error {

	req := &protoBackend.PackageReleaseSetLabelsReq{Labels: labels}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.setReleaseLabels(ctx, req)
}

// EnableRelease enables release active by generation, release type, platform and version.
func (h *handler) EnableRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	req := &protoBackend.PackageReleaseEnableReq{}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.enableRelease(ctx, req)
}

// DisableRelease disables release disactive by generation, release type, platform and version.
func (h *handler) DisableRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	req := &protoBackend.PackageReleaseDisableReq{}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.disableRelease(ctx, req)
}

// SetAsDefaultRelease sets the release as default.
func (h *handler) SetAsDefaultRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	req := &protoBackend.PackageReleaseSetAsDefaultReq{}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.setAsDefaultRelease(ctx, req)
}

// CancelAsDefaultRelease cancels the release as default.
func (h *handler) CancelAsDefaultRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	req := &protoBackend.PackageReleaseCancelAsDefaultReq{}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.cancelAsDefaultRelease(ctx, req)
}

// DeleteRelease deletes release by generation, release type, platform and version.
func (h *handler) DeleteRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	req := &protoBackend.PackageReleaseDeleteReq{}
	req.SetIdentifer(gen, releaseType, plat, version)

	return h.cli.deleteRelease(ctx, req)
}
