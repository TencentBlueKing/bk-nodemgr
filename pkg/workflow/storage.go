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
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// IStorage defines the storage handler.
type IStorage interface {
	IStorageActionInstance
	IStorageOperation
	IStorageOperationInstance
	IStorageTrigger
}

// IStorageActionInstance defines the storage handler for action instance.
type IStorageActionInstance interface {
	// GetActionInstanceData gets full action instance data.
	GetActionInstanceData(ctx context.Context, operationInstanceID, actionName string) (*action.InstanceData, error)

	// GetActionInstanceLifecycle gets action instance lifecycle.
	GetActionInstanceLifecycle(ctx context.Context, operationInstanceID, actionName string) (*action.Lifecycle, error)

	// UpdateActionInstanceLifecycle updates action instance lifecycle.
	UpdateActionInstanceLifecycle(
		ctx context.Context, operationInstanceID, actionName string, lifecycle *action.Lifecycle) error

	// PushActionInstanceMessage pushes action instance message.
	PushActionInstanceMessage(
		ctx context.Context, operationInstanceID, actionName string, messages ...action.Message) error
}

// IStorageOperation defines the storage handler for operation.
type IStorageOperation interface {
	// UpsertOperation upserts operation.
	UpsertOperation(ctx context.Context, oper *operation.Operation) error

	// GetOperation gets operation.
	GetOperation(ctx context.Context, operationID string) (*operation.Operation, error)

	// ListOperation lists operation.
	ListOperation(ctx context.Context, page types.Page, triggerID string) ([]*operation.Operation, int64, error)

	// ListEmptyOperation lists empty operation.
	ListEmptyOperation(ctx context.Context, page types.Page, triggerID string) ([]*operation.Operation, int64, error)
}

// IStorageOperationInstance defines the storage handler for operation instance.
type IStorageOperationInstance interface {
	// GetOperationInstanceData gets full operation instance data.
	GetOperationInstanceFullData(ctx context.Context, operationID string) (*operation.InstanceData, error)

	// GetOperationInstanceBriefData gets brief operation instance data.
	GetOperationInstanceBriefData(ctx context.Context, operationID string) (*operation.InstanceBriefData, error)

	// ListOperationInstanceBriefData lists operation instance brief data. without action instance data.
	ListOperationInstanceBriefData(
		ctx context.Context, page types.Page, triggerID string, states ...operation.State) ([]*operation.InstanceBriefData, error)

	// CountOperationInstance counts operation instance.
	CountOperationInstance(ctx context.Context, triggerID string, states ...operation.State) (int64, error)

	// UpsertOperationInstanceData upserts operation instance data.
	UpsertOperationInstanceData(ctx context.Context, operationInstanceData *operation.InstanceData) error

	// UpdateOperationInstanceLifecycle updates operation instance lifecycle.
	UpdateOperationInstanceLifecycle(ctx context.Context, operationID string, lifecycle *operation.Lifecycle) error

	// WatchOperInstStopping watches operation instance stopping.
	WatchOperInstStopping(ctx context.Context, operationID string) <-chan struct{}
}

// IStorageTrigger defines the storage handler for trigger.
type IStorageTrigger interface {
	// CreateTrigger creates trigger.
	CreateTrigger(ctx context.Context, trig *trigger.Trigger) error

	// GetTrigger gets trigger.
	GetTrigger(ctx context.Context, triggerID string) (*trigger.Trigger, error)

	// UpdateTrigger updates trigger.
	UpdateTrigger(ctx context.Context, trig *trigger.Trigger) error

	// ListAliveTrigger lists alive triggers by given category.
	ListAliveTrigger(ctx context.Context, category trigger.Category) ([]*trigger.Trigger, error)
}
