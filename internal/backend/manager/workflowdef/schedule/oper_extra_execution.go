/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package schedule

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperExtraExecutionName defines the operation instance extra execution name.
	OperExtraExecutionName = "schedule_operation_extra_execution"
)

// NewOperationExtraExecution creates a new operation extra execution.
func NewOperationExtraExecution(cap *Capability) operation.ExtraExecution {
	return &extraExecution{
		workflowStg: cap.StorageWorkflow,
	}
}

// ExtraExecutionParam defines the action's param.
type ExtraExecutionParam struct {
	WorkflowID string `json:"workflow_id"`
	TenantID   string `json:"tenant_id"`
	Operator   string `json:"operator"`
}

type extraExecution struct {
	workflowStg workflow.IStorage
}

// Name returns the name of the action.
func (exec *extraExecution) Name() string {
	return OperExtraExecutionName
}

// Do this func define what the action will do.
func (exec *extraExecution) Do(nCtx contextx.IContext, instance *operation.InstanceBriefData) error {
	param := new(ExtraExecutionParam)
	err := conv.MapToStruct(instance.Metadata.InitContent, param)
	if err != nil {
		return err
	}

	switch instance.Lifecycle.State {
	case operation.StateLaunched:
		if err := exec.preprocess(nCtx, instance, param); err != nil {
			logger.G.Sys().
				WithErr(err).
				With("operation", instance.Metadata.OperationDefName, "workflow-id", param.WorkflowID).
				Error("failed to do preprocess for scheduled workflow")

			instance.Lifecycle.End(action.StateFailed)

			return err
		}

		return nil

	case operation.StateSuccess, operation.StateFailed, operation.StateTimeout, operation.StateTerminated:
		if err := exec.postprocess(nCtx, instance, param); err != nil {
			logger.G.Sys().
				WithErr(err).
				With("operation", instance.Metadata.OperationDefName, "workflow-id", param.WorkflowID).
				Error("failed to do postprocess for scheduled workflow")

			instance.Lifecycle.End(action.StateFailed)

			return err
		}

		return nil

	default:
		logger.G.Sys().
			With("operation", instance.Metadata.OperationDefName, "workflow-id", param.WorkflowID).
			Error("unexpected operation instance state: %s", instance.Lifecycle.State)

		instance.Lifecycle.End(action.StateFailed)

		return fmt.Errorf("unexpected operation instance state: %s", instance.Lifecycle.State)
	}
}

func (exec *extraExecution) preprocess(nCtx contextx.IContext, instance *operation.InstanceBriefData, param *ExtraExecutionParam) error {
	sw, err := exec.workflowStg.GetScheduledWorkflow(nCtx, param.WorkflowID)
	if err != nil {
		logger.G.Sys().
			WithErr(err).
			With("operation", instance.Metadata.OperationDefName, "workflow-id", param.WorkflowID).
			Error("failed to get scheduled workflow")

		return err
	}

	value, ok := sw.PrivateData[managedOperationIDKey]
	if !ok {
		return nil
	}
	operationIDs, ok := value.([]string)
	if !ok {
		return nil
	}

	operInstList, _, err := exec.workflowStg.ListOperInstanceBriefWithoutActionInstByOperationID(nCtx, types.UnlimitedPage(), operationIDs...)
	if err != nil {
		logger.G.Sys().
			WithErr(err).
			With("operation", instance.Metadata.OperationDefName, "workflow-id", param.WorkflowID).
			Error("failed to list operation instance by operation id")

		return err
	}

	for _, operInst := range operInstList {
		if operation.CheckStateFinished(operInst.Lifecycle.State) {
			continue
		}

		return fmt.Errorf("operation instance(%s) state is %s, retry later",
			operInst.Metadata.OperationInstanceID, operInst.Lifecycle.State)
	}

	return nil
}

func (exec *extraExecution) postprocess(nCtx contextx.IContext, instance *operation.InstanceBriefData, param *ExtraExecutionParam) error {
	// get all managed operations.
	parentOperationIDs := []string{instance.Metadata.OperationID}
	managedOperationIDMap := make(map[string]struct{}, 0)
	for len(parentOperationIDs) > 0 {
		operations, _, err := exec.workflowStg.ListOperationByParentOperationID(nCtx, types.UnlimitedPage(), parentOperationIDs...)
		if err != nil {
			logger.G.Sys().
				WithErr(err).
				With("operation", instance.Metadata.OperationDefName, "parent-operation-ids", parentOperationIDs).
				Error("failed to list operations by parent operation id")

			return err
		}

		parentOperationIDs = make([]string, 0)
		for _, operation := range operations {
			parentOperationIDs = append(parentOperationIDs, operation.OperationID)
			managedOperationIDMap[operation.OperationID] = struct{}{}
		}
	}

	// update managed operation ids into private data.
	if err := exec.workflowStg.UpdateScheduledWorkflowPrivateData(
		nCtx,
		param.WorkflowID,
		map[string]any{managedOperationIDKey: conv.MapKeyToSlice(managedOperationIDMap)},
	); err != nil {
		logger.G.Sys().
			WithErr(err).
			With("operation", instance.Metadata.OperationDefName, "workflow-id", param.WorkflowID).
			Error("failed to update managed operation ids into private data")

		return err
	}

	return nil
}

const (
	managedOperationIDKey = "managed-operation-ids"
)
