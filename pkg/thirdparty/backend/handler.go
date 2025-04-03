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

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Handler is interface for nodeman backend handler.
type Handler interface {
	IHandlerNetworkArea
	IHandlerNetworkUnit

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
	// @param request the host distinct request.
	// @param condition the filter conditions.
	// @return the host distinct result.
	DistinctHost(
		ctx context.Context, request types.HostDistinctRequest, condition *types.HostCondition) (
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
	// @param request the topo-event distinct request.
	// @param condition the filter conditions.
	// @return the topo-event distinct result.
	DistinctTopoEvent(
		ctx context.Context, request types.TopoEventDistinctRequest, condition *types.TopoEventCondition) (
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
	ctx context.Context, request types.HostDistinctRequest, condition *types.HostCondition) (
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
	ctx context.Context, request types.TopoEventDistinctRequest, condition *types.TopoEventCondition) (
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
