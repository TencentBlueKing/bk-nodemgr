/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package topo

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// CreateNetworkArea creates a new network-area.
func (h *handler) CreateNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaCreateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create networkarea, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if authErr := h.authorizer.Check(rCtx, auth.ActionNetworkAreaCreate, nil); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to create networkarea, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	networkArea, err := h.cmdbHandler.CreateNetworkArea(rCtx, req.GetBkNetworkareaName(), req.GetCloudVendor())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create networkarea, failed to create networkarea via cmdb")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}

	// record event.
	h.recordTopoEvents(rCtx, &types.TopoEvent{
		TenantID:        rCtx.TenantID(),
		Type:            types.TopoEventNetworkAreaCreate,
		NetworkAreaID:   networkArea.ID,
		NetworkAreaName: networkArea.Name,
		OperateTime:     time.Now(),
		Operator:        rCtx.BKUsername(),
	})

	if err := h.storage.UpsertManyNetworkArea(rCtx,
		req.ConvertNetworkAreaToTypes(rCtx.TenantID(), networkArea.ID)); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to create networkarea, failed to upsert networkarea")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	logger.G.Biz(rCtx).With("networkarea-id", networkArea.ID).Info("created networkarea")

	resource := buildNetworkAreaResources(networkArea.ID)[0]
	if err := h.managerRoleGranter.GrantManagerRole(rCtx, rCtx.BKUsername(), resource); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("resource-type", resource.Type, "resource-id", resource.ID,
			"creator", rCtx.BKUsername(), "tenant-id", rCtx.TenantID()).Error("failed to grant resource creator permissions")
	}

	resp := new(protoBackend.TopoNetworkAreaCreateResp)
	resp.ConvertNetworkAreaFromTypes(networkArea.ID)

	return resp.GetData(), nil
}

// UpdateNetworkArea updates an existing network-area.
func (h *handler) UpdateNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaUpdateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update networkarea, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if authErr := h.authorizer.Check(rCtx, auth.ActionNetworkAreaEdit,
		buildNetworkAreaResources([]int64{req.GetBkNetworkareaId()}...)); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to update networkarea, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	networkArea := req.ConvertNetworkAreaToTypes(rCtx.TenantID())
	if err := h.storage.UpdateManyNetworkArea(rCtx, networkArea); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to update networkarea, failed to upsert networkarea")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordTopoEvents(rCtx, &types.TopoEvent{
		TenantID:        rCtx.TenantID(),
		Type:            types.TopoEventNetworkAreaUpdate,
		NetworkAreaID:   networkArea.ID,
		NetworkAreaName: networkArea.Name,
		OperateTime:     time.Now(),
		Operator:        rCtx.BKUsername(),
	})

	logger.G.Biz(rCtx).With("networkarea-id", networkArea.ID).Info("updated networkarea")

	resp := new(protoBackend.TopoNetworkAreaUpdateResp)
	resp.ConvertNetworkAreaFromTypes(req.GetBkNetworkareaId())

	return resp.GetData(), nil
}

// GetNetworkArea gets an existing network-area.
func (h *handler) GetNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get networkarea, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if authErr := h.authorizer.Check(rCtx, auth.ActionNetworkAreaView,
		buildNetworkAreaResources([]int64{req.GetBkNetworkareaId()}...)); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to get networkarea, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	networkArea, err := h.storage.GetNetworkArea(rCtx, req.GetBkNetworkareaId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get networkarea")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoNetworkAreaGetResp)
	resp.ConvertNetworkAreaFromTypes(networkArea)

	return resp.GetData(), nil
}

// ListNetworkArea lists network-area.
func (h *handler) ListNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkarea, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition := req.ConvertConditionsToTypes()

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkarea, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkAreas, num, err := h.storage.ListNetworkArea(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list networkarea")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoNetworkAreaListResp)
	resp.ConvertNetworkAreasFromTypes(num, networkAreas)

	return resp.GetData(), nil
}

// DeleteNetworkArea deletes an existing network-area.
func (h *handler) DeleteNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoNetworkAreaDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete networkarea, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if authErr := h.authorizer.Check(rCtx, auth.ActionNetworkAreaDelete,
		buildNetworkAreaResources([]int64{req.GetBkNetworkareaId()}...)); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to delete networkarea, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	networkAreaID := req.GetBkNetworkareaId()

	networkArea, err := h.storage.GetNetworkArea(rCtx, networkAreaID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete networkarea. failed to get networkarea")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := h.storage.DeleteManyNetworkArea(rCtx, networkAreaID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete networkarea")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record event.
	h.recordTopoEvents(rCtx, &types.TopoEvent{
		TenantID:        rCtx.TenantID(),
		Type:            types.TopoEventNetworkAreaDelete,
		NetworkAreaID:   networkAreaID,
		NetworkAreaName: networkArea.Name,
		OperateTime:     time.Now(),
		Operator:        rCtx.BKUsername(),
	})

	logger.G.Biz(rCtx).With("networkarea-id", networkArea.ID).Info("deleted networkarea")

	resp := new(protoBackend.TopoNetworkAreaDeleteResp)
	resp.ConvertNetworkUnitFromTypes(req.GetBkNetworkareaId())

	return resp.GetData(), nil
}
