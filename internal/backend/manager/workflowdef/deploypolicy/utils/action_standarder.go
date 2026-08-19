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

// NewDeployPolicyActionStandarder creates a new DeployPolicyActionStandarder.
func NewDeployPolicyActionStandarder() *DeployPolicyActionStandarder {
	return &DeployPolicyActionStandarder{}
}

// DeployPolicyActionStandarder defines the standard parameters of deploy policy action.
type DeployPolicyActionStandarder struct {
	instanceContext *action.InstanceContext
	param           DeployPolicyActionStandardParam

	ctx contextx.IContext
}

// Initialize initializes the DeployPolicyActionStandarder.
func (std *DeployPolicyActionStandarder) Initialize(instanceContext *action.InstanceContext, param DeployPolicyActionStandardParam) error {
	std.instanceContext = instanceContext
	std.param = param

	std.ctx = contextx.From(std.instanceContext.Ctx, contextx.WithTenantID(std.param.TenantID), contextx.WithBKUsername(std.param.Operator))

	return nil
}

// Context returns the tenant user context.
func (std *DeployPolicyActionStandarder) Context() contextx.IContext {
	return std.ctx
}

// TenantID returns the tenant ID.
func (std *DeployPolicyActionStandarder) TenantID() string {
	return std.param.TenantID
}

// Operator returns the operator.
func (std *DeployPolicyActionStandarder) Operator() string {
	return std.param.Operator
}

// InstanceData returns the instance data.
func (std *DeployPolicyActionStandarder) InstanceData() *action.InstanceData {
	return std.instanceContext.Data
}

// DeployPolicyActionStandardParam defines the standard parameters of deploy policy action.
type DeployPolicyActionStandardParam struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}
