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

// Package sync ...
package sync

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// SyncSharedReleases starts a task to sync shared releases.
func (h *handler) SyncSharedReleases(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.SyncSharedReleasesReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("sync shared releases decode request body failed")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	triggerID, err := h.manager.LaunchSyncSharedReleases(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("trigger-id", triggerID).Error("start sync shared releases task failed")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}
	logger.G.Biz(rCtx).With("trigger-id", triggerID).Info("started sync shared releases task")

	resp := &protoBackend.SyncSharedReleasesResp_Data{
		TriggerId: triggerID,
	}

	return resp, nil
}
