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

package utils

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// NewSyncDataActionStandarder creates a new SyncDataActionStandarder.
func NewSyncDataActionStandarder() *SyncDataActionStandarder {
	return &SyncDataActionStandarder{}
}

// SyncDataActionStandarder defines the standard parameters of node action.
type SyncDataActionStandarder struct {
	instanceContext *action.InstanceContext
	param           SyncDataActionStandardParam

	ctx contextx.IContext
}

// Initialize initializes the NodeActionStandarder.
func (std *SyncDataActionStandarder) Initialize(instanceContext *action.InstanceContext, param SyncDataActionStandardParam) error {
	std.instanceContext = instanceContext
	std.param = param

	std.ctx = contextx.From(std.instanceContext.Ctx, contextx.WithTenantID(std.param.TenantID), contextx.WithBKUsername(std.param.Operator))

	return nil
}

// Context returns the tenant user context.
func (std *SyncDataActionStandarder) Context() contextx.IContext {
	return std.ctx
}

// TenantID returns the tenant ID.
func (std *SyncDataActionStandarder) TenantID() string {
	return std.param.TenantID
}

// Operator returns the operator.
func (std *SyncDataActionStandarder) Operator() string {
	return std.param.Operator
}

// InstanceData returns the instance data.
func (std *SyncDataActionStandarder) InstanceData() *action.InstanceData {
	return std.instanceContext.Data
}

// SyncDataActionStandardParam defines the standard parameters of sync data action.
type SyncDataActionStandardParam struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}
