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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IExecutor define the logic of executor.
type IExecutor interface {
	// Execute executes change tasks and records policy workflow associations.
	Execute(nCtx contextx.IContext, execution ExecutionParam, changeTasks ...*ChangeTask) error
}

// ExecutionParam carries operation identity and runtime policy workflow associations.
type ExecutionParam struct {
	OperationID  string
	TriggerID    string
	WorkflowIDs  map[int64]string
	PolicyGroups map[int64]int64
}

func (execution ExecutionParam) workflowIDs(tasks []*ChangeTask) []string {
	groups := make(map[int64]struct{})
	for _, task := range tasks {
		if group, ok := execution.PolicyGroups[task.DeployPolicyID]; ok {
			groups[group] = struct{}{}
		}
	}
	ids := make([]string, 0)
	for policyID, group := range execution.PolicyGroups {
		if _, ok := groups[group]; !ok {
			continue
		}
		if workflowID := execution.WorkflowIDs[policyID]; workflowID != "" {
			ids = append(ids, workflowID)
		}
	}
	sort.Strings(ids)

	return ids
}

var _ IExecutor = &Executor{}

// Executor defines the executor.
type Executor struct {
	nodeManager             managerIface.INodeManager
	pluginManager           managerIface.IPluginManager
	daoPlugin               plugin.IDaoPlugin
	daoProcessConfig        plugin.IDaoProcessConfig
	daoDeployPolicyWorkflow deploypolicy.IDaoDeployPolicyWorkflow
}

// ExecutorConfig defines the config of executor.
type ExecutorConfig struct {
	NodeManager             managerIface.INodeManager
	PluginManager           managerIface.IPluginManager
	DaoPlugin               plugin.IDaoPlugin
	DaoProcessConfig        plugin.IDaoProcessConfig
	DaoDeployPolicyWorkflow deploypolicy.IDaoDeployPolicyWorkflow
}

type pluginPkgTaskParam struct {
	pluginName          string
	pluginPkgName       string
	version             string
	customConfigContext map[string]any
}

// NewExecutor create a new executor.
func NewExecutor(conf *ExecutorConfig) *Executor {
	return &Executor{
		nodeManager:             conf.NodeManager,
		pluginManager:           conf.PluginManager,
		daoPlugin:               conf.DaoPlugin,
		daoProcessConfig:        conf.DaoProcessConfig,
		daoDeployPolicyWorkflow: conf.DaoDeployPolicyWorkflow,
	}
}

// Execute preserves batching while associating children with participating policy groups.
// nolint: gocognit,gocyclo,cyclop
func (executor *Executor) Execute(nCtx contextx.IContext, execution ExecutionParam, changeTasks ...*ChangeTask) error {
	tasksByAction := make(map[ChangeAction][]*ChangeTask)
	for _, changeTask := range changeTasks {
		if _, ok := execution.PolicyGroups[changeTask.DeployPolicyID]; !ok {
			return fmt.Errorf("missing execution group for policy %d", changeTask.DeployPolicyID)
		}
		if execution.WorkflowIDs[changeTask.DeployPolicyID] == "" {
			return fmt.Errorf("missing execution workflow for policy %d", changeTask.DeployPolicyID)
		}
		tasksByAction[changeTask.Action] = append(tasksByAction[changeTask.Action], changeTask)
	}

	for action, tasks := range tasksByAction {
		switch action {
		// ===============================================================================
		// Agent Related Change Actions
		// ===============================================================================
		case ChangeActionAgentInstall:
			err := executor.executeChangeActionAgentInstall(nCtx, execution, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionAgentUninstall:
			err := executor.executeChangeActionAgentUninstall(nCtx, execution, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionAgentUpgrade:
			err := executor.executeChangeActionAgentUpgrade(nCtx, execution, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		// ===============================================================================
		// Plugin Related Change Actions
		// ===============================================================================
		case ChangeActionPluginInstall:
			err := executor.executeChangeActionPluginInstall(nCtx, execution, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginUninstall:
			err := executor.executeChangeActionPluginUninstall(nCtx, execution, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginUpgrade:
			err := executor.executeChangeActionPluginUpgrade(nCtx, execution, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		// ===============================================================================
		// Plugin Sub Config Related Change Actions
		// ===============================================================================
		case ChangeActionPluginApplySubConfig:
			err := executor.executeChangeActionPluginApplySubConfig(nCtx, execution, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginDeleteSubConfig:
			err := executor.executeChangeActionPluginDeleteSubConfig(nCtx, execution, tasks)
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
			err := executor.executeChangeActionPluginPkgInstall(nCtx, execution, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginPkgUpgrade:
			err := executor.executeChangeActionPluginPkgUpgrade(nCtx, execution, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginPkgUninstall:
			err := executor.executeChangeActionPluginPkgUninstall(nCtx, execution, tasks)
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

func (executor *Executor) recordWorkflowChild(nCtx contextx.IContext, execution ExecutionParam,
	tasks []*ChangeTask, domain types.WorkflowDomain, workflowID string) error {

	workflowIDs := execution.workflowIDs(tasks)
	if len(workflowIDs) == 0 {
		return errors.New("deploy policy workflow associations are required")
	}
	if executor.daoDeployPolicyWorkflow == nil {
		return errors.New("deploy policy workflow storage is required")
	}
	if workflowID == "" {
		return errors.New("launched workflow id is empty")
	}

	child := types.DeployPolicyWorkflowChild{
		WorkflowID:     workflowID,
		WorkflowDomain: domain,
	}
	if err := executor.daoDeployPolicyWorkflow.RecordDeployPolicyWorkflowChild(nCtx, workflowIDs, child); err != nil {
		return fmt.Errorf("failed to record deploy policy workflow child: %w", err)
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

func (executor *Executor) executeChangeActionAgentInstall(nCtx contextx.IContext, execution ExecutionParam, tasks []*ChangeTask) error {
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
	if err := executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainNode, workflowID); err != nil {
		return fmt.Errorf("failed to execute change action agent install: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action agent install")

	return nil
}

func (executor *Executor) executeChangeActionAgentUninstall(nCtx contextx.IContext, execution ExecutionParam,
	tasks []*ChangeTask) error {

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
	if err := executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainNode, workflowID); err != nil {
		return fmt.Errorf("failed to execute change action agent uninstall: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action agent uninstall")

	return nil
}

func (executor *Executor) executeChangeActionAgentUpgrade(nCtx contextx.IContext, execution ExecutionParam, tasks []*ChangeTask) error {
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
	if err := executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainNode, workflowID); err != nil {
		return fmt.Errorf("failed to execute change action agent upgrade: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action agent upgrade")

	return nil
}

// ===============================================================================
// Plugin Related Change Actions
// ===============================================================================

func (executor *Executor) executeChangeActionPluginInstall(nCtx contextx.IContext, execution ExecutionParam, tasks []*ChangeTask) error {
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
			ConfigSource: getTaskConfigSource(task),
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
	if err := executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainPlugin, workflowID); err != nil {
		return fmt.Errorf("failed to execute change action plugin install: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action plugin install")

	return nil
}

func (executor *Executor) executeChangeActionPluginUninstall(nCtx contextx.IContext, execution ExecutionParam,
	tasks []*ChangeTask) error {

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
			ConfigSource: getTaskConfigSource(task),
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
	if err := executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainPlugin, workflowID); err != nil {
		return fmt.Errorf("failed to execute change action plugin uninstall: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).Info("successful to execute change action plugin uninstall")

	return nil
}

func (executor *Executor) executeChangeActionPluginUpgrade(nCtx contextx.IContext, execution ExecutionParam, tasks []*ChangeTask) error {
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
			ConfigSource: getTaskConfigSource(task),
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
	if err := executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainPlugin, workflowID); err != nil {
		return fmt.Errorf("failed to execute change action plugin upgrade: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).Info("successful to execute change action plugin upgrade")

	return nil
}

// ===============================================================================
// Plugin Sub Config Related Change Actions
// ===============================================================================

func (executor *Executor) executeChangeActionPluginApplySubConfig(nCtx contextx.IContext, execution ExecutionParam,
	tasks []*ChangeTask) error {

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
			ConfigSource: getTaskConfigSource(task),
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
	if err := executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainPlugin, workflowID); err != nil {
		return fmt.Errorf("failed to execute change action plugin apply sub config: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action plugin apply sub config")

	return nil
}

func (executor *Executor) executeChangeActionPluginDeleteSubConfig(nCtx contextx.IContext, execution ExecutionParam,
	tasks []*ChangeTask) error {

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
			ConfigSource: getTaskConfigSource(task),
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
	if err := executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainPlugin, workflowID); err != nil {
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

func (executor *Executor) executeChangeActionPluginPkgInstall(nCtx contextx.IContext, execution ExecutionParam,
	tasks []*ChangeTask) error {

	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg install: %w", err)
	}

	// 1. convert tasks to plugins.
	plugins := make([]*types.Plugin, len(tasks))
	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := getPluginPkgTaskParam(task)
		if err != nil {
			return fmt.Errorf("failed to get plugin pkg task param: %w", err)
		}

		plugins[idx] = &types.Plugin{
			TenantID: nCtx.TenantID(),
			Name:     param.pluginName,
			PkgName:  param.pluginPkgName,
			Group:    fmt.Sprintf("%d", task.DeployPolicyID),
			Memo:     fmt.Sprintf("this plugin is created by deploy policy %d", task.DeployPolicyID),
		}

		pluginDeployments[idx] = types.NewPluginDeployment(&types.PluginDeploymentInfo{
			Process: types.Process{
				TenantID:   nCtx.TenantID(),
				HostID:     task.Target.Host.HostID,
				PluginName: plugins[idx].Name,
			},
			ConfigSource: getTaskConfigSource(task),
			InstallOptions: types.PluginDeploymentInstallOptions{
				Version: param.version,
			},
		}, &types.PluginDeploymentPluginConf{
			CustomConfigContext: param.customConfigContext,
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
	if err := executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainPlugin, workflowID); err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg install: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action plugin pkg install")

	return nil
}

func (executor *Executor) executeChangeActionPluginPkgUpgrade(nCtx contextx.IContext, execution ExecutionParam,
	tasks []*ChangeTask) error {

	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg upgrade: %w", err)
	}

	// 1. convert tasks to plugins.
	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := getPluginPkgTaskParam(task)
		if err != nil {
			return fmt.Errorf("failed to get plugin pkg task param: %w", err)
		}

		pluginDeployments[idx] = types.NewPluginDeployment(&types.PluginDeploymentInfo{
			Process: types.Process{
				TenantID:   nCtx.TenantID(),
				HostID:     task.Target.Host.HostID,
				PluginName: param.pluginName,
			},
			ConfigSource: getTaskConfigSource(task),
			InstallOptions: types.PluginDeploymentInstallOptions{
				Version: param.version,
			},
		}, &types.PluginDeploymentPluginConf{
			CustomConfigContext: param.customConfigContext,
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
	if err := executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainPlugin, workflowID); err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg upgrade: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).Info("successful to execute change action plugin pkg upgrade")

	return nil
}

func (executor *Executor) executeChangeActionPluginPkgUninstall(nCtx contextx.IContext, execution ExecutionParam,
	tasks []*ChangeTask) error {

	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg uninstall: %w", err)
	}

	// 1. convert tasks to plugins.
	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		param, err := getPluginPkgTaskParam(task)
		if err != nil {
			return fmt.Errorf("failed to get plugin pkg task param: %w", err)
		}

		pluginDeployments[idx] = types.NewPluginDeployment(&types.PluginDeploymentInfo{
			Process: types.Process{
				TenantID:   nCtx.TenantID(),
				HostID:     task.Target.Host.HostID,
				PluginName: param.pluginName,
			},
			ConfigSource: getTaskConfigSource(task),
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
	if err := executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainPlugin, workflowID); err != nil {
		return fmt.Errorf("failed to execute change action plugin pkg uninstall: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).Info("successful to execute change action plugin pkg uninstall")

	return nil
}

func getPluginPkgTaskParam(task *ChangeTask) (*pluginPkgTaskParam, error) {
	if task == nil {
		return nil, fmt.Errorf("task is nil")
	}

	if task.Spec == nil {
		return nil, fmt.Errorf("task spec is nil")
	}

	switch task.Spec.Type() {
	case types.DeploySpecTypeSpecifyPluginPkg:
		return getSpecifyPluginPkgTaskParam(task)
	case types.DeploySpecTypeProjectPluginPkgToHosts:
		return getProjectPluginPkgToHostsTaskParam(task)
	default:
		return nil, fmt.Errorf("unsupported plugin pkg spec type, type(%s)", task.Spec.Type())
	}
}

func getSpecifyPluginPkgTaskParam(task *ChangeTask) (*pluginPkgTaskParam, error) {
	param, err := task.Spec.GetSpecifyPluginPkgParam()
	if err != nil {
		return nil, fmt.Errorf("failed to get specify plugin pkg param: %w", err)
	}

	return &pluginPkgTaskParam{
		pluginName: genPluginNameForSpecifyPluginPkg(
			param.PluginPkgName,
			task.DeployPolicyID,
			task.Target.ServiceInstance.ModuleID,
		),
		pluginPkgName:       param.PluginPkgName,
		version:             param.Version,
		customConfigContext: param.CustomConfigContext,
	}, nil
}

func getProjectPluginPkgToHostsTaskParam(task *ChangeTask) (*pluginPkgTaskParam, error) {
	param, err := task.Spec.GetProjectPluginPkgToHostsParam()
	if err != nil {
		return nil, fmt.Errorf("failed to get project plugin pkg to hosts param: %w", err)
	}

	return &pluginPkgTaskParam{
		pluginName: genPluginNameForProjectPluginPkgToHosts(
			param.PluginPkgName,
			task.DeployPolicyID,
			task.Target.ServiceInstance.ModuleID,
			task.Target.ServiceInstance.HostID,
		),
		pluginPkgName:       param.PluginPkgName,
		version:             param.Version,
		customConfigContext: param.CustomConfigContext,
	}, nil
}

func getTaskConfigSource(task *ChangeTask) types.Target {
	if !needPluginDeploymentConfigSource(task) {
		return types.Target{}
	}

	return *task.ConfigSource
}

func needPluginDeploymentConfigSource(task *ChangeTask) bool {
	if task == nil || task.ConfigSource == nil {
		return false
	}

	configSource := task.ConfigSource
	if configSource.Host.HostID <= 0 {
		return false
	}
	if task.Target == nil || configSource.Host.HostID != task.Target.Host.HostID {
		return true
	}

	return configSource.ServiceInstance.ID > 0 || len(configSource.MatchedTopoRelations) > 0
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
