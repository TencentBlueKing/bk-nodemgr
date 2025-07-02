/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package agent ...
package agent

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// AgentOperationRetry retry the operation of node agent.
func (h *handler) AgentOperationRetry(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to retry operation, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.NodeAgentOperationRetryReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to retry operation, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if req.Validate() != nil {
		h.logger.ErrorCtxf(sCtx, "failed to retry operation, operation ID is required")
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	instanceIDs, err := h.manager.RetryOperationNode(sCtx, manager.RetryOperationNodeParam{
		WorkflowID:   req.GetWorkflowId(),
		RetryMod:     types.NodeOperationRetryMode(req.GetRetryMod()),
		OperationIDs: req.GetOperationId(),
	})
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to retry operation,: %v", err)
		return nil, errf.ErrWrap(errf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentOperationRetryResp)
	resp.ConvertOperInstanceID(instanceIDs)

	h.logger.InfoCtxf(sCtx, "launched to retry operation, instance: %v", instanceIDs)

	return resp.GetData(), nil
}
