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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	customTopoSetBatchSize     = 500
	customTopoHostCountTimeout = 30 * time.Minute
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

	bizIDs := req.GetBkBizId()
	narrowedBizIDs, scopeIsAny, authErr := h.narrowAuthorizedBizIDsForHostList(rCtx, bizIDs, nil)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to get business host count, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	condition := narrowHostConditionByBiz(&types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			BizID: bizIDs,
		},
	}, narrowedBizIDs, scopeIsAny)

	counts, err := h.storage.CountHostGroupByBizID(rCtx, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get business host count")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoBusinessHostCountGetResp)
	resp.ConvertHostCountFromTypes(counts)

	return resp.GetData(), nil
}

// GetBusinessInstTopo gets business instance topology with aggregated host count.
func (h *handler) GetBusinessInstTopo(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoBusinessInstTopoGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get business inst topo, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	bizID := req.GetBkBizId()
	narrowedBizIDs, scopeIsAny, authErr := h.narrowAuthorizedBizIDsForHostList(rCtx, []int64{bizID}, nil)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to get business inst topo, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	topoNodes, err := h.cmdbHandler.GetBizBriefCacheTopo(rCtx, bizID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get business inst topo, failed to get business brief cache topo from cmdb")
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
	bizHostCount, setHostCount, moduleHostCount, customHostCount, err := h.getBusinessInstTopoHostCounts(rCtx, topoNode, narrowedBizIDs, scopeIsAny)
	if err != nil {
		return nil, err
	}

	fillBusinessInstTopoHostCount(topoNode, bizHostCount, setHostCount, moduleHostCount, customHostCount)
	resp.ConvertBusinessInstTopoFromTypes(topoNode)

	return resp.GetData(), nil
}

type businessInstTopoIDs struct {
	bizIDs           []int64
	setIDs           []int64
	moduleIDs        []int64
	customTopoSetIDs map[*types.TopoNodeInfo][]int64
}

func collectBusinessInstTopoIDs(topoNode *types.TopoNodeInfo) businessInstTopoIDs {
	result := businessInstTopoIDs{
		bizIDs:           make([]int64, 0),
		setIDs:           make([]int64, 0),
		moduleIDs:        make([]int64, 0),
		customTopoSetIDs: make(map[*types.TopoNodeInfo][]int64),
	}
	collectBusinessInstTopoNodeIDs(topoNode, &result)
	result.bizIDs = conv.SliceUnique(result.bizIDs)
	result.setIDs = conv.SliceUnique(result.setIDs)
	result.moduleIDs = conv.SliceUnique(result.moduleIDs)

	return result
}

func collectBusinessInstTopoNodeIDs(topoNode *types.TopoNodeInfo, result *businessInstTopoIDs) []int64 {
	if topoNode == nil {
		return nil
	}

	descendantSetIDs := make([]int64, 0)
	switch topoNode.ObjID {
	case cmdb.TopoNodeObjIDBiz:
		result.bizIDs = append(result.bizIDs, topoNode.InstID)
	case cmdb.TopoNodeObjIDSet:
		result.setIDs = append(result.setIDs, topoNode.InstID)
		descendantSetIDs = append(descendantSetIDs, topoNode.InstID)
	case cmdb.TopoNodeObjIDModule:
		result.moduleIDs = append(result.moduleIDs, topoNode.InstID)
	}

	for _, child := range topoNode.Children {
		descendantSetIDs = append(descendantSetIDs, collectBusinessInstTopoNodeIDs(child, result)...)
	}

	descendantSetIDs = conv.SliceUnique(descendantSetIDs)
	if len(descendantSetIDs) == 0 || cmdb.IsMainlineObject(topoNode.ObjID) {
		return descendantSetIDs
	}
	result.customTopoSetIDs[topoNode] = descendantSetIDs

	return descendantSetIDs
}

// nolint: nonamedreturns
func (h *handler) getBusinessInstTopoHostCounts(
	rCtx restserver.IContext,
	topoNode *types.TopoNodeInfo,
	narrowedBizIDs []int64,
	scopeIsAny bool,
) (bizHostCount, setHostCount, moduleHostCount map[int64]int64, customHostCount map[*types.TopoNodeInfo]int64, err error) {

	topoIDs := collectBusinessInstTopoIDs(topoNode)
	bizHostCount, setHostCount, moduleHostCount, err = h.countStandardBusinessInstTopoHosts(
		rCtx, topoIDs, narrowedBizIDs, scopeIsAny,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	customHostCount, err = h.countCustomBusinessInstTopoHosts(
		rCtx, topoIDs.customTopoSetIDs, narrowedBizIDs, scopeIsAny,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return bizHostCount, setHostCount, moduleHostCount, customHostCount, nil
}

// nolint: nonamedreturns
func (h *handler) countStandardBusinessInstTopoHosts(
	rCtx restserver.IContext,
	topoIDs businessInstTopoIDs,
	narrowedBizIDs []int64,
	scopeIsAny bool,
) (bizHostCount, setHostCount, moduleHostCount map[int64]int64, err error) {

	if len(topoIDs.bizIDs) != 0 {
		bizCond := narrowHostConditionByBiz(&types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{BizID: topoIDs.bizIDs},
		}, narrowedBizIDs, scopeIsAny)
		bizHostCount, err = h.storage.CountHostGroupByBizID(rCtx, bizCond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count host group by biz id")
			return nil, nil, nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}
	}

	if len(topoIDs.setIDs) != 0 {
		setCond := narrowHostConditionByBiz(&types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{SetID: topoIDs.setIDs},
		}, narrowedBizIDs, scopeIsAny)
		setHostCount, err = h.storage.CountHostGroupBySetID(rCtx, setCond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count host group by set id")
			return nil, nil, nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}
	}

	if len(topoIDs.moduleIDs) != 0 {
		moduleCond := narrowHostConditionByBiz(&types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{ModuleID: topoIDs.moduleIDs},
		}, narrowedBizIDs, scopeIsAny)
		moduleHostCount, err = h.storage.CountHostGroupByModuleID(rCtx, moduleCond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count host group by module id")
			return nil, nil, nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}
	}

	return bizHostCount, setHostCount, moduleHostCount, nil
}

func (h *handler) countCustomBusinessInstTopoHosts(
	rCtx restserver.IContext,
	customTopoSetIDs map[*types.TopoNodeInfo][]int64,
	narrowedBizIDs []int64,
	scopeIsAny bool,
) (map[*types.TopoNodeInfo]int64, error) {

	customHostCount := make(map[*types.TopoNodeInfo]int64, len(customTopoSetIDs))
	customSetIDList := make([]int64, 0)
	for _, setIDs := range customTopoSetIDs {
		customSetIDList = append(customSetIDList, setIDs...)
	}
	customSetIDList = conv.SliceUnique(customSetIDList)
	if len(customSetIDList) == 0 {
		return customHostCount, nil
	}

	hostIDsBySetID := make(map[int64]map[int64]struct{}, len(customSetIDList))
	err := batchexecutor.Execute(
		rCtx,
		customSetIDList,
		func(nCtx contextx.IContext, batchSetIDs []int64) error {
			customSetCondition := narrowHostConditionByBiz(&types.HostCondition{
				StaticExactInclude: &types.HostStaticExactFields{SetID: batchSetIDs},
			}, narrowedBizIDs, scopeIsAny)
			hosts, _, err := h.storage.ListHostWithFields(
				nCtx,
				types.UnlimitedPage(),
				&types.HostFieldSelection{HostID: true, Topo: true},
				customSetCondition,
			)
			if err != nil {
				return err
			}

			addCustomTopoHostIDs(hostIDsBySetID, hosts, batchSetIDs)

			return nil
		},
		batchexecutor.WithBatchSize(customTopoSetBatchSize),
		batchexecutor.WithTimeout(customTopoHostCountTimeout),
	)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to count hosts under custom topology levels")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	return countCustomTopoHostsFromSetIDs(customTopoSetIDs, hostIDsBySetID), nil
}

func addCustomTopoHostIDs(hostIDsBySetID map[int64]map[int64]struct{}, hosts []*types.Host, setIDs []int64) {
	setIDSet := make(map[int64]struct{}, len(setIDs))
	for _, setID := range setIDs {
		setIDSet[setID] = struct{}{}
	}

	for _, host := range hosts {
		if host == nil || host.Static == nil {
			continue
		}

		for _, topo := range host.Static.Topo {
			if _, ok := setIDSet[topo.SetID]; !ok {
				continue
			}

			if _, ok := hostIDsBySetID[topo.SetID]; !ok {
				hostIDsBySetID[topo.SetID] = make(map[int64]struct{})
			}
			hostIDsBySetID[topo.SetID][host.HostID] = struct{}{}
		}
	}
}

func countCustomTopoHostsFromSetIDs(
	customTopoSetIDs map[*types.TopoNodeInfo][]int64,
	hostIDsBySetID map[int64]map[int64]struct{},
) map[*types.TopoNodeInfo]int64 {

	customHostCount := make(map[*types.TopoNodeInfo]int64, len(customTopoSetIDs))
	for topoNode, setIDs := range customTopoSetIDs {
		hostIDs := make(map[int64]struct{})
		for _, setID := range setIDs {
			for hostID := range hostIDsBySetID[setID] {
				hostIDs[hostID] = struct{}{}
			}
		}
		customHostCount[topoNode] = int64(len(hostIDs))
	}

	return customHostCount
}

func fillBusinessInstTopoHostCount(
	topoNode *types.TopoNodeInfo,
	bizHostCount, setHostCount, moduleHostCount map[int64]int64,
	customHostCount map[*types.TopoNodeInfo]int64,
) {

	var fillNodeHostCount func(*types.TopoNodeInfo) int64
	fillNodeHostCount = func(topoNode *types.TopoNodeInfo) int64 {
		if topoNode == nil {
			return 0
		}

		childHostCount := int64(0)
		for _, child := range topoNode.Children {
			childHostCount += fillNodeHostCount(child)
		}

		switch topoNode.ObjID {
		case cmdb.TopoNodeObjIDBiz:
			// if host is under idle pool, it will not be counted in module host count, but will be counted in biz host count
			// so we need to use biz host count from storage directly to avoid host count missing caused by idle pool.
			topoNode.HostCount = bizHostCount[topoNode.InstID]
		case cmdb.TopoNodeObjIDSet:
			topoNode.HostCount = setHostCount[topoNode.InstID]
		case cmdb.TopoNodeObjIDModule:
			topoNode.HostCount = moduleHostCount[topoNode.InstID]
		default:
			if hostCount, ok := customHostCount[topoNode]; ok {
				topoNode.HostCount = hostCount
				break
			}
			topoNode.HostCount += childHostCount
		}

		return topoNode.HostCount
	}

	fillNodeHostCount(topoNode)
}
