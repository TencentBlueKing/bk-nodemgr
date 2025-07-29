/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflow ...
package workflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// WorkflowOperationRetry retry the operation of node workflow.
func (h *handler) WorkflowOperationRetry(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationRetryReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to retry operation, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to retry operation, invalid request parameters. err: %v", req.Validate())
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	instanceIDs, err := h.manager.RetryOperationNode(ctx, manager.RetryOperationNodeParam{
		WorkflowID:   req.GetWorkflowId(),
		RetryMod:     types.NodeOperationRetryMode(req.GetRetryMod()),
		OperationIDs: req.GetOperationId(),
	})
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to retry operation,err: %w", err)
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationRetryResp)
	resp.ConvertOperInstanceID(instanceIDs)

	h.logger.InfoCtxf(ctx, "launched to retry operation, instance: %v", instanceIDs)

	return resp.GetData(), nil
}
