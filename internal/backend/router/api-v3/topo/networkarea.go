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
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// not max limit in networkarea.
	// return all data in one request.
	maxNetworkAreaLimit = 0
)

// CreateNetworkArea creates a new network-area.
func (h *handler) CreateNetworkArea(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaCreateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create networkarea, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkArea, err := h.cmdbHandler.CreateNetworkArea(ctx, req.GetBkNetworkareaName(), req.GetCloudVendor())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create networkarea, failed to create networkarea via cmdb. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	// record event.
	go func() {
		if err := h.storage.CreateManyTopoEvent(context.Background(),
			&types.TopoEvent{
				TenantID:        ctx.TenantID,
				Type:            types.TopoEventNetworkAreaCreate,
				NetworkAreaID:   networkArea.ID,
				NetworkAreaName: networkArea.Name,
				OperateTime:     time.Now(),
				Operator:        ctx.LoginName,
			}); err != nil {
			h.logger.Warnf("failed to record topo event in networkarea create. networkarea-id(%d), err: %v",
				networkArea.ID, err)
		}
	}()

	if err := h.storage.UpsertManyNetworkArea(ctx,
		req.ConvertNetworkAreaToTypes(ctx.TenantID, networkArea.ID)); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to create networkarea, failed to upsert networkarea. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	h.logger.Infof("created networkarea. networkarea-id(%d)", networkArea.ID)

	resp := new(protoBackend.TopoNetworkAreaCreateResp)
	resp.ConvertNetworkAreaFromTypes(networkArea.ID)

	return resp.GetData(), nil
}

// UpdateNetworkArea updates an existing network-area.
func (h *handler) UpdateNetworkArea(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaUpdateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to update networkarea, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkArea := req.ConvertNetworkAreaToTypes(ctx.TenantID)
	if err := h.storage.UpdateManyNetworkArea(ctx, networkArea); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to update networkarea, failed to upsert networkarea. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record event.
	go func() {
		if err := h.storage.CreateManyTopoEvent(context.Background(),
			&types.TopoEvent{
				TenantID:        ctx.TenantID,
				Type:            types.TopoEventNetworkAreaUpdate,
				NetworkAreaID:   networkArea.ID,
				NetworkAreaName: networkArea.Name,
				OperateTime:     time.Now(),
				Operator:        ctx.LoginName,
			}); err != nil {
			h.logger.Warnf("failed to record topo event in networkarea update. networkarea-id(%d), err: %v",
				networkArea.ID, err)
		}
	}()

	h.logger.InfoCtxf(ctx, "updated networkarea. networkarea-id(%d)", networkArea.ID)

	resp := new(protoBackend.TopoNetworkAreaUpdateResp)
	resp.ConvertNetworkAreaFromTypes(req.GetBkNetworkareaId())

	return resp.GetData(), nil
}

// GetNetworkArea gets an existing network-area.
func (h *handler) GetNetworkArea(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get networkarea, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkArea, err := h.storage.GetNetworkArea(ctx, req.GetBkNetworkareaId())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get networkarea. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoNetworkAreaGetResp)
	resp.ConvertNetworkAreaFromTypes(networkArea)

	return resp.GetData(), nil
}

// ListNetworkArea lists network-area.
func (h *handler) ListNetworkArea(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list networkarea, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkAreas, num, err := h.storage.ListNetworkArea(
		ctx,
		req.ConvertPageToTypes(maxNetworkAreaLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list networkarea. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoNetworkAreaListResp)
	resp.ConvertNetworkAreasFromTypes(num, networkAreas)

	return resp.GetData(), nil
}

// DeleteNetworkArea deletes an existing network-area.
func (h *handler) DeleteNetworkArea(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete networkarea, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkAreaID := req.GetBkNetworkareaId()

	networkArea, err := h.storage.GetNetworkArea(ctx, networkAreaID)
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete networkarea. failed to get networkarea. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := h.storage.DeleteManyNetworkArea(ctx, networkAreaID); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to delete networkarea. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record event.
	go func() {
		if err := h.storage.CreateManyTopoEvent(context.Background(),
			&types.TopoEvent{
				TenantID:        ctx.TenantID,
				Type:            types.TopoEventNetworkAreaDelete,
				NetworkAreaID:   networkAreaID,
				NetworkAreaName: networkArea.Name,
				OperateTime:     time.Now(),
				Operator:        ctx.LoginName,
			}); err != nil {
			h.logger.Warnf("failed to record topo event in networkarea delete. networkarea-id(%d), err: %v", networkArea.ID, err)
		}
	}()

	h.logger.InfoCtxf(ctx, "deleted networkarea. networkarea-id(%d)", networkAreaID)

	resp := new(protoBackend.TopoNetworkAreaDeleteResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkareaId())

	return resp.GetData(), nil
}
