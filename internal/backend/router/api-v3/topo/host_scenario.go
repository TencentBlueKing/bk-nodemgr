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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	hostFieldSelectionMaxPageSize = 5000
	hostFieldSelectionTimeout     = 1 * time.Minute
)

// SelectHostID select host id.
func (h *handler) SelectHostID(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoHostSelectHostIDReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select host id, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	executor := pageexecutor.NewPageExecutor[*types.Host](hostFieldSelectionMaxPageSize, hostFieldSelectionTimeout)
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		hosts, err := h.storage.ListHostWithFieldsWithoutCount(
			nCtx,
			p,
			&types.HostFieldSelection{
				HostID: true},
			req.ConvertConditionsToTypes())

		return hosts, err
	}

	pageResult, err := executor.Execute(rCtx, types.UnlimitedPage(), fn)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host with fields using page executor")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoHostSelectHostIDResp)
	resp.ConvertHostIDFromTypes(pageResult.Items)

	return resp.GetData(), nil
}

// SelectInnerIP selects host inner ip with conditions.
func (h *handler) SelectInnerIP(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoHostSelectInnerIPReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select inner ip, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	executor := pageexecutor.NewPageExecutor[*types.Host](hostFieldSelectionMaxPageSize, hostFieldSelectionTimeout)
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		hosts, err := h.storage.ListHostWithFieldsWithoutCount(
			nCtx,
			p,
			&types.HostFieldSelection{
				InnerIPList: true},
			req.ConvertConditionsToTypes())

		return hosts, err
	}

	pageResult, err := executor.Execute(rCtx, types.UnlimitedPage(), fn)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host with fields using page executor")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoHostSelectInnerIPResp)
	resp.ConvertInnerIPFromTypes(pageResult.Items)

	return resp.GetData(), nil
}

// SelectInnerIPV6 selects host inner ipv6 with conditions.
func (h *handler) SelectInnerIPV6(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoHostSelectInnerIPV6Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select inner ipv6, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	executor := pageexecutor.NewPageExecutor[*types.Host](hostFieldSelectionMaxPageSize, hostFieldSelectionTimeout)
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		hosts, err := h.storage.ListHostWithFieldsWithoutCount(
			nCtx,
			p,
			&types.HostFieldSelection{
				InnerIPV6List: true},
			req.ConvertConditionsToTypes())

		return hosts, err
	}

	pageResult, err := executor.Execute(rCtx, types.UnlimitedPage(), fn)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host with fields using page executor")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoHostSelectInnerIPV6Resp)
	resp.ConvertInnerIPV6FromTypes(pageResult.Items)

	return resp.GetData(), nil
}

// SelectNetWorkareaIDAndInnerIP selects host networkarea id and inner ip with conditions.
func (h *handler) SelectNetWorkareaIDAndInnerIP(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select networkarea id and inner ip, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	executor := pageexecutor.NewPageExecutor[*types.Host](hostFieldSelectionMaxPageSize, hostFieldSelectionTimeout)
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		hosts, err := h.storage.ListHostWithFieldsWithoutCount(
			nCtx,
			p,
			&types.HostFieldSelection{
				NetworkAreaID: true,
				InnerIPList:   true},
			req.ConvertConditionsToTypes())

		return hosts, err
	}

	pageResult, err := executor.Execute(rCtx, types.UnlimitedPage(), fn)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host with fields using page executor")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPResp)
	resp.ConvertNetWorkareaIDAndInnerIPFromTypes(pageResult.Items)

	return resp.GetData(), nil
}

// SelectNetWorkareaIDAndInnerIPV6 selects host networkarea id and inner ipv6 with conditions.
func (h *handler) SelectNetWorkareaIDAndInnerIPV6(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPV6Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select networkarea id and inner ipv6, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	executor := pageexecutor.NewPageExecutor[*types.Host](hostFieldSelectionMaxPageSize, hostFieldSelectionTimeout)
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.Host, error) {
		hosts, err := h.storage.ListHostWithFieldsWithoutCount(
			nCtx,
			p,
			&types.HostFieldSelection{
				NetworkAreaID: true,
				InnerIPV6List: true},
			req.ConvertConditionsToTypes())

		return hosts, err
	}

	pageResult, err := executor.Execute(rCtx, types.UnlimitedPage(), fn)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list host with fields using page executor")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoHostSelectNetWorkareaIDAndInnerIPV6Resp)
	resp.ConvertNetWorkareaIDAndInnerIPV6FromTypes(pageResult.Items)

	return resp.GetData(), nil
}
