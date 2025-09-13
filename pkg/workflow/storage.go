/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/schedule"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// IStorageActionInstance defines the storage handler for action instance.
type IStorageActionInstance interface {
	// GetActionInstanceData gets full action instance data.
	GetActionInstanceData(ctx contextx.IContext, operationInstanceID, actionName string) (*action.InstanceData, error)

	// GetActionInstanceLifecycle gets action instance lifecycle.
	GetActionInstanceLifecycle(ctx contextx.IContext, operationInstanceID, actionName string) (*action.Lifecycle, error)

	// UpdateActionInstanceLifecycle updates action instance lifecycle.
	UpdateActionInstanceLifecycle(
		ctx contextx.IContext, operationInstanceID, actionName string, lifecycle *action.Lifecycle) error

	// UpdateActionInstanceContent updates action instance content.
	UpdateActionInstanceContent(
		ctx contextx.IContext, operationInstanceID, actionName string, content map[string]any) error

	// UpdateOperInstActionStatus will update the oper inst action status.
	UpdateOperInstActionStatus(ctx contextx.IContext, operInstID string, actionName string, status action.State) error

	// UpsertActionInstancePrivateData upserts action instance private data.
	UpsertActionInstancePrivateData(
		ctx contextx.IContext, operInstID string, actionName string, privateData map[string]any) error

	// PushActionInstanceMessage pushes action instance message.
	PushActionInstanceMessage(
		ctx contextx.IContext, operationInstanceID, actionName string, messages ...common.Message) error

	// GetActionInstancePrivateData gets action instance private data.
	GetActionInstancePrivateData(ctx contextx.IContext, operationInstanceID, actionName string) (map[string]any, error)
}

// IStorageOperation defines the storage handler for operation.
type IStorageOperation interface {
	// UpsertOperation upserts operation.
	UpsertOperation(ctx contextx.IContext, oper *operation.Operation) error

	// GetOperation gets operation.
	GetOperation(ctx contextx.IContext, operationID string) (*operation.Operation, error)

	// ListOperationByTriggerID lists operation.
	ListOperationByTriggerID(ctx contextx.IContext, page types.Page, triggerID ...string) (
		[]*operation.Operation, int64, error)

	// ListOperationByOperationID lists operation by operation ID.
	ListOperationByOperationID(ctx contextx.IContext, operationID ...string) ([]*operation.Operation, int64, error)

	// ListEmptyOperation lists empty operation.
	ListEmptyOperation(ctx contextx.IContext, page types.Page, triggerID string) ([]*operation.Operation, int64, error)

	// DeleteOperations deletes operations.
	DeleteOperations(ctx contextx.IContext, operationID ...string) error

	// PullOperationInstanceIDsFromOperation pulls operation instance IDs from operation.
	PullOperationInstanceIDsFromOperation(ctx contextx.IContext, operationID string, operInstIDs ...string) error
}

// IStorageOperationInstance defines the storage handler for operation instance.
type IStorageOperationInstance interface {
	// GetOperationInstanceFullData gets full operation instance data.
	GetOperationInstanceFullData(ctx contextx.IContext, operationInstanceID string) (*operation.InstanceData, error)

	// GetOperationInstanceBriefData gets brief operation instance data.
	GetOperationInstanceBriefData(ctx contextx.IContext, operationInstanceID string) (*operation.InstanceBriefData, error)

	// ListOperationInstanceBriefData lists operation instance brief data. without action instance data.
	ListOperationInstanceBriefData(
		ctx contextx.IContext, page types.Page, conditions ...*types.OperInstDataCondition) (
		[]*operation.InstanceBriefData, int64, error)

	// ListOperInstanceBriefByOperationID lists operation instance brief data.
	ListOperInstanceBriefByOperationID(ctx contextx.IContext, page types.Page, operationID ...string) (
		[]*operation.InstanceBriefData, int64, error)

	// CountOperationInstance counts operation instance.
	CountOperationInstance(ctx contextx.IContext, triggerID string, states ...operation.State) (int64, error)

	// UpsertOperationInstanceData upserts operation instance data.
	UpsertOperationInstanceData(ctx contextx.IContext, operationInstanceData *operation.InstanceData) error

	// UpdateOperationInstanceLifecycle updates operation instance lifecycle.
	UpdateOperationInstanceLifecycle(ctx contextx.IContext, operationInstanceID string, lifecycle *operation.Lifecycle) error

	// UpdateOperationInstanceExtraExecutionMessages updates operation instance extra execution messages.
	UpdateOperationInstanceExtraExecutionMessages(
		ctx contextx.IContext, operationInstanceID string, messages ...common.Message) error

	// WatchOperInstStopping watches operation instance stopping.
	WatchOperInstStopping(ctx contextx.IContext, operationInstanceID string) <-chan struct{}

	// DeleteOperationInstances deletes operation instances.
	DeleteOperationInstances(ctx contextx.IContext, operationInstanceID ...string) error
}

// IStorageTrigger defines the storage handler for trigger.
type IStorageTrigger interface {
	// CreateTrigger creates trigger.
	CreateTrigger(ctx contextx.IContext, trig *trigger.Trigger) error

	// GetTrigger gets trigger.
	GetTrigger(ctx contextx.IContext, triggerID string) (*trigger.Trigger, error)

	// UpdateTrigger updates trigger.
	UpdateTrigger(ctx contextx.IContext, trig *trigger.Trigger) error

	// UpdateTriggerState updates trigger state.
	UpdateTriggerState(ctx contextx.IContext, triggerID string, state trigger.State) error

	// ListAliveTrigger lists alive triggers by given category.
	ListAliveTrigger(ctx contextx.IContext, category trigger.Category) ([]*trigger.Trigger, error)

	// DeleteTriggers deletes triggers by given trigger IDs.
	DeleteTriggers(ctx contextx.IContext, triggerIDs ...string) error
}

// IStorageSchedule defines the interface of schedule workflow storage.
type IStorageSchedule interface {
	// ListScheduleWorkflow lists schedule workflow by page and conditions.
	ListScheduleWorkflow(ctx contextx.IContext, page types.Page, conditions ...*types.ScheduleWorkflowCondition) (
		[]*schedule.Schedule, int64, error)

	// CountScheduleWorkflow counts schedule workflow by conditions.
	CountScheduleWorkflow(ctx contextx.IContext, conditions ...*types.ScheduleWorkflowCondition) (int64, error)

	// GetScheduleWorkflow gets a schedule workflow by workflow-id.
	GetScheduleWorkflow(ctx contextx.IContext, workflowID string) (*schedule.Schedule, error)

	// CreateScheduleWorkflow creates a new schedule workflow.
	CreateScheduleWorkflow(ctx contextx.IContext, workflow *schedule.Schedule) error
}
