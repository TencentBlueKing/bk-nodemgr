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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
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
func (h *handler) ListHost(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoHostListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountHost(
			rCtx,
			req.ConvertConditionsToTypes())
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list host, failed to count host")
			return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
		}

		resp := new(protoApplication.TopoHostListResp)
		resp.ConvertHostsFromTypes(num, nil, types.TopoNameMapping{})

		return resp.GetData(), nil
	}

	hosts, num, err := h.backendHandler.ListHost(
		rCtx,
		req.ConvertPageToTypes(maxHostLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	var mapping types.TopoNameMapping
	gp := gopool.NewPool()
	gp.Go(func() error {
		_ = h.completeNetworkAreaName(rCtx, hosts, &mapping)
		return nil
	})
	gp.Go(func() error {
		_ = h.completeNetworkUnitName(rCtx, hosts, &mapping)
		return nil
	})
	_ = gp.Wait()

	resp := new(protoApplication.TopoHostListResp)
	resp.ConvertHostsFromTypes(num, hosts, mapping)

	return resp.GetData(), nil
}

// SimpleListHost lists hosts with field selection.
func (h *handler) SimpleListHost(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoHostSimpleListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to simple list host, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, num, err := h.backendHandler.SimpleListHost(
		rCtx,
		req.ConvertConditionsToTypes(),
		req.ConvertFieldSelectionToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to simple list host")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoHostSimpleListResp)
	resp.ConvertHostSelectedFiledFromTypes(num, hosts)

	return resp.GetData(), nil
}

func (h *handler) completeNetworkAreaName(
	rCtx restserver.IContext, hosts []*types.Host, mapping *types.TopoNameMapping) error {

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
		rCtx,
		types.Page{Limit: len(ids)},
		&types.NetworkAreaCondition{
			ExactInclude: &types.NetworkAreaExactFields{
				NetworkAreaID: ids,
			},
		},
	)

	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to complete networkarea name")
		return err
	}

	logger.G.Biz(rCtx).With("items", len(items)).Info("completed networkarea name")
	mapping.NetworkArea = make(map[int64]string)
	for _, item := range items {
		mapping.NetworkArea[item.ID] = item.Name
	}

	return nil
}

func (h *handler) completeNetworkUnitName(
	rCtx restserver.IContext, hosts []*types.Host, mapping *types.TopoNameMapping) error {

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
		rCtx,
		types.Page{Limit: len(ids)},
		&types.NetworkUnitCondition{
			ExactInclude: &types.NetworkUnitExactFields{
				NetworkUnitID: ids,
			},
		},
	)

	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to complete networkunit name")
		return err
	}

	logger.G.Biz(rCtx).With("items", len(items)).Info("completed networkunit name")
	mapping.NetworkUnit = make(map[int64]string)
	for _, item := range items {
		mapping.NetworkUnit[item.ID] = item.Name
	}

	return nil
}

// DistinctHost get distinct host fields.
func (h *handler) DistinctHost(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoHostDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct host, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.backendHandler.DistinctHost(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct host. failed to distinct host fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.TopoHostDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
