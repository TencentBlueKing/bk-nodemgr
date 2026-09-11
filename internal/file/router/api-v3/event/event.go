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

// Package event provides package event query routes.
package event

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg      *gin.RouterGroup
	manager manager.IManager
}

func newHandler(rg *gin.RouterGroup, opt *options.Capability) *handler {
	return &handler{
		rg:      rg.Group("/event"),
		manager: opt.Manager,
	}
}

// Load enables package event query routes.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListPackageEvent))
	h.rg.POST("/distinct", restserver.Handler(h.DistinctPackageEvent))
}

// ListPackageEvent lists package events with page and conditions.
func (h *handler) ListPackageEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PackageEventListReq)
	if err := rCtx.BindJSON(req); err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if req.GetOnlyCount() {
		num, err := h.manager.CountPackageEvent(rCtx, conditions...)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to count package events")

			return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to count package events: %w", err))
		}
		resp := new(protoFile.PackageEventListResp)
		resp.ConvertPackageEventsFromTypes(num, nil)

		return resp.GetData(), nil
	}
	page, err := req.ConvertPageToTypes()
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	events, num, err := h.manager.ListPackageEvent(rCtx, page, conditions...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list package events")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to list package events: %w", err))
	}
	resp := new(protoFile.PackageEventListResp)
	resp.ConvertPackageEventsFromTypes(num, events)

	return resp.GetData(), nil
}

// DistinctPackageEvent distincts package event fields.
func (h *handler) DistinctPackageEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoFile.PackageEventDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	request := req.ConvertDistinctRequestToTypes()
	result, err := h.manager.DistinctPackageEvent(rCtx, request, conditions...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct package events")

		return nil, resterrf.ErrWrap(resterrf.Aborted, fmt.Errorf("failed to distinct package events: %w", err))
	}
	resp := new(protoFile.PackageEventDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
