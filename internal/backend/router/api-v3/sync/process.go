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

package sync

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// CleanOrphanProcess start an operation to clean the processes whose host no longer exists.
func (h *handler) CleanOrphanProcess(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.CleanOrphanProcessReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("clean orphan process decode request body failed")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	triggerID, err := h.manager.LaunchCleanOrphanProcess(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("trigger-id", triggerID).Error("start clean orphan process task failed")
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	logger.G.Biz(rCtx).With("trigger-id", triggerID).Info("started clean orphan process task")

	resp := &protoBackend.CleanOrphanProcessResp_Data{
		TriggerId: triggerID,
	}

	return resp, nil
}
