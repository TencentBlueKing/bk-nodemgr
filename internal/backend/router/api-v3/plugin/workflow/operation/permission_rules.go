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

package operation

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (h *handler) authorizedPluginHistoryView(rCtx restserver.IContext, workflowID string) error {
	pluginWorkflow, err := h.daoPluginWorkflow.GetPluginWorkflow(rCtx, workflowID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check workflow history view permission, failed to get the target plugin workflow")
		return resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resources := authRouter.BuildBizResources(pluginWorkflow.BizIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionPluginHistoryView, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to check workflow history view permission, permission denied")
		return resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	return nil
}

func (h *handler) authorizedPluginOperationHistoryView(rCtx restserver.IContext, operationIDs []string) error {
	operations, _, err := h.storageWorkflow.ListOperationByOperationID(rCtx, operationIDs...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check operation history view permission, failed to list operations")
		return resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	triggerByOperationID := make(map[string]string, len(operations))
	for _, op := range operations {
		triggerByOperationID[op.OperationID] = op.TriggerID
	}
	for _, operationID := range operationIDs {
		if _, exists := triggerByOperationID[operationID]; !exists {
			logger.G.Biz(rCtx).With("operation-id", operationID).Error("failed to check operation history view permission, operation not found")
			return resterrf.ErrWrap(resterrf.DBExecCmdFailed, fmt.Errorf("operation not found for operation_id: %s", operationID))
		}
	}

	// Workflow collections are tenant-scoped, unlike the operation collection.
	workflows, _, err := h.daoPluginWorkflow.ListPluginWorkflow(rCtx, types.UnlimitedPage())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check operation history view permission, failed to list workflows")
		return resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	workflowByTriggerID := make(map[string]string, len(workflows))
	for _, workflow := range workflows {
		if workflow.TriggerID == "" {
			continue
		}
		workflowByTriggerID[workflow.TriggerID] = workflow.WorkflowID
	}

	checkedWorkflowIDs := make(map[string]struct{}, len(operationIDs))
	for _, operationID := range operationIDs {
		triggerID := triggerByOperationID[operationID]
		workflowID, exists := workflowByTriggerID[triggerID]
		if !exists {
			logger.G.Biz(rCtx).With("trigger-id", triggerID).Error("failed to check operation history view permission, workflow not found")
			return resterrf.ErrWrap(resterrf.DBExecCmdFailed, fmt.Errorf("workflow not found for trigger_id: %s", triggerID))
		}
		if _, checked := checkedWorkflowIDs[workflowID]; checked {
			continue
		}
		if err := h.authorizedPluginHistoryView(rCtx, workflowID); err != nil {
			return err
		}
		checkedWorkflowIDs[workflowID] = struct{}{}
	}

	return nil
}

func (h *handler) authorizedPluginOperate(rCtx restserver.IContext, workflowID string) error {
	pluginWorkflow, err := h.daoPluginWorkflow.GetPluginWorkflow(rCtx, workflowID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check workflow operate permission, failed to get the target plugin workflow")
		return resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	bizIDs := pluginWorkflow.BizIDs
	resources := authRouter.BuildBizResources(bizIDs...)

	if authErr := h.authorizer.Check(rCtx, auth.ActionPluginOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to check workflow operate permission, permission denied")
		return resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	return nil
}
