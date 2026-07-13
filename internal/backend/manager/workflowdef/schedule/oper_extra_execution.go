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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperExtraExecutionName defines the operation instance extra execution name.
	OperExtraExecutionName = "schedule_operation_extra_execution"
)

// NewOperationExtraExecution creates a new operation extra execution.
func NewOperationExtraExecution(capability *Capability) operation.ExtraExecution {
	return &extraExecution{
		workflowStg: capability.StorageWorkflow,
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
			logger.G.Sys().Ctx(nCtx).
				WithErr(err).
				With("operation", instance.Metadata.OperationDefName, "workflow-id", param.WorkflowID).
				Error("failed to do preprocess for scheduled workflow")

			instance.Lifecycle.End(operation.StateFailed)

			return err
		}

		return nil

	case operation.StateSuccess, operation.StateFailed, operation.StateTimeout, operation.StateTerminated:
		if err := exec.postprocess(nCtx, instance, param); err != nil {
			logger.G.Sys().Ctx(nCtx).
				WithErr(err).
				With("operation", instance.Metadata.OperationDefName, "workflow-id", param.WorkflowID).
				Error("failed to do postprocess for scheduled workflow")

			instance.Lifecycle.End(operation.StateFailed)

			return err
		}

		return nil

	default:
		logger.G.Sys().Ctx(nCtx).
			With("operation", instance.Metadata.OperationDefName, "workflow-id", param.WorkflowID).
			Error("unexpected operation instance state: %s", instance.Lifecycle.State)

		instance.Lifecycle.End(operation.StateFailed)

		return fmt.Errorf("unexpected operation instance state: %s", instance.Lifecycle.State)
	}
}

const (
	operationListMaxPageSize = 1000
	operationListTimeout     = 30 * time.Second
)

func (exec *extraExecution) preprocess(nCtx contextx.IContext, instance *operation.InstanceBriefData, param *ExtraExecutionParam) error {
	sw, err := exec.workflowStg.GetScheduledWorkflow(nCtx, param.WorkflowID)
	if err != nil {
		logger.G.Sys().Ctx(nCtx).
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

	executor := pageexecutor.NewPageExecutor[*operation.InstanceBriefData](operationListMaxPageSize, operationListTimeout)
	fn := func(nCtx contextx.IContext, p types.Page) ([]*operation.InstanceBriefData, error) {
		operInsts, err := exec.workflowStg.ListOperInstanceBriefWithoutActionInstByOperationIDWithoutCount(nCtx, p, operationIDs...)

		return operInsts, err
	}
	result, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		logger.G.Sys().Ctx(nCtx).
			WithErr(err).
			With("operation", instance.Metadata.OperationDefName, "workflow-id", param.WorkflowID).
			Error("failed to list operation instance by operation id")

		return err
	}

	for _, operInst := range result.Items {
		if operation.CheckStateFinished(operInst.Lifecycle.State) {
			continue
		}

		return fmt.Errorf("operation instance(%s) state is %s, retry later",
			operInst.Metadata.OperationInstanceID, operInst.Lifecycle.State)
	}

	return nil
}

func (exec *extraExecution) postprocess(nCtx contextx.IContext, instance *operation.InstanceBriefData, param *ExtraExecutionParam) error {
	// get current scheduled operation instance.
	executor := pageexecutor.NewPageExecutor[string](operationListMaxPageSize, operationListTimeout)
	fn := func(nCtx contextx.IContext, p types.Page) ([]string, error) {
		return exec.workflowStg.ListOperationIDByParentOperInstID(nCtx, p, instance.Metadata.OperationInstanceID)
	}
	result, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		logger.G.Sys().Ctx(nCtx).
			WithErr(err).
			With("operation", instance.Metadata.OperationDefName, "parent-oper-inst-id", instance.Metadata.OperationInstanceID).
			Error("failed to list operation ids by parent operation instance id")

		return err
	}

	// get all managed sub operations.
	subOperationIDs := result.Items
	managedOperationIDMap := make(map[string]struct{}, 0)
	for len(subOperationIDs) > 0 {
		for _, opID := range subOperationIDs {
			managedOperationIDMap[opID] = struct{}{}
		}

		parentOperationIDs := subOperationIDs

		fn := func(nCtx contextx.IContext, p types.Page) ([]string, error) {
			return exec.workflowStg.ListOperationIDByParentOperationID(nCtx, p, parentOperationIDs...)
		}
		result, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
		if err != nil {
			logger.G.Sys().Ctx(nCtx).
				WithErr(err).
				With("operation", instance.Metadata.OperationDefName, "parent-operation-ids", parentOperationIDs).
				Error("failed to list operation ids by parent operation id")

			return err
		}
		subOperationIDs = result.Items
	}

	// update managed operation ids into private data.
	if err := exec.workflowStg.UpdateScheduledWorkflowPrivateData(
		nCtx,
		param.WorkflowID,
		map[string]any{managedOperationIDKey: conv.MapKeyToSlice(managedOperationIDMap)},
	); err != nil {
		logger.G.Sys().Ctx(nCtx).
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
