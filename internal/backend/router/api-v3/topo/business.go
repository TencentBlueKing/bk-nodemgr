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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListBusiness list business with specified conditions.
func (h *handler) ListBusiness(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoBusinessListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list business, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list business, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	bizs, num, err := h.storage.ListBusinesses(rCtx, page, req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list business")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoBusinessListResp)
	resp.ConvertBusinessFromTypes(num, bizs)

	return resp.GetData(), nil
}

// GetBusinessHostCount gets host count grouped by business id.
func (h *handler) GetBusinessHostCount(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoBusinessHostCountGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get business host count, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition := &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			BizID: req.GetBkBizId(),
		},
	}

	counts, err := h.storage.CountHostGroupByBizID(rCtx, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get business host count")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoBusinessHostCountGetResp)
	resp.ConvertHostCountFromTypes(counts)

	return resp.GetData(), nil
}

const (
	// businessInstTopoObjIDBiz is the CMDB topology object id for business nodes.
	businessInstTopoObjIDBiz = "biz"
	// businessInstTopoObjIDModule is the CMDB topology object id for module nodes.
	businessInstTopoObjIDModule = "module"
)

// GetBusinessInstTopo gets business instance topology with aggregated host count.
func (h *handler) GetBusinessInstTopo(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoBusinessInstTopoGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get business inst topo, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	bizID := req.GetBkBizId()
	topoNodes, err := h.cmdbHandler.SearchBizInstTopo(rCtx, bizID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get business inst topo, failed to search business inst topo from cmdb")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.TopoBusinessInstTopoGetResp)
	if len(topoNodes) == 0 {
		logger.G.Biz(rCtx).With("biz-id", bizID).Info("business inst topo is empty")
		resp.Data = &protoBackend.TopoBusinessInstTopoGetResp_Data{Items: nil}

		return resp.GetData(), nil
	}

	if len(topoNodes) > 1 {
		logger.G.Biz(rCtx).With("biz-id", bizID).Error("invalid business inst topo, more than 1 root nodes")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, errors.New("invalid business inst topo, more than 1 root nodes"))
	}

	topoNode := topoNodes[0]
	bizHostCount, moduleHostCount, err := h.getBusinessInstTopoHostCounts(rCtx, topoNode)
	if err != nil {
		return nil, err
	}

	fillBusinessInstTopoHostCount(topoNode, bizHostCount, moduleHostCount)
	resp.ConvertBusinessInstTopoFromTypes(topoNode)

	return resp.GetData(), nil
}

// nolint: nonamedreturns
func (h *handler) getBusinessInstTopoHostCounts(rCtx restserver.IContext, topoNode *types.TopoNodeInfo) (
	bizHostCount map[int64]int64, moduleHostCount map[int64]int64, err error) {

	moduleIDList := make([]int64, 0)
	bizIDList := make([]int64, 0)

	var topoInstDFS func(*types.TopoNodeInfo)
	topoInstDFS = func(topoNode *types.TopoNodeInfo) {
		if topoNode == nil {
			return
		}

		switch topoNode.ObjID {
		case businessInstTopoObjIDBiz:
			bizIDList = append(bizIDList, topoNode.InstID)
		case businessInstTopoObjIDModule:
			moduleIDList = append(moduleIDList, topoNode.InstID)
		}

		for _, child := range topoNode.Children {
			topoInstDFS(child)
		}
	}

	topoInstDFS(topoNode)

	if len(bizIDList) != 0 {
		bizHostCount, err = h.storage.CountHostGroupByBizID(rCtx, &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				BizID: bizIDList,
			},
		})
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count host group by biz id")
			return nil, nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}
	}

	if len(moduleIDList) != 0 {
		moduleHostCount, err = h.storage.CountHostGroupByModuleID(rCtx, &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				ModuleID: moduleIDList,
			},
		})
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count host group by module id")
			return nil, nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}
	}

	return bizHostCount, moduleHostCount, nil
}

func fillBusinessInstTopoHostCount(topoNode *types.TopoNodeInfo, bizHostCount, moduleHostCount map[int64]int64) {
	var fillNodeHostCount func(*types.TopoNodeInfo) int64
	fillNodeHostCount = func(topoNode *types.TopoNodeInfo) int64 {
		if topoNode == nil {
			return 0
		}

		var childHostCount int64
		for _, child := range topoNode.Children {
			childHostCount += fillNodeHostCount(child)
		}

		switch topoNode.ObjID {
		case businessInstTopoObjIDBiz:
			topoNode.HostCount = childHostCount

			// if host is under idle pool, it will not be counted in module host count, but will be counted in biz host count
			// so we need to take the max value of biz host count and child host count to avoid the case that
			// host count is less than the sum of child host count.
			if bizHostCount[topoNode.InstID] > childHostCount {
				topoNode.HostCount = bizHostCount[topoNode.InstID]
			}
		case businessInstTopoObjIDModule:
			topoNode.HostCount = moduleHostCount[topoNode.InstID]
		default:
			topoNode.HostCount += childHostCount
		}

		return topoNode.HostCount
	}

	fillNodeHostCount(topoNode)
}
