/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb provides the CMDB mock API handler.
package cmdb

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/common"
	"github.com/gin-gonic/gin"
)

const (
	identifierPrefix = "mock_identifier_"
)

//###############################################################################
// Business API
//###############################################################################

// SearchBusiness search businesses.
func (h *handler) SearchBusiness(gCtx *gin.Context) {
	var req cmdb.SearchBusinessReq
	if err := common.BindJSON(gCtx, &req); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind request")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid request body: %v", err))

		return
	}

	// search businesses.
	businesses, err := h.store.SearchBusiness(req.Page)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to search businesses")
		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to search businesses: %v", err))

		return
	}

	respondSuccess(gCtx, &cmdb.SearchBusinessResp{
		Count: h.store.GetBusinessCount(),
		Info:  businesses,
	})
}

//###############################################################################
// Object Attribute API
//###############################################################################

// SearchObjectAttribute searches object attributes by object id.
func (h *handler) SearchObjectAttribute(gCtx *gin.Context) {
	var req cmdb.SearchObjectAttributeReq
	if err := common.BindJSON(gCtx, &req); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind request")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid request body: %v", err))

		return
	}

	if req.BKBizID != cmdb.CCNoBusinessID {
		respondError(gCtx, CodeInvalidParameter,
			fmt.Sprintf("failed to search object attribute, bk_biz_id(%d) is not supported", req.BKBizID))

		return
	}

	attrs, err := h.store.SearchObjectAttribute(req.BKObjID)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to search object attribute")
		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to search object attribute: %v", err))

		return
	}

	respondSuccess(gCtx, cmdb.SearchObjectAttributeResp(attrs))
}

//###############################################################################
// Cloud Area API
//###############################################################################

// SearchCloudArea searches cloud areas.
func (h *handler) SearchCloudArea(gCtx *gin.Context) {
	var req cmdb.SearchCloudAreaReq
	if err := common.BindJSON(gCtx, &req); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind request")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid request body: %v", err))

		return
	}

	areas, err := h.store.SearchCloudAreas(req.Page)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to search cloud areas")
		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to search cloud areas: %v", err))

		return
	}

	respondSuccess(gCtx, &cmdb.SearchCloudAreaResp{
		Count: h.store.GetCloudAreaCount(),
		Info:  areas,
	})
}

//###############################################################################
// Host API
//###############################################################################

// ListBizHosts lists business hosts.
func (h *handler) ListBizHosts(gCtx *gin.Context) {
	var req cmdb.ListBizHostsReq
	if err := common.BindJSON(gCtx, &req); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind request")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid request body: %v", err))

		return
	}

	bizID := req.BKBizID
	hosts, err := h.store.ListBizHosts(bizID, req.Page)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list biz hosts")
		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to list biz hosts: %v", err))

		return
	}

	respondSuccess(gCtx, &cmdb.ListBizHostsResp{
		Count: h.store.GetBizHostCount(bizID),
		Info:  hosts,
	})
}

// AddHostToBusinessIdle adds hosts to business idle.
func (h *handler) AddHostToBusinessIdle(gCtx *gin.Context) {
	var req cmdb.AddHostToBusinessIdleReq
	if err := common.BindJSON(gCtx, &req); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind request")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid request body: %v", err))

		return
	}

	hostIDs, err := h.store.AddHostToBusinessIdle(req.BKBizID, req.BKHostList)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to add host to business idle")
		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to add host to business idle: %v", err))

		return
	}

	respondSuccess(gCtx, &cmdb.AddHostToBusinessIdleResp{
		BKHostIDs: hostIDs,
	})
}

// BindHostAgent binds host agent relation.
func (h *handler) BindHostAgent(gCtx *gin.Context) {
	var req cmdb.BindHostAgentReq
	if err := common.BindJSON(gCtx, &req); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind request")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid request body: %v", err))

		return
	}

	if err := h.store.BindHostAgent(req.List); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind host agent")
		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to bind host agent: %v", err))

		return
	}

	respondSuccess(gCtx, cmdb.BindHostAgentResp(""))
}

//###############################################################################
// Host Identifier API
//###############################################################################

// PushHostIdentifier pushes host identifier.
//
// NOTE: This is a FAKE implementation that only simulates the API response at memory level.
// It does NOT:
//   - Query agent status from GSE
//   - Actually push identifier files to GSE
//   - Write identifier files to agent machines
//   - Poll GSE for real task results
func (h *handler) PushHostIdentifier(gCtx *gin.Context) {
	var req cmdb.PushHostIdentifierReq
	if err := common.BindJSON(gCtx, &req); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind request")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid request body: %v", err))

		return
	}

	taskID := h.store.PushHostIdentifier(req.BKHostIDs)
	resp := &cmdb.PushHostIdentifierResp{
		TaskID:    taskID,
		HostInfos: conv.SliceToSlice(req.BKHostIDs, buildHostIdentifierInfo),
	}

	respondSuccess(gCtx, resp)
}

func buildHostIdentifierInfo(hostID int64) struct {
	BKHostID       int64  `json:"bk_host_id"`
	Identification string `json:"identification"`
} {
	// mock identification.
	identification := fmt.Sprintf("%s%d", identifierPrefix, hostID)
	return struct {
		BKHostID       int64  `json:"bk_host_id"`
		Identification string `json:"identification"`
	}{
		BKHostID:       hostID,
		Identification: identification,
	}
}

// FindHostIdentifierPushResult finds host identifier push result.
//
// NOTE: This is a FAKE implementation that always returns all hosts as successful.
// It does NOT query real GSE task results, so:
//   - successList always contains all requested hostIDs
//   - failedList is always empty
//   - pendingList is always empty
func (h *handler) FindHostIdentifierPushResult(gCtx *gin.Context) {
	var req cmdb.FindHostIdentifierPushResultReq
	if err := common.BindJSON(gCtx, &req); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind request")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid request body: %v", err))

		return
	}

	successList := h.store.FindHostIdentifierPushResult(req.TaskID)
	respondSuccess(gCtx, &cmdb.FindHostIdentifierPushResultResp{
		SuccessList: successList,
		FailedList:  []int64{},
		PendingList: []int64{},
	})
}

//###############################################################################
// Watch API
//###############################################################################

// WatchHostResource watches host resource events.
func (h *handler) WatchHostResource(gCtx *gin.Context) {
	var req cmdb.ResourceWatchReq
	if err := common.BindJSON(gCtx, &req); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind request")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid request body: %v", err))

		return
	}

	resource := ResourceType(req.BKResource)
	if resource != ResourceTypeHost {
		logger.G.Sys().With("resource", req.BKResource).Error("unsupported resource type")
		respondError(gCtx, CodeInvalidParameter, "resource parameter mismatch")

		return
	}

	events, err := h.store.WatchResource(resource, req.BKCursor)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to watch resource")
		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to watch resource: %v", err))

		return
	}

	respondSuccess(gCtx, &cmdb.ResourceWatchResp{
		BKWatched: len(events) > 0,
		BKEvents:  convertWatchEventsToCMDBFormat(events),
	})
}

// WatchHostRelationResource watches host relation resource events.
func (h *handler) WatchHostRelationResource(gCtx *gin.Context) {
	var req cmdb.ResourceWatchReq
	if err := common.BindJSON(gCtx, &req); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind request")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid request body: %v", err))

		return
	}

	resource := ResourceType(req.BKResource)
	if resource != ResourceTypeHostRelation {
		logger.G.Sys().With("resource", req.BKResource).Error("unsupported resource type")
		respondError(gCtx, CodeInvalidParameter, "resource parameter mismatch")

		return
	}

	events, err := h.store.WatchResource(resource, req.BKCursor)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to watch resource")
		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to watch resource: %v", err))

		return
	}

	respondSuccess(gCtx, &cmdb.ResourceWatchResp{
		BKWatched: len(events) > 0,
		BKEvents:  convertWatchEventsToCMDBFormat(events),
	})
}

func convertWatchEventsToCMDBFormat(events []*watchEvent) []*map[string]any {
	result := make([]*map[string]any, len(events))
	for idx, event := range events {
		m := map[string]any{
			WatchFieldCursor:    event.Cursor,
			WatchFieldResource:  string(event.Resource),
			WatchFieldEventType: string(event.EventType),
			WatchFieldDetail:    event.Detail,
		}
		result[idx] = &m
	}

	return result
}
