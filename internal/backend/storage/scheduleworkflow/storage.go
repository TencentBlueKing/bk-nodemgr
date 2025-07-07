/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package scheduleworkflow provides storage for schedule workflow.
package scheduleworkflow

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the interface of schedule workflow storage.
type IStorage interface {
	basestorage.Interface

	// ListScheduleWorkflow lists schedule workflow by page and conditions.
	ListScheduleWorkflow(ctx context.Context, page types.Page, conditions ...*types.ScheduleWorkflowCondition) (
		[]*types.ScheduleWorkflow, int64, error)

	// CountScheduleWorkflow counts schedule workflow by conditions.
	CountScheduleWorkflow(ctx context.Context, conditions ...*types.ScheduleWorkflowCondition) (int64, error)

	// DistinctScheduleWorkflow distincts schedule workflow fields.
	DistinctScheduleWorkflow(ctx context.Context, request types.ScheduleWorkflowDistinctRequest,
		conditions ...*types.ScheduleWorkflowCondition) (*types.ScheduleWorkflowDistinctResult, error)

	// GetScheduleWorkflow gets a schedule workflow by workflow-id.
	GetScheduleWorkflow(ctx context.Context, workflowID string) (*types.ScheduleWorkflow, error)

	// CreateScheduleWorkflow creates a new schedule workflow.
	CreateScheduleWorkflow(ctx context.Context, workflow *types.ScheduleWorkflow) error
}
