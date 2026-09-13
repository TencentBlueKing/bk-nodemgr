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
	IDaoProcessConfig

	IDomainPlugin
}

// IDaoPluginDeployment defines the plugin deployment dao interface.
type IDaoPluginDeployment interface {
	// CreatePluginDeployment create plugin deployment.
	CreatePluginDeployment(nCtx contextx.IContext, pluginDeployment *types.PluginDeployment) error

	// ListPluginDeployment list plugin deployment.
	ListPluginDeployment(nCtx contextx.IContext, page types.Page, conditions ...*types.PluginDeploymentCondition) (
		[]*types.PluginDeployment, int64, error)

	// UpdatePluginDeploymentInfo update plugin deployment info.
	UpdatePluginDeploymentInfo(nCtx contextx.IContext, token string, pluginDeploymentInfo *types.PluginDeploymentInfo) error

	// GetPluginDeploymentInfo get plugin deployment info.
	GetPluginDeploymentInfo(nCtx contextx.IContext, token string) (*types.PluginDeploymentInfo, error)

	// GetPluginDeploymentPluginConf get plugin deployment plugin conf.
	GetPluginDeploymentPluginConf(ctx contextx.IContext, token string) (*types.PluginDeploymentPluginConf, error)

	// UpdatePluginDeploymentPluginConf set plugin deployment plugin conf.
	UpdatePluginDeploymentPluginConf(ctx contextx.IContext, token string, pluginConf *types.PluginDeploymentPluginConf) error

	// GetPluginDeploymentPluginConfConfigFilesDetail get plugin deployment plugin conf config detail.
	GetPluginDeploymentPluginConfConfigFilesDetail(ctx contextx.IContext, token string) ([]*types.PluginConfigDetail, error)

	// UpsertPluginDeploymentPluginConfConfigFilesDetail update plugin deployment plugin conf config detail.
	UpsertPluginDeploymentPluginConfConfigFilesDetail(ctx contextx.IContext, token string, configDetails ...*types.PluginConfigDetail) error
}

// IDaoPluginWorkflow defines the dao interface.
type IDaoPluginWorkflow interface {
	// GetPluginWorkflow gets a plugin workflow by workflow-id.
	GetPluginWorkflow(nCtx contextx.IContext, workflowID string) (*types.PluginWorkflow, error)

	// GetPluginWorkflowStatus gets the status of a plugin workflow.
	GetPluginWorkflowStatus(nCtx contextx.IContext, workflowID string) (types.PluginWorkflowStatus, error)

	// CreatePluginWorkflow creates a new plugin workflow.
	CreatePluginWorkflow(nCtx contextx.IContext, workflow *types.PluginWorkflow) error

	// UpdatePluginWorkflowStatus updates the status of a plugin workflow.
	UpdatePluginWorkflowStatus(nCtx contextx.IContext, workflowID string, status types.PluginWorkflowStatus) error

	// CountPluginWorkflow count plugin workflows.
	CountPluginWorkflow(nCtx contextx.IContext, conditions ...*types.PluginWorkflowCondition) (int64, error)

	// ListPluginWorkflow list plugin workflows.
	ListPluginWorkflow(nCtx contextx.IContext, page types.Page, conditions ...*types.PluginWorkflowCondition) ([]*types.PluginWorkflow, int64, error)

	// DistinctPluginWorkflow distinct plugin workflows.
	DistinctPluginWorkflow(nCtx contextx.IContext, request types.PluginWorkflowDistinctRequest, conditions ...*types.PluginWorkflowCondition) (
		*types.PluginWorkflowDistinctResult, error)
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

	// SetPluginMemo set plugin memo by plugin name.
	SetPluginMemo(nCtx contextx.IContext, pluginName string, memo string) error

	// UpsertManyPlugins upsert many plugins.
	UpsertManyPlugins(nCtx contextx.IContext, plugins ...*types.Plugin) error
}

// IDaoProcess defines the process dao interface.
// nolint: interfacebloat
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

	// UpdateManyProcessInfo batch update process info by process ID.
	UpdateManyProcessInfo(nCtx contextx.IContext, processInfoDeltas []*types.ProcessInfoDelta) error

	// UpdateProcessManyHostBizID update process biz id for many host.
	UpdateProcessManyHostBizID(nCtx contextx.IContext, bizID int64, hostID ...int64) error

	// DeleteProcess delete process.
	DeleteProcess(nCtx contextx.IContext, hostID int64, pluginName string) error

	// ExistProcess exist process.
	ExistProcess(nCtx contextx.IContext, hostID int64, pluginName string) (bool, error)

	// GetProcess get process by host id and plugin name.
	GetProcess(nCtx contextx.IContext, hostID int64, pluginName string) (*types.Process, error)

	// GetProcessDistributionByHostID get process distribution by host ID.
	GetProcessDistributionByHostID(nCtx contextx.IContext, condition ...*types.ProcessCondition) (map[int64]int64, error)

	// GetProcessDistributionByPluginName get process distribution by plugin name.
	GetProcessDistributionByPluginName(nCtx contextx.IContext, condition ...*types.ProcessCondition) (map[string]int64, error)

	// DistinctProcess get distinct process.
	DistinctProcess(nCtx contextx.IContext, request types.ProcessDistinctSelector, condition ...*types.ProcessCondition) (
		*types.ProcessDistinctResult, error)
}

// IDaoProcessConfig defines the process config dao interface.
type IDaoProcessConfig interface {
	// CreateProcessConfig create process config record.
	CreateProcessConfig(nCtx contextx.IContext, config *types.ProcessConfig) error

	// GetProcessConfig get process config record.
	GetProcessConfig(nCtx contextx.IContext, processUniqueKey *types.ProcessUniqueKey, name string) (*types.ProcessConfig, error)

	// CountProcessConfigs count process config records.
	CountProcessConfigs(nCtx contextx.IContext, conditions ...*types.ProcessConfigCondition) (int64, error)

	// ListProcessConfigs list process config records.
	ListProcessConfigs(nCtx contextx.IContext, page types.Page, conditions ...*types.ProcessConfigCondition) ([]*types.ProcessConfig, int64, error)

	// UpsertProcessConfigs upsert many process config record.
	UpsertProcessConfigs(nCtx contextx.IContext, configs ...*types.ProcessConfig) error

	// DeleteProcessConfigsByProcessUniqueKey delete many process config record by process unique key.
	DeleteProcessConfigsByProcessUniqueKey(nCtx contextx.IContext, processUniqueKeys ...*types.ProcessUniqueKey) error

	// DeleteProcessConfigs delete many process config record.
	DeleteProcessConfigs(nCtx contextx.IContext, processUniqueKey *types.ProcessUniqueKey, names ...string) error
}

// IDomainPlugin defines the plugin domain interface.
type IDomainPlugin interface {
	// GetPluginVisibleBizIDs get plugin visible biz ids by plugin name.
	GetPluginVisibleBizIDs(nCtx contextx.IContext, pluginName string) ([]int64, error)

	// ListVisiblePluginByBizIDs list visible plugin by biz ids.
	ListVisiblePluginByBizIDs(nCtx contextx.IContext, bizIDs []int64) ([]*types.Plugin, error)

	// EnsureDefaultPlugin creates a missing default plugin from the release without changing existing plugins.
	EnsureDefaultPlugin(nCtx contextx.IContext, release *types.ReleasePlugin) error
}
