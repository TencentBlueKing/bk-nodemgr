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
	"context"

	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	maxHostLimit = 1000
)

// ListHost lists hosts with page and conditions.
func (h *handler) ListHost(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.TopoHostListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list host, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountHost(
			ctx,
			req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to list host, failed to count host. err: %v", err)
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.TopoHostListResp)
		resp.ConvertHostsFromTypes(num, nil, types.TopoNameMapping{})

		return resp.GetData(), nil
	}

	hosts, num, err := h.backendHandler.ListHost(
		ctx,
		req.ConvertPageToTypes(maxHostLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list host. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	var mapping types.TopoNameMapping
	gp := gopool.NewPool()
	gp.Go(func() error {
		_ = h.completeNetworkAreaName(ctx, hosts, &mapping)
		return nil
	})
	gp.Go(func() error {
		_ = h.completeNetworkUnitName(ctx, hosts, &mapping)
		return nil
	})
	_ = gp.Wait()

	resp := new(protoApplication.TopoHostListResp)
	resp.ConvertHostsFromTypes(num, hosts, mapping)

	return resp.GetData(), nil
}

func (h *handler) completeNetworkAreaName(
	ctx context.Context, hosts []*types.Host, mapping *types.TopoNameMapping) error {

	idMap := make(map[int64]bool)
	for _, host := range hosts {
		idMap[host.Static.NetworkAreaID] = true
	}

	ids := make([]int64, len(idMap))
	index := 0
	for id := range idMap {
		ids[index] = id
		index++
	}

	items, _, err := h.backendHandler.ListNetworkArea(
		ctx,
		types.Page{Limit: len(ids)},
		&types.NetworkAreaCondition{
			ExactInclude: &types.NetworkAreaExactFields{
				NetworkAreaID: ids,
			},
		},
	)

	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to complete networkarea name. err: %v", err)
		return err
	}

	h.logger.InfoCtxf(ctx, "completed networkarea name: %d", len(items))
	mapping.NetworkArea = make(map[int64]string)
	for _, item := range items {
		mapping.NetworkArea[item.ID] = item.Name
	}

	return nil
}

func (h *handler) completeNetworkUnitName(
	ctx context.Context, hosts []*types.Host, mapping *types.TopoNameMapping) error {

	idMap := make(map[int64]bool)
	for _, host := range hosts {
		idMap[host.Dynamic.NetworkUnitID] = true
	}

	ids := make([]int64, len(idMap))
	index := 0
	for id := range idMap {
		ids[index] = id
		index++
	}

	items, _, err := h.backendHandler.ListNetworkUnit(
		ctx,
		types.Page{Limit: len(ids)},
		&types.NetworkUnitCondition{
			ExactInclude: &types.NetworkUnitExactFields{
				NetworkUnitID: ids,
			},
		},
	)

	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to complete networkunit name. err: %v", err)
		return err
	}

	h.logger.InfoCtxf(ctx, "completed networkunit name: %d", len(items))
	mapping.NetworkUnit = make(map[int64]string)
	for _, item := range items {
		mapping.NetworkUnit[item.ID] = item.Name
	}

	return nil
}

// DistinctHost get distinct host fields.
func (h *handler) DistinctHost(ctx *restserver.Context) (interface{}, error) {
	req := new(protoApplication.TopoHostDistinctReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to distinct host, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.backendHandler.DistinctHost(
		ctx,
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to distinct host. failed to distinct host fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.TopoHostDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
