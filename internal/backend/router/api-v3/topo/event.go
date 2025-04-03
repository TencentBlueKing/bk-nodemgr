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
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	maxEventLimit = 1000
)

// ListEvent lists events with page and conditions.
func (h *handler) ListEvent(ctx *rest.Context) (interface{}, error) {
	req := new(protoBackend.TopoEventListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to list event, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list event, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountTopoEvent(
			sCtx,
			req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.Errorf("failed to list event. failed to count event. err: %v", err)
			return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.TopoEventListResp)
		resp.ConvertTopoEventsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	events, num, err := h.storage.ListTopoEvent(
		sCtx,
		req.ConvertPageToTypes(maxEventLimit),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.Errorf("failed to list event. err: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoEventListResp)
	resp.ConvertTopoEventsFromTypes(num, events)

	return resp.GetData(), nil
}

// DistinctEvent distincts events with conditions.
func (h *handler) DistinctEvent(ctx *rest.Context) (interface{}, error) {
	req := new(protoBackend.TopoEventDistinctReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to distinct topoevent, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to distinct topoevent, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	result, err := h.storage.DistinctTopoEvent(
		sCtx,
		types.NewTopoEventDistinctRequestAllSet(),
		req.ConvertConditionsToTypes())
	if err != nil {
		h.logger.Errorf("failed to distinct topoevent. failed to distinct host fields: %v", err)
		return nil, errf.ErrWrap(errf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.TopoEventDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
