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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

// SyncCmdbHost start an operation to sync business and host from cmdb.
func (h *handler) SyncCmdbHost(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to sync cmdb host, failed to get request context. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.SyncCmdbHostReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to sync cmdb host, failed to decode request body. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	triggerID, err := h.manager.Execute(sCtx,
		&syncdata.OperInstSyncBizAndHostFromCMDB{
			TenantID: ctx.TenantID,
		},
	)
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to start sync cmdb host operation. trigger-id(%s), err: %v",
			triggerID, err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	resp := &protoBackend.SyncCmdbHostResp_Data{
		WorkflowId: triggerID,
	}

	return resp, nil
}

// SyncCmdbNetworkArea start an operation to sync networkarea from cmdb.
func (h *handler) SyncCmdbNetworkArea(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to sync cmdb networkarea, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.SyncCmdbNetworkAreaReq)
	if err := ctx.BindJSON(req); err != nil {
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	triggerID, err := h.manager.Execute(sCtx, &syncdata.OperSyncNetworkAreaFromCMDB{
		TenantID: ctx.TenantID,
	})
	if err != nil {
		h.logger.Errorf("failed to start sync cmdb networkarea operation. trigger-id(%s), err: %v", triggerID, err)
		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	resp := &protoBackend.SyncCmdbNetworkAreaResp_Data{
		WorkflowId: triggerID,
	}

	return resp, nil
}
