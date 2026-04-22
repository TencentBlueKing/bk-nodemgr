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
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerHost defines the host Handler.
// nolint: interfacebloat
type IHandlerHost interface {
	// ListHost list host within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return host list with page and the total count with filter and error.
	ListHost(nCtx contextx.IContext, page types.Page, condition *types.HostCondition) ([]*types.Host, int64, error)

	// SelectHostID select host id by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host id list and error.
	SelectHostID(nCtx contextx.IContext, condition *types.HostCondition) ([]int64, error)

	// SelectInnerIP select host inner ip by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host inner ip list and error.
	SelectInnerIP(nCtx contextx.IContext, condition *types.HostCondition) ([]string, error)

	// SelectInnerIPV6 select host inner ipv6 by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host inner ipv6 list and error.
	SelectInnerIPV6(nCtx contextx.IContext, condition *types.HostCondition) ([]string, error)

	// SelectNetWorkareaIDAndInnerIP select host networkarea id and inner ip by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host networkarea id and inner ip list and error.
	SelectNetWorkareaIDAndInnerIP(nCtx contextx.IContext, condition *types.HostCondition) ([]string, error)

	// SelectNetWorkareaIDAndInnerIPV6 select host networkarea id and inner ipv6 by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host networkarea id and inner ipv6 list and error.
	SelectNetWorkareaIDAndInnerIPV6(nCtx contextx.IContext, condition *types.HostCondition) ([]string, error)

	// DistinctHost distinct host by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host distinct result and error.
	DistinctHost(nCtx contextx.IContext, condition *types.HostCondition) (*types.HostDistinctResult, error)

	// CountHost count host within specified tenant in contextx.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host count with filter and error.
	CountHost(nCtx contextx.IContext, condition *types.HostCondition) (int64, error)

	// GetHostDistributionByNodeRole get host distribution by node role.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host distribution by node role and error.
	GetHostDistributionByNodeRole(nCtx contextx.IContext, condition *types.HostCondition) (map[types.NodeRole]int64, error)

	// GetHostDistributionByNetworkAreaID get host distribution by node role.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host distribution by node role and error.
	GetHostDistributionByNetworkAreaID(nCtx contextx.IContext, condition *types.HostCondition) (map[int64]int64, error)

	// GetHostDistributionByNodeVersion get host distribution by node version.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the host distribution by node version and error.
	GetHostDistributionByNodeVersion(nCtx contextx.IContext, condition *types.HostCondition) (map[string]int64, error)
}

// ListHost list host within specified tenant in contextx.
// nolint: funlen
func (h *Handler) ListHost(nCtx contextx.IContext, page types.Page, condition *types.HostCondition) ([]*types.Host, int64, error) {
	req := &protoBackend.TopoHostListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listHost(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	total, hosts := resp.ConvertHostsToTypes()

	return hosts, total, nil
}

// SelectHostID select host id within specified tenant in contextx.
func (h *Handler) SelectHostID(nCtx contextx.IContext, condition *types.HostCondition) ([]int64, error) {
	req := &protoBackend.TopoHostSelectHostIDReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.selectHostID(nCtx, req)
	if err != nil {
		return nil, err
	}

	hosts := resp.ConvertHostIDToTypes()

	return hosts, nil
}

// SelectInnerIP select host inner ip within specified tenant in contextx.
func (h *Handler) SelectInnerIP(nCtx contextx.IContext, condition *types.HostCondition) ([]string, error) {
	req := &protoBackend.TopoHostSelectInnerIPReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.selectInnerIP(nCtx, req)
	if err != nil {
		return nil, err
	}

	items := resp.ConvertInnerIPToTypes()

	return items, nil
}

// SelectInnerIPV6 select host inner ipv6 within specified tenant in contextx.
func (h *Handler) SelectInnerIPV6(nCtx contextx.IContext, condition *types.HostCondition) ([]string, error) {
	req := &protoBackend.TopoHostSelectInnerIPV6Req{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.selectInnerIPV6(nCtx, req)
	if err != nil {
		return nil, err
	}

	items := resp.ConvertInnerIPV6ToTypes()

	return items, nil
}

// SelectNetWorkareaIDAndInnerIP select host networkarea id and inner ip within specified tenant in contextx.
func (h *Handler) SelectNetWorkareaIDAndInnerIP(nCtx contextx.IContext, condition *types.HostCondition) ([]string, error) {
	req := &protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.selectNetWorkareaIDAndInnerIP(nCtx, req)
	if err != nil {
		return nil, err
	}

	items := resp.ConvertNetWorkareaIDAndInnerIPToTypes()

	return items, nil
}

// SelectNetWorkareaIDAndInnerIPV6 select host networkarea id and inner ipv6 within specified tenant in contextx.
func (h *Handler) SelectNetWorkareaIDAndInnerIPV6(nCtx contextx.IContext, condition *types.HostCondition) ([]string, error) {
	req := &protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPV6Req{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.selectNetWorkareaIDAndInnerIPV6(nCtx, req)
	if err != nil {
		return nil, err
	}

	items := resp.ConvertNetWorkareaIDAndInnerIPV6ToTypes()

	return items, nil
}

// DistinctHost distinct host within specified tenant in contextx.
func (h *Handler) DistinctHost(nCtx contextx.IContext, condition *types.HostCondition) (*types.HostDistinctResult, error) {
	req := &protoBackend.TopoHostDistinctReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctHost(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// CountHost count host within specified tenant in contextx.
func (h *Handler) CountHost(nCtx contextx.IContext, condition *types.HostCondition) (int64, error) {
	req := &protoBackend.TopoHostListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listHost(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// GetHostDistributionByNodeRole count host within specified tenant in contextx.
func (h *Handler) GetHostDistributionByNodeRole(nCtx contextx.IContext, condition *types.HostCondition) (map[types.NodeRole]int64, error) {
	req := &protoBackend.TopoGetHostDistributionByNodeRoleReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.getHostDistributionByNodeRole(nCtx, req)
	if err != nil {
		return nil, err
	}

	hostDistributionByNodeRole := make(map[types.NodeRole]int64)
	for nodeRole, hostCount := range resp.GetData() {
		hostDistributionByNodeRole[types.NodeRole(nodeRole)] = hostCount
	}

	return hostDistributionByNodeRole, nil
}

// GetHostDistributionByNetworkAreaID count host within specified tenant in contextx.
func (h *Handler) GetHostDistributionByNetworkAreaID(nCtx contextx.IContext, condition *types.HostCondition) (map[int64]int64, error) {
	req := &protoBackend.TopoGetHostDistributionByNetworkAreaIDReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.getHostDistributionByNetworkAreaID(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetData(), nil
}

// GetHostDistributionByNodeVersion count host within specified tenant in contextx.
func (h *Handler) GetHostDistributionByNodeVersion(nCtx contextx.IContext, condition *types.HostCondition) (map[string]int64, error) {
	req := &protoBackend.TopoGetHostDistributionByNodeVersionReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.getHostDistributionByNodeVersion(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}
