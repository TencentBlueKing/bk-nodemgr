/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pkg

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	maxEventLimit = 1000
)

// ListPackageEvent lists events with page and conditions.
func (h *handler) ListPackageEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageEventListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, failed to convert conditions")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// only count.
	if req.GetOnlyCount() {
		num, err := h.storage.CountPackageEvent(
			rCtx,
		)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list event, failed to count event")

			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PackageEventListResp)
		resp.ConvertPackageEventsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	events, num, err := h.storage.ListPackageEvent(
		rCtx,
		req.ConvertPageToTypes(maxEventLimit),
		conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list event")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageEventListResp)
	resp.ConvertPackageEventsFromTypes(num, events)

	return resp.GetData(), nil
}

// DistinctPackageEvent distincts events with conditions.
// nolint: dupl
func (h *handler) DistinctPackageEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageEventDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct topoevent, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct topoevent, failed to convert conditions")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.storage.DistinctPackageEvent(
		rCtx,
		types.NewPackageEventDistinctRequestAllSet(),
		conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct topoevent, failed to distinct host fields")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageEventDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
