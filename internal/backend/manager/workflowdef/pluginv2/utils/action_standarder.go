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
	"fmt"

	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// NewPluginActionStandarder creates a new PluginActionStandarder.
func NewPluginActionStandarder(daoPluginDeployment pluginStg.IDaoPluginDeployment) *PluginActionStandarder {
	return &PluginActionStandarder{
		daoPluginDeployment: daoPluginDeployment,
	}
}

// PluginActionStandarder defines the standard parameters of plugin action.
type PluginActionStandarder struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment

	instanceContext *action.InstanceContext
	param           PluginActionStandardParam

	ctx  contextx.IContext
	info *types.PluginDeploymentInfo
}

// Initialize initializes the PluginActionStandarder.
func (std *PluginActionStandarder) Initialize(instanceContext *action.InstanceContext, param PluginActionStandardParam) error {
	std.instanceContext = instanceContext
	std.param = param

	var err error
	std.info, err = std.daoPluginDeployment.GetPluginDeploymentInfo(std.instanceContext.Ctx, std.param.Token)
	if err != nil {
		return fmt.Errorf("failed to get plugin deployment info: %w", err)
	}

	std.ctx = contextx.From(std.instanceContext.Ctx, contextx.WithTenantID(std.info.Process.TenantID), contextx.WithBKUsername(std.param.Operator))

	return nil
}

// Save saves the PluginActionStandarder data.
func (std *PluginActionStandarder) Save() error {
	if err := std.daoPluginDeployment.UpdatePluginDeploymentInfo(std.instanceContext.Ctx, std.param.Token, std.info); err != nil {
		return fmt.Errorf("failed to update plugin deployment info: %w", err)
	}

	return nil
}

// SaveBlockingActionName saves the BlockingActionName data.
func (std *PluginActionStandarder) SaveBlockingActionName(actionName string) error {
	std.info.BlockingActionName = actionName
	if err := std.daoPluginDeployment.UpdatePluginDeploymentInfo(std.instanceContext.Ctx, std.param.Token, std.info); err != nil {
		return fmt.Errorf("failed to update plugin deployment info: %w", err)
	}

	return nil
}

// DeployInfo returns the plugin deployment info.
func (std *PluginActionStandarder) DeployInfo() *types.PluginDeploymentInfo {
	return std.info
}

// Context returns the tenant user context.
func (std *PluginActionStandarder) Context() contextx.IContext {
	return std.ctx
}

// Token returns the token.
func (std *PluginActionStandarder) Token() string {
	return std.param.Token
}

// Operator returns the operator.
func (std *PluginActionStandarder) Operator() string {
	return std.param.Operator
}

// InstanceData returns the instance data.
func (std *PluginActionStandarder) InstanceData() *action.InstanceData {
	return std.instanceContext.Data
}

// PluginActionStandardParam defines the standard parameters of plugin action.
type PluginActionStandardParam struct {
	TenantID string `json:"tenant_id"`
	Token    string `json:"token"`
	Operator string `json:"operator"`
}
