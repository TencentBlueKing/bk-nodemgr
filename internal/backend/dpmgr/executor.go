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

package dpmgr

import (
	"errors"
	"fmt"
	"sort"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IExecutor define the logic of executor.
type IExecutor interface {
	// Execute execute the change tasks.
	Execute(nCtx contextx.IContext, changeTasks ...*ChangeTask) error
}

var _ IExecutor = &Executor{}

// Executor defines the executor.
type Executor struct {
	nodeManager      managerIface.INodeManager
	pluginManager    managerIface.IPluginManager
	daoPlugin        plugin.IDaoPlugin
	daoProcessConfig plugin.IDaoProcessConfig
}

// ExecutorConfig defines the config of executor.
type ExecutorConfig struct {
	NodeManager      managerIface.INodeManager
	PluginManager    managerIface.IPluginManager
	DaoPlugin        plugin.IDaoPlugin
	DaoProcessConfig plugin.IDaoProcessConfig
}

// NewExecutor create a new executor.
func NewExecutor(conf *ExecutorConfig) *Executor {
	return &Executor{
		nodeManager:      conf.NodeManager,
		pluginManager:    conf.PluginManager,
		daoPlugin:        conf.DaoPlugin,
		daoProcessConfig: conf.DaoProcessConfig,
	}
}

// Execute execute the change tasks.
// nolint: gocognit,gocyclo,cyclop
func (executor *Executor) Execute(nCtx contextx.IContext, changeTasks ...*ChangeTask) error {
	m := make(map[ChangeAction][]*ChangeTask)
	for _, changeTask := range changeTasks {
		m[changeTask.Action] = append(m[changeTask.Action], changeTask)
	}

	for action, tasks := range m {
		switch action {
		// ===============================================================================
		// Agent Related Change Actions
		// ===============================================================================
		case ChangeActionAgentInstall:
			err := executor.executeChangeActionAgentInstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionAgentUninstall:
			err := executor.executeChangeActionAgentUninstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionAgentUpgrade:
			err := executor.executeChangeActionAgentUpgrade(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		// ===============================================================================
		// Plugin Related Change Actions
		// ===============================================================================
		case ChangeActionPluginInstall:
			err := executor.executeChangeActionPluginInstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginUninstall:
			err := executor.executeChangeActionPluginUninstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginUpgrade:
			err := executor.executeChangeActionPluginUpgrade(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		// ===============================================================================
		// Plugin Sub Config Related Change Actions
		// ===============================================================================
		case ChangeActionPluginApplySubConfig:
			err := executor.executeChangeActionPluginApplySubConfig(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginDeleteSubConfig:
			err := executor.executeChangeActionPluginDeleteSubConfig(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginDeleteSubConfigRecord:
			err := executor.executeChangeActionPluginDeleteSubConfigRecord(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		// ===============================================================================
		// Plugin Pkg Sub Config Related Change Actions
		// ===============================================================================
		case ChangeActionPluginPkgInstall:
			err := executor.executeChangeActionPluginPkgInstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginPkgUpgrade:
			err := executor.executeChangeActionPluginPkgUpgrade(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginPkgUninstall:
			err := executor.executeChangeActionPluginPkgUninstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		// ===============================================================================
		// Proxy Related Change Actions
		// ===============================================================================
		case ChangeActionProxyInstall:
			err := executor.executeChangeActionProxyInstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionProxyUninstall:
			err := executor.executeChangeActionProxyUninstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionProxyUpgrade:
			err := executor.executeChangeActionProxyUpgrade(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		default:
			return fmt.Errorf("failed to schedule and execute change action: unknown change action, change-action(%s)", action)
		}
	}

	return nil
}

func collectDeployPolicyIDs(tasks []*ChangeTask) []int64 {
	policyIDMap := make(map[int64]struct{})
	for _, task := range tasks {
		policyIDMap[task.DeployPolicyID] = struct{}{}
	}

	policyIDs := conv.MapKeyToSlice(policyIDMap)
	sort.Slice(policyIDs, func(i, j int) bool {
		return policyIDs[i] < policyIDs[j]
	})

	return policyIDs
}

func collectTargetBizIDs(tasks []*ChangeTask) []int64 {
	bizIDMap := make(map[int64]struct{})
	for _, task := range tasks {
		bizIDMap[task.Target.Host.Static.BizID] = struct{}{}
	}

	bizIDs := conv.MapKeyToSlice(bizIDMap)
	sort.Slice(bizIDs, func(i, j int) bool {
		return bizIDs[i] < bizIDs[j]
	})

	return bizIDs
}

func (executor *Executor) executeChangeActionAgentInstall(nCtx contextx.IContext, tasks []*ChangeTask) error {
	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action agent install: %w", err)
	}

	nodeDeployments := make([]*types.NodeDeployment, len(tasks))
	bizMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := task.Spec.GetSpecifyAgentParam()
		if err != nil {
			return fmt.Errorf("failed to schedule and execute change action: %w", err)
		}

		nodeDeployments[idx] = types.NewNodeDeployment(&types.DeploymentInfo{
			Host: types.Host{
				HostID:   task.Target.Host.HostID,
				TenantID: task.Target.Host.TenantID,
				Static:   task.Target.Host.Static,
				Dynamic: &types.HostDynamic{
					NodeRole:         types.NodeRoleAgent,
					NodeStatus:       types.NodeStatusInit,
					NodeVersion:      param.NodeVersion,
					NodeGeneration:   task.Target.Host.Dynamic.NodeGeneration,
					NodeCPUArch:      task.Target.Host.Dynamic.NodeCPUArch,
					NodeOsType:       task.Target.Host.Dynamic.NodeOsType,
					AgentID:          task.Target.Host.Dynamic.AgentID,
					NetworkUnitID:    task.Target.Host.Dynamic.NetworkUnitID,
					LoginIP:          task.Target.Host.Dynamic.LoginIP,
					LoginPort:        task.Target.Host.Dynamic.LoginPort,
					LoginUser:        task.Target.Host.Dynamic.LoginUser,
					LoginMode:        task.Target.Host.Dynamic.LoginMode,
					LoginCreditID:    task.Target.Host.Dynamic.LoginCreditID,
					LoginCreditValid: task.Target.Host.Dynamic.LoginCreditValid,
					ExportIP:         task.Target.Host.Dynamic.ExportIP,
					AdvertiseIP:      task.Target.Host.Dynamic.AdvertiseIP,
				},
			},
		})

		bizMap[task.Target.Host.Static.BizID] = struct{}{}
	}

	bizIDs := conv.MapKeyToSlice(bizMap)
	workflowID, err := executor.nodeManager.LaunchInstallNode(nCtx, types.InstallNodeParam{
		Type:            types.NodeWorkflowTypeInstallAgent,
		BizIDs:          bizIDs,
		Operator:        operator,
		DeployPolicyIDs: collectDeployPolicyIDs(tasks),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action agent install: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action agent install")

	return nil
}

func (executor *Executor) executeChangeActionAgentUninstall(nCtx contextx.IContext, tasks []*ChangeTask) error {
	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action agent uninstall: %w", err)
	}

	nodeDeployments := make([]*types.NodeDeployment, len(tasks))
	bizMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := task.Spec.GetSpecifyAgentParam()
		if err != nil {
			return fmt.Errorf("failed to schedule and execute change action: %w", err)
		}

		nodeDeployments[idx] = types.NewNodeDeployment(&types.DeploymentInfo{
			Host: types.Host{
				HostID:   task.Target.Host.HostID,
				TenantID: task.Target.Host.TenantID,
				Static:   task.Target.Host.Static,
				Dynamic: &types.HostDynamic{
					NodeRole:         types.NodeRoleAgent,
					NodeStatus:       types.NodeStatusInit,
					NodeVersion:      param.NodeVersion,
					NodeGeneration:   task.Target.Host.Dynamic.NodeGeneration,
					NodeCPUArch:      task.Target.Host.Dynamic.NodeCPUArch,
					NodeOsType:       task.Target.Host.Dynamic.NodeOsType,
					AgentID:          task.Target.Host.Dynamic.AgentID,
					NetworkUnitID:    task.Target.Host.Dynamic.NetworkUnitID,
					LoginIP:          task.Target.Host.Dynamic.LoginIP,
					LoginPort:        task.Target.Host.Dynamic.LoginPort,
					LoginUser:        task.Target.Host.Dynamic.LoginUser,
					LoginMode:        task.Target.Host.Dynamic.LoginMode,
					LoginCreditID:    task.Target.Host.Dynamic.LoginCreditID,
					LoginCreditValid: task.Target.Host.Dynamic.LoginCreditValid,
					ExportIP:         task.Target.Host.Dynamic.ExportIP,
					AdvertiseIP:      task.Target.Host.Dynamic.AdvertiseIP,
				},
			},
		})

		bizMap[task.Target.Host.Static.BizID] = struct{}{}
	}

	bizIDs := conv.MapKeyToSlice(bizMap)
	workflowID, err := executor.nodeManager.LaunchUninstallNode(nCtx, types.UninstallNodeParam{
		Type:            types.NodeWorkflowTypeUninstallAgent,
		BizIDs:          bizIDs,
		Operator:        operator,
		DeployPolicyIDs: collectDeployPolicyIDs(tasks),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action agent uninstall: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action agent uninstall")

	return nil
}

func (executor *Executor) executeChangeActionAgentUpgrade(nCtx contextx.IContext, tasks []*ChangeTask) error {
	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action agent upgrade: %w", err)
	}

	nodeDeployments := make([]*types.NodeDeployment, len(tasks))
	bizMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := task.Spec.GetSpecifyAgentParam()
		if err != nil {
			return fmt.Errorf("failed to schedule and execute change action: %w", err)
		}

		nodeDeployments[idx] = types.NewNodeDeployment(&types.DeploymentInfo{
			Host: types.Host{
				HostID:   task.Target.Host.HostID,
				TenantID: task.Target.Host.TenantID,
				Static:   task.Target.Host.Static,
				Dynamic: &types.HostDynamic{
					NodeRole:         types.NodeRoleAgent,
					NodeStatus:       types.NodeStatusInit,
					NodeVersion:      param.NodeVersion,
					NodeGeneration:   task.Target.Host.Dynamic.NodeGeneration,
					NodeCPUArch:      task.Target.Host.Dynamic.NodeCPUArch,
					NodeOsType:       task.Target.Host.Dynamic.NodeOsType,
					AgentID:          task.Target.Host.Dynamic.AgentID,
					NetworkUnitID:    task.Target.Host.Dynamic.NetworkUnitID,
					LoginIP:          task.Target.Host.Dynamic.LoginIP,
					LoginPort:        task.Target.Host.Dynamic.LoginPort,
					LoginUser:        task.Target.Host.Dynamic.LoginUser,
					LoginMode:        task.Target.Host.Dynamic.LoginMode,
					LoginCreditID:    task.Target.Host.Dynamic.LoginCreditID,
					LoginCreditValid: task.Target.Host.Dynamic.LoginCreditValid,
					ExportIP:         task.Target.Host.Dynamic.ExportIP,
					AdvertiseIP:      task.Target.Host.Dynamic.AdvertiseIP,
				},
			},
		})

		bizMap[task.Target.Host.Static.BizID] = struct{}{}
	}

	bizIDs := conv.MapKeyToSlice(bizMap)
	workflowID, err := executor.nodeManager.LaunchUpgradeNode(nCtx, types.UpgradeNodeParam{
		Type:            types.NodeWorkflowTypeUpgradeAgent,
		BizIDs:          bizIDs,
		Operator:        operator,
		DeployPolicyIDs: collectDeployPolicyIDs(tasks),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action agent upgrade: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action agent upgrade")

	return nil
}

// ===============================================================================
// Plugin Related Change Actions
// ===============================================================================

func (executor *Executor) executeChangeActionPluginInstall(nCtx contextx.IContext, tasks []*ChangeTask) error {
	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin install: %w", err)
	}

	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := task.Spec.GetSpecifyPluginParam()
		if err != nil {
			return fmt.Errorf("failed to schedule and execute change action: %w", err)
		}

		pluginDeployments[idx] = types.NewPluginDeployment(&types.PluginDeploymentInfo{
			Process: types.Process{
				TenantID:   nCtx.TenantID(),
				HostID:     task.Target.Host.HostID,
				PluginName: param.PluginName,
			},
			InstallOptions: types.PluginDeploymentInstallOptions{
				Version: param.Version,
			},
		}, &types.PluginDeploymentPluginConf{
			CustomConfigContext: param.CustomConfigContext,
		})

		hostMap[task.Target.Host.HostID] = struct{}{}
	}

	hostIDs := conv.MapKeyToSlice(hostMap)
	workflowID, err := executor.pluginManager.LaunchInstallPlugin(nCtx, types.InstallPluginParam{
		Type:              types.PluginWorkflowTypeInstall,
		HostIDs:           hostIDs,
		BizIDs:            collectTargetBizIDs(tasks),
		Operator:          operator,
		DeployPolicyIDs:   collectDeployPolicyIDs(tasks),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin install: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action plugin install")

	return nil
}

func (executor *Executor) executeChangeActionPluginUninstall(nCtx contextx.IContext, tasks []*ChangeTask) error {
	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin uninstall: %w", err)
	}

	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := task.Spec.GetSpecifyPluginParam()
		if err != nil {
			return fmt.Errorf("failed to schedule and execute change action: %w", err)
		}

		pluginDeployments[idx] = types.NewPluginDeployment(&types.PluginDeploymentInfo{
			Process: types.Process{
				TenantID:   nCtx.TenantID(),
				HostID:     task.Target.Host.HostID,
				PluginName: param.PluginName,
			},
		}, &types.PluginDeploymentPluginConf{})

		hostMap[task.Target.Host.HostID] = struct{}{}
	}

	hostIDs := conv.MapKeyToSlice(hostMap)
	workflowID, err := executor.pluginManager.LaunchUninstallPlugin(nCtx, types.UninstallPluginParam{
		Type:              types.PluginWorkflowTypeUninstall,
		HostIDs:           hostIDs,
		BizIDs:            collectTargetBizIDs(tasks),
		Operator:          operator,
		DeployPolicyIDs:   collectDeployPolicyIDs(tasks),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin uninstall: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).Info("successful to execute change action plugin uninstall")

	return nil
}

func (executor *Executor) executeChangeActionPluginUpgrade(nCtx contextx.IContext, tasks []*ChangeTask) error {
	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin upgrade: %w", err)
	}

	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := task.Spec.GetSpecifyPluginParam()
		if err != nil {
			return fmt.Errorf("failed to schedule and execute change action: %w", err)
		}

		pluginDeployments[idx] = types.NewPluginDeployment(&types.PluginDeploymentInfo{
			Process: types.Process{
				TenantID:   nCtx.TenantID(),
				HostID:     task.Target.Host.HostID,
				PluginName: param.PluginName,
			},
			InstallOptions: types.PluginDeploymentInstallOptions{
				Version: param.Version,
			},
		}, &types.PluginDeploymentPluginConf{
			CustomConfigContext: param.CustomConfigContext,
		})

		hostMap[task.Target.Host.HostID] = struct{}{}
	}

	hostIDs := conv.MapKeyToSlice(hostMap)
	workflowID, err := executor.pluginManager.LaunchUpgradePlugin(nCtx, types.UpgradePluginParam{
		Type:              types.PluginWorkflowTypeUpgrade,
		HostIDs:           hostIDs,
		BizIDs:            collectTargetBizIDs(tasks),
		Operator:          operator,
		DeployPolicyIDs:   collectDeployPolicyIDs(tasks),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin upgrade: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).Info("successful to execute change action plugin upgrade")

	return nil
}

// ===============================================================================
// Plugin Sub Config Related Change Actions
// ===============================================================================

func (executor *Executor) executeChangeActionPluginApplySubConfig(nCtx contextx.IContext, tasks []*ChangeTask) error {
	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin apply sub config: %w", err)
	}

	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := task.Spec.GetSpecifyPluginSubConfigParam()
		if err != nil {
			return fmt.Errorf("failed to schedule and execute change action: %w", err)
		}

		pluginDeployments[idx] = types.NewPluginDeployment(&types.PluginDeploymentInfo{
			Process: types.Process{
				TenantID:   nCtx.TenantID(),
				HostID:     task.Target.Host.HostID,
				PluginName: param.PluginName,
			},
		}, &types.PluginDeploymentPluginConf{
			Set:                 genDeployPolicyProcessConfigSet(task.DeployPolicyID),
			ConfigFilesDetail:   param.ConfigFilesDetail,
			CustomConfigContext: param.CustomConfigContext,
		})

		hostMap[task.Target.Host.HostID] = struct{}{}
	}

	hostIDs := conv.MapKeyToSlice(hostMap)
	workflowID, err := executor.pluginManager.LaunchApplyPluginSubConfig(nCtx, types.ApplyPluginSubConfigParam{
		Type:              types.PluginWorkflowTypeApplyPluginSubConfig,
		HostIDs:           hostIDs,
		BizIDs:            collectTargetBizIDs(tasks),
		Operator:          operator,
		DeployPolicyIDs:   collectDeployPolicyIDs(tasks),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin apply sub config: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action plugin apply sub config")

	return nil
}

func (executor *Executor) executeChangeActionPluginDeleteSubConfig(nCtx contextx.IContext, tasks []*ChangeTask) error {
	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin delete sub config: %w", err)
	}

	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := task.Spec.GetSpecifyPluginSubConfigParam()
		if err != nil {
			return fmt.Errorf("failed to schedule and execute change action: %w", err)
		}

		removeConfigFileNames := conv.SliceToSlice(param.ConfigFilesDetail, func(detail *types.PluginConfigDetail) string {
			return detail.Name
		})
		pluginDeployments[idx] = types.NewPluginDeployment(&types.PluginDeploymentInfo{
			Process: types.Process{
				TenantID:   nCtx.TenantID(),
				HostID:     task.Target.Host.HostID,
				PluginName: param.PluginName,
			},
		}, &types.PluginDeploymentPluginConf{
			RemoveConfigFileName: removeConfigFileNames,
		})

		hostMap[task.Target.Host.HostID] = struct{}{}
	}

	hostIDs := conv.MapKeyToSlice(hostMap)
	workflowID, err := executor.pluginManager.LaunchRemovePluginSubConfig(nCtx, types.RemovePluginSubConfigParam{
		Type:              types.PluginWorkflowTypeRemovePluginSubConfig,
		HostIDs:           hostIDs,
		BizIDs:            collectTargetBizIDs(tasks),
		Operator:          operator,
		DeployPolicyIDs:   collectDeployPolicyIDs(tasks),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin delete sub config: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action plugin delete sub config")

	return nil
}

func (executor *Executor) executeChangeActionPluginDeleteSubConfigRecord(nCtx contextx.IContext, tasks []*ChangeTask) error {
	for _, task := range tasks {
		param, err := task.Spec.GetSpecifyPluginSubConfigParam()
		if err != nil {
			return fmt.Errorf("failed to execute change action plugin delete sub config record: %w", err)
		}

		configNames := conv.SliceToSlice(param.ConfigFilesDetail, func(detail *types.PluginConfigDetail) string {
			return detail.Name
		})
		processUniqueKey := &types.ProcessUniqueKey{
			HostID: task.Target.Host.HostID,
			Name:   param.PluginName,
		}
		if err := executor.daoProcessConfig.DeleteProcessConfigs(nCtx, processUniqueKey, configNames...); err != nil {
			return fmt.Errorf("failed to execute change action plugin delete sub config record: %w", err)
		}
	}

	logger.G.Sys().With("task-count", len(tasks)).
		Info("successful to execute change action plugin delete sub config record")

	return nil
}

// ===============================================================================
// Plugin Pkg Sub Config Related Change Actions
// ===============================================================================

func (executor *Executor) executeChangeActionPluginPkgInstall(nCtx contextx.IContext, tasks []*ChangeTask) error {
	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg install: %w", err)
	}

	// 1. convert tasks to plugins.
	plugins := make([]*types.Plugin, len(tasks))
	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := task.Spec.GetSpecifyPluginPkgParam()
		if err != nil {
			return fmt.Errorf("failed to get specify plugin pkg param for task: %w", err)
		}

		plugins[idx] = &types.Plugin{
			TenantID: nCtx.TenantID(),
			Name:     genPluginNameForSpecifyPluginPkg(param.PluginPkgName, task.DeployPolicyID, task.Target.ServiceInstance.ModuleID),
			PkgName:  param.PluginPkgName,
			Group:    fmt.Sprintf("%d", task.DeployPolicyID),
			Memo:     fmt.Sprintf("this plugin is created by deploy policy %d", task.DeployPolicyID),
		}

		pluginDeployments[idx] = types.NewPluginDeployment(&types.PluginDeploymentInfo{
			Process: types.Process{
				TenantID:   nCtx.TenantID(),
				HostID:     task.Target.Host.HostID,
				PluginName: plugins[idx].Name,
			},
			InstallOptions: types.PluginDeploymentInstallOptions{
				Version: param.Version,
			},
		}, &types.PluginDeploymentPluginConf{
			CustomConfigContext: param.CustomConfigContext,
		})

		hostMap[task.Target.Host.HostID] = struct{}{}
	}

	// 2. upsert plugins.
	if err := executor.daoPlugin.UpsertManyPlugins(nCtx, plugins...); err != nil {
		return fmt.Errorf("failed to upsert plugins: %w", err)
	}

	// 3. build plugin deployments.
	hostIDs := conv.MapKeyToSlice(hostMap)
	workflowID, err := executor.pluginManager.LaunchInstallPlugin(nCtx, types.InstallPluginParam{
		Type:              types.PluginWorkflowTypeInstall,
		HostIDs:           hostIDs,
		BizIDs:            collectTargetBizIDs(tasks),
		Operator:          operator,
		DeployPolicyIDs:   collectDeployPolicyIDs(tasks),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg install: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action plugin pkg install")

	return nil
}

func (executor *Executor) executeChangeActionPluginPkgUpgrade(nCtx contextx.IContext, tasks []*ChangeTask) error {
	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg upgrade: %w", err)
	}

	// 1. convert tasks to plugins.
	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := task.Spec.GetSpecifyPluginPkgParam()
		if err != nil {
			return fmt.Errorf("failed to get specify plugin pkg param for task: %w", err)
		}

		pluginDeployments[idx] = types.NewPluginDeployment(&types.PluginDeploymentInfo{
			Process: types.Process{
				TenantID:   nCtx.TenantID(),
				HostID:     task.Target.Host.HostID,
				PluginName: genPluginNameForSpecifyPluginPkg(param.PluginPkgName, task.DeployPolicyID, task.Target.ServiceInstance.ModuleID),
			},
			InstallOptions: types.PluginDeploymentInstallOptions{
				Version: param.Version,
			},
		}, &types.PluginDeploymentPluginConf{
			CustomConfigContext: param.CustomConfigContext,
		})

		hostMap[task.Target.Host.HostID] = struct{}{}
	}

	// 2. build plugin deployments.
	hostIDs := conv.MapKeyToSlice(hostMap)
	workflowID, err := executor.pluginManager.LaunchUpgradePlugin(nCtx, types.UpgradePluginParam{
		Type:              types.PluginWorkflowTypeUpgrade,
		HostIDs:           hostIDs,
		BizIDs:            collectTargetBizIDs(tasks),
		Operator:          operator,
		DeployPolicyIDs:   collectDeployPolicyIDs(tasks),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg upgrade: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).Info("successful to execute change action plugin pkg upgrade")

	return nil
}

func (executor *Executor) executeChangeActionPluginPkgUninstall(nCtx contextx.IContext, tasks []*ChangeTask) error {
	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg uninstall: %w", err)
	}

	// 1. convert tasks to plugins.
	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := task.Spec.GetSpecifyPluginPkgParam()
		if err != nil {
			return fmt.Errorf("failed to get specify plugin pkg param for task: %w", err)
		}

		pluginDeployments[idx] = types.NewPluginDeployment(&types.PluginDeploymentInfo{
			Process: types.Process{
				TenantID:   nCtx.TenantID(),
				HostID:     task.Target.Host.HostID,
				PluginName: genPluginNameForSpecifyPluginPkg(param.PluginPkgName, task.DeployPolicyID, task.Target.ServiceInstance.ModuleID),
			},
		}, &types.PluginDeploymentPluginConf{})

		hostMap[task.Target.Host.HostID] = struct{}{}
	}

	// 2. build plugin deployments.
	hostIDs := conv.MapKeyToSlice(hostMap)
	workflowID, err := executor.pluginManager.LaunchUninstallPlugin(nCtx, types.UninstallPluginParam{
		Type:              types.PluginWorkflowTypeUninstall,
		HostIDs:           hostIDs,
		BizIDs:            collectTargetBizIDs(tasks),
		Operator:          operator,
		DeployPolicyIDs:   collectDeployPolicyIDs(tasks),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg uninstall: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).Info("successful to execute change action plugin pkg uninstall")

	return nil
}

// ===============================================================================
// Proxy Related Change Actions
// ===============================================================================

func (executor *Executor) executeChangeActionProxyInstall(_ contextx.IContext, _ []*ChangeTask) error {
	return errors.New("not implemented")
}

func (executor *Executor) executeChangeActionProxyUninstall(_ contextx.IContext, _ []*ChangeTask) error {
	return errors.New("not implemented")
}

func (executor *Executor) executeChangeActionProxyUpgrade(_ contextx.IContext, _ []*ChangeTask) error {
	return errors.New("not implemented")
}
