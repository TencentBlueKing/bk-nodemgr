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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerTopo defines the topo Handler.
type IHandlerTopo interface {
	IHandlerAccessPoint
	IHandlerBusiness
	IHandlerConstant
	IHandlerGraph
	IHandlerNetworkArea
	IHandlerNetworkUnit
}

// IHandlerAccessPoint defines the access point Handler.
type IHandlerAccessPoint interface {
	// ListAccessPoint list access points by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the access-point list with page and the total count with filter and error.
	ListAccessPoint(nCtx contextx.IContext, page types.Page, condition *types.AccessPointCondition) ([]*types.AccessPoint, int64, error)
}

// IHandlerBusiness defines the business Handler.
type IHandlerBusiness interface {
	// ListBusiness list business within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return business list with page and the total count with filter and error.
	ListBusiness(nCtx contextx.IContext, page types.Page, condition *types.BusinessCondition) ([]*types.Business, int64, error)
}

// IHandlerConstant defines the constant Handler.
type IHandlerConstant interface {
	// GetConstant get constant by fields.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param fields describes the fields to get.
	// @return the constant result and error.
	GetConstant(nCtx contextx.IContext, fields types.TopoConstantFields) (*types.TopoConstant, error)
}

// IHandlerGraph defines the graph Handler.
type IHandlerGraph interface {
	// GetGraphNode get graph node.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param networkUnitIDs the network unit ids.
	// @return the graph node list and error.
	GetGraphNode(nCtx contextx.IContext, networkUnitIDs []int64) ([]*types.GraphNodeInfo, error)
}

// IHandlerNetworkArea defines the network area Handler.
type IHandlerNetworkArea interface {
	// CreateNetworkArea create network area within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param networkArea the network area to create.
	// @return the network-area id and error.
	CreateNetworkArea(nCtx contextx.IContext, networkArea *types.NetworkArea) (int64, error)

	// UpdateNetworkArea update network area within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param networkArea the network area to update.
	// @return the error.
	UpdateNetworkArea(nCtx contextx.IContext, networkArea *types.NetworkArea) error

	// ListNetworkArea list network area within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the network-area list with page and the total count with filter and error.
	ListNetworkArea(nCtx contextx.IContext, page types.Page, condition *types.NetworkAreaCondition) ([]*types.NetworkArea, int64, error)

	// DeleteNetworkArea delete network area within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param networkAreaID the network area id.
	// @return the error.
	DeleteNetworkArea(nCtx contextx.IContext, networkAreaID int64) error

	// GetNetworkArea get specific network area.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param networkAreaID the network area id.
	// @return the network-area and error.
	GetNetworkArea(nCtx contextx.IContext, networkAreaID int64) (*types.NetworkArea, error)
}

// IHandlerNetworkUnit defines the network unit Handler.
type IHandlerNetworkUnit interface {
	// CreateNetworkUnit create network unit within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param networkUnit the network unit to create.
	// @param accessPoints the access points of the network unit.
	// @return the network-unit id and error.
	CreateNetworkUnit(nCtx contextx.IContext, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (int64, error)

	// UpdateNetworkUnit update network unit within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param networkUnit the network unit to update.
	// @param accessPoints the access points of the network unit.
	// @return the error.
	UpdateNetworkUnit(
		nCtx contextx.IContext, fields types.NetworkUnitUpdateFields, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) error

	// ListNetworkUnit list network unit within specified tenant in contextx.
	// @param nCtx content, contains tenant-id.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the network-unit list with page and the total count with filter and error.
	ListNetworkUnit(nCtx contextx.IContext, page types.Page, condition *types.NetworkUnitCondition) ([]*types.NetworkUnit, int64, error)

	// ListNetworkUnitBrief list brief network unit information within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the brief network-unit list with page and the total count with filter and error.
	ListNetworkUnitBrief(nCtx contextx.IContext, page types.Page, condition *types.NetworkUnitCondition) ([]*types.NetworkUnit, int64, error)

	// GetNetworkUnit get specific network unit.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param networkUnitID the network unit id.
	// @return the network-unit and its accesspoints and error.
	GetNetworkUnit(nCtx contextx.IContext, networkUnitID int64) (*types.NetworkUnit, map[int64]*types.AccessPoint, error)

	// DeleteNetworkUnit delete network unit within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param networkUnitID the network unit id.
	// @return the error.
	DeleteNetworkUnit(nCtx contextx.IContext, networkUnitID int64) error

	// GetNetworkUnitDistributionByNetworkAreaID get networkunit distribution by network area id.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the networkunit distribution map and error.
	GetNetworkUnitDistributionByNetworkAreaID(nCtx contextx.IContext, condition *types.NetworkUnitCondition) (map[int64]int64, error)
}

// ===============================================================================
// Access Point Related Interfaces
// ===============================================================================

// ListAccessPoint list access point within specified tenant in contextx.
func (h *Handler) ListAccessPoint(nCtx contextx.IContext, page types.Page, condition *types.AccessPointCondition) (
	[]*types.AccessPoint, int64, error) {

	req := new(protoBackend.TopoAccessPointListReq)
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	var (
		total        int64
		accessPoints []*types.AccessPoint
	)
	executor := pageexecutor.NewPageExecutor[*types.AccessPoint](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.AccessPoint, error) {
		req.Page = convertPage(page)
		resp, err := h.cli.listAccessPoint(nCtx, req)
		if err != nil {
			return nil, err
		}

		total, accessPoints = resp.ConvertAccessPointsToTypes()

		return accessPoints, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, total, nil
}

// ===============================================================================
// Business Related Interfaces
// ===============================================================================

// ListBusiness list business within specified tenant in contextx.
func (h *Handler) ListBusiness(nCtx contextx.IContext, page types.Page, condition *types.BusinessCondition) ([]*types.Business, int64, error) {
	req := new(protoBackend.TopoBusinessListReq)
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	var (
		total      int64
		businesses []*types.Business
	)
	executor := pageexecutor.NewPageExecutor[*types.Business](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.Business, error) {
		req.Page = convertPage(page)
		resp, err := h.cli.listBusiness(nCtx, req)
		if err != nil {
			return nil, err
		}

		total, businesses = resp.ConvertBusinessToTypes()

		return businesses, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, total, nil
}

// ===============================================================================
// Topo Constant Related Interfaces
// ===============================================================================

// GetConstant get constant by fields.
func (h *Handler) GetConstant(nCtx contextx.IContext, fields types.TopoConstantFields) (*types.TopoConstant, error) {
	req := &protoBackend.TopoConstantGetReq{}
	if err := req.ConvertFieldsFromTypes(fields); err != nil {
		return nil, err
	}

	resp, err := h.cli.getConstant(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertConstantToTypes(), nil
}

// ===============================================================================
// Graph Node Related Interfaces
// ===============================================================================

// GetGraphNode get graph node.
func (h *Handler) GetGraphNode(nCtx contextx.IContext,
	networkUnitIDs []int64) ([]*types.GraphNodeInfo, error) {

	req := &protoBackend.TopoGraphNodeGetReq{
		BkNetworkunitId: networkUnitIDs,
	}

	resp, err := h.cli.getGraphNode(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// ===============================================================================
// Network Area Related Interfaces
// ===============================================================================

// CreateNetworkArea creates a new networkarea.
func (h *Handler) CreateNetworkArea(nCtx contextx.IContext, networkArea *types.NetworkArea) (int64, error) {
	req := &protoBackend.TopoNetworkAreaCreateReq{
		BkNetworkareaName: networkArea.Name,
		CloudVendor:       networkArea.CloudVendor,
	}

	resp, err := h.cli.createNetworkArea(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetBkNetworkareaId(), nil
}

// UpdateNetworkArea updates an existing networkarea.
func (h *Handler) UpdateNetworkArea(nCtx contextx.IContext, networkArea *types.NetworkArea) error {
	req := &protoBackend.TopoNetworkAreaUpdateReq{
		BkNetworkareaId:   networkArea.ID,
		BkNetworkareaName: networkArea.Name,
		CloudVendor:       networkArea.CloudVendor,
	}

	_, err := h.cli.updateNetworkArea(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// ListNetworkArea list network area within specified tenant in contextx.
func (h *Handler) ListNetworkArea(nCtx contextx.IContext, page types.Page, condition *types.NetworkAreaCondition) (
	[]*types.NetworkArea, int64, error) {

	req := new(protoBackend.TopoNetworkAreaListReq)
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	var (
		total        int64
		networkAreas []*types.NetworkArea
	)
	executor := pageexecutor.NewPageExecutor[*types.NetworkArea](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.NetworkArea, error) {
		req.Page = convertPage(page)
		resp, err := h.cli.listNetworkArea(nCtx, req)
		if err != nil {
			return nil, err
		}

		total, networkAreas = resp.ConvertNetworkAreasToTypes()

		return networkAreas, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, total, nil
}

// GetNetworkArea gets an existing networkarea.
func (h *Handler) GetNetworkArea(nCtx contextx.IContext, networkAreaID int64) (*types.NetworkArea, error) {
	req := &protoBackend.TopoNetworkAreaGetReq{
		BkNetworkareaId: networkAreaID,
	}

	resp, err := h.cli.getNetworkArea(nCtx, req)
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
func (h *Handler) DeleteNetworkArea(nCtx contextx.IContext, networkAreaID int64) error {
	req := &protoBackend.TopoNetworkAreaDeleteReq{
		BkNetworkareaId: networkAreaID,
	}

	_, err := h.cli.deleteNetworkArea(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// ===============================================================================
// Network Unit Related Interfaces
// ===============================================================================

// CreateNetworkUnit creates a new networkunit.
func (h *Handler) CreateNetworkUnit(nCtx contextx.IContext, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (int64, error) {
	req := new(protoBackend.TopoNetworkUnitCreateReq)
	req.ConvertNetworkUnitFromTypes(networkUnit, accessPoints...)

	resp, err := h.cli.createNetworkUnit(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetBkNetworkunitId(), nil
}

// UpdateNetworkUnit updates an existing networkunit.
func (h *Handler) UpdateNetworkUnit(
	nCtx contextx.IContext,
	fields types.NetworkUnitUpdateFields,
	networkUnit *types.NetworkUnit,
	accessPoints ...*types.AccessPoint) error {

	req := new(protoBackend.TopoNetworkUnitUpdateReq)
	req.ConvertNetworkUnitFromTypes(fields, networkUnit, accessPoints...)

	_, err := h.cli.updateNetworkUnit(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// GetNetworkUnit gets an existing networkunit.
func (h *Handler) GetNetworkUnit(nCtx contextx.IContext, networkUnitID int64) (*types.NetworkUnit, map[int64]*types.AccessPoint, error) {
	req := &protoBackend.TopoNetworkUnitGetReq{
		BkNetworkunitId: networkUnitID,
	}

	resp, err := h.cli.getNetworkUnit(nCtx, req)
	if err != nil {
		return nil, nil, err
	}

	networkUnit, accessPoints := resp.ConvertNetworkUnitToTypes()

	return networkUnit, accessPoints, nil
}

// ListNetworkUnit list network unit within specified tenant in contextx.
func (h *Handler) ListNetworkUnit(nCtx contextx.IContext, page types.Page, condition *types.NetworkUnitCondition) (
	[]*types.NetworkUnit, int64, error) {

	req := new(protoBackend.TopoNetworkUnitListReq)
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	var (
		total        int64
		networkUnits []*types.NetworkUnit
	)
	executor := pageexecutor.NewPageExecutor[*types.NetworkUnit](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.NetworkUnit, error) {
		req.Page = convertPage(page)
		resp, err := h.cli.listNetworkUnit(nCtx, req)
		if err != nil {
			return nil, err
		}

		total, networkUnits = resp.ConvertNetworkUnitsToTypes()

		return networkUnits, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, total, nil
}

func (h *Handler) ListNetworkUnitBrief(nCtx contextx.IContext, page types.Page, condition *types.NetworkUnitCondition) (
	[]*types.NetworkUnit, int64, error) {

	req := new(protoBackend.TopoNetworkUnitListBriefReq)
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	var (
		total        int64
		networkUnits []*types.NetworkUnit
	)
	executor := pageexecutor.NewPageExecutor[*types.NetworkUnit](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.NetworkUnit, error) {
		req.Page = convertPage(page)
		resp, err := h.cli.listNetworkUnitBrief(nCtx, req)
		if err != nil {
			return nil, err
		}

		total, networkUnits = resp.ConvertNetworkUnitsToTypes()

		return networkUnits, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, 0, err
	}

	return result.Items, total, nil
}

// DeleteNetworkUnit deletes network unit within specified tenant in contextx.
func (h *Handler) DeleteNetworkUnit(nCtx contextx.IContext, networkUnitID int64) error {
	req := &protoBackend.TopoNetworkUnitDeleteReq{
		BkNetworkunitId: networkUnitID,
	}

	_, err := h.cli.deleteNetworkUnit(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// GetNetworkUnitDistributionByNetworkAreaID get networkunit distribution by network area id.
func (h *Handler) GetNetworkUnitDistributionByNetworkAreaID(
	nCtx contextx.IContext,
	condition *types.NetworkUnitCondition,
) (map[int64]int64, error) {
	req := &protoBackend.TopoGetNetworkUnitDistributionByNetworkAreaIDReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.getNetworkUnitDistributionByNetworkAreaID(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}
