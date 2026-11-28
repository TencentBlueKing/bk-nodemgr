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

package workflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// IStorage defines the interface of schedule workflow storage.
type IStorage interface {
	basestorage.Interface
	workflow.IStorageTrigger
	workflow.IStorageActionInstance
	workflow.IStorageOperation
	workflow.IStorageOperationInstance

	IStorageScheduledWorkflow
	IDomainScheduledWorkflowQuery
}

// IDomainScheduledWorkflowQuery restricts execution queries to the current tenant's scheduled task chains.
type IDomainScheduledWorkflowQuery interface {
	// ListScheduledWorkflowOperation lists root operations or one instance's direct child operations.
	ListScheduledWorkflowOperation(nCtx contextx.IContext, page types.Page, workflowID, parentOperInstID string,
		onlyCount bool) ([]*operation.Operation, int64, error)

	// ListScheduledWorkflowOperationInstance lists instances after validating all supplied identifiers.
	ListScheduledWorkflowOperationInstance(nCtx contextx.IContext, operationIDs, operInstIDs []string,
		onlyCount bool) ([]*operation.InstanceBriefData, int64, error)

	// GetScheduledWorkflowOperationInstanceLog gets an instance in a scheduled task chain.
	GetScheduledWorkflowOperationInstanceLog(nCtx contextx.IContext, operInstID string) (*operation.InstanceData, error)
}

// IStorageScheduledWorkflow defines the interface of scheduled workflow storage.
type IStorageScheduledWorkflow interface {
	// ListScheduledWorkflow lists scheduled workflow by page and conditions.
	ListScheduledWorkflow(nCtx contextx.IContext, page types.Page, conditions ...*types.ScheduledWorkflowCondition) (
		[]*types.ScheduledWorkflow, int64, error)

	// CountScheduledWorkflow counts scheduled workflow by conditions.
	CountScheduledWorkflow(nCtx contextx.IContext, conditions ...*types.ScheduledWorkflowCondition) (int64, error)

	// GetScheduledWorkflow gets a scheduled workflow by workflow-id.
	GetScheduledWorkflow(nCtx contextx.IContext, workflowID string) (*types.ScheduledWorkflow, error)

	// CreateScheduledWorkflow creates a new scheduled workflow.
	CreateScheduledWorkflow(nCtx contextx.IContext, workflow *types.ScheduledWorkflow) error

	// UpdateScheduledWorkflowTriggerID updates a scheduled workflow's trigger ID.
	UpdateScheduledWorkflowTriggerID(nCtx contextx.IContext, workflowID, triggerID string) error

	// UpdateScheduledWorkflowPrivateData updates a scheduled workflow's private data.
	UpdateScheduledWorkflowPrivateData(nCtx contextx.IContext, workflowID string, privateData map[string]any) error

	// UpdateScheduleWorkflowInterval updates a scheduled workflow's interval and operator metadata.
	UpdateScheduleWorkflowInterval(nCtx contextx.IContext, workflowID, interval string) error

	// SwitchScheduleWorkflow enables or disables a scheduled workflow.
	SwitchScheduleWorkflow(nCtx contextx.IContext, workflowID string, enable bool) error
}
