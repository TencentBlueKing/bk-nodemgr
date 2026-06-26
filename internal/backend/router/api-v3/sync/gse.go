/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package sync ...
package sync

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// SyncAgentState start an operation to sync agent state from gse.
func (h *handler) SyncAgentState(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.SyncAgentStateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("sync agent state decode request body failed")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	triggerID, err := h.manager.LaunchSyncAgentState(rCtx, req.GetHostIds()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("trigger-id", triggerID).Error("start sync cmdb host operation")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := &protoBackend.SyncAgentStateResp_Data{
		TriggerId: triggerID,
	}

	return resp, nil
}

// SyncAllAgentState start an operation to sync all agent state from gse.
func (h *handler) SyncAllAgentState(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.SyncAllAgentStateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("sync all agent state decode request body failed")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	triggerID, err := h.manager.LaunchSyncAllAgentState(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("trigger-id", triggerID).Error("start sync all agent state operation")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := &protoBackend.SyncAllAgentStateResp_Data{
		TriggerId: triggerID,
	}

	return resp, nil
}

// SyncAgentInfo start an operation to sync agent info from gse.
func (h *handler) SyncAgentInfo(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.SyncAgentInfoReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("sync agent info decode request body failed")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	triggerID, err := h.manager.LaunchSyncAgentInfo(rCtx, req.GetHostIds()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("trigger-id", triggerID).Error("start sync cmdb host operation")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := &protoBackend.SyncAgentInfoResp_Data{
		TriggerId: triggerID,
	}

	return resp, nil
}

// SyncCorrectAgentID start an operation to correct agent id.
func (h *handler) SyncCorrectAgentID(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.SyncCorrectAgentIDReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("sync correct agent id decode request body failed")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	triggerID, err := h.manager.LaunchSyncCorrectAgentID(rCtx, req.GetHostIds()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("trigger-id", triggerID).Error("start sync correct agent id operation")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := &protoBackend.SyncCorrectAgentIDResp_Data{
		TriggerId: triggerID,
	}

	return resp, nil
}

// SyncAliveHostAgentInfo start an operation to sync alive host agent info from gse.
func (h *handler) SyncAliveHostAgentInfo(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.SyncAliveHostAgentInfoReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("sync alive host agent info decode request body failed")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	triggerID, err := h.manager.LaunchSyncAliveHostAgentInfo(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("trigger-id", triggerID).Error("start sync alive host agent info operation")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := &protoBackend.SyncAliveHostAgentInfoResp_Data{
		TriggerId: triggerID,
	}

	return resp, nil
}

// SyncAlivePluginProcessInfo start an operation to sync alive host plugin process info from gse.
func (h *handler) SyncAlivePluginProcessInfo(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.SyncAlivePluginProcessInfoReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("sync alive plugin process info decode request body failed")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	triggerID, err := h.manager.LaunchSyncAlivePluginProcessInfo(rCtx, req.GetHostIds()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("trigger-id", triggerID).Error("start sync alive plugin process info operation")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := &protoBackend.SyncAlivePluginProcessInfoResp_Data{
		TriggerId: triggerID,
	}

	return resp, nil
}

// SyncAllAlivePluginProcessInfo start an operation to sync all alive host plugin process info from gse.
func (h *handler) SyncAllAlivePluginProcessInfo(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.SyncAllAlivePluginProcessInfoReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("sync all alive plugin process info decode request body failed")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	triggerID, err := h.manager.LaunchSyncAllAlivePluginProcessInfo(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("trigger-id", triggerID).Error("start sync all alive plugin process info operation")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := &protoBackend.SyncAllAlivePluginProcessInfoResp_Data{
		TriggerId: triggerID,
	}

	return resp, nil
}
