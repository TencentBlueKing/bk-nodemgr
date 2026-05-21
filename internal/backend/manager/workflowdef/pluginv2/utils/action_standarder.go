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
	"fmt"

	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
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
	TenantID   string `json:"tenant_id"`
	Token      string `json:"token"`
	Operator   string `json:"operator"`
	SkipAction bool   `json:"skip_action"`
}

// GetActualExistingProcess gets the existing plugin process and refreshes its AgentID with the latest host value.
func GetActualExistingProcess(
	nCtx contextx.IContext,
	daoProcess pluginStg.IDaoProcessV2,
	daoHost topoStg.IStorageHost,
	hostID int64,
	pluginName string,
) (*types.Process, error) {

	process, err := daoProcess.GetProcessV2(nCtx, hostID, pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin process, process-name(%s), host-id(%d): %w", pluginName, hostID, err)
	}

	host, err := daoHost.GetHostByID(nCtx, process.HostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get host by id, host-id(%d): %w", process.HostID, err)
	}
	if host.Dynamic == nil || host.Dynamic.AgentID == "" {
		return nil, fmt.Errorf("host dynamic agent id is empty, host-id(%d)", process.HostID)
	}

	actualProcess := *process
	actualProcess.Info.AgentID = host.Dynamic.AgentID

	return &actualProcess, nil
}
