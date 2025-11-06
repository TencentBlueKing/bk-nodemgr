/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the storage interface.
type IStorage interface {
	basestorage.Interface

	IDaoPluginDeployment
	IDaoPluginWorkflow
	IDaoPlugin
	IDaoProcess
}

// IDaoPluginDeployment defines the plugin deployment dao interface.
type IDaoPluginDeployment interface {
	// CreatePluginDeployment create plugin deployment.
	CreatePluginDeployment(nCtx contextx.IContext, pluginDeployment *types.PluginDeployment) error

	// UpdatePluginDeploymentInfo update plugin deployment info.
	UpdatePluginDeploymentInfo(nCtx contextx.IContext, token string, pluginDeploymentInfo *types.PluginDeploymentInfo) error

	// GetPluginDeploymentInfo get plugin deployment info.
	GetPluginDeploymentInfo(nCtx contextx.IContext, token string) (*types.PluginDeploymentInfo, error)

	// GetPluginDeploymentPluginConfConfigFilesDetail get plugin deployment plugin conf config detail.
	GetPluginDeploymentPluginConfConfigFilesDetail(ctx contextx.IContext, token string) ([]*types.PluginConfigDetail, error)

	// UpdatePluginDeploymentPluginConfConfigFilesDetail set plugin deployment plugin conf config detail.
	UpdatePluginDeploymentPluginConfConfigFilesDetail(ctx contextx.IContext, token string, configs ...*types.PluginConfigDetail) error

	// GetPluginDeploymentPluginConfCustomConfigContext get plugin deployment plugin conf custom config context.
	GetPluginDeploymentPluginConfCustomConfigContext(ctx contextx.IContext, token string) (map[string]any, error)
}

// IDaoPluginWorkflow defines the dao interface.
type IDaoPluginWorkflow interface {
	// GetPluginWorkflow gets a plugin workflow by workflow-id.
	GetPluginWorkflow(nCtx contextx.IContext, workflowID string) (*types.PluginWorkflow, error)

	// CreatePluginWorkflow creates a new plugin workflow.
	CreatePluginWorkflow(nCtx contextx.IContext, workflow *types.PluginWorkflow) error

	// UpdatePluginWorkflowStatus updates the status of a plugin workflow.
	UpdatePluginWorkflowStatus(nCtx contextx.IContext, workflowID string, status types.PluginWorkflowStatus) error
}

// IDaoPlugin defines the plugin dao interface.
type IDaoPlugin interface {
	// GetPlugin get plugin by id.
	GetPlugin(nCtx contextx.IContext, pluginName string) (*types.Plugin, error)

	// CountPlugins count plugins.
	CountPlugins(nCtx contextx.IContext, conditions ...*types.PluginCondition) (int64, error)

	// ListPlugins list plugins.
	ListPlugins(nCtx contextx.IContext, page types.Page, conditions ...*types.PluginCondition) ([]*types.Plugin, int64, error)

	// ExistPluginByPluginName check if plugin exist by plugin name.
	ExistPluginByPluginName(nCtx contextx.IContext, pluginName string) (bool, error)

	// ExistDefaultPluginByPluginPkgName check if default plugin exist by plugin package name.
	ExistDefaultPluginByPluginPkgName(nCtx contextx.IContext, pluginPkgName string) (exist bool, err error)

	// CreatePlugin create a plugin.
	CreatePlugin(nCtx contextx.IContext, plugin *types.Plugin) error
}

// IDaoProcess defines the process dao interface.
type IDaoProcess interface {
	// CountProcesses count processes.
	CountProcesses(nCtx contextx.IContext, condition ...*types.ProcessCondition) (int64, error)

	// ListProcesses list processes.
	ListProcesses(nCtx contextx.IContext, page types.Page, condition ...*types.ProcessCondition) ([]*types.Process, int64, error)

	// CreateProcess create process.
	CreateProcess(nCtx contextx.IContext, process *types.Process) error

	// UpdateProcess update process.
	UpdateProcess(nCtx contextx.IContext, hostID int64, pluginName string, process *types.Process) error

	// UpdateProcessInfo update process.
	UpdateProcessInfo(nCtx contextx.IContext, hostID int64, pluginName string, processInfo *types.ProcessInfo) error

	// DeleteProcess delete process.
	DeleteProcess(nCtx contextx.IContext, hostID int64, pluginName string) error

	// ExistProcess exist process.
	ExistProcess(nCtx contextx.IContext, hostID int64, pluginName string) (bool, error)

	// UpdateManyProcessInfo batch update process info by process ID.
	UpdateManyProcessInfo(nCtx contextx.IContext, processInfoDeltas []*types.ProcessInfoDelta) error
	
	// GetProcess get process by host id and plugin name.
	GetProcess(nCtx contextx.IContext, hostID int64, pluginName string) (*types.Process, error)
}
