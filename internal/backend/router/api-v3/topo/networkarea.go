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
	"time"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// not max limit in networkarea.
	// return all data in one request.
	maxNetworkAreaLimit = 0
)

// CreateNetworkArea creates a new network-area.
func (h *handler) CreateNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaCreateReq)
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

	// record event.
	go func() {
		if err := h.storage.CreateManyTopoEvent(context.Background(), &types.TopoEvent{
			TenantID:        ctx.TenantID,
			Type:            types.TopoEventNetworkAreaCreate,
			NetworkAreaID:   networkArea.ID,
			NetworkAreaName: networkArea.Name,
			OperateTime:     time.Now(),
			Operator:        ctx.Username,
		}); err != nil {
			h.logger.Warnf("failed to record topo event in networkarea create. networkarea-id(%d), err: %v", networkArea.ID, err)
		}
	}()

	if err := h.storage.UpsertManyNetworkArea(sCtx,
		req.ConvertNetworkAreaToTypes(ctx.TenantID, networkArea.ID)); err != nil {
		h.logger.Errorf("failed to create networkarea, failed to upsert networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	h.logger.Infof("successfully created networkarea. networkarea-id(%d)", networkArea.ID)

	resp := new(protoBackend.TopoNetworkAreaCreateResp)
	resp.ConvertNetworkAreaFromTypes(networkArea.ID)

	return resp.GetData(), nil
}

// UpdateNetworkArea updates an existing network-area.
func (h *handler) UpdateNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaUpdateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to update networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to update networkarea, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkArea := req.ConvertNetworkAreaToTypes(ctx.TenantID)
	if err := h.storage.UpdateManyNetworkArea(sCtx, networkArea); err != nil {
		h.logger.Errorf("failed to update networkarea, failed to upsert networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event.
	go func() {
		if err := h.storage.CreateManyTopoEvent(context.Background(), &types.TopoEvent{
			TenantID:        ctx.TenantID,
			Type:            types.TopoEventNetworkAreaUpdate,
			NetworkAreaID:   networkArea.ID,
			NetworkAreaName: networkArea.Name,
			OperateTime:     time.Now(),
			Operator:        ctx.Username,
		}); err != nil {
			h.logger.Warnf("failed to record topo event in networkarea update. networkarea-id(%d), err: %v", networkArea.ID, err)
		}
	}()

	h.logger.Infof("successfully updated networkarea. networkarea-id(%d)", networkArea.ID)

	resp := new(protoBackend.TopoNetworkAreaUpdateResp)
	resp.ConvertNetworkAreaFromTypes(req.GetBkNetworkareaId())

	return resp.GetData(), nil
}

// GetNetworkArea gets an existing network-area.
func (h *handler) GetNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaGetReq)
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

	resp := new(protoBackend.TopoNetworkAreaGetResp)
	resp.ConvertNetworkAreaFromTypes(networkArea)

	return resp.GetData(), nil
}

// ListNetworkArea lists network-area.
func (h *handler) ListNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaListReq)
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

	resp := new(protoBackend.TopoNetworkAreaListResp)
	resp.ConvertNetworkAreasFromTypes(num, networkAreas)

	return resp.GetData(), nil
}

// DeleteNetworkArea deletes an existing network-area.
func (h *handler) DeleteNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to delete networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to delete networkarea, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkAreaID := req.GetBkNetworkareaId()

	networkArea, err := h.storage.GetNetworkArea(sCtx, networkAreaID)
	if err != nil {
		h.logger.Errorf("failed to delete networkarea. failed to get networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	if err := h.storage.DeleteManyNetworkArea(sCtx, networkAreaID); err != nil {
		h.logger.Errorf("failed to delete networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	// record event.
	go func() {
		if err := h.storage.CreateManyTopoEvent(context.Background(), &types.TopoEvent{
			TenantID:        ctx.TenantID,
			Type:            types.TopoEventNetworkAreaDelete,
			NetworkAreaID:   networkAreaID,
			NetworkAreaName: networkArea.Name,
			OperateTime:     time.Now(),
			Operator:        ctx.Username,
		}); err != nil {
			h.logger.Warnf("failed to record topo event in networkarea delete. networkarea-id(%d), err: %v", networkArea.ID, err)
		}
	}()

	h.logger.Infof("successfully deleted networkarea. networkarea-id(%d)", networkAreaID)

	resp := new(protoBackend.TopoNetworkAreaDeleteResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkareaId())

	return resp.GetData(), nil
}
