/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package utils

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// NewScheduleActionStandarder creates a new ScheduleActionStandarder.
func NewScheduleActionStandarder() *ScheduleActionStandarder {
	return &ScheduleActionStandarder{}
}

// ScheduleActionStandarder defines the standard parameters of node action.
type ScheduleActionStandarder struct {
	instanceContext *action.InstanceContext
	param           ScheduleActionStandardParam

	ctx contextx.IContext
}

// Initialize initializes the NodeActionStandarder.
func (std *ScheduleActionStandarder) Initialize(instanceContext *action.InstanceContext, param ScheduleActionStandardParam) error {
	std.instanceContext = instanceContext
	std.param = param

	std.ctx = contextx.From(std.instanceContext.Ctx, contextx.WithTenantID(std.param.TenantID), contextx.WithBKUsername(std.param.Operator))

	return nil
}

// Context returns the tenant user context.
func (std *ScheduleActionStandarder) Context() contextx.IContext {
	return std.ctx
}

// WorkflowID returns the workflow ID.
func (std *ScheduleActionStandarder) WorkflowID() string {
	return std.param.WorkflowID
}

// TenantID returns the tenant ID.
func (std *ScheduleActionStandarder) TenantID() string {
	return std.param.TenantID
}

// Operator returns the operator.
func (std *ScheduleActionStandarder) Operator() string {
	return std.param.Operator
}

// InstanceData returns the instance data.
func (std *ScheduleActionStandarder) InstanceData() *action.InstanceData {
	return std.instanceContext.Data
}

// ScheduleActionStandardParam defines the standard parameters of sync data action.
type ScheduleActionStandardParam struct {
	WorkflowID string `json:"workflow_id"`
	TenantID   string `json:"tenant_id"`
	Operator   string `json:"operator"`
}
