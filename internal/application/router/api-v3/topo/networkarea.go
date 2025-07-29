/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	"errors"
	"fmt"

	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// not max limit in networkarea.
	// return all data in one request.
	maxNetworkAreaLimit = 0
)

// CreateNetworkArea creates a new network-area.
func (h *handler) CreateNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaCreateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkAreaID, err := h.backendHandler.CreateNetworkArea(ctx, req.ConvertNetworkAreaToTypes(ctx.TenantID, -1))
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create networkarea, failed to create networkarea via cmdb. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaCreateResp)
	resp.ConvertNetworkAreaFromTypes(networkAreaID)

	return resp.GetData(), nil
}

// UpdateNetworkArea updates an existing network-area.
func (h *handler) UpdateNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaUpdateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to update networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.backendHandler.UpdateNetworkArea(ctx, req.ConvertNetworkAreaToTypes(ctx.TenantID)); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to update networkarea, failed to upsert networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaUpdateResp)
	resp.ConvertNetworkAreaFromTypes(req.GetBkNetworkareaId())

	return resp.GetData(), nil
}

// GetNetworkArea gets an existing network-area.
func (h *handler) GetNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkArea, err := h.backendHandler.GetNetworkArea(ctx, req.GetBkNetworkareaId())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaCreateResp)
	resp.ConvertNetworkAreaFromTypes(networkArea.ID)

	return resp.GetData(), nil
}

// ListNetworkArea lists network-area.
func (h *handler) ListNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkAreas, num, err := h.backendHandler.ListNetworkArea(
		ctx,
		req.ConvertPageToTypes(maxNetworkAreaLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaListResp)
	resp.ConvertNetworkAreasFromTypes(num, networkAreas)

	return resp.GetData(), nil
}

// StatisticsNetworkArea statistics network-area.
// nolint: funlen, gocognit
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (h *handler) StatisticsNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaStatisticsReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to statistics networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkunits, _, err := h.backendHandler.ListNetworkUnit(
		ctx,
		types.Page{Limit: 0},
		req.ConvertNetworkUnitConditionToTypes(),
	)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to statistics networkarea, failed to list networkunit. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// result is a map of networkareaID and networkareaStatics
	result := make(map[int64]*protoApplication.NetworkAreaStatistics)
	for _, networkunit := range networkunits {
		if _, ok := result[networkunit.NetworkAreaID]; !ok {
			result[networkunit.NetworkAreaID] = &protoApplication.NetworkAreaStatistics{
				NetworkAreaID: networkunit.NetworkAreaID,
			}
		}
		result[networkunit.NetworkAreaID].NetworkUnitCount++
	}

	networkAreaIDs := req.GetBkNetworkareaId()
	gp := gopool.NewPool()
	for idx := range networkAreaIDs {
		networkAreaID := networkAreaIDs[idx]

		if _, ok := result[networkAreaID]; !ok {
			result[networkAreaID] = &protoApplication.NetworkAreaStatistics{
				NetworkAreaID: networkAreaID,
			}
		}

		// count agent.
		gp.Go(func() error {
			num, err := h.backendHandler.CountHost(ctx, &types.HostCondition{
				ExactInclude: &types.HostExactFields{
					NetworkUnitID: []int64{networkAreaID},
					NodeRole:      []types.NodeRole{types.NodeRoleAgent},
				},
			})
			if err != nil {
				return errors.Join(err, fmt.Errorf("failed to count agent, networkunit-id: %d", networkAreaID))
			}

			result[networkAreaID].AgentCount = num

			return nil
		})

		// count proxy.
		gp.Go(func() error {
			num, err := h.backendHandler.CountHost(ctx, &types.HostCondition{
				ExactInclude: &types.HostExactFields{
					NetworkUnitID: []int64{networkAreaID},
					NodeRole:      []types.NodeRole{types.NodeRoleProxy},
				},
			})
			if err != nil {
				return errors.Join(err, fmt.Errorf("failed to count proxy, networkunit-id: %d", networkAreaID))
			}

			result[networkAreaID].ProxyCount = num

			return nil
		})

		gp.Go(func() error {
			events, _, err := h.backendHandler.ListTopoEvent(ctx, types.Page{Limit: 1}, &types.TopoEventCondition{
				ExactInclude: &types.TopoEventExactFields{
					NetworkAreaID: []int64{networkAreaID},
				},
			})
			if err != nil {
				return errors.Join(err, fmt.Errorf("failed to list topo event, networkunit-id: %d", networkAreaID))
			}

			if len(events) == 0 {
				return nil
			}

			result[networkAreaID].LastOperator = events[0].Operator
			result[networkAreaID].LastOperateTime = events[0].OperateTime

			return nil
		})
	}

	// wait until all servers stopped or application error.
	if err := gp.Wait(); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to statistics networkarea, failed to count host: %v", err)

		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaStatisticsResp)
	resp.ConvertNetworkAreaStatisticsFromResult(result)

	return resp.GetData(), nil
}

// DeleteNetworkArea deletes an existing network-area.
func (h *handler) DeleteNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(protoApplication.TopoNetworkAreaDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.backendHandler.DeleteNetworkArea(ctx, req.GetBkNetworkareaId()); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.TopoNetworkAreaDeleteResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkareaId())

	return resp.GetData(), nil
}
