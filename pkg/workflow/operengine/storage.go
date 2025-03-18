/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operengine ...
package operengine

import (
	"context"
)

// OperInstStorage represents a OperInst operInstMgr storage handler.
// it will be used to store the custom records during OperInst scheduling.
type OperInstStorage interface {
	// GetOperInstData will get the OperInst param.
	GetOperInstData(ctx context.Context, operationInstID string) (*OperInstData, error)

	// GetOperInstDataWithoutActionData will get the OperInst param without action data.
	GetOperInstDataWithoutActionData(ctx context.Context, operationInstID string) (*OperInstData, error)

	// UpsertOperInstData will insert/update the OperInst param.
	UpsertOperInstData(ctx context.Context, data *OperInstData) error

	// MarkOperInstStopping will markActionRunning the OperInst is stopping.
	MarkOperInstStopping(ctx context.Context, operationInstID string) error

	// WatchOperInstStopping will return a chan, when the OperInst is stopping, it will close the chan.
	WatchOperInstStopping(ctx context.Context, operationInstID string) <-chan struct{}

	// GetActionInstData will get the action_inst_data.
	GetActionInstData(ctx context.Context, operInstID string, actionName string) (*ActionInstData, error)

	// GetActInstLifecycle will get the action_inst_data's lifecycle.
	GetActInstLifecycle(ctx context.Context, operInstID string, actionName string) (*ActInstLifeCycle, error)

	// UpdateActInstLifecycle will update the action_inst_data's lifecycle.
	UpdateActInstLifecycle(ctx context.Context, operInstID string, actionName string, lifecycle *ActInstLifeCycle) error

	// UpdateActionInstContent will update the action_inst_data's content.
	UpdateActionInstContent(ctx context.Context, operInstID string, actionName string, content map[string]any) error

	// UpdateLifecycle will update the OperInst's Lifecycle.
	UpdateLifecycle(ctx context.Context, operInstID string, data *Lifecycle) error

	// PushActInstMsgs will push a message to the action_inst_data's msg queue.
	PushActInstMsgs(ctx context.Context, operInstID string, actionName string, msgs ...Message) error
}

// OperationStorage represents a Operation operInstMgr storage handler.
type OperationStorage interface {
	// GetOperation will get an Operation from the database.
	GetOperation(ctx context.Context, operationID string) (*Operation, error)

	// UpsertOperation will insert or update an Operation in the database.
	UpsertOperation(ctx context.Context, operation *Operation) error
}
