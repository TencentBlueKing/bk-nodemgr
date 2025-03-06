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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
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

	if err := validateTopoNetworkAreaCreateReq(req); err != nil {
		h.logger.Errorf("failed to create networkarea, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to create networkarea, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkArea, err := h.cmdbHandler.CreateNetworkArea(sCtx, req.GetBkNetworkareaName())
	if err != nil {
		h.logger.Errorf("failed to create networkarea, failed to create networkarea via cmdb. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	if err := h.storage.UpsertManyNetworkArea(sCtx, &types.NetworkArea{
		TenantID:    ctx.TenantID,
		ID:          networkArea.ID,
		Name:        networkArea.Name,
		CloudVendor: req.GetBkCloudVendor(),
	}); err != nil {
		h.logger.Errorf("failed to create networkarea, failed to upsert networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := &proto.TopoNetworkAreaCreateResp_Data{BkNetworkareaId: new(int64)}
	*resp.BkNetworkareaId = networkArea.ID

	return resp, nil
}

// UpdateNetworkArea updates an existing network-area.
func (h *handler) UpdateNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkAreaUpdateReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to update networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := validateTopoNetworkAreaUpdateReq(req); err != nil {
		h.logger.Errorf("failed to update networkarea, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to update networkarea, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := h.storage.UpdateManyNetworkArea(sCtx, &types.NetworkArea{
		TenantID:    ctx.TenantID,
		ID:          req.GetBkNetworkareaId(),
		Name:        req.GetBkNetworkareaName(),
		CloudVendor: req.GetBkCloudVendor(),
	}); err != nil {
		h.logger.Errorf("failed to update networkarea, failed to upsert networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := &proto.TopoNetworkAreaUpdateResp_Data{BkNetworkareaId: new(int64)}
	*resp.BkNetworkareaId = req.GetBkNetworkareaId()

	return resp, nil
}

// GetNetworkArea gets an existing network-area.
func (h *handler) GetNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkAreaGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to get networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := validateTopoNetworkAreaGetReq(req); err != nil {
		h.logger.Errorf("failed to get networkarea, failed to validate request body. err: %v", err)
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

	data := newEmptyNetworkArea()
	*data.TenantId = networkArea.TenantID
	*data.BkNetworkareaId = networkArea.ID
	*data.BkNetworkareaName = networkArea.Name
	*data.BkCloudVendor = networkArea.CloudVendor

	return data, nil
}

// ListNetworkArea lists network-area.
func (h *handler) ListNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkAreaListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to list networkarea, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := validateTopoNetworkAreaListReq(req); err != nil {
		h.logger.Errorf("failed to list networkarea, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list networkarea, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	networkAreas, num, err := h.storage.ListNetworkArea(
		sCtx,
		generatePage(req.GetPage(), maxNetworkAreaLimit),
		generateNetworkAreaConditions(req))
	if err != nil {
		h.logger.Errorf("failed to list networkarea. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	items := make([]*proto.NetworkArea, len(networkAreas))
	for idx, networkArea := range networkAreas {
		item := newEmptyNetworkArea()
		*item.TenantId = networkArea.TenantID
		*item.BkNetworkareaId = networkArea.ID
		*item.BkNetworkareaName = networkArea.Name
		*item.BkCloudVendor = networkArea.CloudVendor

		items[idx] = item
	}

	resp := &proto.TopoNetworkAreaListResp_Data{
		Total: num,
		Items: items,
	}

	return resp, nil
}

// DeleteNetworkArea deletes an existing network-area.
func (h *handler) DeleteNetworkArea(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoNetworkAreaDeleteReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to delete networkarea, failed to decode request body. err: %v", err)
		return nil, err
	}

	if err := validateTopoNetworkAreaDeleteReq(req); err != nil {
		h.logger.Errorf("failed to delete networkarea, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
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

	resp := &proto.TopoNetworkAreaDeleteResp_Data{BkNetworkareaId: new(int64)}
	*resp.BkNetworkareaId = req.GetBkNetworkareaId()

	return resp, nil
}

func validateTopoNetworkAreaCreateReq(req *proto.TopoNetworkAreaCreateReq) error {
	if req.GetBkNetworkareaName() == "" {
		return errors.New("bk_networkarea_name is required")
	}

	if req.GetBkCloudVendor() == "" {
		return errors.New("bk_cloud_vendor is required")
	}

	return nil
}

func validateTopoNetworkAreaUpdateReq(req *proto.TopoNetworkAreaUpdateReq) error {
	if req.GetBkNetworkareaId() < 0 {
		return errors.New("bk_networkarea_id is invalid")
	}

	if req.GetBkNetworkareaName() == "" {
		return errors.New("bk_networkarea_name is required")
	}

	if req.GetBkCloudVendor() == "" {
		return errors.New("bk_cloud_vendor is required")
	}

	return nil
}

func validateTopoNetworkAreaGetReq(req *proto.TopoNetworkAreaGetReq) error {
	if req.GetBkNetworkareaId() < 0 {
		return errors.New("bk_networkarea_id is invalid")
	}

	return nil
}

func validateTopoNetworkAreaListReq(req *proto.TopoNetworkAreaListReq) error {
	if err := validateTopoPage(req.GetPage()); err != nil {
		return err
	}

	return nil
}

func validateTopoNetworkAreaDeleteReq(req *proto.TopoNetworkAreaDeleteReq) error {
	if req.GetBkNetworkareaId() < 0 {
		return errors.New("bk_networkarea_id is invalid")
	}

	return nil
}

func generateNetworkAreaConditions(req *proto.TopoNetworkAreaListReq) topo.NetworkAreaCondition {
	// exact conditions.
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		conditions := topo.NetworkAreaCondition{
			Type: types.ConditionTypeExactInclude,
		}
		conditions.Exact = &topo.NetworkAreaExactFields{
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			CloudVendor:   exactCond.GetBkCloudVendor(),
		}

		return conditions
	}

	// fuzzy conditions.
	if fuzzyCond := req.GetFuzzyIncludeConditions(); fuzzyCond != nil {
		conditions := topo.NetworkAreaCondition{
			Type: types.ConditionTypeFuzzyInclude,
		}
		conditions.Fuzzy = &topo.NetworkAreaFuzzyFields{
			NetworkAreaName: fuzzyCond.GetBkNetworkareaName(),
		}

		return conditions
	}

	// default empty conditions.
	return topo.NetworkAreaCondition{
		Type: types.ConditionTypeExactInclude,
	}
}

func newEmptyNetworkArea() *proto.NetworkArea {
	return &proto.NetworkArea{
		TenantId:          new(string),
		BkNetworkareaId:   new(int64),
		BkNetworkareaName: new(string),
		BkCloudVendor:     new(string),
	}
}
