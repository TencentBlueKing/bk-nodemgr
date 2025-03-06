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
package backend

import (
	"context"

	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Handler is interface for nodeman backend handler.
type Handler interface {
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

	// ListNetworkArea list network area within specified tenant in context.
	// @param ctx context, contains tenant-id.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the network-area list with page and the total count with filter.
	ListNetworkArea(ctx context.Context, page types.Page, condition *types.NetworkAreaCondition) (
		[]*types.NetworkArea, int64, error)

	// ListNetworkUnit list network unit within specified tenant in context.
	// @param ctx content, contains tenant-id.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the network-unit list with page and the total count with filter.
	ListNetworkUnit(ctx context.Context, page types.Page, condition *types.NetworkUnitCondition) (
		[]*types.NetworkUnit, int64, error)

	// GetNetworkArea get specific network area.
	// @param ctx context, contains tenant-id.
	// @param networkAreaID the network area id.
	// @return the network-area.
	GetNetworkArea(ctx context.Context, networkAreaID int64) (*types.NetworkArea, error)

	// GetNetworkUnit get specific network unit.
	// @param ctx context, contains tenant-id.
	// @param networkUnitID the network unit id.
	// @return the network-unit and its accesspoints.
	GetNetworkUnit(ctx context.Context, networkUnitID int64) (*types.NetworkUnit, map[int64]*types.AccessPoint, error)
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

	req := &proto.TopoBusinessListReq{
		Page: convertPage(page),
	}

	if condition != nil {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				req.ExactIncludeConditions = &proto.TopoBusinessListReq_ExactConditions{
					BkBizId: condition.Exact.BizID,
				}
			}

		case types.ConditionTypeFuzzyInclude:
			if condition.Fuzzy != nil {
				req.FuzzyIncludeConditions = &proto.TopoBusinessListReq_FuzzyConditions{
					BkBizName: condition.Fuzzy.BizName,
				}
			}

		default:
			return nil, 0, ErrConditionTypeNotSupport()
		}
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

	req := &proto.TopoHostListReq{
		Page: convertPage(page),
	}

	if condition != nil {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				req.ExactIncludeConditions = &proto.TopoHostListReq_ExactConditions{
					BkHostId:        condition.Exact.HostID,
					BkBizId:         condition.Exact.BizID,
					BkNetworkareaId: condition.Exact.NetworkAreaID,
					BkOsType:        condition.Exact.OSType,
					NodeRole:        types.NodeRoleListToStringList(condition.Exact.NodeRole),
					NodeStatus:      types.NodeStatusListToStringList(condition.Exact.NodeStatus),
					NodeVersion:     condition.Exact.NodeVersion,
					BkAgentId:       condition.Exact.AgentID,
				}
			}

		case types.ConditionTypeFuzzyInclude:
			if condition.Fuzzy != nil {
				req.FuzzyIncludeConditions = &proto.TopoHostListReq_FuzzyConditions{
					BkHostName:      condition.Fuzzy.HostName,
					DeptName:        condition.Fuzzy.DeptName,
					BkHostInnerip:   condition.Fuzzy.InnerIP,
					BkHostInneripV6: condition.Fuzzy.InnerIPV6,
					BkHostOuterip:   condition.Fuzzy.OuterIP,
					BkHostOuteripV6: condition.Fuzzy.OuterIPV6,
				}
			}

		default:
			return nil, 0, ErrConditionTypeNotSupport()
		}
	}

	resp, err := h.cli.listHost(ctx, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	items := resp.GetItems()
	data := make([]*types.Host, len(items))
	for idx, item := range items {
		host := &types.Host{
			TenantID: item.GetTenantId(),
			HostID:   item.GetBkHostId(),
			Static:   &types.HostStatic{},
			Dynamic:  &types.HostDynamic{},
		}

		if info := item.GetInfo(); info != nil {
			host.Static = &types.HostStatic{
				BizID:         info.GetBkBizId(),
				NetworkAreaID: info.GetBkNetworkareaId(),
				HostName:      info.GetBkHostName(),
				DeptName:      info.GetDeptName(),
				InnerIP:       info.GetBkHostInnerip(),
				InnerIPV6:     info.GetBkHostInneripV6(),
				OuterIP:       info.GetBkHostOuterip(),
				OuterIPV6:     info.GetBkHostOuteripV6(),
				Mac:           info.GetBkMac(),
				OSType:        info.GetBkOsType(),
			}
		}

		if state := item.GetState(); state != nil {
			host.Dynamic = &types.HostDynamic{
				AgentID:     state.GetBkAgentId(),
				NodeRole:    types.NodeRole(state.GetNodeRole()),
				NodeStatus:  types.NodeStatus(state.GetNodeStatus()),
				NodeVersion: state.GetNodeVersion(),
			}
		}

		data[idx] = host
	}

	return data, resp.GetTotal(), nil
}

// ListNetworkArea list network area within specified tenant in context.
func (h *handler) ListNetworkArea(ctx context.Context, page types.Page, condition *types.NetworkAreaCondition) (
	[]*types.NetworkArea, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	req := &proto.TopoNetworkAreaListReq{
		Page: convertPage(page),
	}

	if condition != nil {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				req.ExactIncludeConditions = &proto.TopoNetworkAreaListReq_ExactConditions{
					BkNetworkareaId: condition.Exact.NetworkAreaID,
					BkCloudVendor:   condition.Exact.CloudVendor,
				}
			}

		case types.ConditionTypeFuzzyInclude:
			if condition.Fuzzy != nil {
				req.FuzzyIncludeConditions = &proto.TopoNetworkAreaListReq_FuzzyConditions{
					BkNetworkareaName: condition.Fuzzy.NetworkAreaName,
				}
			}

		default:
			return nil, 0, ErrConditionTypeNotSupport()
		}
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
			CloudVendor: item.GetBkCloudVendor(),
		}
	}

	return data, resp.GetTotal(), nil
}

// ListNetworkUnit list network unit within specified tenant in context.
func (h *handler) ListNetworkUnit(ctx context.Context, page types.Page, condition *types.NetworkUnitCondition) (
	[]*types.NetworkUnit, int64, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, 0, err
	}

	req := &proto.TopoNetworkUnitListReq{
		Page: convertPage(page),
	}

	if condition != nil {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				req.ExactIncludeConditions = &proto.TopoNetworkUnitListReq_ExactConditions{
					BkNetworkunitId: condition.Exact.NetworkUnitID,
					BkNetworkareaId: condition.Exact.NetworkAreaID,
				}
			}

		default:
			return nil, 0, ErrConditionTypeNotSupport()
		}
	}

	resp, err := h.cli.listNetworkUnit(ctx, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	items := resp.GetItems()
	data := make([]*types.NetworkUnit, len(items))
	for idx, item := range items {
		data[idx] = &types.NetworkUnit{
			TenantID:      item.GetTenantId(),
			NetworkAreaID: item.GetBkNetworkareaId(),
			ID:            item.GetBkNetworkunitId(),
			Name:          item.GetBkNetworkunitName(),
			AccessPoints:  item.GetAccessPoints(),
			Links:         convertLinks(item.GetLinks()),
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

	req := &proto.TopoNetworkAreaGetReq{
		BkNetworkareaId: networkAreaID,
	}

	resp, err := h.cli.getNetworkArea(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}

	return &types.NetworkArea{
		TenantID:    resp.GetTenantId(),
		ID:          resp.GetBkNetworkareaId(),
		Name:        resp.GetBkNetworkareaName(),
		CloudVendor: resp.GetBkCloudVendor(),
	}, nil
}

// GetNetworkUnit gets an existing networkunit.
func (h *handler) GetNetworkUnit(ctx context.Context, networkUnitID int64) (
	*types.NetworkUnit, map[int64]*types.AccessPoint, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, nil, err
	}

	req := &proto.TopoNetworkUnitGetReq{
		BkNetworkunitId: networkUnitID,
	}

	resp, err := h.cli.getNetworkUnit(ctx, tenantID, req)
	if err != nil {
		return nil, nil, err
	}

	accessPoints := convertAccssPoints(resp.GetTenantId(), resp.GetBkNetworkareaId(), resp.GetAccessPoints())
	accessPointsMap := make(map[int64]*types.AccessPoint)
	accessPointIDs := make([]int64, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		accessPointsMap[accessPoint.ID] = accessPoint
		accessPointIDs[idx] = accessPoint.ID
	}

	return &types.NetworkUnit{
		TenantID:      resp.GetTenantId(),
		NetworkAreaID: resp.GetBkNetworkareaId(),
		ID:            resp.GetBkNetworkunitId(),
		Name:          resp.GetBkNetworkunitName(),
		AccessPoints:  accessPointIDs,
		Links:         convertLinks(resp.GetLinks()),
	}, accessPointsMap, nil
}

func convertAccssPoints(
	tenantID string,
	networkAreaID int64,
	accessPoints []*proto.AccessPoint) []*types.AccessPoint {

	data := make([]*types.AccessPoint, len(accessPoints))
	for idx, accessPoint := range accessPoints {
		data[idx] = &types.AccessPoint{
			TenantID:      tenantID,
			NetworkAreaID: networkAreaID,
			Name:          accessPoint.GetAccesspointName(),
		}

		if endpoints := accessPoint.GetEndpoints(); endpoints != nil {
			data[idx].Endpoints.Cluster = endpoints.GetCluster()
			data[idx].Endpoints.File = endpoints.GetFile()
			data[idx].Endpoints.Data = endpoints.GetData()
		}
	}

	return data
}

func convertLinks(links *proto.Links) types.Links {
	data := types.Links{}
	if links != nil {
		if reqLink := links.GetCluster(); reqLink != nil {
			data.Cluster = &types.Link{AccessPointID: reqLink.GetAccesspointId()}
		}
		if reqLink := links.GetFile(); reqLink != nil {
			data.File = &types.Link{AccessPointID: reqLink.GetAccesspointId()}
		}
		if reqLink := links.GetData(); reqLink != nil {
			data.Data = &types.Link{AccessPointID: reqLink.GetAccesspointId()}
		}
	}

	return data
}
