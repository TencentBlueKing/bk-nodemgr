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
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListHost lists hosts with page and conditions.
func (h *handler) ListHost(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoHostListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition := req.ConvertConditionsToTypes()
	var bizIDs []int64
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		bizIDs = exactCond.GetBkBizId()
	}
	narrowedBizIDs, scopeIsAny, authErr := h.narrowAuthorizedBizIDsForHostList(rCtx, bizIDs, condition)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list host, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}
	condition = narrowHostConditionByBiz(condition, narrowedBizIDs, scopeIsAny)

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountHost(
			rCtx,
			condition)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list host. failed to count host")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.TopoHostListResp)
		resp.ConvertHostsFromTypes(num, nil, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, num, err := h.storage.ListHostOrderByUpdateTime(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// get credit status.
	creditIDMap := make(map[string]struct{})
	for _, host := range hosts {
		if host.Dynamic.LoginCreditID != "" {
			creditIDMap[host.Dynamic.LoginCreditID] = struct{}{}
		}
	}
	hostCredits, err := h.storageHostCredit.CheckHostCreditValid(rCtx, conv.MapKeyToSlice(creditIDMap)...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host. failed to get host credit status")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoHostListResp)
	resp.ConvertHostsFromTypes(num, hosts, hostCredits)

	return resp.GetData(), nil
}

// DistinctHost get distinct host fields.
func (h *handler) DistinctHost(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoHostDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct host, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.storage.DistinctHost(
		rCtx,
		types.NewHostDistinctRequestAllSet(),
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct host. failed to distinct host fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoHostDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// GetHostDistributionByNodeRole get host distribution by node role.
func (h *handler) GetHostDistributionByNodeRole(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoGetHostDistributionByNodeRoleReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get host distribution by node role, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.storage.GetHostDistributionByNodeRole(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			Error("failed to get host distribution by node role. failed to get host distribution by node role fields: %v", err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoGetHostDistributionByNodeRoleResp)

	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// GetHostDistributionByNetworkAreaID get host distribution by network area id.
func (h *handler) GetHostDistributionByNetworkAreaID(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoGetHostDistributionByNetworkAreaIDReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get host distribution by network area id, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.storage.GetHostDistributionByNetworkAreaID(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			Error("failed to get host distribution by network area id. failed to get host distribution by network area id fields: %v", err)

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoGetHostDistributionByNetworkAreaIDResp)

	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
