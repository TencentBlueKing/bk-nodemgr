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
	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

const (
	maxNetworkAreaLimit = 1000
)

// CreateNetworkArea creates a new network-area.
func (h *handler) CreateNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkAreaCreateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to create networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to create networkarea, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkArea, err := h.cmdbHandler.CreateNetworkArea(sCtx, req.GetBkNetworkareaName(), req.GetBkCloudVendor())
	if err != nil {
		h.logger.Errorf("failed to create networkarea, failed to create networkarea via cmdb. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	if err := h.storage.UpsertManyNetworkArea(sCtx, req.ConvertNetworkAreaToTypes(ctx.TenantID, networkArea.ID)); err != nil {
		h.logger.Errorf("failed to create networkarea, failed to upsert networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(proto.TopoNetworkAreaCreateResp)
	resp.ConvertNetworkAreaFromTypes(networkArea.ID)

	return resp.Data, nil
}

// UpdateNetworkArea updates an existing network-area.
func (h *handler) UpdateNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkAreaUpdateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to update networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to update networkarea, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.storage.UpdateManyNetworkArea(sCtx, req.ConvertNetworkAreaToTypes(ctx.TenantID)); err != nil {
		h.logger.Errorf("failed to update networkarea, failed to upsert networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(proto.TopoNetworkAreaUpdateResp)
	resp.ConvertNetworkAreaFromTypes(req.GetBkNetworkareaId())

	return resp, nil
}

// GetNetworkArea gets an existing network-area.
func (h *handler) GetNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkAreaGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to get networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to get networkarea, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkArea, err := h.storage.GetNetworkArea(sCtx, req.GetBkNetworkareaId())
	if err != nil {
		h.logger.Errorf("failed to get networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(proto.TopoNetworkAreaGetResp)
	resp.ConvertNetworkAreaFromTypes(networkArea)

	return resp.Data, nil
}

// ListNetworkArea lists network-area.
func (h *handler) ListNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkAreaListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to list networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list networkarea, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkAreas, num, err := h.storage.ListNetworkArea(
		sCtx,
		req.ConvertPageToTypes(maxNetworkAreaLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.Errorf("failed to list networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(proto.TopoNetworkAreaListResp)
	resp.ConvertNetworkAreasFromTypes(num, networkAreas)

	return resp.Data, nil
}

// DeleteNetworkArea deletes an existing network-area.
func (h *handler) DeleteNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkAreaDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to delete networkarea, failed to decode request body. err: %v", err)
		return nil, err
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to delete networkarea, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.storage.DeleteManyNetworkArea(sCtx, req.GetBkNetworkareaId()); err != nil {
		h.logger.Errorf("failed to delete networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(proto.TopoNetworkAreaDeleteResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkareaId())

	return resp, nil
}
