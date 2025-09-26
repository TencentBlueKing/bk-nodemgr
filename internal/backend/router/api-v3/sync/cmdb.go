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

// SyncCmdbHost start an operation to sync business and host from cmdb.
func (h *handler) SyncCmdbHost(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.SyncCmdbHostReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to sync cmdb host, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	triggerID, err := h.manager.LaunchSyncBizAndHost(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("trigger-id", triggerID).Error("failed to start sync cmdb host operation")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := &protoBackend.SyncCmdbHostResp_Data{
		WorkflowId: triggerID,
	}

	return resp, nil
}

// SyncCmdbNetworkArea start an operation to sync networkarea from cmdb.
func (h *handler) SyncCmdbNetworkArea(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.SyncCmdbNetworkAreaReq)
	if err := rCtx.BindJSON(req); err != nil {
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	triggerID, err := h.manager.LaunchSyncNetworkArea(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("trigger-id", triggerID).Error("failed to start sync cmdb networkarea operation")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := &protoBackend.SyncCmdbNetworkAreaResp_Data{
		WorkflowId: triggerID,
	}

	return resp, nil
}
