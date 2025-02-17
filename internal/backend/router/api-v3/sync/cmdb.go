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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef"
	types "github.com/TencentBlueKing/bk-nodemgr/internal/backend/types/router/api-v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"github.com/google/uuid"
)

// SyncCmdbHost ...
func (h *handler) SyncCmdbHost(ctx *rest.Context) (interface{}, error) {
	req := new(types.SyncCmdbHostReq)
	if err := ctx.BindJSON(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	triggerID := uuid.New().String()
	tenantID := ctx.TenantID

	err := h.manager.ExecuteOperation(workflowdef.OperDefNameSyncBizAndHost, triggerID, &operengine.OperInstParam{
		Timeout: 1 * time.Minute,
		InitContent: map[string]any{
			workflowdef.CKeyTenantID: tenantID,
		},
	})
	if err != nil {
		h.logger.Errorf("failed to start sync cmdb host operation, err: %v", err)
		return nil, err
	}

	resp := &types.SyncCmdbHostResp{
		TriggerID: triggerID,
	}

	return resp, nil
}
