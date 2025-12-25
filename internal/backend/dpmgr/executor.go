/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package dpmgr

import (
	"errors"
	"fmt"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
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
	nodeManager   managerIface.INodeManager
	pluginManager managerIface.IPluginManager
}

// ExecutorConfig defines the config of executor.
type ExecutorConfig struct {
	NodeManager   managerIface.INodeManager
	PluginManager managerIface.IPluginManager
}

// NewExecutor create a new executor.
func NewExecutor(conf *ExecutorConfig) *Executor {
	return &Executor{
		nodeManager:   conf.NodeManager,
		pluginManager: conf.PluginManager,
	}
}

// Execute execute the change tasks.
func (s *Executor) Execute(nCtx contextx.IContext, changeTasks ...*ChangeTask) error {
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
			err := s.executeChangeActionAgentInstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionAgentUninstall:
			err := s.executeChangeActionAgentUninstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionAgentUpgrade:
			err := s.executeChangeActionAgentUpgrade(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		// ===============================================================================
		// Plugin Related Change Actions
		// ===============================================================================
		case ChangeActionPluginInstall:
			err := s.executeChangeActionPluginInstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginUninstall:
			err := s.executeChangeActionPluginUninstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginUpgrade:
			err := s.executeChangeActionPluginUpgrade(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		// ===============================================================================
		// Plugin Sub Config Related Change Actions
		// ===============================================================================
		case ChangeActionPluginApplySubConfig:
			err := s.executeChangeActionPluginApplySubConfig(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionPluginDeleteSubConfig:
			err := s.executeChangeActionPluginDeleteSubConfig(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		// ===============================================================================
		// Proxy Related Change Actions
		// ===============================================================================
		case ChangeActionProxyInstall:
			err := s.executeChangeActionProxyInstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionProxyUninstall:
			err := s.executeChangeActionProxyUninstall(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		case ChangeActionProxyUpgrade:
			err := s.executeChangeActionProxyUpgrade(nCtx, tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule and execute change action: %w", err)
			}
		default:
			return fmt.Errorf("failed to schedule and execute change action: unknown change action, change-action(%s)", action)
		}
	}

	return nil
}

func (s *Executor) executeChangeActionAgentInstall(nCtx contextx.IContext, tasks []*ChangeTask) error {
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
	workflowID, err := s.nodeManager.LaunchInstallNode(nCtx, types.InstallNodeParam{
		Type:            types.NodeWorkflowTypeInstallAgent,
		BizIDs:          bizIDs,
		Operator:        access.GetVirtualUser(),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action agent install: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action agent install")

	return nil
}

func (s *Executor) executeChangeActionAgentUninstall(nCtx contextx.IContext, tasks []*ChangeTask) error {
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
	workflowID, err := s.nodeManager.LaunchUninstallNode(nCtx, types.UninstallNodeParam{
		Type:            types.NodeWorkflowTypeUninstallAgent,
		BizIDs:          bizIDs,
		Operator:        access.GetVirtualUser(),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action agent uninstall: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action agent uninstall")

	return nil
}

func (s *Executor) executeChangeActionAgentUpgrade(nCtx contextx.IContext, tasks []*ChangeTask) error {
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
	workflowID, err := s.nodeManager.LaunchUpgradeNode(nCtx, types.UpgradeNodeParam{
		Type:            types.NodeWorkflowTypeUpgradeAgent,
		BizIDs:          bizIDs,
		Operator:        access.GetVirtualUser(),
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

func (s *Executor) executeChangeActionPluginInstall(nCtx contextx.IContext, tasks []*ChangeTask) error {
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
				Version: param.PluginVersion,
			},
		}, &types.PluginDeploymentPluginConf{
			CustomConfigContext: param.CustomConfigContext,
		})

		hostMap[task.Target.Host.HostID] = struct{}{}
	}

	hostIDs := conv.MapKeyToSlice(hostMap)
	workflowID, err := s.pluginManager.LaunchInstallPlugin(nCtx, types.InstallPluginParam{
		Type:              types.PluginWorkflowTypeInstall,
		HostIDs:           hostIDs,
		Operator:          access.GetVirtualUser(),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin install: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action plugin install")

	return nil
}

func (s *Executor) executeChangeActionPluginUninstall(_ contextx.IContext, _ []*ChangeTask) error {
	return errors.New("not implemented")
}

func (s *Executor) executeChangeActionPluginUpgrade(_ contextx.IContext, _ []*ChangeTask) error {
	return errors.New("not implemented")
}

// ===============================================================================
// Plugin Sub Config Related Change Actions
// ===============================================================================

func (s *Executor) executeChangeActionPluginApplySubConfig(nCtx contextx.IContext, tasks []*ChangeTask) error {
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
			ConfigFilesDetail:   param.ConfigFilesDetail,
			CustomConfigContext: param.CustomConfigContext,
		})

		hostMap[task.Target.Host.HostID] = struct{}{}
	}

	hostIDs := conv.MapKeyToSlice(hostMap)
	workflowID, err := s.pluginManager.LaunchApplyPluginSubConfig(nCtx, types.ApplyPluginSubConfigParam{
		Type:              types.PluginWorkflowTypeApplyPluginSubConfig,
		HostIDs:           hostIDs,
		Operator:          access.GetVirtualUser(),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin apply sub config: %w", err)
	}

	logger.G.Sys().With("workflow-id", workflowID).
		Info("successful to execute change action plugin apply sub config")

	return nil
}

func (s *Executor) executeChangeActionPluginDeleteSubConfig(_ contextx.IContext, _ []*ChangeTask) error {
	// TODO: implement me.
	return errors.New("not implemented")
}

// ===============================================================================
// Proxy Related Change Actions
// ===============================================================================

func (s *Executor) executeChangeActionProxyInstall(_ contextx.IContext, _ []*ChangeTask) error {
	return errors.New("not implemented")
}

func (s *Executor) executeChangeActionProxyUninstall(_ contextx.IContext, _ []*ChangeTask) error {
	return errors.New("not implemented")
}

func (s *Executor) executeChangeActionProxyUpgrade(_ contextx.IContext, _ []*ChangeTask) error {
	return errors.New("not implemented")
}
