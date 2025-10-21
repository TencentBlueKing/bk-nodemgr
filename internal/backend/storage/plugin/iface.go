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

	// GetPluginDeploymentMainConfig get plugin deployment main config.
	GetPluginDeploymentMainConfig(nCtx contextx.IContext, token string) ([]byte, error)

	// UpdatePluginDeploymentMainConfig set plugin deployment main config.
	UpdatePluginDeploymentMainConfig(nCtx contextx.IContext, token string, mainConfig []byte) error
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
	// GetPluginByID get plugin by id.
	GetPluginByID(nCtx contextx.IContext, pluginID string) (*types.Plugin, error)
}

// IDaoProcess defines the process dao interface.
type IDaoProcess interface {
	// ListProcesses list processes.
	ListProcesses(nCtx contextx.IContext, page types.Page, condition ...*types.ProcessCondition) ([]*types.Process, int64, error)

	// GetProcessByID get process by id.
	GetProcessByID(nCtx contextx.IContext, processID string) (*types.Process, error)

	// CreateProcess create process.
	CreateProcess(nCtx contextx.IContext, process *types.Process) error

	// UpdateProcess update process.
	UpdateProcess(nCtx contextx.IContext, processID string, process *types.Process) error

	// UpdateProcessInfo update process.
	UpdateProcessInfo(nCtx contextx.IContext, processID string, processInfo *types.ProcessInfo) error

	// DeleteProcess delete process.
	DeleteProcess(nCtx contextx.IContext, processID string) error

	// ExistProcess exist process.
	ExistProcess(nCtx contextx.IContext, processID string) (bool, error)
}
