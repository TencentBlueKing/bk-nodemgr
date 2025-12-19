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
)

// SelectHostID selects host id with conditions.
func (h *handler) SelectHostID(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoHostSelectHostIDReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select host, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, err := h.backendHandler.SelectHostID(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select host")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoHostSelectHostIDResp)
	resp.ConvertHostID(hosts)

	return resp.GetData(), nil
}

// SelectInnerIP selects host inner ip with conditions.
func (h *handler) SelectInnerIP(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoHostSelectInnerIPReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select inner ip, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	items, err := h.backendHandler.SelectInnerIP(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select inner ip")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoHostSelectInnerIPResp)
	resp.ConvertInnerIP(items)

	return resp.GetData(), nil
}

// SelectInnerIPV6 selects host inner ipv6 with conditions.
func (h *handler) SelectInnerIPV6(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoHostSelectInnerIPV6Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select inner ipv6, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	items, err := h.backendHandler.SelectInnerIPV6(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select inner ipv6")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoHostSelectInnerIPV6Resp)
	resp.ConvertInnerIPV6(items)

	return resp.GetData(), nil
}

// SelectNetWorkareaIDAndInnerIP selects host networkarea id and inner ip with conditions.
func (h *handler) SelectNetWorkareaIDAndInnerIP(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoHostSelectNetWorkareaIDAndInnerIPReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select networkarea id and inner ip, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	items, err := h.backendHandler.SelectNetWorkareaIDAndInnerIP(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select networkarea id and inner ip")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoHostSelectNetWorkareaIDAndInnerIPResp)
	resp.ConvertNetWorkareaIDAndInnerIP(items)

	return resp.GetData(), nil
}

// SelectNetWorkareaIDAndInnerIPV6 selects host networkarea id and inner ipv6 with conditions.
func (h *handler) SelectNetWorkareaIDAndInnerIPV6(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.TopoHostSelectNetWorkareaIDAndInnerIPV6Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select networkarea id and inner ipv6, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	items, err := h.backendHandler.SelectNetWorkareaIDAndInnerIPV6(
		rCtx,
		req.ConvertConditionsToTypes())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to select networkarea id and inner ipv6")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	resp := new(protoApplication.TopoHostSelectNetWorkareaIDAndInnerIPV6Resp)
	resp.ConvertNetWorkareaIDAndInnerIPV6(items)

	return resp.GetData(), nil
}
