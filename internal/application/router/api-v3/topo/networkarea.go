/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package topo

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// CreateNetworkArea creates a new network-area.
func (h *handler) CreateNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaCreateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create networkarea, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkAreaID, err := h.backendHandler.CreateNetworkArea(rCtx, req.ConvertNetworkAreaToTypes(rCtx.TenantID(), -1))
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create networkarea, failed to create networkarea via cmdb")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaCreateResp)
	resp.ConvertNetworkAreaFromTypes(networkAreaID)

	return resp.GetData(), nil
}

// UpdateNetworkArea updates an existing network-area.
func (h *handler) UpdateNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaUpdateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update networkarea, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.backendHandler.UpdateNetworkArea(rCtx, req.ConvertNetworkAreaToTypes(rCtx.TenantID())); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update networkarea, failed to upsert networkarea")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaUpdateResp)
	resp.ConvertNetworkAreaFromTypes(req.GetBkNetworkareaId())

	return resp.GetData(), nil
}

// GetNetworkArea gets an existing network-area.
func (h *handler) GetNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get networkarea, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkArea, err := h.backendHandler.GetNetworkArea(rCtx, req.GetBkNetworkareaId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get networkarea")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaCreateResp)
	resp.ConvertNetworkAreaFromTypes(networkArea.ID)

	return resp.GetData(), nil
}

// ListNetworkArea lists network-area.
func (h *handler) ListNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkarea, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkarea, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// special logic:
	// networkarea request with limit 0 means unlimited
	if page.Limit == 0 {
		page = types.UnlimitedPage()
	}

	networkAreas, num, err := h.backendHandler.ListNetworkArea(rCtx, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkarea")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaListResp)
	resp.ConvertNetworkAreasFromTypes(num, networkAreas)

	return resp.GetData(), nil
}

// StatisticsNetworkArea statistics network-area.
// nolint: funlen, gocognit
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (h *handler) StatisticsNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaStatisticsReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics networkarea, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkUnitDistributionByNetworkAreaID, err := h.backendHandler.GetNetworkUnitDistributionByNetworkAreaID(
		rCtx,
		req.ConvertNetworkUnitConditionToTypes(),
	)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics networkarea, failed to get networkunit distribution by network area id")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	// result is a map of networkareaID and networkareaStatics
	result := make(map[int64]*protoApplication.NetworkAreaStatistics)
	for networkAreaID, networkUnitCount := range networkUnitDistributionByNetworkAreaID {
		result[networkAreaID] = &protoApplication.NetworkAreaStatistics{
			NetworkAreaID:    networkAreaID,
			NetworkUnitCount: networkUnitCount,
		}
	}

	networkAreaIDs := req.GetBkNetworkareaId()
	for _, id := range networkAreaIDs {
		if _, ok := result[id]; !ok {
			result[id] = &protoApplication.NetworkAreaStatistics{
				NetworkAreaID: id,
			}
		}
	}

	gp := gopool.NewPool()
	// get agent host count
	gp.Go(func() error {
		hostDistributionByNetworkAreaID, err := h.backendHandler.GetHostDistributionByNetworkAreaID(rCtx, &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				NetworkAreaID: networkAreaIDs,
			},
			DynamicExactInclude: &types.HostDynamicExactFields{
				NodeRole: []types.NodeRole{types.NodeRoleAgent},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to count agent: %w", err)
		}

		for networkAreaID, hostCount := range hostDistributionByNetworkAreaID {
			if _, ok := result[networkAreaID]; ok {
				result[networkAreaID].AgentCount = hostCount
			}
		}

		return nil
	})

	// get proxy host count
	gp.Go(func() error {
		hostDistributionByNetworkAreaID, err := h.backendHandler.GetHostDistributionByNetworkAreaID(rCtx, &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				NetworkAreaID: networkAreaIDs,
			},
			DynamicExactInclude: &types.HostDynamicExactFields{
				NodeRole: []types.NodeRole{types.NodeRoleProxy},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to count agent: %w", err)
		}

		for networkAreaID, hostCount := range hostDistributionByNetworkAreaID {
			if _, ok := result[networkAreaID]; ok {
				result[networkAreaID].ProxyCount = hostCount
			}
		}

		return nil
	})

	// wait until all servers stopped or application error.
	if err := gp.Wait(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to statistics networkarea, failed to count host: %v", err)

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaStatisticsResp)
	resp.ConvertNetworkAreaStatisticsFromResult(result)

	return resp.GetData(), nil
}

// DeleteNetworkArea deletes an existing network-area.
func (h *handler) DeleteNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete networkarea, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.backendHandler.DeleteNetworkArea(rCtx, req.GetBkNetworkareaId()); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete networkarea")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaDeleteResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkareaId())

	return resp.GetData(), nil
}
