/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package operation

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// TerminateOperation terminate node workflow operation.
func (h *handler) TerminateOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowOperationTerminateReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to terminate operation, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID := req.GetWorkflowId()

	// Check permission before terminate
	if err := h.checkWorkflowOperatePermission(rCtx, workflowID); err != nil {
		return nil, err
	}

	err := h.nodeMgrIface.TerminateNodeOperationLastInstance(rCtx, types.TerminateNodeWorkflowOperationParam{
		WorkflowID:   workflowID,
		OperationIDs: req.GetOperationIds(),
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to terminate operation")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowOperationTerminateResp)

	return resp.GetData(), nil
}
