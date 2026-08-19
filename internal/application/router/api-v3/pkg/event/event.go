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

// Package event provides the api for event.
package event

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	backendHandler backend.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/event"),
		backendHandler: capability.BackendHandler,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	// package event apis.
	h.rg.POST("/list", restserver.Handler(h.ListPackageEvent))
	h.rg.POST("/distinct", restserver.Handler(h.DistinctPackageEvent))
}

// ListPackageEvent lists events with page and conditions.
func (h *handler) ListPackageEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageEventListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list package event, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list package event, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	// only count.
	if req.GetOnlyCount() {
		num, err := h.backendHandler.CountPackageEvent(
			rCtx,
			conditions)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list package event. failed to count event")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoApplication.PackageEventListResp)
		resp.ConvertPackageEventsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list package event, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	events, num, err := h.backendHandler.ListPackageEvent(rCtx, page, conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list package event")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.PackageEventListResp)
	resp.ConvertPackageEventsFromTypes(num, events)

	return resp.GetData(), nil
}

// DistinctEvent distincts events with conditions.
// nolint: dupl
func (h *handler) DistinctPackageEvent(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoApplication.PackageEventDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct package event, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	conditions, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct package event, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	result, err := h.backendHandler.DistinctPackageEvent(
		rCtx,
		conditions)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct package event. failed to distinct package fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoApplication.PackageEventDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
